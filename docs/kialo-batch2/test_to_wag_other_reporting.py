"""The replay's label for a claim coded Other: cases T1-T9 (d086-first-gold.md §6; ChatGPT's patch 03; T8
after step 1's planted defects, T9 after its independent review); and for a changed conclusion, cases
T10-T19 with T6 corrected (F3 of the external review of exchange 3, ChatGPT's patch 01 with the check's
changes). A capturing transport stands in for the server: the tests read labels only. It concludes as the
engine does: a move about a statement or a premise, that formula or its negation; an answer at an
exception, the exception for an attack and its negation for a defence; a new argument, the formula given."""
import contextlib, importlib.util, io, json, sys, tempfile, unittest
from pathlib import Path
from unittest import mock
sys.dont_write_bytecode = True  # no __pycache__ beside to_wag.py: the tree stays clean
spec = importlib.util.spec_from_file_location('replay', Path(__file__).parent / 'to_wag.py')
replay = importlib.util.module_from_spec(spec)
spec.loader.exec_module(replay)

class OtherReporting(unittest.TestCase):

    def route(self, wid='other', role='premise_attack', polarity='against', cq=None, refuse_cq=False, new_statement=False, conclusion_kind="matched", lose_reply=False, part='premise', new_conclusion='new_statement',
              conclusion=None, parent_wid='negative_consequences', parent_concludes=None, board=None, cq_error='S-test: scripted CQ refusal'):
        # conclusion: the reply's annotated conclusion, given (T21, T22); parent_wid: the parent's scheme (one the
        # game lacks plays it with Other, T24); parent_concludes: the formula the parent's argument concludes in
        # place of its annotated one (T22); board: with it, a move about a statement that is neither the thesis,
        # an argument's conclusion nor one of these is refused (S7), as the game refuses it (T22); cq_error: the
        # refusal of a scripted CQ route (T24).
        claim = 'should_be_performed(tax_meat)'
        premise = 'brings_about(tax_meat,starvation_from_reduced_meat_access)'

        def inst(cid, wid, formula, bindings):
            return {'id': cid + '#1', 'primary': True, 'scheme': {'walton_id': wid, 'fit': 'reconstructed'}, 'bindings': bindings, 'premises': [], 'conclusion': {'formula': formula, 'en': formula or 'Reply'}}
        ann = {'items': [{'kialo_claim_id': 'thesis', 'instances': [inst('thesis', 'none', claim, {})]}, {'kialo_claim_id': 'parent', 'literal': ['A reconstructed consequence argument.'], 'instances': [inst('parent', parent_wid, '¬' + claim, {'A': 'tax_meat', 'G': 'starvation_from_reduced_meat_access'})], 'role': {'kind': 'independent_con', 'polarity': 'against'}, 'target': {'actual': {'claim': 'thesis', 'part': 'conclusion', 'formula': claim}}}, {'kialo_claim_id': 'reply', 'literal': ['The people said to be harmed do not consume the taxed good.'], 'instances': [inst('reply', wid, ('¬' if polarity == 'against' else '') + premise, {})], 'role': {'kind': role, 'polarity': polarity, 'cq': cq}, 'target': {'actual': None if new_statement else {'claim': 'parent', 'instance': 'parent#1', 'part': part, 'formula': None if part == 'inference' else '¬' + claim if part == 'conclusion' else premise}}}]}
        if part == 'exception':
            # R8: a claim at an exception concludes it against the argument, its negation for it.
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = ('¬' if polarity == 'for' else '') + premise
        if conclusion_kind == "different":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = "background_fact"
        elif conclusion_kind == "null":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = None
        elif conclusion_kind == "spaced":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = "¬ brings_about( tax_meat , starvation_from_reduced_meat_access )"
        elif conclusion_kind == "not":
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = "not(brings_about(tax_meat,starvation_from_reduced_meat_access))"
        if conclusion is not None:
            ann["items"][-1]["instances"][0]["conclusion"]["formula"] = conclusion
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
                target = action.get('target_stmt_id')
                if board is not None and target and target != claim and target not in board and \
                        target not in (a['conclusion'] for a in state['arguments'].values()):
                    return {'result': {'accepted': False, 'error': f'S7: statement "{target}" is not on the board'}}
                if refuse_cq and action.get('answers_cq_id'):
                    return {'result': {'accepted': False, 'error': cq_error}}
                aid = 'a' + str(len(state['arguments']) + 1)
                if lose_reply and action.get('target_stmt_id') == premise:
                    # T8: the move is accepted, but its argument is not in the state, so no conclusion shows.
                    state['active'] = 'Con' if state['active'] == 'Pro' else 'Pro'
                    return {'result': {'accepted': True, 'arg_ids': ['a_lost']}}
                formula = action.get('target_stmt_id') or action.get('target_premise_id')
                if not formula and action.get('answers_cq_id') and part == 'exception':
                    formula = premise  # the exception: E for an attack, ¬E for a defence (answers.go)
                    formula = formula if action.get('type') == 'attack' else '¬' + formula
                elif action.get('attack_type') == 'undercutter':
                    formula = 'undercut(' + action['target_arg_id'] + ')'
                elif not formula:
                    formula = new_conclusion
                elif action.get('type') == 'attack':
                    formula = '¬' + formula if not formula.startswith('¬') else formula[1:]
                if parent_concludes and action.get('text') == 'A reconstructed consequence argument.':
                    formula = parent_concludes
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
            report = json.loads((p / 'report.json').read_text())
            self.summary = report['integration']
            self.rows = {x['claim']: x for x in report['moves']}
            result = next((x for x in report['moves'] if x['claim'] == 'reply'))
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
        # T6, corrected by F3: a changed conclusion has a label of its own, for Other as for a named
        # scheme (it was failed, which the README keeps for a move the game refused).
        result, _ = self.route(conclusion_kind="different")
        self.assertEqual(result["played"], "other")
        self.assertNotEqual(result["intended"]["conclusion"], result["executed"]["conclusion"])
        self.assertEqual(result["integration"], "conclusion_changed")
        self.assertEqual(result["conclusion_change"], {"annotated": "background_fact",
                                                       "executed": "¬brings_about(tax_meat,starvation_from_reduced_meat_access)"})
        self.assertIs(result["conclusion_preserved"], False)

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

    def test_t9_spelling_of_the_same_conclusion_is_intended(self):
        for kind in ("spaced", "not"):
            with self.subTest(kind=kind):
                result, _ = self.route(conclusion_kind=kind)
                self.assertEqual(result["played"], "other")
                self.assertNotEqual(result["intended"]["conclusion"], result["executed"]["conclusion"])
                self.assertEqual(result["executed"]["conclusion"], "¬brings_about(tax_meat,starvation_from_reduced_meat_access)")
                self.assertEqual(result["integration"], "as_intended")

    def test_t10_named_scheme_changed_conclusion(self):
        result, _ = self.route(wid='negative_consequences', conclusion_kind='different')
        self.assertEqual(result['played'], 'statement')
        self.assertEqual(result['integration'], 'conclusion_changed')
        self.assertEqual(result['conclusion_change']['annotated'], 'background_fact')
        self.assertEqual(result['conclusion_change']['executed'], '¬brings_about(tax_meat,starvation_from_reduced_meat_access)')
        self.assertIs(result['conclusion_preserved'], False)
        # The summary counts the label (one not among its keys would vanish from the counts).
        self.assertEqual(self.summary['conclusion_changed'], 1)
        self.assertEqual(self.summary['as_intended'], 1)  # the parent

    def test_t11_named_scheme_same_conclusion_stays_intended(self):
        result, _ = self.route(wid='negative_consequences')
        self.assertEqual(result['played'], 'statement')
        self.assertEqual(result['integration'], 'as_intended')
        self.assertIs(result['conclusion_preserved'], True)
        self.assertNotIn('conclusion_change', result)

    def test_t12_named_scheme_null_conclusion_is_failed(self):
        result, _ = self.route(wid='negative_consequences', conclusion_kind='null')
        self.assertEqual(result['played'], 'statement')
        self.assertEqual(result['integration'], 'failed')
        self.assertIsNone(result['conclusion_preserved'])

    def test_t13_answer_through_its_question_changed_conclusion(self):
        for wid in ('negative_consequences', 'other'):
            with self.subTest(wid=wid):
                result, _ = self.route(wid=wid, role='cq_answer', cq='CQ:likelihood', conclusion_kind='different')
                self.assertEqual(result['played'], 'cq-other' if wid == 'other' else 'cq')
                self.assertEqual(result['integration'], 'conclusion_changed')
                self.assertEqual(result['conclusion_change']['annotated'], 'background_fact')

    def test_t14_answer_through_its_question_same_conclusion(self):
        for wid in ('negative_consequences', 'other'):
            with self.subTest(wid=wid):
                result, _ = self.route(wid=wid, role='cq_answer', cq='CQ:likelihood')
                self.assertEqual(result['played'], 'cq-other' if wid == 'other' else 'cq')
                self.assertEqual(result['integration'], 'as_intended')
                self.assertIs(result['conclusion_preserved'], True)

    def test_t15_answer_at_an_exception_attack_and_defence(self):
        # R8: a claim at an exception concludes it against the argument, its negation for it; the engine
        # concludes the exception for an attack and its negation for a defence (answers.go).
        for polarity in ('against', 'for'):
            with self.subTest(polarity=polarity):
                result, actions = self.route(wid='negative_consequences', role='cq_answer', cq='CQ:side_effect',
                                             polarity=polarity, part='exception')
                self.assertEqual(result['played'], 'cq')
                self.assertEqual(actions[-1]['type'], 'attack' if polarity == 'against' else 'support')
                self.assertEqual(result['integration'], 'as_intended')
                self.assertIs(result['conclusion_preserved'], True)

    def test_t16_a_new_argument(self):
        premise = 'brings_about(tax_meat,starvation_from_reduced_meat_access)'
        for made, label in ((premise, 'as_intended'), ('another_formula', 'conclusion_changed')):
            with self.subTest(made=made):
                result, _ = self.route(wid='negative_consequences', new_statement=True, role='independent_pro',
                                       polarity='for', new_conclusion=made)
                self.assertEqual(result['played'], 'new')
                self.assertEqual(result['integration'], label)

    def test_t17_conclusion_preserved_on_rows_of_other_labels(self):
        # A coverage gap keeps its label, and says whether its argument concludes the claim's conclusion.
        result, _ = self.route(wid='verbal_classification', conclusion_kind='different')
        self.assertEqual(result['integration'], 'coverage_gap')
        self.assertIs(result['conclusion_preserved'], False)
        result, _ = self.route(wid='verbal_classification')
        self.assertIs(result['conclusion_preserved'], True)
        # A challenge makes no argument: nothing is compared.
        result, _ = self.route(wid='none')
        self.assertEqual(result['played'], 'challenge')
        self.assertNotIn('conclusion_preserved', result)

    def test_t19_an_undercutter_of_an_inference_is_not_compared(self):
        result, actions = self.route(wid='negative_consequences', role='undercutter', part='inference')
        self.assertEqual(result['played'], 'inference')
        self.assertEqual(result['executed']['conclusion'], 'undercut(a1)')
        self.assertEqual(result['integration'], 'as_intended')
        self.assertNotIn('conclusion_preserved', result)


    def test_t20_a_new_other_argument_is_not_compared(self):
        # Cases T20-T24: the independent check of F3, F5 and F6 (d086-first-gold.md §6). A new Other
        # argument concludes the statement the game makes from the claim's English (its C): null, not
        # compared; its label as before (T5). In the batch, 13292.14, .46 and .61.
        for wid in ('other', 'verbal_classification'):
            with self.subTest(wid=wid):
                result, actions = self.route(wid=wid, new_statement=True, role='independent_pro', polarity='for',
                                             new_conclusion='A synthetic claim of its own.')
                self.assertEqual(result['played'], 'other-new')
                self.assertEqual(actions[-1]['variable_bindings']['C'], result['intended']['conclusion'])
                self.assertEqual(result['integration'], 'coverage_gap')
                self.assertIn('conclusion_preserved', result)
                self.assertIsNone(result['conclusion_preserved'])
                self.assertNotIn('conclusion_change', result)

    def test_t21_go_whitespace_in_a_conclusion_is_no_change(self):
        # As T9: the formulas are compared without Go's whitespace (a tab, U+00A0, U+3000), as
        # validate.py and agreement.py compare them.
        premise = 'brings_about(tax_meat,starvation_from_reduced_meat_access)'
        for space in ('\t', '\u00a0', '\u3000', ' \t\u3000'):
            for wid in ('other', 'negative_consequences'):
                with self.subTest(space=repr(space), wid=wid):
                    result, _ = self.route(wid=wid, conclusion='¬' + space + premise.replace(',', ',' + space))
                    self.assertEqual(result['executed']['conclusion'], '¬' + premise)
                    self.assertEqual(result['integration'], 'as_intended')
                    self.assertIs(result['conclusion_preserved'], True)

    def test_t22_a_claim_at_the_conclusion_of_a_changed_conclusion(self):
        # Its target's argument concludes another formula (conclusion_changed): the move is about the
        # target's annotated conclusion, which is not on the board: a cascade.
        claim = 'should_be_performed(tax_meat)'
        result, actions = self.route(wid='negative_consequences', role='independent_con', polarity='against',
                                     part='conclusion', conclusion=claim, parent_concludes='other_formula(tax_meat)',
                                     board=set())
        self.assertEqual(self.rows['parent']['integration'], 'conclusion_changed')
        self.assertTrue(result['played'].startswith('no: '), result['played'])
        self.assertIn('S7', result['played'])
        self.assertEqual({a.get('target_stmt_id') for a in actions[-2:]}, {'¬' + claim})
        self.assertEqual(result['integration'], 'cascade')
        # Where the annotated conclusion is on the board, the move is made about it, as intended.
        result, actions = self.route(wid='negative_consequences', role='independent_con', polarity='against',
                                     part='conclusion', conclusion=claim, parent_concludes='other_formula(tax_meat)',
                                     board={'¬' + claim})
        self.assertEqual(result['played'], 'statement')
        self.assertEqual(actions[-1]['target_stmt_id'], '¬' + claim)
        self.assertEqual(result['integration'], 'as_intended')
        self.assertIs(result['conclusion_preserved'], True)
        # A target whose conclusion is kept: the move is about its argument's conclusion, as before.
        result, actions = self.route(wid='negative_consequences', role='independent_con', polarity='against',
                                     part='conclusion', conclusion=claim, board=set())
        self.assertEqual(self.rows['parent']['integration'], 'as_intended')
        self.assertEqual(result['played'], 'statement')
        self.assertEqual(result['integration'], 'as_intended')

    def test_t23_an_annotated_conclusion_and_none_executed_is_failed(self):
        # T8 with an annotated conclusion: the executed argument shows none, so nothing is compared.
        for wid in ('other', 'negative_consequences'):
            with self.subTest(wid=wid):
                result, _ = self.route(wid=wid, lose_reply=True)
                self.assertIsNotNone(result['intended']['conclusion'])
                self.assertIsNone(result['executed']['conclusion'])
                self.assertEqual(result['integration'], 'failed')
                self.assertIsNone(result['conclusion_preserved'])
                self.assertNotIn('conclusion_change', result)

    def test_t24_a_later_route_after_a_refused_question(self):
        # Corrected on the tests' first run (d086-first-gold.md §6): a row labelled conclusion_changed
        # is always the first route's move, so no refusal precedes it and the cascade rule cannot reach
        # it. What can occur: the question route refused (S8, the target played with Other), the
        # statement route taken, concluding another formula: not as intended (T4), so a cascade.
        result, actions = self.route(wid='negative_consequences', role='cq_answer', cq='CQ:likelihood',
                                     refuse_cq=True, conclusion_kind='different', parent_wid='verbal_classification',
                                     cq_error='S8: premise "brings_about(tax_meat,starvation_from_reduced_meat_access)" not found in argument "a1"')
        self.assertEqual(self.rows['parent']['integration'], 'coverage_gap')
        self.assertEqual(result['played'], 'statement')
        self.assertTrue(any('S8' in e for e in result['fallback_after']))
        self.assertIs(result['conclusion_preserved'], False)
        self.assertEqual(result['integration'], 'cascade')
        # The same without the refusal: the question route is the first, its move as intended in role,
        # its argument concluding another formula: conclusion_changed, and never a cascade.
        result, _ = self.route(wid='negative_consequences', role='cq_answer', cq='CQ:likelihood',
                               conclusion_kind='different', parent_wid='verbal_classification')
        self.assertEqual(result['played'], 'cq')
        self.assertNotIn('fallback_after', result)
        self.assertEqual(result['integration'], 'conclusion_changed')


if __name__ == '__main__':
    unittest.main(verbosity=2)
