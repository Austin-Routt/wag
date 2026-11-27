"""The replay's label for a claim coded Other: cases T1-T7 (d086-first-gold.md §6; ChatGPT's patch 03).
A capturing transport stands in for the server: the tests read labels only."""
import contextlib, importlib.util, io, json, sys, tempfile, unittest
from pathlib import Path
from unittest import mock
sys.dont_write_bytecode = True  # no __pycache__ beside to_wag.py: the tree stays clean
spec = importlib.util.spec_from_file_location('replay', Path(__file__).parent / 'to_wag.py')
replay = importlib.util.module_from_spec(spec)
spec.loader.exec_module(replay)

class OtherReporting(unittest.TestCase):

    def route(self, wid='other', role='premise_attack', polarity='against', cq=None, refuse_cq=False, new_statement=False, conclusion_kind="matched", lose_reply=False):
        claim = 'should_be_performed(tax_meat)'
        premise = 'brings_about(tax_meat,starvation_from_reduced_meat_access)'

        def inst(cid, wid, formula, bindings):
            return {'id': cid + '#1', 'primary': True, 'scheme': {'walton_id': wid, 'fit': 'reconstructed'}, 'bindings': bindings, 'premises': [], 'conclusion': {'formula': formula, 'en': formula or 'Reply'}}
        ann = {'items': [{'kialo_claim_id': 'thesis', 'instances': [inst('thesis', 'none', claim, {})]}, {'kialo_claim_id': 'parent', 'literal': ['A reconstructed consequence argument.'], 'instances': [inst('parent', 'negative_consequences', '¬' + claim, {'A': 'tax_meat', 'G': 'starvation_from_reduced_meat_access'})], 'role': {'kind': 'independent_con', 'polarity': 'against'}, 'target': {'actual': {'claim': 'thesis', 'part': 'conclusion', 'formula': claim}}}, {'kialo_claim_id': 'reply', 'literal': ['The people said to be harmed do not consume the taxed good.'], 'instances': [inst('reply', wid, ('¬' if polarity == 'against' else '') + premise, {})], 'role': {'kind': role, 'polarity': polarity, 'cq': cq}, 'target': {'actual': None if new_statement else {'claim': 'parent', 'instance': 'parent#1', 'part': 'premise', 'formula': premise}}}]}
        if conclusion_kind == "different":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = "background_fact"
        elif conclusion_kind == "null":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = None
        items = {'discussion': {'thesis_claim_id': 'thesis'}, 'items': [{'kialo_claim_id': x['kialo_claim_id'], 'text': 'Synthetic routing fixture', 'kialo_parents': []} for x in ann['items']]}
        state = {'id': 'test', 'active': 'Pro', 'arguments': {}, 'statements': {claim: {'label': 'out'}}}
        actions = []

        def transport(base, path, body=None):
            if path == '/api/meta':
                return {'schemes': [{'id': 'AS:negative_consequences', 'carneades_id': 'negative_consequences'}, {'id': 'AS:other', 'carneades_id': 'other', 'optional': True}]}
            if path.endswith('/export'):
                return {'data_sha256': 'capturing-transport-only'}
            if path.endswith('/move'):
                action = body['action']
                actions.append(action)
                if action.get('type') == 'abstain':
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True}}
                if refuse_cq and action.get('answers_cq_id'):
                    return {'result': {'accepted': False, 'error': 'S-test: scripted CQ refusal'}}
                aid = 'a' + str(len(state['arguments']) + 1)
                if lose_reply and action.get('target_stmt_id') == premise:
                    # T8: the move is accepted, but its argument is not in the state, so no conclusion shows.
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True, 'arg_ids': ['a_lost']}}
                formula = action.get('target_stmt_id') or 'new_statement'
                if action.get('type') == 'attack':
                    formula = '¬' + formula if not formula.startswith('¬') else formula[1:]
                state['arguments'][aid] = {'conclusion': formula}
                state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                return {'result': {'accepted': True, 'arg_ids': [aid]}}
            return state
        with tempfile.TemporaryDirectory() as d:
            p = Path(d)
            for n, v in [('annotations.json', ann), ('items.json', items), ('case.json', {})]:
                (p / n).write_text(json.dumps(v))
            argv = ['to_wag.py', str(p / 'annotations.json'), str(p / 'items.json'), 'http://capture', '--report', str(p / 'report.json')]
            replay.WALTON_TO_WAG.clear()
            with mock.patch.object(replay, 'call', transport), mock.patch.object(sys, 'argv', argv), contextlib.redirect_stdout(io.StringIO()):
                replay.main()
            result = next((x for x in json.loads((p / 'report.json').read_text())['moves'] if x['claim'] == 'reply'))
        return (result, actions)

    def test_t1_explicit_other_premise_attack_is_intended(self):
        result, _ = self.route()
        self.assertEqual(result['played'], 'other')
        self.assertEqual(result['integration'], 'as_intended')
        self.assertEqual(result['executed']['conclusion'], result['intended']['conclusion'])

    def test_t2_explicit_other_premise_support_is_intended(self):
        result, _ = self.route(role='premise_backing', polarity='for')
        self.assertEqual(result['integration'], 'as_intended')

    def test_t3_unavailable_scheme_stays_coverage_gap(self):
        result, _ = self.route(wid='verbal_classification')
        self.assertEqual(result['integration'], 'coverage_gap')

    def test_t4_lost_named_cq_route_stays_failed(self):
        result, actions = self.route(role='cq_answer', cq='CQ:likelihood', refuse_cq=True)
        self.assertEqual(result['played'], 'other')
        self.assertEqual(result['integration'], 'failed')
        self.assertTrue(any((a.get('answers_cq_id') for a in actions)))

    def test_t5_free_new_statement_not_promoted_by_this_fix(self):
        result, _ = self.route(new_statement=True, role='independent_pro', polarity='for')
        self.assertEqual(result['played'], 'other-new')
        self.assertEqual(result['integration'], 'coverage_gap')
    def test_t6_different_own_conclusion_is_not_intended(self):
        result, _ = self.route(conclusion_kind="different")
        self.assertEqual(result["played"], "other")
        self.assertNotEqual(result["intended"]["conclusion"], result["executed"]["conclusion"])
        self.assertEqual(result["integration"], "failed")

    def test_t7_null_own_conclusion_is_not_intended(self):
        result, _ = self.route(conclusion_kind="null")
        self.assertEqual(result["played"], "other")
        self.assertIsNone(result["intended"]["conclusion"])
        self.assertEqual(result["integration"], "failed")

    def test_t8_two_missing_conclusions_are_not_the_same(self):
        result, _ = self.route(conclusion_kind="null", lose_reply=True)
        self.assertEqual(result["played"], "other")
        self.assertIsNone(result["intended"]["conclusion"])
        self.assertIsNone(result["executed"]["conclusion"])
        self.assertEqual(result["integration"], "failed")

if __name__ == '__main__':
    unittest.main(verbosity=2)
