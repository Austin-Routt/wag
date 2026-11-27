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
with open(os.path.join(DOC, 'items.json'), encoding='utf-8') as fh:
    ITEMS = json.load(fh)
ERROR = re.compile(r'(\d+\.\d+|case\.json|inputs): ')
# The walton.yml the format-1 gold was made with (WAG 96c3c5a), and the copy
# of it that validate.py reads for a format-1 file (format1/walton.yml; D092:
# checked as it was made; proposal D008-D074).
WALTON_OF_FORMAT1 = 'bb1edbcc69536bad04dc33ce8e8c3578d3a610e9abb6fadf1ea42d67d1badf42'
with open(os.path.join(DOC, 'format1', 'walton.yml'), 'rb') as fh:
    WALTON_FORMAT1_COPY = hashlib.sha256(fh.read()).hexdigest()
# The game's NASCQ files the format-1 gold was made with (WAG 96c3c5a), which
# validate.py reads for a format-1 file from the frozen data of 2026-10-15 to
# step C5.1's switch (step C5.2 asks questions where they asked none).
REPO = os.path.normpath(os.path.join(DOC, '..', '..'))
NASCQ_OF_FORMAT1 = {'schemes.json': '8616f56cde01fbb4372adb745410f7766cfda756889a4ce422d5979a8cb398b5',
                    'critical_questions.json': 'cbc86ff8be867a5ae21376da4403e566306e9c294e0fb1255ecb133720ab5227',
                    'edges.json': '6f367c9e1efbc3ed16da385f206e4867af5ae4b4e51d1cd5de5c2f23a1c416c8'}


def run(gold, case=None, pin=True, items=None):
    """validate.py's exit status, its error lines and its whole output. With
    a case, it is written beside the gold and named with --case, and the
    gold's pin follows it unless pin is false; so with items (--items)."""
    with tempfile.TemporaryDirectory() as d:
        path = os.path.join(d, 'gold.json')
        cmd = [sys.executable, '-I', '-B', os.path.join(DOC, 'validate.py'), path]
        for name, value, flag in (('case.json', case, '--case'), ('items.json', items, '--items')):
            if value is None:
                continue
            cpath = os.path.join(d, name)
            text = json.dumps(value, ensure_ascii=False, indent=1)
            with open(cpath, 'w', encoding='utf-8') as fh:
                fh.write(text)
            if pin and 'inputs' in gold:
                gold = dict(gold, inputs=dict(gold['inputs'], **{name: hashlib.sha256(text.encode('utf-8')).hexdigest()}))
            cmd += [flag, cpath]
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
    # A fragment's null conclusion is no formula: a record with no formula is not the instance's.
    mine = ({p['formula'] for p in ins['premises']} | {ins['conclusion']['formula']}) - {None}
    ins['bindings'] = {k: (new if v == old else v) for k, v in ins['bindings'].items()}
    for p in ins['premises']:
        p['formula'] = sub(p['formula'])
    if ins['conclusion']['formula'] is not None:  # a fragment's unstated conclusion stays null
        ins['conclusion']['formula'] = sub(ins['conclusion']['formula'])
    for c in it['challenges'] if records else []:
        # A format-2 record names its instance: only the renamed one's follow.
        if c.get('formula') in mine and c.get('instance', ins.get('id')) == ins.get('id'):
            c['formula'] = sub(c['formula'])


def renumber(g, old, new):
    """Give an instance another id, and every record that names it."""
    for it in g['items']:
        for x in it['instances']:
            if x['id'] == old:
                x['id'] = new
        for c in it['challenges']:
            if c.get('instance') == old:
                c['instance'] = new
        t = it['target']
        for o in [t.get('actual') or {}] + (t.get('other_kialo_parents') or []):
            if o.get('instance') == old:
                o['instance'] = new


def as_annotator(g, name='model-x'):
    """The gold as an annotator's record: no adjudication, and each claim's
    primary instance first, as <claim>#1, as an annotator numbers it (R12); a
    gold keeps the ids its history gave them (a recoded claim's bridge is #2
    or #4), so the primary and #1 swap ids, and every record that names them
    follows."""
    g['annotator'] = name
    for it in g['items']:
        it.pop('adjudication', None)
        first = it['kialo_claim_id'] + '#1'
        pid = next(x['id'] for x in it['instances'] if x['primary'])
        if pid != first:
            tmp = it['kialo_claim_id'] + '#0'
            renumber(g, pid, tmp)
            renumber(g, first, pid)
            renumber(g, tmp, first)
        it['instances'].sort(key=lambda x: not x['primary'])
    return g


def next_id(it):
    used = [int(x['id'].split('#')[1]) for x in it['instances']] + [int(x.split('#')[1]) for x in it.get('retired', [])]
    return f"{it['kialo_claim_id']}#{max(used) + 1}"


