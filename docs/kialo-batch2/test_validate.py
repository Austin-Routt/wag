"""Planted-defect tests of validate.py's rules: R8 to R11 (decisions D075 and
D089; proposal D004-D075), and the identity contract's R12 to R17 with the
format-2 forms of R4, R9 and R10 (decision D092, D039;
docs/annotation-records.md).

    python3 -I test_validate.py

Each test writes a copy of gold.json with one defect planted (and, where the
test needs one, a copy of case.json, the gold's pin following it), runs
validate.py on it, and checks that the rule reports it and nothing else
does; the gold itself must pass, and so must the format-1 gold kept beside it.
"""
import copy
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
import unittest

DOC = os.path.dirname(os.path.abspath(__file__))
with open(os.path.join(DOC, 'gold.json'), encoding='utf-8') as fh:
    GOLD = json.load(fh)
with open(os.path.join(DOC, 'gold-format1.json'), encoding='utf-8') as fh:
    GOLD1 = json.load(fh)
with open(os.path.join(DOC, 'case.json'), encoding='utf-8') as fh:
    CASE = json.load(fh)
ERROR = re.compile(r'(\d+\.\d+|case\.json|inputs): ')
# The walton.yml the format-1 gold was made with (WAG 96c3c5a), and today's.
WALTON_OF_FORMAT1 = 'bb1edbcc69536bad04dc33ce8e8c3578d3a610e9abb6fadf1ea42d67d1badf42'
with open(os.path.join(DOC, '..', 'scheme-registry', 'walton-clean', 'walton.yml'), 'rb') as fh:
    WALTON_NOW = hashlib.sha256(fh.read()).hexdigest()


def run(gold, case=None, pin=True):
    """validate.py's exit status, its error lines and its whole output. With
    a case, it is written beside the gold and named with --case, and the
    gold's pin follows it unless pin is false."""
    with tempfile.TemporaryDirectory() as d:
        path = os.path.join(d, 'gold.json')
        cmd = [sys.executable, '-I', '-B', os.path.join(DOC, 'validate.py'), path]
        if case is not None:
            cpath = os.path.join(d, 'case.json')
            text = json.dumps(case, ensure_ascii=False, indent=1)
            with open(cpath, 'w', encoding='utf-8') as fh:
                fh.write(text)
            if pin and 'inputs' in gold:
                gold = dict(gold, inputs=dict(gold['inputs'], **{'case.json': hashlib.sha256(text.encode('utf-8')).hexdigest()}))
            cmd += ['--case', cpath]
        with open(path, 'w', encoding='utf-8') as fh:
            json.dump(gold, fh, ensure_ascii=False)
        r = subprocess.run(cmd, capture_output=True, text=True)
    errors = [line for line in r.stdout.splitlines() if ERROR.match(line)]
    return r.returncode, errors, r.stdout + r.stderr


def item(g, cid):
    return next(i for i in g['items'] if i['kialo_claim_id'] == cid)


def primary(it):
    return next(x for x in it['instances'] if x.get('primary'))


def inst(g, iid):
    return next(x for i in g['items'] for x in i['instances'] if x.get('id') == iid)


def rename(it, ins, old, new, records=True):
    """Replace an atom in an instance's bindings and formulas (and in the
    claim's challenge records that name the instance's formulas), keeping
    the instance walton.yml's."""
    def sub(f):
        return re.sub(r'(?<![A-Za-z0-9_])' + re.escape(old) + r'(?![A-Za-z0-9_])', new, f)
    mine = {p['formula'] for p in ins['premises']} | {ins['conclusion']['formula']}
    ins['bindings'] = {k: (new if v == old else v) for k, v in ins['bindings'].items()}
    for p in ins['premises']:
        p['formula'] = sub(p['formula'])
    ins['conclusion']['formula'] = sub(ins['conclusion']['formula'])
    for c in it['challenges'] if records else []:
        # A format-2 record names its instance: only the renamed one's follow.
        if c.get('formula') in mine and c.get('instance', ins.get('id')) == ins.get('id'):
            c['formula'] = sub(c['formula'])


def next_id(it):
    used = [int(x['id'].split('#')[1]) for x in it['instances']] + [int(x.split('#')[1]) for x in it.get('retired', [])]
    return f"{it['kialo_claim_id']}#{max(used) + 1}"


def conclusion_target(g, claim, iid=None):
    """A format-2 target at an instance's conclusion (the claim's primary
    instance unless another is named)."""
    ins = inst(g, iid) if iid else primary(item(g, claim))
    return {'claim': claim, 'instance': ins['id'], 'part': 'conclusion', 'formula': ins['conclusion']['formula']}


