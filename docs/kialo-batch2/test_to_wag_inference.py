"""Tests of to_wag.py's routes to an argument's inference (OC18: a claim whose target is the
inference of another claim's argument): an answer to the argument's question there, with the
claim's scheme or with Other, or an undercutter of that argument; a target not on the board; a role
that cannot be played there; and a claim whose scheme or question the game lacks.

A capturing transport stands in for the server and records the actions the replay sends; it tests
the routes, not the engine. The native probes of the reviewer's patch and of its check
(wag-workspace verify/ethix-2026-11-27/PATCHES.md) ran these routes against a wag-server.

usage: python3 -I test_to_wag_inference.py
"""
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest import mock

sys.dont_write_bytecode = True  # no __pycache__ beside to_wag.py: the tree stays clean

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('replay_under_test', HERE / 'to_wag.py')
replay = importlib.util.module_from_spec(spec)
spec.loader.exec_module(replay)


class ReplayInferenceRoutes(unittest.TestCase):
    def route(self, role='cq_answer', cq='CQ:other_event', scheme='other', polarity='against', target_exists=True):
        claim = 'true_in_situation(rain)'
        def instance(cid, wid, formula, bindings):
            return {'id': cid + '#1', 'primary': True, 'scheme': {'walton_id': wid, 'fit': 'reconstructed'},
                    'bindings': bindings, 'premises': [], 'conclusion': {'formula': formula, 'en': formula or 'reply'}}
        ann = {'items': [
            {'kialo_claim_id': 'thesis', 'instances': [instance('thesis', 'none', claim, {})]},
            {'kialo_claim_id': 'parent', 'literal': ['Synthetic sign.'],
             'instances': [instance('parent', 'sign', claim, {'A': 'wet_streets', 'B': 'rain'})],
             'role': {'kind': 'independent_pro', 'polarity': 'for'},
             'target': {'actual': {'claim': 'thesis', 'part': 'conclusion', 'formula': claim}}},
            {'kialo_claim_id': 'reply', 'literal': ['Synthetic other cause.'],
             'instances': [instance('reply', scheme, None,
                                   {'W': 'source', 'D': 'causes'} if scheme == 'expert_opinion' else {})],
             'role': {'kind': role, 'cq': cq, 'polarity': polarity},
             'target': {'actual': {'claim': 'parent' if target_exists else 'missing',
                                   'instance': 'parent#1' if target_exists else 'missing#1',
                                   'part': 'inference', 'formula': None}}}
        ]}
        items = {'discussion': {'thesis_claim_id': 'thesis'},
                 'items': [{'kialo_claim_id': x['kialo_claim_id'], 'text': 'synthetic', 'kialo_parents': []}
                           for x in ann['items']]}
        state = {'id': 'test', 'active': 'Pro', 'arguments': {}, 'statements': {claim: {'label': 'in'}}}
        actions = []
        def transport(base, path, body=None):
            if path == '/api/meta':
                return {'schemes': [{'id': 'AS:sign', 'carneades_id': 'sign'},
                                    {'id': 'AS:expert_opinion', 'carneades_id': 'expert_opinion'},
                                    {'id': 'AS:other', 'carneades_id': 'other', 'optional': True}]}
            if path.endswith('/export'):
                return {'data_sha256': 'capturing-transport-only'}
            if path.endswith('/move'):
                action = body['action']
                actions.append(action)
                if action.get('type') == 'abstain':
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True}}
                if action.get('answers_cq_id') == 'CQ:does_not_exist':
                    return {'result': {'accepted': False, 'error': 'S-test: unknown CQ'}}
                aid = 'a' + str(len(state['arguments']) + 1)
                if action.get('answers_cq_id') or action.get('attack_type') == 'undercutter':
                    conclusion = 'undercut(' + action['target_arg_id'] + ')'
                else:
                    conclusion = action.get('target_stmt_id') or action.get('variable_bindings', {}).get('C', 'new_statement')
                state['arguments'][aid] = {'conclusion': conclusion}
                state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                return {'result': {'accepted': True, 'arg_ids': [aid]}}
            return state
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name, value in [('annotations.json', ann), ('items.json', items), ('case.json', {})]:
                (root / name).write_text(json.dumps(value))
            argv = ['to_wag.py', str(root / 'annotations.json'), str(root / 'items.json'), 'http://capture',
                    '--report', str(root / 'report.json')]
            replay.WALTON_TO_WAG.clear()
            with mock.patch.object(replay, 'call', transport), mock.patch.object(sys, 'argv', argv), contextlib.redirect_stdout(io.StringIO()):
                replay.main()
            report = json.loads((root / 'report.json').read_text())
            result = next(x for x in report['moves'] if x['claim'] == 'reply')
        return result, actions

    def test_other_answer_keeps_argument_locus_cq(self):
        result, actions = self.route()
        self.assertEqual(result['played'], 'cq-other')
        self.assertEqual(result['integration'], 'as_intended')
        self.assertEqual(result['executed']['cq'], 'CQ:other_event')
        self.assertEqual(result['executed']['conclusion'], 'undercut(a1)')
        self.assertTrue(any(x.get('answers_cq_id') == 'CQ:other_event' for x in actions))

    def test_native_scheme_answer_keeps_argument_locus_cq(self):
        result, _ = self.route(scheme='expert_opinion')
        self.assertEqual(result['played'], 'cq')
        self.assertEqual(result['integration'], 'as_intended')
        self.assertEqual(result['executed']['scheme_id'], 'AS:expert_opinion')
        self.assertEqual(result['executed']['cq'], 'CQ:other_event')

    def test_generic_undercutter_stays_argument_scoped(self):
        result, _ = self.route(role='undercutter', cq=None)
        self.assertEqual(result['played'], 'inference-other')
        self.assertEqual(result['executed']['attack_type'], 'undercutter')
        self.assertEqual(result['executed']['conclusion'], 'undercut(a1)')

    def test_rejected_cq_is_not_bypassed_by_generic_fallback(self):
        result, actions = self.route(cq='CQ:does_not_exist')
        self.assertIsNone(result['executed'])
        reply_actions = [x for x in actions if x.get('scheme_id') == 'AS:other']
        self.assertEqual(len(reply_actions), 1)
        self.assertEqual(reply_actions[0]['answers_cq_id'], 'CQ:does_not_exist')

    def test_missing_argument_target_does_not_become_new_statement(self):
        result, _ = self.route(target_exists=False)
        self.assertIsNone(result['executed'])
        self.assertEqual(result['integration'], 'cascade')

    def test_wrong_inference_role_is_not_reinterpreted(self):
        result, _ = self.route(role='independent_con', cq=None)
        self.assertIsNone(result['executed'])
        self.assertEqual(result['integration'], 'failed')

    def test_pro_undercutter_is_refused(self):
        result, _ = self.route(role='undercutter', cq=None, polarity='for')
        self.assertIsNone(result['executed'])
        self.assertEqual(result['integration'], 'failed')

    def test_answer_whose_scheme_the_game_lacks_goes_through_other(self):
        result, actions = self.route(scheme='cause_to_effect')
        self.assertEqual(result['played'], 'cq-other')
        self.assertEqual(result['integration'], 'coverage_gap')
        self.assertEqual(result['executed']['cq'], 'CQ:other_event')
        self.assertEqual(result['executed']['conclusion'], 'undercut(a1)')
        self.assertTrue(all(x.get('scheme_id') == 'AS:other' for x in actions if x.get('answers_cq_id')))

    def test_refused_answer_whose_scheme_the_game_lacks_is_not_played(self):
        # As for a move about a statement: a claim whose scheme the game lacks, and whose move with
        # Other the game refuses, is not played.
        result, actions = self.route(scheme='cause_to_effect', cq='CQ:does_not_exist')
        self.assertIsNone(result['executed'])
        self.assertEqual(result['integration'], 'not_played')
        self.assertEqual([x['answers_cq_id'] for x in actions if x.get('answers_cq_id')], ['CQ:does_not_exist'])

    def test_answer_to_a_question_the_game_lacks_is_a_coverage_gap(self):
        result, actions = self.route(scheme='expert_opinion', cq='Walton CQ2')
        self.assertIsNone(result['executed'])
        self.assertEqual(result['integration'], 'coverage_gap')
        self.assertFalse(any(x.get('target_arg_id') for x in actions if x.get('type') != 'abstain'))


if __name__ == '__main__':
    unittest.main(verbosity=2)