def plant_two_person(g, goal):
    """Plant in 13292.108 a secondary instance of two-person practical
    reasoning (governments, who intend to realize the goal, told by
    nutrition science that the tax is the way), and retarget 13292.104, a
    better means, at its better-means exception, 23/CQ2 (D093(b))."""
    it = item(g, '13292.108')
    iid = next_id(it)
    A, B, X, Y = 'governments', 'nutrition_science', goal, 'tax_meat'
    p = lambda role, f: {'role': role, 'formula': f, 'en': f, 'status': 'supplied'}
    it['instances'].append({
        'id': iid, 'primary': False,
        'scheme': {'walton_id': 'two_person_practical', 'compendium': '23', 'wag_id': None, 'fit': 'reconstructed',
                   'confidence': 0.4, 'alternatives': []},
        'bindings': {'A': A, 'B': B, 'X': X, 'Y': Y},
        'premises': [p('premise', f'tells_intend({A},{X},{B})'), p('premise', f'situation_necessity({B},{Y},{X},{A})'),
                     p('assumption', f'possible_bring_about({A},{Y})'),
                     p('exception', f'realize_more_effectively({A},{X},{Y})'), p('exception', f'incompatible_goal({A},{X})'),
                     p('exception', f'bad_outweighs({Y},{X})')],
        'conclusion': {'formula': f'should_do({A},{Y})', 'en': f'should_do({A},{Y})'},
    })
    ans = item(g, '13292.104')
    ans['target']['actual'] = {'claim': '13292.108', 'instance': iid, 'part': 'exception',
                               'formula': f'realize_more_effectively({A},{X},{Y})'}
    ans['role']['cq'] = '23 CQ2'
    return g


def plant_bridge(g, goal):
    """Plant in 13292.108 a secondary instance of the bridge from continuing
    (governments ought to continue trying to realize the goal, and the tax
    would continue the effort), and retarget 13292.104, a better means, at its
    borrowed better-means exception: an undercutter, with no cq, since no
    question of the book or the game targets it (guidelines §3.1, OC6)."""
    it = item(g, '13292.108')
    iid = next_id(it)
    P_, G, A = 'governments', goal, 'tax_meat'
    p = lambda role, f: {'role': role, 'formula': f, 'en': f, 'status': 'supplied'}
    it['instances'].append({
        'id': iid, 'primary': False,
        'scheme': {'walton_id': 'bridge_from_continuing', 'compendium': None, 'wag_id': None, 'fit': 'reconstructed',
                   'confidence': 0.4, 'alternatives': []},
        'bindings': {'P': P_, 'G': G, 'A': A},
        'premises': [p('premise', f'ought_to_continue({P_},{G})'), p('premise', f'continues_trying({A},{G})'),
                     p('premise', f'deliberating_agent({P_},{A})'), p('assumption', f'possible_bring_about({P_},{A})'),
                     p('exception', f'realize_more_effectively({P_},{G},{A})'), p('exception', f'bad_outweighs({A},{G})')],
        'conclusion': {'formula': f'should_be_performed({A})', 'en': f'should_be_performed({A})'},
    })
    ans = item(g, '13292.104')
    ans['target']['actual'] = {'claim': '13292.108', 'instance': iid, 'part': 'exception',
                               'formula': f'realize_more_effectively({P_},{G},{A})'}
    ans['role']['kind'], ans['role']['cq'] = 'undercutter', None
    return g


def conclusion_target(g, claim, iid=None):
    """A format-2 target at an instance's conclusion (the claim's primary
    instance unless another is named)."""
    ins = inst(g, iid) if iid else primary(item(g, claim))
    return {'claim': claim, 'instance': ins['id'], 'part': 'conclusion', 'formula': ins['conclusion']['formula']}


