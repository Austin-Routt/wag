"""An unresolved target produces no move: cases U1-U8 (d086-first-gold.md §6; D102; ChatGPT's patch 04;
U7-U8 after the independent review of step 1)."""
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

class UnresolvedTargets(unittest.TestCase):
    def route(self, fields=(), role='cq_answer', cq='CQ:likelihood', wid='verbal_classification',
              polarity='for', part='premise', malformed=None):
        claim = 'should_be_performed(tax_meat)'
        premise = 'brings_about(tax_meat,reduced_unhealthy_meat_consumption)'
        def instance(cid, scheme, formula):
            return {'id':cid+'#1','primary':True,'scheme':{'walton_id':scheme,'fit':'reconstructed'},
                    'bindings':{},'premises':[],'conclusion':{'formula':formula,'en':formula}}
        target = {'claim':'parent','instance':'parent#1','part':part,
                  'formula':None if part=='inference' else premise}
        for field in fields:target[field]={'unresolved':'The exact '+field+' is not established.'}
        for field,value in (malformed or {}).items():target[field]=copy.deepcopy(value)
        ann={'items':[
            {'kialo_claim_id':'thesis','instances':[instance('thesis','none',claim)]},
            {'kialo_claim_id':'parent','literal':['A consequence argument.'],
             'instances':[instance('parent','positive_consequences',claim)],
             'role':{'kind':'independent_pro','polarity':'for'},
             'target':{'actual':{'claim':'thesis','part':'conclusion','formula':claim}}},
            {'kialo_claim_id':'reply','literal':['A background condition only.'],
             'instances':[instance('reply',wid,'background_condition')],
             'role':{'kind':role,'cq':cq,'polarity':polarity},'target':{'actual':target}}]}
        items={'discussion':{'thesis_claim_id':'thesis'},'items':[
            {'kialo_claim_id':x['kialo_claim_id'],'text':'Synthetic routing control','kialo_parents':[]}
            for x in ann['items']]}
        state={'id':'test','active':'Pro','arguments':{},'statements':{claim:{'label':'out'}}}
        moves=[]
        def transport(base,path,body=None):
            if path=='/api/meta':return {'schemes':[
                {'id':'AS:positive_consequences','carneades_id':'positive_consequences'},
                {'id':'AS:expert_opinion','carneades_id':'expert_opinion'},
                {'id':'AS:other','carneades_id':'other','optional':True}]}
            if path.endswith('/export'):return {'data_sha256':'capturing-transport-only'}
            if path.endswith('/move'):
                moves.append(copy.deepcopy(body));act=body['action']
                if act['type']=='abstain':
                    state['active']='Con' if state['active']=='Pro' else 'Pro'
                    return {'result':{'accepted':True}}
                aid='a'+str(len(state['arguments'])+1)
                if act.get('attack_type')=='undercutter' or act.get('answers_cq_id'):
                    conclusion='undercut('+act['target_arg_id']+')'
                else:conclusion=act.get('target_stmt_id') or 'new_statement'
                state['arguments'][aid]={'conclusion':conclusion}
                state['active']='Con' if state['active']=='Pro' else 'Pro'
                return {'result':{'accepted':True,'arg_ids':[aid]}}
            if '/moves?' in path:return []
            return state
        with tempfile.TemporaryDirectory() as d:
            p=Path(d)
            for name,value in [('annotations.json',ann),('items.json',items),('case.json',{})]:
                (p/name).write_text(json.dumps(value))
            before=(p/'annotations.json').read_bytes()
            argv=['to_wag.py',str(p/'annotations.json'),str(p/'items.json'),'http://capture','--report',str(p/'report.json')]
            replay.WALTON_TO_WAG.clear()
            with mock.patch.object(replay,'call',transport),mock.patch.object(sys,'argv',argv),contextlib.redirect_stdout(io.StringIO()):
                replay.main()
            self.assertEqual(before,(p/'annotations.json').read_bytes())
            report=json.loads((p/'report.json').read_text())
        entry=next(x for x in report['moves'] if x['claim']=='reply')
        return entry,[m for m in moves if m.get('why','').startswith('Kialo reply:')]

    def assert_skipped(self,entry,moves,fields):
        self.assertIsNone(entry['executed'])
        self.assertEqual(moves,[])
        self.assertEqual(entry['integration'],'not_played')
        self.assertTrue(entry['played'].startswith('no: unresolved target'))
        self.assertEqual(set(entry['unresolved_target']),set(fields))

    def test_u1_each_marker_blocks_normalization_and_dispatch(self):
        for field in ('formula','part','instance'):
            with self.subTest(field=field):
                entry,moves=self.route((field,))
                self.assert_skipped(entry,moves,(field,))

    def test_u2_all_markers_retain_provisional_annotation(self):
        fields=('instance','part','formula');entry,moves=self.route(fields)
        self.assert_skipped(entry,moves,fields)
        self.assertEqual(entry['intended']['walton_id'],'verbal_classification')
        self.assertEqual(entry['intended']['conclusion'],'background_condition')
        self.assertEqual(entry['intended']['cq'],'CQ:likelihood')

    def test_u3_unresolved_bare_denial_does_not_become_challenge(self):
        entry,moves=self.route(('formula',),wid='none',role='premise_attack',cq=None,polarity='against')
        self.assert_skipped(entry,moves,('formula',))

    def test_u4_unresolved_named_cq_does_not_use_parent_fallback(self):
        entry,moves=self.route(('instance',),wid='expert_opinion')
        self.assert_skipped(entry,moves,('instance',))

    def test_u5_resolved_statement_control_still_plays(self):
        entry,moves=self.route()
        self.assertEqual(entry['played'],'other')
        self.assertEqual(entry['executed']['target_stmt'],'brings_about(tax_meat,reduced_unhealthy_meat_consumption)')
        self.assertEqual(len(moves),1)
        self.assertNotIn('unresolved_target',entry)

    def test_u6_inference_null_formula_is_not_unresolved(self):
        entry,moves=self.route(role='undercutter',cq=None,wid='other',polarity='against',part='inference')
        self.assertEqual(entry['played'],'inference-other')
        self.assertEqual(entry['executed']['conclusion'],'undercut(a1)')
        self.assertEqual(len(moves),1)

    def test_u7_malformed_target_values_are_not_played(self):
        bad=[{'unresolved':'  '},{'unstated':'The formula is not stated.'},['brings_about(tax_meat,x)'],
             {'unresolved':'Not established.','note':'a second key'},{'unresolved':7}]
        for field in ('instance','part','formula'):
            for value in bad:
                for wid in ('expert_opinion','verbal_classification','none'):
                    with self.subTest(field=field,value=value,wid=wid):
                        entry,moves=self.route(wid=wid,malformed={field:value},
                                               role='premise_attack' if wid=='none' else 'cq_answer',
                                               cq=None if wid=='none' else 'CQ:likelihood',
                                               polarity='against' if wid=='none' else 'for')
                        self.assertIsNone(entry['executed'])
                        self.assertEqual(moves,[])
                        self.assertEqual(entry['integration'],'not_played')
                        self.assertTrue(entry['played'].startswith('no: malformed target'))
                        self.assertEqual(entry['malformed_target'],[field])
                        self.assertNotIn('unresolved_target',entry)

    def test_a5_a_reason_go_does_not_trim_is_unresolved(self):
        # Case A5: a reason of U+001C to U+001F is a reason, as Go reads it; Go's white space alone is not.
        entry,moves=self.route(malformed={'formula':{'unresolved':'\x1c\x1d\x1e\x1f'}})
        self.assert_skipped(entry,moves,('formula',))
        entry,moves=self.route(malformed={'formula':{'unresolved':'\u3000\t\x85'}})
        self.assertEqual(entry['malformed_target'],['formula'])
        self.assertEqual(moves,[])

    def test_u8_resolved_named_cq_control_is_answered_at_its_target(self):
        entry,moves=self.route(wid='expert_opinion')
        self.assertEqual(entry['played'],'cq')
        self.assertEqual(entry['integration'],'as_intended')
        self.assertEqual(len(moves),1)
        act=moves[0]['action']
        self.assertEqual(act['answers_cq_id'],'CQ:likelihood')
        self.assertEqual(act['target_premise_id'],'brings_about(tax_meat,reduced_unhealthy_meat_consumption)')
        self.assertNotIn('unresolved_target',entry)

if __name__=='__main__':unittest.main(verbosity=2)
