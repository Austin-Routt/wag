"""The replay plays an instance by the discussion (D096): cases RC11-RC15 (d086-first-gold.md §6;
annotation-records.md §13). A capturing transport stands in for the server."""
import contextlib
import copy
import importlib.util
import io
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest import mock

sys.dont_write_bytecode = True  # no __pycache__ beside to_wag.py: the tree stays clean
spec = importlib.util.spec_from_file_location('replay', Path(__file__).parent / 'to_wag.py')
replay = importlib.util.module_from_spec(spec)
spec.loader.exec_module(replay)

CLAIM = 'should_be_performed(tax_meat)'
PREMISE = 'brings_about(tax_meat,less_emissions)'


def instance(iid, scheme, formula, primary=True, premises=(), by=None, bindings=None):
    x = {'id': iid, 'primary': primary, 'scheme': {'walton_id': scheme, 'fit': 'reconstructed'},
         'bindings': bindings or {}, 'premises': [{'role': 'premise', 'formula': f, 'status': 'supplied'} for f in premises],
         'conclusion': {'formula': formula, 'en': formula}}
    if by:
        x['by'] = by
    return x


class DiscussionReplay(unittest.TestCase):
    def run_replay(self, ann, refuse_scheme=None):
        items = {'discussion': {'thesis_claim_id': 'thesis'}, 'items': [
            {'kialo_claim_id': x['kialo_claim_id'], 'text': 'Synthetic routing fixture', 'kialo_parents': []}
            for x in ann['items']]}
        state = {'id': 'test', 'active': 'Pro', 'arguments': {}, 'statements': {CLAIM: {'label': 'out'}}}
        moves = []

        def transport(base, path, body=None):
            if path == '/api/meta':
                return {'schemes': [{'id': 'AS:positive_consequences', 'carneades_id': 'positive_consequences'},
                                    {'id': 'AS:other', 'carneades_id': 'other', 'optional': True}]}
            if path.endswith('/export'):
                return {'data_sha256': 'capturing-transport-only'}
            if path.endswith('/move'):
                moves.append(copy.deepcopy(body))
                act = body['action']
                if act['type'] == 'abstain':
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True}}
                if act['type'] == 'question':
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True}}
                if refuse_scheme and act.get('scheme_id') == refuse_scheme:
                    return {'result': {'accepted': False, 'error': 'S-test: scripted refusal'}}
                aid = 'a' + str(len(state['arguments']) + 1)
                stmt = act.get('target_stmt_id') or 'new_statement'
                if act.get('type') == 'attack':
                    stmt = '¬' + stmt if not stmt.startswith('¬') else stmt[1:]
                state['arguments'][aid] = {'conclusion': stmt}
                state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                return {'result': {'accepted': True, 'arg_ids': [aid]}}
            if '/moves?' in path:
                return []
            return state

        with tempfile.TemporaryDirectory() as d:
            p = Path(d)
            for name, value in [('annotations.json', ann), ('items.json', items), ('case.json', {})]:
                (p / name).write_text(json.dumps(value))
            argv = ['to_wag.py', str(p / 'annotations.json'), str(p / 'items.json'), 'http://capture',
                    '--report', str(p / 'report.json')]
            replay.WALTON_TO_WAG.clear()
            with mock.patch.object(replay, 'call', transport), mock.patch.object(sys, 'argv', argv), \
                    contextlib.redirect_stdout(io.StringIO()):
                replay.main()
            report = json.loads((p / 'report.json').read_text())
        return report, [m for m in moves if m['action']['type'] != 'abstain']

    def thesis_with_discussion(self):
        disc = instance('thesis#2', 'positive_consequences', CLAIM, primary=False, premises=[PREMISE],
                        by='the discussion', bindings={'A': 'tax_meat', 'G': 'less_emissions'})
        thesis = {'kialo_claim_id': 'thesis', 'instances': [instance('thesis#1', 'none', CLAIM), disc]}
        reply = {'kialo_claim_id': 'reply', 'literal': ['The tax would not cut emissions.'],
                 'instances': [instance('reply#1', 'other', '¬' + PREMISE)],
                 'role': {'kind': 'premise_attack', 'polarity': 'against', 'cq': None},
                 'target': {'actual': {'claim': 'thesis', 'instance': 'thesis#2', 'part': 'premise', 'formula': PREMISE}}}
        return {'items': [thesis, reply]}

    def test_rc11_the_theses_discussion_instance_is_played_first(self):
        report, moves = self.run_replay(self.thesis_with_discussion())
        first = moves[0]['action']
        self.assertEqual((first['type'], first.get('target_stmt_id'), first.get('scheme_id')),
                         ('support', CLAIM, 'AS:positive_consequences'))
        self.assertEqual([d['instance'] for d in report['discussion']], ['thesis#2'])
        self.assertTrue(report['discussion'][0]['played'])
        entry = next(x for x in report['moves'] if x['claim'] == 'reply')
        self.assertNotEqual(entry['integration'], 'cascade')
        self.assertEqual(entry['executed']['target_stmt'], PREMISE)
        self.assertEqual([x['claim'] for x in report['moves']], ['reply'])

    def test_rc12_another_claims_discussion_instance_follows_its_claim(self):
        parent_stmt = 'brings_about(tax_meat,cheaper_veg)'
        sub = 'brings_about(cheaper_veg,healthier_diets)'
        disc = instance('parent#2', 'positive_consequences', parent_stmt, primary=False, premises=[sub],
                        by='the discussion', bindings={'A': 'cheaper_veg', 'G': 'healthier_diets'})
        parent = {'kialo_claim_id': 'parent', 'literal': ['The tax makes vegetables cheaper.'],
                  'instances': [instance('parent#1', 'other', parent_stmt), disc],
                  'role': {'kind': 'independent_pro', 'polarity': 'for', 'cq': None},
                  'target': {'actual': {'claim': 'thesis', 'instance': 'thesis#1', 'part': 'conclusion', 'formula': CLAIM}}}
        reply = {'kialo_claim_id': 'reply', 'literal': ['Cheaper vegetables would not change diets.'],
                 'instances': [instance('reply#1', 'other', '¬' + sub)],
                 'role': {'kind': 'premise_attack', 'polarity': 'against', 'cq': None},
                 'target': {'actual': {'claim': 'parent', 'instance': 'parent#2', 'part': 'premise', 'formula': sub}}}
        thesis = {'kialo_claim_id': 'thesis', 'instances': [instance('thesis#1', 'none', CLAIM)]}
        report, moves = self.run_replay({'items': [thesis, parent, reply]})
        whys = [m.get('why', '') for m in moves]
        i_parent = next(i for i, w in enumerate(whys) if w.startswith('Kialo parent:'))
        i_disc = next(i for i, w in enumerate(whys) if 'parent#2' in w)
        i_reply = next(i for i, w in enumerate(whys) if w.startswith('Kialo reply:'))
        self.assertTrue(i_parent < i_disc < i_reply, whys)
        d = moves[i_disc]['action']
        self.assertEqual((d['type'], d.get('scheme_id'), d.get('target_stmt_id')),
                         ('support', 'AS:positive_consequences', parent_stmt))
        entry = next(x for x in report['moves'] if x['claim'] == 'reply')
        self.assertNotEqual(entry['integration'], 'cascade')

    def with_fragment(self, stated):
        """The thesis's discussion instance, a fragment that backs its premise PREMISE but states
        `stated`, and a bare denial of what the fragment states."""
        ann = self.thesis_with_discussion()
        ann['items'] = ann['items'][:1]
        frag_inst = instance('frag#1', 'positive_consequences', None, premises=[stated], bindings={'G': 'less_emissions'})
        frag_inst['scheme']['fit'] = 'fragment'
        frag_inst['premises'][0]['status'] = 'explicit'
        frag = {'kialo_claim_id': 'frag', 'literal': ['The tax would cut emissions.'], 'instances': [frag_inst],
                'role': {'kind': 'premise_backing', 'polarity': 'for', 'cq': None},
                'target': {'actual': {'claim': 'thesis', 'instance': 'thesis#2', 'part': 'premise', 'formula': PREMISE}}}
        deny = {'kialo_claim_id': 'deny', 'literal': ['It would not.'],
                'instances': [instance('deny#1', 'none', 'no_cut')],
                'role': {'kind': 'premise_attack', 'polarity': 'against', 'cq': None},
                'target': {'actual': {'claim': 'frag', 'instance': 'frag#1', 'part': 'premise', 'formula': stated}}}
        deny['instances'][0]['scheme']['fit'] = 'none'
        ann['items'] += [frag, deny]
        return ann

    def test_rc14_a_denial_of_what_a_fragment_states_is_a_challenge_where_it_is_a_premise(self):
        report, moves = self.run_replay(self.with_fragment(PREMISE))
        disc_arg = report['discussion'][0]['arg']
        frag = next(x for x in report['moves'] if x['claim'] == 'frag')
        self.assertEqual(frag['integration'], 'fragment')
        self.assertNotEqual(frag['arg'], disc_arg)
        q = [m['action'] for m in moves if m['action']['type'] == 'question']
        self.assertEqual(len(q), 1, moves)
        self.assertEqual((q[0]['target_arg_id'], q[0]['target_premise_id']), (disc_arg, PREMISE))
        deny = next(x for x in report['moves'] if x['claim'] == 'deny')
        self.assertEqual((deny['played'], deny['integration'], deny['executed']['target_arg']),
                         ('challenge', 'as_intended', disc_arg))

    def test_rc15_another_premise_of_a_fragment_is_not_moved(self):
        other = 'brings_about(tax_meat,less_methane)'
        report, moves = self.run_replay(self.with_fragment(other))
        frag = next(x for x in report['moves'] if x['claim'] == 'frag')
        q = [m['action'] for m in moves if m['action']['type'] == 'question']
        self.assertEqual(len(q), 1, moves)
        self.assertEqual((q[0]['target_arg_id'], q[0]['target_premise_id']), (frag['arg'], other))

    def test_rc13_a_refused_discussion_instance(self):
        report, moves = self.run_replay(self.thesis_with_discussion(), refuse_scheme='AS:positive_consequences')
        d = report['discussion'][0]
        self.assertFalse(d['played'])
        self.assertIn('scripted refusal', d['error'])
        entry = next(x for x in report['moves'] if x['claim'] == 'reply')
        self.assertEqual(entry['integration'], 'cascade')


if __name__ == '__main__':
    unittest.main(verbosity=2)