class Rules(unittest.TestCase):
    def fails(self, g, cid, rules, text, case=None, pin=True):
        """validate.py fails, on cid only, by the rules given, one error
        containing text."""
        code, errors, out = run(g, case, pin)
        self.assertEqual(code, 1, out)
        for e in errors:
            self.assertTrue(any(e.startswith(f'{cid}: {r} ') for r in rules), f'unexpected error: {e}\n{out}')
        self.assertTrue(any(text in e for e in errors), out)

    def passes(self, g, case=None):
        code, errors, out = run(g, case)
        self.assertEqual((code, errors), (0, []), out)
        return out

    def test_the_gold_passes(self):
        out = self.passes(GOLD)
        self.assertIn('format 2', out)
        self.assertNotIn('UNRESOLVED', out)

    def test_the_format1_gold_passes_as_format1(self):
        # The format-1 gold is kept as it was (D039); it is checked against
        # the walton.yml it was made with only while that is the current
        # one. After a change to walton.yml, it is read as format 1 and its
        # verdict said, but formulas that changed since are not its fault.
        if WALTON_NOW != WALTON_OF_FORMAT1:
            code, errors, out = run(GOLD1)
            self.assertIn('format 1', out)
            self.assertNotIn('Traceback', out)
            return
        out = self.passes(GOLD1)
        self.assertIn('format 1', out)

    def test_format1_named(self):
        g = copy.deepcopy(GOLD1)
        g['format'] = 'wag-annotation/1'
        code, errors, out = run(g)
        self.assertIn('format 1', out)
        if WALTON_NOW == WALTON_OF_FORMAT1:
            self.assertEqual((code, errors), (0, []), out)

    # R8: a claim at an exception concludes it.
    def test_r8_an_argument_for_another_course(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        for x in it['instances']:
            x['primary'] = x['scheme']['walton_id'] == 'practical_reasoning2'
        self.fails(g, '13292.123', ['R8'], 'concludes should_be_performed(abolish_meat_subsidies)')

    def test_r8_whatever_the_role(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.127')
        it['role'].update(kind='rebuttal', cq=None)
        it['target']['actual']['formula'] = 'incompatible_goal(governments,environmental_harm_from_meat_reduced)'
        self.fails(g, '13292.127', ['R2', 'R8'], 'R8 an answer at the exception incompatible_goal(')

    def test_r8_a_defence_concludes_the_negation(self):
        # Polarity for at an exception argues that it does not hold, as the
        # replay plays it: the claim concludes the exception's negation.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.127')
        f = it['target']['actual']['formula']
        for x in it['instances']:
            x['primary'] = False
        it['instances'].insert(0, {
            'id': next_id(it), 'primary': True,
            'scheme': {'walton_id': 'other', 'compendium': None, 'wag_id': None, 'fit': 'reconstructed', 'confidence': 0.5,
                       'alternatives': []},
            'bindings': {},
            'premises': [{'role': 'premise', 'formula': 'brings_about(tax_meat,starvation_from_reduced_meat_access)',
                          'en': '…', 'status': 'explicit'}],
            'conclusion': {'formula': '¬' + f, 'en': '…'}})
        it['role']['polarity'] = 'for'
        self.passes(g)
        it['instances'][0]['conclusion']['formula'] = f
        self.fails(g, '13292.127', ['R8'], 'not ¬' + f)

    def test_r8_spaces_and_the_order_of_instances(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.127')
        it['target']['actual']['formula'] = it['target']['actual']['formula'].replace(',', ', ')
        it['instances'].reverse()
        self.passes(g)

    # R9: a better means is another action, records its comparison and, as
    # the answer, has the answered argument's goal premise.
    def test_r9_the_same_action(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        ins = primary(it)
        rename(it, ins, ins['bindings']['A2'], ins['bindings']['A'])
        self.fails(g, '13292.123', ['R9'], 'compares tax_meat with itself')

    def test_r9_an_alias_is_the_same_action(self):
        # Two names for one action do not make another course (D063, D091).
        case = copy.deepcopy(CASE)
        next(a for a in case['actions'] if a['id'] == 'tax_meat')['aliases'] = ['tax_damaging_activity']
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        ins = primary(it)
        rename(it, ins, ins['bindings']['A2'], 'tax_damaging_activity')
        self.fails(g, '13292.123', ['R9'], 'compares tax_meat with itself (tax_damaging_activity is its alias)', case=case)

    def test_r9_no_comparison(self):
        for value in (None, 'better', 'more_surely'):
            with self.subTest(value=value):
                g = copy.deepcopy(GOLD)
                ins = primary(item(g, '13292.104'))
                if value is None:
                    del ins['comparison']
                else:
                    ins['comparison'] = value
                self.fails(g, '13292.104', ['R9'], 'records no comparison')

    def test_r9_each_comparison(self):
        for value in ('more surely', 'sooner', 'at lower cost'):
            with self.subTest(value=value):
                g = copy.deepcopy(GOLD)
                primary(item(g, '13292.104'))['comparison'] = value
                self.passes(g)

    def test_r9_another_goal(self):
        # Away from the exception, where R8 does not apply (R11 refuses that
        # too), and whatever the role.
        for kind, cq in (('cq_answer', '22.1 CQ2'), ('rebuttal', None)):
            with self.subTest(kind=kind):
                g = copy.deepcopy(GOLD)
                it = item(g, '13292.123')
                it['target']['actual'] = conclusion_target(g, '13292.118')
                it['role'].update(kind=kind, cq=cq)
                ins = primary(it)
                rename(it, ins, ins['bindings']['G'], 'deforestation_slowed')
                self.fails(g, '13292.123', ['R9', 'R11'], "is not the answered argument's")

    def test_r9_a_secondary_better_means(self):
        # Only the answer itself must share the answered argument's goal.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.104')
        dup = copy.deepcopy(primary(it))
        dup['primary'], dup['id'] = False, next_id(it)
        rename(it, dup, dup['bindings']['G'], 'deforestation_slowed', records=False)
        it['instances'].append(dup)
        self.passes(g)

    # R11: an answer scheme's primary instance is at an exception.
    def test_r11_an_answer_scheme_away_from_the_exception(self):
        for cid in ('13292.127', '13292.43', '13292.123'):
            with self.subTest(claim=cid):
                g = copy.deepcopy(GOLD)
                it = item(g, cid)
                a = it['target']['actual']
                it['target']['actual'] = conclusion_target(g, a['claim'], a['instance'])
                self.fails(g, cid, ['R11'], 'is not at an exception')

    def test_r11_a_fragment_backing_an_answer_schemes_premise(self):
        # A claim that states one premise of its parent's answer carries the
        # parent's scheme as a fragment (§2, §3), at that premise: no R9 or
        # R11 for it. The same instance elsewhere, or not a fragment, is.
        def plant(fit, part):
            g = copy.deepcopy(GOLD)
            parent = primary(item(g, '13292.104'))
            it = item(g, '13292.208')
            frag = copy.deepcopy(parent)
            frag['id'] = next_id(it)
            frag['scheme'] = {'walton_id': 'better_means', 'compendium': None, 'wag_id': None, 'fit': fit,
                              'confidence': 0.5, 'alternatives': []}
            frag.pop('comparison', None)
            it['instances'] = [frag] + [dict(x, primary=False) for x in it['instances']]
            f = 'brings_about(regulate_animal_farming_industry,environmental_harm_from_meat_reduced)'
            it['target']['actual'] = ({'claim': '13292.104', 'instance': '13292.104#1', 'part': 'premise', 'formula': f}
                                      if part == 'premise' else conclusion_target(g, '13292.104'))
            it['role'].update(kind='premise_backing', polarity='for')
            return g
        self.passes(plant('fragment', 'premise'))
        self.fails(plant('fragment', 'conclusion'), '13292.208', ['R11'], 'is not at an exception')
        self.fails(plant('reconstructed', 'premise'), '13292.208', ['R9', 'R11'], 'is not at an exception')

    def test_r11_a_secondary_answer_scheme(self):
        # Only the claim's own argument must be at the exception.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.127')
        for x in it['instances']:
            x['primary'] = x['scheme']['walton_id'] == 'negative_consequences'
        it['role'].update(kind='rebuttal', cq=None)
        it['target']['actual'] = conclusion_target(g, '13292.108')
        code, errors, out = run(g)
        self.assertFalse([e for e in errors if e.startswith('13292.127: R11')], out)

    # R10: a challenge names an instance, its own or its target's, and a
    # formula of that instance.
    def test_r10_a_formula_renamed_in_walton_yml(self):
        g = copy.deepcopy(GOLD)
        rec = next(c for c in item(g, '13292.108')['challenges'] if c.get('cq') == '22.1 CQ5')
        rec['formula'] = 'side_effects(tax_meat,meat_production_and_consumption_harm_environment)'
        self.fails(g, '13292.108', ['R10'], 'names side_effects(tax_meat,')

    def test_r10_the_targets_other_instance(self):
        g = copy.deepcopy(GOLD)
        it, target = item(g, '13292.43'), item(g, '13292.9')
        other = next(x for x in target['instances'] if not x['primary'])
        f = other['premises'][0]['formula']
        self.assertNotIn(f, [p['formula'] for p in primary(target)['premises']])
        it['challenges'].append({'cq': None, 'wag_cq': None, 'type': 'causal interference', 'instance': other['id'], 'formula': f})
        self.fails(g, '13292.43', ['R10'], f"names the instance '{other['id']}', neither one of its own nor its target's")
        it['challenges'][-1]['instance'] = primary(target)['id']
        self.fails(g, '13292.43', ['R10'], f"names {f}, which is not a formula of {primary(target)['id']}")

    def test_r10_the_wrong_one_of_its_own_instances(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.127')
        rec = next(c for c in it['challenges'] if c.get('formula', '').startswith('matters_more('))
        rec['instance'] = next(x['id'] for x in it['instances'] if not x['primary'])
        self.fails(g, '13292.127', ['R10'], 'names matters_more(')

    def test_r10_a_challenge_of_a_whole_instance(self):
        g = copy.deepcopy(GOLD)
        rec = next(c for c in item(g, '13292.8')['challenges'] if c.get('formula') is None)
        self.assertEqual(rec['instance'], '13292.8#1')
        rec['instance'] = '13292.8#9'
        self.fails(g, '13292.8', ['R10'], "names the instance '13292.8#9'")

    # R4 (D039): in format 2 a cause that has not occurred is the modal
    # scheme's; format 1 keeps the rule of the N8 note.
    def test_r4_a_cause_the_case_deliberates(self):
        # The deliberated action has not occurred, whatever the status says.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.9')
        step = next(x for x in it['instances'] if x['bindings'].get('E1') == 'tax_meat')
        step['scheme']['walton_id'] = 'cause_to_effect'
        p0 = next(p for p in step['premises'] if p['formula'].startswith('may_occur('))
        p0['formula'], p0['status'] = 'has_occurred(tax_meat)', 'inferred'
        step['conclusion']['formula'] = 'will_occur(' + step['bindings']['E2'] + ')'
        self.fails(g, '13292.9', ['R4'], 'an action the case deliberates, which has not occurred')

    def test_r4_a_cause_a_modal_step_concludes(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.9')
        step = next(x for x in it['instances'] if x['bindings'].get('E1') == 'meat_industry_harmed')
        step['scheme']['walton_id'] = 'cause_to_effect'
        p0 = next(p for p in step['premises'] if p['formula'].startswith('may_occur('))
        p0['formula'], p0['status'] = 'has_occurred(meat_industry_harmed)', 'explicit'
        step['conclusion']['formula'] = 'will_occur(tax_revenue_reduced)'
        self.fails(g, '13292.9', ['R4'], 'which a modal step of the claim says only may occur')

    def test_r4_a_supplied_occurrence_that_has_occurred(self):
        # "supplied" says who states the premise, not that the cause is
        # hypothetical: subsidies exist, so cause_to_effect stands.
        g = copy.deepcopy(GOLD)
        ins = next(x for x in item(g, '13292.11')['instances'] if x['scheme']['walton_id'] == 'cause_to_effect')
        next(p for p in ins['premises'] if p['formula'].startswith('has_occurred('))['status'] = 'supplied'
        self.passes(g)

    def test_r4_in_format1_is_the_n8_rule(self):
        if WALTON_NOW != WALTON_OF_FORMAT1:
            self.skipTest('walton.yml has changed since the format-1 gold was made')
        g = copy.deepcopy(GOLD1)
        ins = item(g, '13292.36')['instances'][1]
        self.assertEqual(ins['scheme']['walton_id'], 'cause_to_effect')
        ins['note'] = 'a causal step'
        self.fails(g, '13292.36', ['R4'], 'R4 not applied to a hypothetical cause_to_effect')

    # R12: ids.
    def test_r12_ids(self):
        cases = {
            'another claim': lambda it: it['instances'][0].update(id='13292.9#9'),
            'a leading zero': lambda it: it['instances'][0].update(id='13292.12#01'),
            'zero': lambda it: it['instances'][0].update(id='13292.12#0'),
            'no number': lambda it: it['instances'][0].update(id='13292.12'),
            'twice': lambda it: it['instances'][1].update(id=it['instances'][0]['id']),
            'retired': lambda it: it.update(retired=[it['instances'][2]['id']]),
        }
        for name, plant in cases.items():
            with self.subTest(name):
                g = copy.deepcopy(GOLD)
                it = item(g, '13292.12')
                it['challenges'] = [c for c in it['challenges'] if c['instance'] != it['instances'][0]['id']]
                plant(it)
                # The claims that target 13292.12#1 lose it when its id changes (R13).
                code, errors, out = run(g)
                self.assertEqual(code, 1, out)
                self.assertTrue(any(e.startswith('13292.12: R12 ') for e in errors), out)
                for e in errors:
                    self.assertTrue(e.startswith('13292.12: R12 ') or re.match(r'13292\.(194|45|12): R1[03] ', e), e)

    def test_r12_no_id(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['instances'][0].pop('id')
        code, errors, out = run(g)
        # The item is malformed: said once, and the two claims that target
        # 13292.12#1 do not read it (no cascade).
        self.assertEqual(errors, ['13292.12: R16 instances[0].id is not an instance id'], out)

    def test_r12_a_retired_number_is_not_used_again(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.104')
        it['retired'] = ['13292.104#4']
        self.passes(g)
        dup = copy.deepcopy(it['instances'][2])
        dup['id'] = '13292.104#4'
        it['instances'].append(dup)
        self.fails(g, '13292.104', ['R12'], 'the instance 13292.104#4 has a retired id')

    # R13: exact targets.
    def test_r13_an_instance_that_is_not_there(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['instance'] = '13292.9#7'
        self.fails(g, '13292.12', ['R13'], 'names the instance 13292.9#7, which is not an instance of 13292.9')

    def test_r13_an_instance_of_another_claim(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['instance'] = '13292.7#1'
        self.fails(g, '13292.12', ['R13'], 'not an instance of 13292.9')

    def test_r13_a_formula_of_another_instance_of_the_claim(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['instance'] = '13292.9#2'
        self.fails(g, '13292.12', ['R13'], 'brings_about(tax_meat,tax_revenue_reduced), which is not a premise of 13292.9#2')

    def test_r13_the_wrong_conclusion(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.9')['target']['actual']['formula'] = '¬should_be_performed(tax_meat)'
        self.fails(g, '13292.9', ['R13'], '¬should_be_performed(tax_meat), which is not a conclusion of 13292.1#1')

    def test_r13_a_secondary_instance_can_be_targeted(self):
        # The point of instance ids: a reply to a claim's secondary inference.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.12')
        it['target']['actual'] = {'claim': '13292.9', 'instance': '13292.9#3', 'part': 'premise',
                                  'formula': 'causes(meat_industry_harmed,tax_revenue_reduced)'}
        code, errors, out = run(g)
        self.assertFalse([e for e in errors if ' R13 ' in e], out)
        self.assertTrue(all(e.startswith('13292.12: R2 ') for e in errors), out)

    def test_r13_another_kialo_parent(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.180')['target']['other_kialo_parents'][0]['instance'] = '13292.121#2'
        self.fails(g, '13292.180', ['R13'], 'names the instance 13292.121#2, which is not an instance of 13292.121')
        g = copy.deepcopy(GOLD)
        del item(g, '13292.180')['target']['other_kialo_parents'][0]['instance']
        self.fails(g, '13292.180', ['R16'], 'target.other_kialo_parents[0].instance is not an instance id')

    def test_r13_unresolved_is_kept_not_refused(self):
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['instance'] = {'unresolved': 'the step or the consequence argument'}
        out = self.passes(g)
        self.assertIn('UNRESOLVED (1)', out)
        self.assertIn('13292.12: its target instance (the step or the consequence argument)', out)
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['instance'] = {'unresolved': ''}
        self.fails(g, '13292.12', ['R16'], 'target.actual.instance is not an instance id, or {"unresolved": why}')

    # R14: actors and actions.
    def test_r14_an_agent_who_is_not_an_actor(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.6')
        rename(it, primary(it), 'governments', 'meat_tax_proponents')
        self.fails(g, '13292.6', ['R14'], 'binds the agent P to meat_tax_proponents, which is not an actor of the case')

    def test_r14_an_action_the_case_does_not_have(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        ins = primary(it)
        rename(it, ins, ins['bindings']['A2'], 'switch_to_other_protein')
        self.fails(g, '13292.123', ['R14'], 'binds the action A2 to switch_to_other_protein, which is not an action of the case')

    def test_r14_another_actors_action(self):
        # The wrong actor: people can switch to other protein; governments
        # cannot, so it is no better means of theirs.
        case = copy.deepcopy(CASE)
        case['actors'].append({'id': 'meat_tax_proponents', 'text': 'the proponents'})
        case['actions'].append({'id': 'switch_to_other_protein', 'actor': 'meat_tax_proponents', 'aliases': []})
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        ins = primary(it)
        rename(it, ins, ins['bindings']['A2'], 'switch_to_other_protein')
        self.fails(g, '13292.123', ['R14'], "which is meat_tax_proponents's action, not governments's", case=case)

    def test_r14_a_consequence_scheme_binds_an_action(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.7')
        rename(it, primary(it), 'tax_meat', 'meat_tax_case')
        code, errors, out = run(g)
        self.assertTrue(any(e.startswith('13292.7: R14 ') and 'binds the action A to meat_tax_case' in e for e in errors), out)

    def test_r14_an_alias_is_its_action(self):
        case = copy.deepcopy(CASE)
        # An atom of the lexicon the gold does not use (meat_subsidies, the
        # state, is a cause in 13292.11: no alias of an action).
        next(a for a in case['actions'] if a['id'] == 'abolish_meat_subsidies')['aliases'] = ['cheap_open_meat_trade']
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.123')
        ins = next(x for x in it['instances'] if x['scheme']['walton_id'] == 'practical_reasoning2')
        rename(it, ins, 'abolish_meat_subsidies', 'cheap_open_meat_trade')
        out = self.passes(g, case=case)
        self.assertIn('format 2', out)

    def test_r14_the_case_record(self):
        def plant(change):
            case = copy.deepcopy(CASE)
            change(case)
            return case
        cases = {
            'an alias that is an action': (lambda c: c['actions'][0]['aliases'].append('raise_import_standards'),
                                           'cannot be an alias of tax_meat'),
            'an actor the case lacks': (lambda c: c['actions'][1].update(actor='people'),
                                        "is performed by 'people', not an actor of the case"),
            'an action not in the lexicon': (lambda c: c['actions'].append({'id': 'tax_fish', 'actor': 'governments', 'aliases': []}),
                                             'the action tax_fish is not an atom of the lexicon'),
            'a thesis action the case lacks': (lambda c: c['thesis'].update(action='tax_fish'),
                                               "the thesis's action 'tax_fish' is not an action of the case"),
        }
        for name, (change, text) in cases.items():
            with self.subTest(name):
                self.fails(copy.deepcopy(GOLD), 'case.json', ['R14'], text, case=plant(change))

    def test_r14_the_thesis_concludes_the_case_action(self):
        case = copy.deepcopy(CASE)
        case['thesis']['action'] = 'abolish_meat_subsidies'
        self.fails(copy.deepcopy(GOLD), '13292.1', ['R14'], 'not should_be_performed(abolish_meat_subsidies)', case=case)

    # R15: contexts.
    def test_r15_a_context_the_case_lacks(self):
        g = copy.deepcopy(GOLD)
        primary(item(g, '13292.7'))['context'] = 'denmark_2005'
        code, errors, out = run(g)
        self.assertTrue(any(e.startswith("13292.7: R15 the instance 13292.7#1 is in the context 'denmark_2005'") for e in errors), out)

    def test_r15_an_answer_in_another_context(self):
        # The wrong context: a better means found for another situation
        # does not answer this one's argument.
        case = copy.deepcopy(CASE)
        case['contexts'].append({'id': 'denmark_2005', 'text': 'Denmark in 2005'})
        g = copy.deepcopy(GOLD)
        primary(item(g, '13292.104'))['context'] = 'denmark_2005'
        code, errors, out = run(g, case)
        # 13292.208, which replies to 13292.104's answer, is now in another
        # context than its target too.
        self.assertEqual(sorted(errors), [
            "13292.104: R15 its primary instance is in the context 'denmark_2005', its target 13292.108#1 in 'present'",
            "13292.208: R15 its primary instance is in the context 'present', its target 13292.104#1 in 'denmark_2005'"], out)
        # Both in the other context: no complaint.
        primary(item(g, '13292.108'))['context'] = 'denmark_2005'
        code, errors, out = run(g, case)
        self.assertFalse([e for e in errors if e.startswith('13292.104: R15')], out)

    def test_r15_the_default_context(self):
        case = copy.deepcopy(CASE)
        case['default_context'] = 'denmark_2005'
        self.fails(copy.deepcopy(GOLD), 'case.json', ['R15'], "its default context 'denmark_2005' is not one of its contexts", case=case)

    # R16: no value without its kind. The field contract is checked first; a
    # malformed record is reported, never a traceback, and not checked further.
    def test_r16_the_field_contract(self):
        def ins7(g):
            return primary(item(g, '13292.7'))
        cases = {
            're_parented': ('13292.7', lambda g: item(g, '13292.7')['target'].update(re_parented=False), 'target.re_parented is recorded'),
            'no polarity': ('13292.7', lambda g: item(g, '13292.7')['role'].update(polarity=None), 'role.polarity is None'),
            'a polarity for the thesis': ('13292.1', lambda g: item(g, '13292.1')['role'].update(polarity='for'), 'the thesis has no polarity'),
            'no actual target': ('13292.7', lambda g: item(g, '13292.7')['target'].update(actual=None), 'only the thesis has no actual target'),
            'an empty actual target': ('13292.7', lambda g: item(g, '13292.7')['target'].update(actual={}), 'target.actual is {}'),
            'a target for the thesis': ('13292.1', lambda g: item(g, '13292.1')['target'].update(actual=conclusion_target(g, '13292.7')), 'the thesis has no target'),
            'another relation': ('13292.7', lambda g: item(g, '13292.7')['target'].update(kialo_relation='pro'), "target.kialo_relation is 'pro', not 'con'"),
            'no relation': ('13292.7', lambda g: item(g, '13292.7')['target'].update(kialo_relation=None), 'target.kialo_relation is None'),
            'no kialo_parent': ('13292.7', lambda g: item(g, '13292.7')['target'].pop('kialo_parent'), 'target.kialo_parent is missing'),
            'a binding left null': ('13292.7', lambda g: ins7(g)['bindings'].update(G=None), '13292.7#1.bindings.G is None'),
            'unstated outside a fragment': ('13292.7', lambda g: ins7(g)['bindings'].update(G={'unstated': 'not said'}), "13292.7#1.bindings.G is {'unstated'"),
            'unresolved in a binding': ('13292.7', lambda g: ins7(g)['bindings'].update(G={'unresolved': 'which harm'}), "13292.7#1.bindings.G is {'unresolved'"),
            'no primary flag': ('13292.7', lambda g: ins7(g).pop('primary'), '13292.7#1.primary is not true or false'),
            'no compendium': ('13292.7', lambda g: ins7(g)['scheme'].pop('compendium'), '13292.7#1.scheme.compendium is not text, or null'),
            'no fit': ('13292.7', lambda g: ins7(g)['scheme'].pop('fit'), '13292.7#1.scheme.fit is None'),
            'a null confidence': ('13292.7', lambda g: ins7(g)['scheme'].update(confidence=None), '13292.7#1.scheme.confidence is not a number'),
            'null alternatives': ('13292.7', lambda g: ins7(g)['scheme'].update(alternatives=None), '13292.7#1.scheme.alternatives is not a list'),
            'a marker for a scheme': ('13292.7', lambda g: ins7(g)['scheme'].update(walton_id={'unresolved': 'x'}), '13292.7#1.scheme.walton_id is not a scheme id'),
            'a marker for an id': ('13292.7', lambda g: ins7(g).update(id={'unresolved': 'x'}), 'instances[0].id is not an instance id'),
            'a marker for a context': ('13292.7', lambda g: ins7(g).update(context={'unresolved': 'x'}), '13292.7#1.context is not a context id'),
            'no bindings': ('13292.7', lambda g: ins7(g).pop('bindings'), '13292.7#1.bindings is not an object'),
            'no premises': ('13292.7', lambda g: ins7(g).pop('premises'), '13292.7#1.premises is not a list'),
            'a marker for a premise formula': ('13292.7', lambda g: ins7(g)['premises'][0].update(formula={'unresolved': 'x'}), '13292.7#1.premises[0].formula is not a formula'),
            'a null premise status': ('13292.7', lambda g: ins7(g)['premises'][0].update(status=None), '13292.7#1.premises[0].status is None'),
            'a null premise English': ('13292.7', lambda g: ins7(g)['premises'][0].update(en=None), '13292.7#1.premises[0].en is not text'),
            'a conclusion without English': ('13292.7', lambda g: ins7(g)['conclusion'].pop('en'), '13292.7#1.conclusion is not an object'),
            'a null conclusion': ('13292.7', lambda g: ins7(g)['conclusion'].update(formula=None), 'only other, none and a fragment'),
            'a null literal': ('13292.7', lambda g: item(g, '13292.7').update(literal=None), "literal is not a list of the claim's propositions"),
            'no role kind': ('13292.7', lambda g: item(g, '13292.7')['role'].pop('kind'), 'role.kind is None'),
            'an unknown role kind': ('13292.7', lambda g: item(g, '13292.7')['role'].update(kind='banana'), "role.kind is 'banana'"),
            'a marker for a role': ('13292.7', lambda g: item(g, '13292.7')['role'].update(kind={'unresolved': 'x'}), "role.kind is {'unresolved'"),
            'a null role confidence': ('13292.7', lambda g: item(g, '13292.7')['role'].update(confidence=None), 'role.confidence is not a number'),
            'an empty cq off cq_answer': ('13292.7', lambda g: item(g, '13292.7')['role'].update(cq=''), "role.cq is '': null"),
            'no cq': ('13292.7', lambda g: item(g, '13292.7')['role'].pop('cq'), 'role.cq is missing'),
            'no challenges': ('13292.7', lambda g: item(g, '13292.7').pop('challenges'), 'challenges is not a list'),
            'a challenge without formula': ('13292.7', lambda g: item(g, '13292.7')['challenges'][0].pop('formula'), 'challenges[0].formula is not a formula, or null'),
            'a marker for a challenge formula': ('13292.7', lambda g: item(g, '13292.7')['challenges'][0].update(formula={'unresolved': 'x'}), 'challenges[0].formula is not a formula'),
            'a null challenge type': ('13292.7', lambda g: item(g, '13292.7')['challenges'][0].update(type=None), 'challenges[0].type is not a type'),
            'a challenge without cq': ('13292.7', lambda g: item(g, '13292.7')['challenges'][0].pop('cq'), "challenges[0].cq is not a question's label, or null when there is none"),
            'a marker for a wag_cq': ('13292.7', lambda g: item(g, '13292.7')['challenges'][0].update(wag_cq={'unresolved': 'x'}), "challenges[0].wag_cq is not a question's label"),
            'no evidence urls': ('13292.7', lambda g: item(g, '13292.7')['evidence'].pop('urls'), 'evidence needs status'),
            'no candidate_pattern': ('13292.7', lambda g: item(g, '13292.7')['notes'].pop('candidate_pattern'), 'notes needs text and candidate_pattern'),
            'no actual claim': ('13292.12', lambda g: item(g, '13292.12')['target']['actual'].pop('claim'), 'target.actual.claim is not a claim'),
            'no actual part': ('13292.12', lambda g: item(g, '13292.12')['target']['actual'].pop('part'), 'target.actual.part is None'),
            'a marker for an actual claim': ('13292.12', lambda g: item(g, '13292.12')['target']['actual'].update(claim={'unresolved': 'x'}), 'target.actual.claim is not a claim'),
            'an inference with a formula': ('13292.12', lambda g: item(g, '13292.12')['target']['actual'].update(part='inference'),
                                            'an inference target has no formula'),
            'an inference target without its formula': ('13292.12', lambda g: (item(g, '13292.12')['target']['actual'].update(part='inference'),
                                                                 item(g, '13292.12')['target']['actual'].pop('formula')), 'target.actual.formula is missing'),
            'other parents for one parent': ('13292.7', lambda g: item(g, '13292.7')['target'].update(other_kialo_parents=[]), 'the claim has one Kialo parent or none (R7)'),
            'retired not a list': ('13292.7', lambda g: item(g, '13292.7').update(retired='13292.7#2'), "retired is not a list of ids of 13292.7's instances"),
            "a retired id of another claim": ('13292.7', lambda g: item(g, '13292.7').update(retired=['13292.9#9']), "retired is not a list of ids of 13292.7's instances"),
        }
        for name, (cid, plant, text) in cases.items():
            with self.subTest(name):
                g = copy.deepcopy(GOLD)
                plant(g)
                code, errors, out = run(g)
                self.assertEqual(code, 1, out)
                self.assertNotIn('Traceback', out)
                self.assertTrue(any(e.startswith(f'{cid}: R16 ') and text in e for e in errors), f'{name}:\n{out}')
                for e in errors:
                    self.assertTrue(e.startswith(f'{cid}: R16 ') or ' R13 ' in e or ' R10 ' in e or ' R15 ' in e, f'{name}: {e}')

    def test_r16_other_parents_missing(self):
        # A claim with more than one Kialo parent lists the others (R7 then
        # checks the list is complete).
        g = copy.deepcopy(GOLD)
        item(g, '13292.180')['target'].pop('other_kialo_parents')
        self.fails(g, '13292.180', ['R16'], 'target.other_kialo_parents is missing: the claim has more than one Kialo parent (R7)')

    def test_r16_a_challenge_without_a_printed_question(self):
        # A challenge the book prints no question for has cq null (§7, §8).
        g = copy.deepcopy(GOLD)
        self.assertTrue(any(c['cq'] is None for i in g['items'] for c in i['challenges']))
        self.passes(g)

    def test_r16_the_cases_id(self):
        g = copy.deepcopy(GOLD)
        g['case'] = 'kialo-99999'
        self.fails(g, 'case.json', ['R16'], "the file names the case 'kialo-99999'")

    def test_r13_another_kialo_parents_conclusion(self):
        g = copy.deepcopy(GOLD)
        o = item(g, '13292.180')['target']['other_kialo_parents'][0]
        o.update(part='conclusion', formula='should_be_performed(tax_meat)', role='rebuttal', cq=None)
        self.fails(g, '13292.180', ['R13'], 'should_be_performed(tax_meat), which is not a conclusion of 13292.121#1')

    def test_r16_the_pinned_inputs(self):
        g = copy.deepcopy(GOLD)
        self.assertEqual(len(g['inputs']), 10, 'every input validate.py reads is pinned')
        g['inputs']['walton.yml'] = '0' * 64
        self.fails(g, 'inputs', ['R16'], 'the gold was checked against another walton.yml')
        g['inputs']['positions.json'] = '0' * 64
        self.fails(g, 'inputs', ['R16'], 'another positions.json')
        g = copy.deepcopy(GOLD)
        del g['inputs']['lexicon.json']
        self.fails(g, 'inputs', ['R16'], 'a gold pins exactly')
        case = copy.deepcopy(CASE)
        case['status'] += ' Changed.'
        self.fails(copy.deepcopy(GOLD), 'inputs', ['R16'], 'another case.json', case=case, pin=False)

    def test_r16_an_annotators_file_pins_nothing(self):
        g = copy.deepcopy(GOLD)
        g['annotator'] = 'A'
        del g['inputs']
        for it in g['items']:
            del it['adjudication']
        self.passes(g)

    def test_r16_a_fragment_leaves_variables_unstated(self):
        # A fragment that states one premise binds only that premise's
        # variables; the others are unstated, not null.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.137')
        frag = copy.deepcopy(primary(item(g, '13292.108')))
        frag.update(id=next_id(it), primary=False)
        frag['scheme'] = {'walton_id': 'practical_reasoning2', 'compendium': '22.1', 'wag_id': 'AS:practical_reasoning_goal',
                          'fit': 'fragment', 'confidence': 0.5, 'alternatives': []}
        for v in ('A', 'S2', 'G', 'P'):
            frag['bindings'][v] = {'unstated': 'the claim states the circumstances only'}
        frag['premises'] = [p for p in frag['premises'] if p['formula'].startswith('current_circumstances(')]
        frag['conclusion'] = {'formula': None, 'en': ''}
        it['instances'].append(frag)
        self.passes(g)
        frag['scheme']['fit'] = 'reconstructed'
        code, errors, out = run(g)
        self.assertTrue(any(e.startswith('13292.137: R16 ') and ".bindings.A is {'unstated'" in e for e in errors), out)

    def test_a_fragment_states_some_of_its_bound_formulas(self):
        # A fragment that binds P, G and S1 may state goal_of_agent(P,G)
        # alone: it need not state current_circumstances(S1). A formula that
        # is not its scheme's is refused all the same.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.137')
        frag = copy.deepcopy(primary(item(g, '13292.108')))
        frag.update(id=next_id(it), primary=False)
        frag['scheme'] = {'walton_id': 'practical_reasoning2', 'compendium': '22.1', 'wag_id': 'AS:practical_reasoning_goal',
                          'fit': 'fragment', 'confidence': 0.5, 'alternatives': []}
        for v in ('A', 'S2'):
            frag['bindings'][v] = {'unstated': 'the claim states the goal only'}
        frag['premises'] = [p for p in frag['premises'] if p['formula'].startswith('goal_of_agent(')]
        frag['conclusion'] = {'formula': None, 'en': ''}
        it['instances'].append(frag)
        self.passes(g)
        frag['premises'].append({'role': 'premise', 'formula': 'possible_goal(meat_production)', 'en': '…', 'status': 'supplied'})
        self.fails(g, '13292.137', ['practical_reasoning2'], 'premises differ from walton.yml')

    # R17: verification.
    def test_r17_verification(self):
        # The revision is computed here as the contract states it (§9), not
        # by validate.py's own function.
        def rev(it):
            x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
            return hashlib.sha256(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8')).hexdigest()
        with open(os.path.join(DOC, 'items.json'), encoding='utf-8') as fh:
            text = next(i['text'] for i in json.load(fh)['items'] if i['kialo_claim_id'] == '13292.25')
        tsha = hashlib.sha256(text.encode('utf-8')).hexdigest()

        def verify(it, **kw):
            it['verified'] = dict({'by': 'owner', 'date': '2026-11-30', 'item_rev': rev(it), 'text_sha256': tsha}, **kw)
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.25')
        verify(it)
        self.passes(g)
        it['adjudication']['reason'] += ' (checked again)'
        self.passes(g)  # the adjudication record is not the item's content
        it['literal'][0] += ' (edited)'
        self.fails(g, '13292.25', ['R17'], 'verify it again')
        # The claim's text changed since: the verification does not hold.
        g = copy.deepcopy(GOLD)
        verify(item(g, '13292.25'), text_sha256='0' * 64)
        self.fails(g, '13292.25', ['R17'], "verified on another text of the claim")
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.25')
        it['target']['actual']['formula'] = {'unresolved': 'which premise'}
        verify(it)
        self.fails(g, '13292.25', ['R17'], 'some of its values are unresolved')
        for bad in ({'date': '30 November'}, {'date': '2026-13-45'}, {'by': 'claude-opus'}, {'text_sha256': None}):
            with self.subTest(bad=bad):
                g = copy.deepcopy(GOLD)
                verify(item(g, '13292.25'), **bad)
                self.fails(g, '13292.25', ['R17'], 'needs by "owner" (D086), a date')

    def test_a_fragment_conclusion_that_names_an_unstated_variable(self):
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.137')
        frag = copy.deepcopy(primary(item(g, '13292.108')))
        frag.update(id=next_id(it), primary=False)
        frag['scheme'] = {'walton_id': 'practical_reasoning2', 'compendium': '22.1', 'wag_id': 'AS:practical_reasoning_goal',
                          'fit': 'fragment', 'confidence': 0.5, 'alternatives': []}
        frag['bindings']['A'] = {'unstated': 'the claim states the circumstances only'}
        frag['premises'] = [p for p in frag['premises'] if p['formula'].startswith('current_circumstances(')]
        frag['conclusion'] = {'formula': 'should_be_performed(A)', 'en': ''}
        it['instances'].append(frag)
        self.fails(g, '13292.137', ['R16'], 'whose conclusion names a variable it leaves unstated')

    def test_an_unknown_format(self):
        g = copy.deepcopy(GOLD)
        g['format'] = 'wag-annotation/3'
        code, errors, out = run(g)
        self.assertEqual(code, 1, out)
        self.assertIn("the file is in format 'wag-annotation/3'", out)


class Repin(unittest.TestCase):
    """--repin pins the inputs only when nothing else is wrong."""
    def repin(self, g):
        with tempfile.TemporaryDirectory() as d:
            path = os.path.join(d, 'gold.json')
            with open(path, 'w', encoding='utf-8') as fh:
                fh.write(json.dumps(g, indent=1, ensure_ascii=False))
            r = subprocess.run([sys.executable, '-I', '-B', os.path.join(DOC, 'validate.py'), path, '--repin'], capture_output=True, text=True)
            with open(path, encoding='utf-8') as fh:
                return r.returncode, r.stdout, json.load(fh)

    def test_stale_pins_are_repinned(self):
        g = copy.deepcopy(GOLD)
        g['inputs']['links.json'] = '0' * 64
        code, out, after = self.repin(g)
        self.assertEqual(code, 0, out)
        self.assertEqual(after['inputs'], GOLD['inputs'])
        self.assertEqual(after['items'], GOLD['items'])

    def test_no_repin_while_something_else_is_wrong(self):
        g = copy.deepcopy(GOLD)
        g['inputs']['links.json'] = '0' * 64
        item(g, '13292.7')['role']['polarity'] = None
        code, out, after = self.repin(g)
        self.assertEqual(code, 1, out)
        self.assertEqual(after['inputs']['links.json'], '0' * 64)


if __name__ == '__main__':
    unittest.main()