class Rules(unittest.TestCase):
    def fails(self, g, cid, rules, text, case=None, pin=True, items=None):
        """validate.py fails, on cid only, by the rules given, one error
        containing text."""
        code, errors, out = run(g, case, pin, items)
        self.assertEqual(code, 1, out)
        for e in errors:
            self.assertTrue(any(e.startswith(f'{cid}: {r} ') for r in rules), f'unexpected error: {e}\n{out}')
        self.assertTrue(any(text in e for e in errors), out)

    def passes(self, g, case=None, items=None):
        code, errors, out = run(g, case, items=items)
        self.assertEqual((code, errors), (0, []), out)
        return out

    def test_the_gold_passes(self):
        out = self.passes(GOLD)
        self.assertIn('format 2', out)
        self.assertNotIn('UNRESOLVED', out)

    def test_the_format1_copy_is_the_walton_it_was_made_with(self):
        self.assertEqual(WALTON_FORMAT1_COPY, WALTON_OF_FORMAT1,
                         'format1/walton.yml is not the walton.yml the format-1 gold was made with')

    def test_the_format1_gold_passes_as_format1(self):
        # The format-1 gold is kept as it was (D039) and checked as it was
        # made (D092): against the copy of the walton.yml it was made with.
        out = self.passes(GOLD1)
        self.assertIn('format 1', out)

    def test_format1_named(self):
        g = copy.deepcopy(GOLD1)
        g['format'] = 'wag-annotation/1'
        code, errors, out = run(g)
        self.assertIn('format 1', out)
        self.assertEqual((code, errors), (0, []), out)

    def test_the_format1_nascq_is_the_one_it_was_made_with(self):
        # A kept version is checked against the game's questions it was made
        # with: the current ones ask CQ:bad_outweighs and CQ:good_outweighs
        # where the format-1 gold names the book's questions (13292.127,
        # 13292.43), so the format-1 gold passing shows which are read.
        for name, want in NASCQ_OF_FORMAT1.items():
            with open(os.path.join(REPO, 'wag', 'data', 'frozen', 'e7141ec8dea379bc', 'nascq', name), 'rb') as fh:
                self.assertEqual(hashlib.sha256(fh.read()).hexdigest(), want, name)

    # R16 (since step C5.2): an instance names the game's scheme for its
    # walton.yml scheme (wag_id), or null when the game plays none.
    def test_r16_wag_id_of_a_scheme_the_game_plays(self):
        for wag in (None, 'AS:bridge_possible_good'):
            g = copy.deepcopy(GOLD)
            inst(g, '13292.9#4')['scheme']['wag_id'] = wag
            self.fails(g, '13292.9', ['R16'], f"13292.9#4 names the game's scheme {wag!r}, but the game plays "
                       'bridge_from_possible_harm as AS:bridge_possible_harm')

    def test_r16_wag_id_of_a_scheme_the_game_does_not_play(self):
        g = copy.deepcopy(GOLD)
        inst(g, '13292.9#2')['scheme']['wag_id'] = 'AS:consequences'
        self.fails(g, '13292.9', ['R16'], 'but the game plays cause_to_effect_modal with none (null)')

    # R2 for a challenge (since step C5.2): where the game asks questions at
    # its formula, it names one of them; elsewhere, one the game asks of the
    # scheme, or none when the game plays no scheme for it.
    def test_r2_challenge_another_question_than_the_game_asks_there(self):
        g = copy.deepcopy(GOLD)
        c = next(c for c in item(g, '13292.108')['challenges'] if c['wag_cq'] == 'CQ:bad_outweighs')
        c['wag_cq'] = 'CQ:side_effect'
        self.fails(g, '13292.108', ['R2'], "names the game's CQ:side_effect, but the game asks ['CQ:bad_outweighs'] there")

    def test_r2_challenge_no_question_where_the_game_asks_one(self):
        g = copy.deepcopy(GOLD)
        c = next(c for c in item(g, '13292.9')['challenges'] if c['wag_cq'] == 'CQ:modal_evidence')
        c['wag_cq'] = None
        self.fails(g, '13292.9', ['R2'], "names no question of the game, but the game asks ['CQ:modal_evidence', "
                   "'CQ:modal_likelihood'] there")

    def test_r2_challenge_at_a_conclusion(self):
        g = copy.deepcopy(GOLD)
        c = next(c for c in item(g, '13292.7')['challenges'] if c['wag_cq'] == 'CQ:opposite_consequences')
        c['wag_cq'] = 'CQ:likelihood'
        self.fails(g, '13292.7', ['R2'], "names the game's CQ:likelihood, but the game asks ['CQ:opposite_consequences'] there")
        c['wag_cq'] = None
        self.fails(g, '13292.7', ['R2'], "names no question of the game, but the game asks ['CQ:opposite_consequences'] there")

    def test_r2_challenge_a_question_the_game_does_not_ask_of_the_scheme(self):
        # The game's analogy asks at formulas of its own (relevant_differences,
        # more_on_point), so at walton.yml's it cannot place the question, but
        # it is one the game asks of the analogy, as the gold's are.
        g = copy.deepcopy(GOLD)
        c = next(c for c in item(g, '13292.61')['challenges'] if c['wag_cq'] == 'CQ:difference')
        c['wag_cq'] = 'CQ:likelihood'
        self.fails(g, '13292.61', ['R2'], "names the game's CQ:likelihood, which the game does not ask of AS:analogy")

    def test_r2_challenge_a_question_the_game_asks_at_another_formula(self):
        # Where the game asks nothing at a challenge's formula, the question
        # it names may not be one the game asks at another formula of the
        # scheme, at its conclusion, or of the whole argument (the
        # independent check of the switch, W-1).
        for cid, iid, formula, cq, where in (
                ('13292.7', '13292.7#1', 'bad(food_less_affordable_for_poor)', 'CQ:likelihood', 'at brings_about(...)'),
                ('13292.9', '13292.9#4', 'matters_to_decision(tax_meat,tax_revenue_reduced)', 'CQ:worth_finding_out_first',
                 'at worth_finding_out_first(...)'),
                ('13292.9', '13292.9#4', '¬should_be_performed(tax_meat)', 'CQ:worth_finding_out_first',
                 'at worth_finding_out_first(...)'),
                ('13292.108', '13292.108#1', 'goal_of_agent(governments,environmental_harm_from_meat_reduced)', 'CQ:other_means',
                 'at realize_more_effectively(...)'),
                ('13292.115', '13292.115#1', 'good(reduced_unhealthy_meat_consumption)', 'CQ:side_effect', 'at its conclusion')):
            g = copy.deepcopy(GOLD)
            item(g, cid)['challenges'].append({'cq': None, 'wag_cq': cq, 'type': 'planted', 'instance': iid, 'formula': formula})
            self.fails(g, cid, ['R2'], f"names the game's {cq}, which the game asks {where}")

    def test_r2_challenge_a_question_of_the_whole_argument(self):
        # Sign's CQ:other_event is asked of the argument: a challenge about
        # the whole instance (formula null) names it, one at a formula does
        # not. The sign instance is planted (the gold has none); its
        # exception more_reliable takes no argument, a formula, not an atom.
        def plant(formula):
            g = copy.deepcopy(GOLD)
            it = item(g, '13292.9')
            iid = next_id(it)
            b = {'A': 'meat_industry_harmed', 'B': 'tax_revenue_reduced'}
            prem = [('premise', 'finding_true(meat_industry_harmed)'), ('premise', 'indicated_true(tax_revenue_reduced,meat_industry_harmed)'),
                    ('assumption', 'strong_correlation(meat_industry_harmed,tax_revenue_reduced)'), ('exception', 'more_reliable')]
            it['instances'].append({'id': iid, 'primary': False, 'bindings': b,
                                    'scheme': {'walton_id': 'sign', 'compendium': None, 'wag_id': 'AS:sign', 'fit': 'reconstructed',
                                               'confidence': 0.4, 'alternatives': []},
                                    'premises': [{'role': r, 'formula': f, 'en': f, 'status': 'supplied'} for r, f in prem],
                                    'conclusion': {'formula': 'true_in_situation(tax_revenue_reduced)', 'en': 'planted'}})
            it['challenges'].append({'cq': None, 'wag_cq': 'CQ:other_event', 'type': 'other cause', 'instance': iid, 'formula': formula})
            return g
        self.passes(plant(None))
        self.fails(plant('finding_true(meat_industry_harmed)'), '13292.9', ['R2'],
                   "names the game's CQ:other_event, which the game asks of the whole argument")

    def test_r2_challenge_of_an_answer_scheme(self):
        # The game plays the answer schemes but asks them nothing.
        g = copy.deepcopy(GOLD)
        item(g, '13292.127')['challenges'].append({'cq': None, 'wag_cq': 'CQ:bad_outweighs', 'type': 'planted',
                                                   'instance': '13292.127#1',
                                                   'formula': 'bad(starvation_from_reduced_meat_access)'})
        self.fails(g, '13292.127', ['R2'], 'the game asks no question of AS:bad_outweighs_good (wag_cq null)')

    def test_r2_challenge_of_a_scheme_the_game_does_not_play(self):
        g = copy.deepcopy(GOLD)
        f = inst(g, '13292.9#3')['premises'][2]['formula']
        item(g, '13292.9')['challenges'].append({'cq': '28 CQ3', 'wag_cq': 'CQ:likelihood', 'type': 'causal interference',
                                                 'instance': '13292.9#3', 'formula': f})
        self.fails(g, '13292.9', ['R2'], 'but the game plays cause_to_effect_modal with no scheme')

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

    def test_r9_the_continuing_bridge_states_its_goal_in_ought_to_continue(self):
        # A better means for the bridge's goal, at its borrowed exception
        # (D093(f)): ought_to_continue(P,G) is its goal premise.
        g = plant_bridge(copy.deepcopy(GOLD), 'environmental_harm_from_meat_reduced')
        self.passes(g)

    def test_r9_the_continuing_bridge_another_goal(self):
        g = plant_bridge(copy.deepcopy(GOLD), 'meat_problems_resolved')
        self.fails(g, '13292.104', ['R8', 'R9'], "is not the answered argument's ['goal_of_agent(governments,meat_problems_resolved)']")

    def test_r9_two_person_practical_states_its_goal_in_tells_intend(self):
        # A better means for two-person practical reasoning's goal (23/CQ2):
        # tells_intend(A,X,B) is its goal premise, goal_of_agent(A,X) (D093(b)).
        g = plant_two_person(copy.deepcopy(GOLD), 'environmental_harm_from_meat_reduced')
        self.passes(g)

    def test_r9_two_person_practical_another_goal(self):
        # The same answer to an argument whose agent intends another goal: R8
        # (the answer concludes another exception) and R9 (another goal).
        g = plant_two_person(copy.deepcopy(GOLD), 'meat_problems_resolved')
        self.fails(g, '13292.104', ['R8', 'R9'], "is not the answered argument's ['goal_of_agent(governments,meat_problems_resolved)']")

    # R11: an answer scheme's primary instance is at an exception.
    def test_r11_an_answer_scheme_away_from_the_exception(self):
        for cid in ('13292.127', '13292.43', '13292.123'):
            with self.subTest(claim=cid):
                g = copy.deepcopy(GOLD)
                it = item(g, cid)
                a = it['target']['actual']
                it['target']['actual'] = conclusion_target(g, a['claim'], a['instance'])
                self.fails(g, cid, ['R11'], 'is not at an exception')

    def test_r11_lists_every_scheme_that_concludes_an_exception(self):
        # R11's answer schemes are the schemes of walton.yml that conclude
        # another scheme's exception, and only those: a new answer scheme
        # (proposal D095's finding_out_first) is not left out of R11.
        import yaml
        src = open(os.path.join(DOC, 'validate.py'), encoding='utf-8').read()
        listed = set(re.findall(r"'(\w+)'", re.search(r'ANSWER_SCHEMES = \(([^)]*)\)', src).group(1)))
        with open(os.path.join(DOC, '..', 'scheme-registry', 'walton-clean', 'walton.yml'), encoding='utf-8') as fh:
            schemes = yaml.safe_load(fh)['argument_schemes']
        pred = lambda f: f.lstrip('¬').split('(')[0]
        exceptions = {pred(e) for s in schemes for e in (s.get('exceptions') or [])}
        concluding = {s['id'] for s in schemes if any(pred(c) in exceptions for c in (s.get('conclusions') or []))}
        self.assertIn('finding_out_first', concluding)
        self.assertEqual(listed, concluding)

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
            frag['scheme'] = {'walton_id': 'better_means', 'compendium': None, 'wag_id': 'AS:better_means', 'fit': fit,
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
        # Its target instance (since D008, 13292.9's modal argument, not its
        # primary bridge) and another instance of the target claim.
        tinst = next(x for x in target['instances'] if x['id'] == it['target']['actual']['instance'])
        other = next(x for x in target['instances'] if x['id'] != tinst['id'])
        f = other['premises'][0]['formula']
        self.assertNotIn(f, [p['formula'] for p in tinst['premises']])
        it['challenges'].append({'cq': None, 'wag_cq': None, 'type': 'causal interference', 'instance': other['id'], 'formula': f})
        self.fails(g, '13292.43', ['R10'], f"names the instance '{other['id']}', neither one of its own nor its target's")
        it['challenges'][-1]['instance'] = tinst['id']
        self.fails(g, '13292.43', ['R10'], f"names {f}, which is not a formula of {tinst['id']}")

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
        p0['formula'], p0['status'], p0['cite'] = 'has_occurred(tax_meat)', 'inferred', '13292.1'
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
        ins['scheme']['fit'] = 'reconstructed'  # a supplied premise makes it a reconstruction (R18)
        self.passes(g)

    # R18 (guidelines §5 item 6): fit explicit when every premise-role formula
    # is explicit. From the reviewer's patch (D086 exchange 1).
    def test_r18_explicit_fit(self):
        g = copy.deepcopy(GOLD)
        x = inst(g, '13292.11#2')
        self.assertEqual(x['scheme']['fit'], 'explicit')
        pm = next(p for p in x['premises'] if p['role'] == 'premise')
        pm['status'] = 'supplied'
        self.fails(g, '13292.11', ['R18'], 'has fit explicit, but its premise')
        pm['status'], pm['cite'] = 'inferred', '13292.118'
        self.fails(g, '13292.11', ['R18'], 'has fit explicit, but its premise')
        # A reconstruction may supply its premises; an explicit fit, its
        # assumptions and exceptions (13292.11#2's interference is supplied).
        x['scheme']['fit'] = 'reconstructed'
        self.passes(g)
        self.assertTrue(any(p['role'] == 'exception' and p['status'] == 'supplied'
                            for p in inst(GOLD, '13292.11#2')['premises']))

    def test_r18_not_in_format1(self):
        # A format-1 file keeps the rules it was made with.
        g = copy.deepcopy(GOLD1)
        x = next(x for it in g['items'] for x in it['instances'] if x['scheme']['fit'] == 'explicit')
        for p in x['premises']:
            if p['role'] == 'premise':
                p['status'] = 'supplied'
        code, errors, out = run(g)
        self.assertFalse(any(' R18 ' in e for e in errors), out)

    # R19 (guidelines §2 item 3): an inferred premise cites the claim that says it.
    def test_r19_an_inferred_premise_cites_its_claim(self):
        g = copy.deepcopy(GOLD)
        pm = next(p for p in inst(g, '13292.118#1')['premises'] if p.get('cite') == '13292.11')
        self.assertEqual(pm['status'], 'inferred')
        for bad in (None, '13292.99999', '13292.118', 7):
            if bad is None:
                pm.pop('cite', None)
            else:
                pm['cite'] = bad
            self.fails(g, '13292.118', ['R19'], 'is inferred but cites no other claim')
        pm['status'] = 'supplied'
        pm.pop('cite')
        self.passes(g)  # a supplied premise cites nothing

    def test_r4_in_format1_is_the_n8_rule(self):
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

    def test_r12_an_annotators_primary_comes_first_as_1(self):
        # An annotator's record (no adjudication) lists its primary instance
        # first, as <claim>#1, so that a later claim can name it before it is
        # shown; a gold's ids stay as its history made them.
        a = as_annotator(copy.deepcopy(GOLD))
        self.passes(a)
        b = copy.deepcopy(a)
        insts = item(b, '13292.9')['instances']
        insts[0], insts[1] = insts[1], insts[0]
        self.fails(b, '13292.9', ['R12'], 'its primary instance is not listed first as 13292.9#1')
        b = copy.deepcopy(a)
        renumber(b, '13292.7#1', '13292.7#2')
        self.fails(b, '13292.7', ['R12'], 'its primary instance is not listed first as 13292.7#1')
        g = copy.deepcopy(GOLD)
        insts = item(g, '13292.9')['instances']
        insts[0], insts[1] = insts[1], insts[0]
        self.passes(g)

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
        # Another instance of the target claim, one that does not have the
        # target's formula (whichever instance the gold names now).
        g = copy.deepcopy(GOLD)
        a = item(g, '13292.12')['target']['actual']
        other = next(x['id'] for x in item(g, '13292.9')['instances']
                     if x['id'] != a['instance'] and a['formula'] not in [p['formula'] for p in x['premises']])
        a['instance'] = other
        self.fails(g, '13292.12', ['R13'], f"{a['formula']}, which is not a premise of {other}")

    def test_d008_a_possible_consequence_concludes_no_more_than_it_may(self):
        # D008: a modal argument concludes that the action may be unwise
        # because of G; the thesis is reached only through its bridge. A modal
        # instance that concludes the categorical negation is refused.
        g = copy.deepcopy(GOLD)
        ins = next(x for x in item(g, '13292.174')['instances'] if x['scheme']['walton_id'] == 'negative_consequences_modal')
        ins['conclusion']['formula'] = '¬should_be_performed(tax_meat)'
        self.fails(g, '13292.174', ['negative_consequences_modal'], 'negative_consequences_modal conclusion differs')
        # Nor can a possible consequence fill the categorical premise.
        g = copy.deepcopy(GOLD)
        ins = next(x for x in item(g, '13292.174')['instances'] if x['scheme']['walton_id'] == 'negative_consequences_modal')
        ins['scheme'].update(walton_id='negative_consequences', wag_id='AS:negative_consequences')
        # (Its challenges name the modal questions, which the categorical
        # scheme is not asked: R2 says so too.)
        self.fails(g, '13292.174', ['negative_consequences', 'R2'], 'negative_consequences premises differ from walton.yml')

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
        missing = next_id(item(g, '13292.121'))
        item(g, '13292.180')['target']['other_kialo_parents'][0]['instance'] = missing
        self.fails(g, '13292.180', ['R13'], f'names the instance {missing}, which is not an instance of 13292.121')
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

    def test_r13_an_unresolved_part_is_kept(self):
        # §8: a target's part may be unresolved; it is kept, not refused.
        g = copy.deepcopy(GOLD)
        item(g, '13292.12')['target']['actual']['part'] = {'unresolved': 'premise or exception'}
        out = self.passes(g)
        self.assertIn('13292.12: its target part (premise or exception)', out)

    def test_r13_another_kialo_parents_inference(self):
        # An other parent's exact target is checked whatever its part, the
        # inference included.
        g = copy.deepcopy(GOLD)
        o = item(g, '13292.180')['target']['other_kialo_parents'][0]
        o.update(part='inference', instance='13292.121#99', formula=None)
        self.fails(g, '13292.180', ['R13'], 'names the instance 13292.121#99, which is not an instance of 13292.121')
        o['instance'] = '13292.121#1'
        self.passes(g)

    def test_r4_a_modal_fragment_that_leaves_its_effect_unstated(self):
        # A modal step's unstated effect is not a cause R4 can name.
        g = copy.deepcopy(GOLD)
        it = item(g, '13292.11')
        it['instances'].append({'id': next_id(it), 'primary': False,
                                'scheme': {'walton_id': 'cause_to_effect_modal', 'compendium': '28', 'wag_id': None, 'fit': 'fragment',
                                           'confidence': 0.4, 'alternatives': []},
                                'bindings': {'E1': 'meat_subsidies', 'E2': {'unstated': 'the effect is not said'}},
                                'premises': [{'role': 'premise', 'formula': 'may_occur(meat_subsidies)', 'en': 'x', 'status': 'explicit'}],
                                'conclusion': {'formula': None, 'en': ''}})
        self.passes(g)

    # R14: actors and actions.
    def test_r14_an_agent_who_is_not_an_actor(self):
        # An item whose primary instance binds the agent and into whose
        # formulas naming the agent no other item points (a target or a
        # challenge record), so that the renamed agent is its error alone.
        # (13292.6 served until D086 round 1 made it a fragment that leaves
        # the agent unstated.)
        g = copy.deepcopy(GOLD)
        def pointed_into(cid):
            ids = {x['id'] for x in item(g, cid)['instances']}
            for o in g['items']:
                if o['kialo_claim_id'] == cid:
                    continue
                t = (o.get('target') or {}).get('actual') or {}
                if t.get('claim') == cid and 'governments' in str(t.get('formula')):
                    return True
                if any(c.get('instance') in ids and 'governments' in str(c.get('formula')) for c in o.get('challenges') or []):
                    return True
            return False
        cid = next(i['kialo_claim_id'] for i in g['items']
                   if primary(i)['bindings'].get('P') == 'governments' and primary(i)['scheme']['walton_id'] != 'better_means'
                   and not pointed_into(i['kialo_claim_id']))
        it = item(g, cid)
        rename(it, primary(it), 'governments', 'meat_tax_proponents')
        self.fails(g, cid, ['R14'], 'binds the agent P to meat_tax_proponents, which is not an actor of the case')

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

    def test_r7_another_parents_relation(self):
        # The source's pro or con of each Kialo parent, as items.json has it
        # (an external review of 2026-11-25): a reversed relation of a
        # secondary parent is refused, as the primary's always was.
        g = copy.deepcopy(GOLD)
        item(g, '13292.180')['target']['other_kialo_parents'][0]['relation'] = 'con'
        self.fails(g, '13292.180', ['R7'], 'R7 other_kialo_parents record 13292.121 con; items.json has 13292.121 pro')

    def test_r7_each_other_parent_in_any_order_as_often_as_items_has_it(self):
        # The other parents with their relations, as items.json has them, in
        # any order and each as often as it has it (the independent check of
        # ID-k, 2026-11-26: compared as a set, or unsorted, every earlier test
        # passed). Batch 2 has no claim with more than two parents, so the
        # order is tried on an items.json with a third.
        g = copy.deepcopy(GOLD)
        others = item(g, '13292.180')['target']['other_kialo_parents']
        others.append(copy.deepcopy(others[0]))
        self.fails(g, '13292.180', ['R7'], 'R7 other_kialo_parents record 13292.121 pro, 13292.121 pro; items.json has 13292.121 pro')
        items = copy.deepcopy(ITEMS)
        next(i for i in items['items'] if i['kialo_claim_id'] == '13292.180')['kialo_parents'].append(
            {'claim_id': '13292.115', 'relation': 'con', 'location_id': '13292.999'})
        g = copy.deepcopy(GOLD)
        item(g, '13292.180')['target']['other_kialo_parents'].insert(0, {'claim_id': '13292.115', 'relation': 'con'})
        self.passes(g, items=items)
        item(g, '13292.180')['target']['other_kialo_parents'][0]['relation'] = 'pro'
        self.fails(g, '13292.180', ['R7'], 'R7 other_kialo_parents record 13292.115 pro, 13292.121 pro; items.json has 13292.115 con, 13292.121 pro', items=items)

    def test_r14_an_actor_is_no_action_and_no_alias(self):
        # An actor's id names the actor: as an alias of an action, it would be
        # read as that action wherever it stands (an external review of
        # 2026-11-25).
        for name, plant, text in (
                ('an alias', lambda c: next(a for a in c['actions'] if a['id'] == 'tax_meat').update(aliases=['governments']),
                 "'governments' is an actor, so it cannot be an alias of tax_meat"),
                ('an action', lambda c: next(a for a in c['actions'] if a['id'] == 'raise_import_standards').update(id='governments'),
                 "'governments' is an actor, so it cannot be an action")):
            with self.subTest(name):
                case = copy.deepcopy(CASE)
                plant(case)
                code, errors, out = run(copy.deepcopy(GOLD), case)
                self.assertEqual(code, 1, out)
                self.assertTrue(any('case.json: R14 ' in e and text in e for e in errors), out)

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
            'no polarity key': ('13292.7', lambda g: item(g, '13292.7')['role'].pop('polarity'), 'role.polarity is missing'),
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
                # A malformed item is said once and read by no other item's rule.
                for e in errors:
                    self.assertTrue(e.startswith(f'{cid}: '), f'{name}: {e}')

    def test_r16_no_traceback(self):
        # What the schema pass does not read is read safely: a gold's
        # adjudication, a malformed thesis, a claim items.json does not have,
        # and items that are not objects.
        cases = [
            ('no adjudication', lambda g: item(g, '13292.7').pop('adjudication'), '13292.7: missing adjudication'),
            ('a null adjudication', lambda g: item(g, '13292.7').update(adjudication=None), '13292.7: bad adjudication'),
            ('no changed_fields', lambda g: item(g, '13292.7')['adjudication'].pop('changed_fields'), '13292.7: bad adjudication'),
            ('a string conclusion of the thesis', lambda g: primary(item(g, '13292.1')).update(conclusion='should_be_performed(tax_meat)'),
             '13292.1: R16 13292.1#1.conclusion is not an object'),
            ('a claim items.json does not have', lambda g: item(g, '13292.7').update(kialo_claim_id='13292.999'),
             '13292.999: not a claim of items.json'),
        ]
        for name, plant, text in cases:
            with self.subTest(name):
                g = copy.deepcopy(GOLD)
                plant(g)
                code, errors, out = run(g)
                self.assertEqual(code, 1, out)
                self.assertNotIn('Traceback', out)
                self.assertTrue(any(text in e for e in errors), f'{name}:\n{out}')
                if name.startswith('a string conclusion'):
                    self.assertTrue(all(e.startswith('13292.1: R16 ') for e in errors), out)
        g = copy.deepcopy(GOLD)
        g['items'][3] = 'not an item'
        code, errors, out = run(g)
        self.assertEqual(code, 1, out)
        self.assertNotIn('Traceback', out)
        self.assertIn('the file\'s items are not a list of objects', out)

    def test_r16_a_malformed_target_is_not_read(self):
        # The claims that challenge a malformed claim's instance are not
        # refused for it: its own errors are said once.
        g = copy.deepcopy(GOLD)
        item(g, '13292.108')['role']['confidence'] = None
        code, errors, out = run(g)
        self.assertEqual(errors, ['13292.108: R16 role.confidence is not a number from 0 to 1'], out)

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
        g = as_annotator(copy.deepcopy(GOLD), 'A')
        del g['inputs']
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

    # Item F (docs/c5-item-f.md, F): schemes the game does not play in C5.2.
    @staticmethod
    def plant_instance(it, scheme, bindings, premises, conclusion):
        iid = next_id(it)
        it['instances'].append({
            'id': iid, 'primary': False, 'bindings': bindings,
            'scheme': {'walton_id': scheme, 'compendium': None, 'wag_id': None, 'fit': 'reconstructed',
                       'confidence': 0.4, 'alternatives': []},
            'premises': [{'role': r, 'formula': f, 'en': f, 'status': 'supplied'} for r, f in premises],
            'conclusion': {'formula': conclusion, 'en': conclusion}})
        return iid

    def test_itemf_f_context_another_actors_action_in_a_bridge_from_agent(self):
        # F-context: a bridge carrying should_do(P,A) where the case record
        # gives A to another actor is refused (R14), whatever supports
        # should_do elsewhere.
        case = copy.deepcopy(CASE)
        case['actors'].append({'id': 'meat_tax_proponents', 'text': 'the proponents'})
        case['actions'].append({'id': 'switch_to_other_protein', 'actor': 'meat_tax_proponents', 'aliases': []})
        g = copy.deepcopy(GOLD)
        P_, A = 'governments', 'switch_to_other_protein'
        self.plant_instance(item(g, '13292.108'), 'bridge_from_agent', {'P': P_, 'A': A},
                            [('premise', f'should_do({P_},{A})'), ('premise', f'deliberating_agent({P_},{A})')],
                            f'should_be_performed({A})')
        self.fails(g, '13292.108', ['R14'], "which is meat_tax_proponents's action, not governments's", case=case)

    def test_itemf_g_unbound_end_and_means(self):
        # G-unbound: an instance of end_and_means_bad with X or Y unbound (or
        # blank) is refused, not read as a wildcard.
        for b in ({'Y': 'tax_revenue_reduced'}, {'X': '', 'Y': 'tax_revenue_reduced'},
                  {'X': 'meat_industry_harmed'}):
            g = copy.deepcopy(GOLD)
            X, Y = b.get('X', 'X'), b.get('Y', 'Y')
            self.plant_instance(item(g, '13292.108'), 'end_and_means_bad', b,
                                [('premise', f'means_to_end({X},{Y})'), ('premise', f'bad({Y})'),
                                 ('exception', f'means_good_outweighs({X},{Y})')], f'bad({X})')
            code, errors, out = run(g)
            self.assertEqual(code, 1, out)
            self.assertTrue(any(e.startswith('13292.108: ') for e in errors), out)
            if set(b) != {'X', 'Y'}:
                self.assertTrue(any(f"binds {sorted(b)} not ['X', 'Y']" in e for e in errors), out)


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
