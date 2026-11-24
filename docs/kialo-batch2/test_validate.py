"""Planted-defect tests of validate.py's rules R8 to R11 (decisions D075 and
D089; proposal D004-D075).

    python3 -I test_validate.py

Each test writes a copy of gold.json with one defect planted, runs
validate.py on it, and checks that the rule reports it and nothing else
does; the gold itself must pass.
"""
import copy
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


def run(gold):
    """validate.py's exit status, its error lines and its whole output."""
    with tempfile.TemporaryDirectory() as d:
        path = os.path.join(d, 'gold.json')
        with open(path, 'w', encoding='utf-8') as fh:
            json.dump(gold, fh, ensure_ascii=False)
        r = subprocess.run([sys.executable, '-I', '-B', os.path.join(DOC, 'validate.py'), path],
                           capture_output=True, text=True)
    errors = [line for line in r.stdout.splitlines() if re.match(r'\d+\.\d+: ', line)]
    return r.returncode, errors, r.stdout + r.stderr


def item(g, cid):
    return next(i for i in g['items'] if i['kialo_claim_id'] == cid)


def primary(it):
    return next(x for x in it['instances'] if x.get('primary'))


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
        if c.get('formula') in mine:
            c['formula'] = sub(c['formula'])


class Rules(unittest.TestCase):
    def fails(self, g, cid, rules, text):
        """validate.py fails, on cid only, by the rules given, one error
        containing text."""
        code, errors, out = run(g)
        self.assertEqual(code, 1, out)
        for e in errors:
            self.assertTrue(any(e.startswith(f'{cid}: {r} ') for r in rules), f'unexpected error: {e}\n{out}')
        self.assertTrue(any(text in e for e in errors), out)

    def passes(self, g):
        code, errors, out = run(g)
        self.assertEqual((code, errors), (0, []), out)

    def test_the_gold_passes(self):
        self.passes(GOLD)

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
            'primary': True, 'scheme': {'walton_id': 'other', 'fit': 'reconstructed'}, 'bindings': {},
            'premises': [{'role': 'premise', 'formula': 'brings_about(tax_meat,starvation_from_reduced_meat_access)',
                          'status': 'explicit'}],
            'conclusion': {'formula': '¬' + f}})
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
                it['target']['actual'] = {'claim': '13292.118', 'part': 'conclusion'}
                it['role'].update(kind=kind, cq=cq)
                ins = primary(it)
                rename(it, ins, ins['bindings']['G'], 'deforestation_slowed')
                self.fails(g, '13292.123', ['R9', 'R11'], "is not the answered argument's")

    def test_r9_a_secondary_better_means(self):
        # Only the answer itself must share the answered argument's goal.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.104')
        dup = copy.deepcopy(primary(it))
        dup['primary'] = False
        rename(it, dup, dup['bindings']['G'], 'deforestation_slowed', records=False)
        it['instances'].append(dup)
        self.passes(g)

    # R11: an answer scheme's primary instance is at an exception.
    def test_r11_an_answer_scheme_away_from_the_exception(self):
        for cid in ('13292.127', '13292.43', '13292.123'):
            with self.subTest(claim=cid):
                g = copy.deepcopy(GOLD)
                it = item(g, cid)
                it['target']['actual'] = {'claim': it['target']['actual']['claim'], 'part': 'conclusion'}
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
            frag['scheme'] = {'walton_id': 'better_means', 'fit': fit, 'confidence': 0.5}
            frag.pop('comparison', None)
            it['instances'] = [frag] + [dict(x, primary=False) for x in it['instances']]
            f = 'brings_about(regulate_animal_farming_industry,environmental_harm_from_meat_reduced)'
            it['target']['actual'] = ({'claim': '13292.104', 'part': 'premise', 'formula': f} if part == 'premise'
                                      else {'claim': '13292.104', 'part': 'conclusion'})
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
        it['target']['actual'] = {'claim': '13292.108', 'part': 'conclusion'}
        code, errors, out = run(g)
        self.assertFalse([e for e in errors if e.startswith('13292.127: R11')], out)

    # R10: a challenge record names a formula of the claim's own instances
    # or of its target's primary instance.
    def test_r10_a_formula_renamed_in_walton_yml(self):
        g = copy.deepcopy(GOLD)
        rec = next(c for c in item(g, '13292.108')['challenges'] if c.get('cq') == '22.1 CQ5')
        rec['formula'] = 'side_effects(tax_meat,meat_production_and_consumption_harm_environment)'
        self.fails(g, '13292.108', ['R10'], 'names side_effects(tax_meat,')

    def test_r10_a_formula_of_the_targets_other_instance(self):
        g = copy.deepcopy(GOLD)
        it, target = item(g, '13292.43'), item(g, '13292.9')
        names = {p['formula'] for x in it['instances'] for p in x['premises']}
        names |= {p['formula'] for p in primary(target)['premises']}
        other = [p['formula'] for x in target['instances'] if not x.get('primary') for p in x['premises']
                 if p['formula'] not in names]
        self.assertTrue(other, "13292.9's other instances have no formula of their own")
        it['challenges'].append({'cq': None, 'wag_cq': None, 'type': 'causal interference', 'formula': other[0]})
        self.fails(g, '13292.43', ['R10'], 'names ' + other[0])


if __name__ == '__main__':
    unittest.main()
