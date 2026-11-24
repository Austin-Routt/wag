"""Recode the batch-2 gold for proposal D004-D075, version 3 (stage 2, item C2).

    python3 -I recode_c2.py

A one-off migration, kept as the record of what changed and why. It applies
the gold effects of proposal D004-D075 (docs/proposals/) to gold.json and
lexicon.json here, and copies them to the annotation engine's copies
(wag/data/annotation/kialo-13292-batch2/). It refuses to run on a gold that
no longer has the old formulas. English renderings come from walton.yml's
templates (with the proposal applied) and lexicon.json, as the gold's were
made.

(1) The eight practical_reasoning2 instances bind P to governments, the
    thesis's agent, and take the goal premise goal_of_agent(P,G) (D005), the
    possibilities possible_bring_about(P,A) and possible_goal(G) (D088), the
    exceptions scoped to the agent (D004, D089) and bad_outweighs(A,G) for
    the side effect (D070, D091). Targets and challenge records that named
    the old formulas follow.
(2) 13292.45 on the repaired necessary-condition scheme (D030, D063, D090):
    G for its goal, S the barriers to cheaper meat imports, the agent in each
    premise, its new exceptions; A2 leaves it. Its challenge records name
    the new formulas, under their labels (22.2's, the scheme's own number, as
    the guidelines label 33.2's questions); 13292.194's eco-tariffs, another
    barrier, are a better means, not the objection of CQ1.
(3) 13292.174 and 13292.5 rank nothing against 13292.108's goal: rebuttals,
    negative consequences at its conclusion (D070; guidelines §3.1, OC2).
(4) 13292.123 and 13292.104 say that another course would do better: each
    gets a primary better_means instance at the exception it answers (D075,
    D089), with the comparison it makes; the other course's own instance is
    secondary. 13292.208 follows 13292.104's argument. 13292.147 offers
    another course without claiming it better: no answer (D006), a proposal
    of another course (D066).
(5) The weighing claims coded Other become instances of C003's pair:
    13292.43 good_outweighs_bad; 13292.127 bad_outweighs_good, against
    13292.108 by OC7's one-instance rule now that practical reasoning is in
    the class.
"""
import json
import os
import re
import shutil
import sys

import yaml

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
WALTON = REPO + 'docs/scheme-registry/walton-clean/walton.yml'
COPY = REPO + 'wag/data/annotation/kialo-13292-batch2/'
GOLD, LEX = DOC + 'gold.json', DOC + 'lexicon.json'
gold = json.load(open(GOLD, encoding='utf-8'))
lex = json.load(open(LEX, encoding='utf-8'))
W = yaml.safe_load(open(WALTON, encoding='utf-8'))
LANG = W['language']
SCHEMES = {s['id']: s for s in W['argument_schemes']}
I = {i['kialo_claim_id']: i for i in gold['items']}
WHEN = '2026-11-24'
BY = 'proposal D004-D075, version 3 (recode_c2.py)'
AG = 'governments'

for k in ('goal_of_agent/2', 'realize_more_effectively/3', 'more_effective_means/4', 'goal_achievable_without_any/3',
          'matters_more/2'):
    if k not in LANG:
        raise SystemExit(f'{WALTON} has no template {k}: apply the proposal to walton.yml first')


def nows(f):
    return re.sub(r'\s+', '', f or '')


def parse(f):
    m = re.fullmatch(r'(¬?)([a-z_0-9]+)\((.*)\)', nows(f))
    if not m:
        raise SystemExit(f'cannot parse {f!r}')
    return m.group(1), m.group(2), m.group(3).split(',')


def en(f):
    neg, pred, args = parse(f)
    if neg:
        raise SystemExit(f'negated formula {f!r}: not expected here')
    return LANG[f'{pred}/{len(args)}'] % tuple("'" + lex[a] + "'" for a in args)


def subst(f, b):
    return re.sub(r'\b([A-Z][A-Z0-9]*)\b', lambda m: b.get(m.group(1), m.group(1)), f)


def since(note, text):
    """An earlier note, with what changed on WHEN after it."""
    return ((note or '').rstrip('.') + '. ' if note else '') + f'Since {WHEN}: {text}'


def record(cid, fields, what):
    I[cid]['adjudication'].setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': fields, 'what': what})


def primary(it):
    p = [x for x in it['instances'] if x.get('primary')]
    if len(p) != 1:
        raise SystemExit(f"{it['kialo_claim_id']}: {len(p)} primary instances")
    return p[0]


def rebuild(ins, wid, notes, renamed):
    """Rebuild an instance's premises and conclusion from walton.yml. A formula
    keeps the status, cite and note it had, under its old spelling if renamed
    maps the new formula to the old one; notes adds to (or gives) them."""
    sch = SCHEMES[wid]
    old = {nows(p['formula']): p for p in ins['premises']}
    out = []
    for role, key in (('premise', 'premises'), ('assumption', 'assumptions'), ('exception', 'exceptions')):
        for f in sch.get(key) or []:
            g = nows(subst(f, ins['bindings']))
            p = {'role': role, 'formula': g, 'en': en(g)}
            src = old.get(renamed.get(g, g))
            if src is not None:
                for k in ('status', 'cite', 'note'):
                    if k in src:
                        p[k] = src[k]
            p.update(notes.get(g, {}))
            if 'status' not in p:
                raise SystemExit(f'{wid}: no status for {g}')
            out.append(p)
    ins['premises'] = out
    c = nows(subst(sch['conclusions'][0], ins['bindings']))
    ins['conclusion'] = {'formula': c, 'en': en(c)}


# Guards: the gold is the one this recode was written for.
PR2 = [(it['kialo_claim_id'], ins) for it in gold['items'] for ins in it['instances']
       if ins['scheme']['walton_id'] == 'practical_reasoning2']
if len(PR2) != 8 or any('P' in ins['bindings'] for _, ins in PR2):
    raise SystemExit(f'expected 8 practical_reasoning2 instances without P, found {len(PR2)}: already recoded?')
for cid in ('13292.127', '13292.43'):
    if primary(I[cid])['scheme']['walton_id'] != 'other':
        raise SystemExit(f'{cid}: not the Other weighing it was')

REMAP = {}  # claim -> {old formula: new formula} of its primary instance
OWN = {}  # claim -> [{old formula: new formula}] of each of its practical_reasoning2 instances

# (1) practical_reasoning2.
for cid, ins in PR2:
    b = ins['bindings']
    b['P'] = AG
    a, g, s1, s2 = b['A'], b['G'], b['S1'], b['S2']
    renamed = {f'possible_bring_about({AG},{a})': f'possible({a})', f'possible_goal({g})': f'possible({g})',
               f'bring_about_more_effectively({AG},{s2},{a})': f'bring_about_more_effectively({s2},{a})',
               f'realize_more_effectively({AG},{g},{a})': f'realize_more_effectively({g},{a})',
               f'bad_outweighs({a},{g})': f'side_effects({a},{s1})',
               f'incompatible_goal({AG},{g})': f'incompatible_goal({g})'}
    old = {nows(p['formula']): p.get('note') for p in ins['premises']}
    notes = {
        f'goal_of_agent({AG},{g})': {'status': 'supplied', 'note': (
            f"the agent's goal (D005, since {WHEN}); the agent is the thesis's, governments, who would take the "
            "action")},
        f'possible_bring_about({AG},{a})': {'note': since(old.get(f'possible({a})'), "that governments can do it (D088)")},
        f'possible_goal({g})': {'note': since(old.get(f'possible({g})'), 'that the goal can be reached (D088)')},
        f'bring_about_more_effectively({AG},{s2},{a})': {'note': since(old.get(f'bring_about_more_effectively({s2},{a})'),
                                                                     'scoped to the agent (D089)')},
        f'realize_more_effectively({AG},{g},{a})': {'note': since(old.get(f'realize_more_effectively({g},{a})'),
                                                                 'scoped to the agent (D089)')},
        f'bad_outweighs({a},{g})': {'note': since(old.get(f'side_effects({a},{s1})'), (
            "the claim that the action's bad consequences outweigh the goal (D070, D091; D055's formula); a plain "
            "side effect is a rebuttal at the conclusion"))},
        f'incompatible_goal({AG},{g})': {'note': since(old.get(f'incompatible_goal({g})'), 'scoped to the agent (D004)')},
    }
    rebuild(ins, 'practical_reasoning2', notes, renamed)
    OWN.setdefault(cid, []).append({nows(o): nows(n) for n, o in renamed.items()})
    if ins.get('primary'):
        REMAP[cid] = OWN[cid][-1]
    record(cid, ['instances.bindings', 'instances.premises', 'challenges'],
           'practical_reasoning2 binds P to governments, with goal_of_agent(P,G), possible_bring_about(P,A), '
           'possible_goal(G), the exceptions scoped to the agent and bad_outweighs(A,G) (D004, D005, D070, D088, D089, '
           'D091)')

# (2) 13292.45 on the repaired necessary-condition scheme.
it45 = I['13292.45']
ins = primary(it45)
if ins['scheme']['walton_id'] != 'practical_necessary_condition' or 'A2' not in ins['bindings']:
    raise SystemExit('13292.45 already recoded')
x, goal, a2 = ins['bindings']['X'], ins['bindings']['A'], ins['bindings']['A2']
if ins['bindings']['P'] != AG:
    raise SystemExit('13292.45: P is not governments')
S = 'barriers_to_cheaper_meat_imports'
lex.setdefault(S, 'the barriers to cheaper meat imports (higher import standards, quotas, tariffs)')
ins['bindings'] = {'X': x, 'G': goal, 'P': AG, 'S': S}
renamed45 = {
    f'goal_of_agent({AG},{goal})': f'goal_is_action({goal})',
    f'alternative_action({AG},{S},{goal})': f'alternative_action({x},{goal})',
    f'necessary_action({AG},{x},{S},{goal})': f'necessary_action({x},{goal})',
    f'practicality_action({AG},{x})': f'practicality_action({x})',
    f'more_acceptable_bringing({AG},{goal},{x})': f'more_acceptable_bringing({goal},{x})',
    f'possible_goal({goal})': f'possible_action({goal})',
    f'acceptable_action({AG},{x})': f'acceptable_action({x})',
    f'possible_bring_about({AG},{x})': f'possible_agent({x},{AG})',
    f'realize_more_effectively({AG},{goal},{x})': f'alternative_means({goal},{x})',
    f'bad_outweighs({x},{goal})': f'side_effect_action({AG},{x})',
    f'incompatible_goal({AG},{goal})': f'conflict_goals({AG},{a2},{goal})',
}
EXC = 'exception: presumed not to hold unless shown'
rebuild(ins, 'practical_necessary_condition', {
    f'alternative_action({AG},{S},{goal})': {'note': 'the set of means is not named; S stands for the barriers to '
                                                     'cheaper imports, of which higher standards are one (D030)'},
    f'realize_more_effectively({AG},{goal},{x})': {'status': 'supplied', 'note': since(EXC, (
        'a strictly better means defeats the recommendation, an equally good one does not (D090, D006)'))},
    f'goal_achievable_without_any({AG},{goal},{S})': {'status': 'supplied', 'note': since(EXC, (
        "that governments could protect domestic producers without any barrier to imports, the necessary-condition "
        "scheme's objection (D090)"))},
    f'bad_outweighs({x},{goal})': {'status': 'supplied', 'note': since(EXC, (
        "the claim that the measure's bad consequences outweigh the goal (D091)"))},
    f'incompatible_goal({AG},{goal})': {'status': 'supplied', 'note': since(EXC, (
        "scoped to the agent and the goal (D004); the plausible conflicting goal the gold bound before, cheap and "
        "open trade in meat, is what an answer would name"))},
}, renamed45)
it45['challenges'] = [
    {'cq': '22.2 CQ1', 'wag_cq': None, 'type': 'other means',
     'formula': f'goal_achievable_without_any({AG},{goal},{S})',
     'note': ("a route that uses no barrier to imports (D090); 13292.194's eco-tariffs are another barrier, a member "
              "of S, not such a route")},
    {'cq': None, 'wag_cq': None, 'type': 'better means', 'formula': f'realize_more_effectively({AG},{goal},{x})',
     'note': "walton/realize_more_effectively (D090): another barrier, such as 13292.194's eco-tariffs, claimed to "
             "do better"},
    {'cq': '22.2 CQ3', 'wag_cq': None, 'type': 'efficacy', 'formula': f'possible_bring_about({AG},{x})',
     'note': "'manageable' is asserted, not shown"},
    {'cq': '22.2 CQ4', 'wag_cq': None, 'type': 'side effects that outweigh', 'formula': f'bad_outweighs({x},{goal})',
     'note': 'trade frictions and dearer imports: a rebuttal at the conclusion unless claimed to outweigh protecting the '
             'producers (D070, D091)'},
    {'cq': '22.2 CQ5', 'wag_cq': None, 'type': 'severity or value', 'formula': f'incompatible_goal({AG},{goal})',
     'note': 'cheap and open trade in meat'},
]
record('13292.45', ['instances.bindings', 'instances.premises', 'instances.conclusion', 'challenges'],
       'the necessary-condition scheme repaired (D030, D063, D088, D090, D091): G, S, the agent in each premise, '
       'realize_more_effectively and goal_achievable_without_any, bad_outweighs(X,G), incompatible_goal(P,G), '
       'should_be_performed(X); A2 leaves it; the challenge records name the new formulas, and 13292.194\'s '
       'eco-tariffs are a better means (walton/realize_more_effectively), not CQ1\'s objection')

# (3) the two side effects become rebuttals at 13292.108's conclusion.
for cid in ('13292.174', '13292.5'):
    it = I[cid]
    if it['role']['kind'] != 'cq_answer' or it['role'].get('cq') != 'CQ:side_effect':
        raise SystemExit(f'{cid}: not the side-effect answer it was')
    it['role'] = {'kind': 'rebuttal', 'confidence': it['role']['confidence'], 'cq': None, 'polarity': 'against',
                  'move': 'side_effect'}
    it['target']['actual'] = {'claim': '13292.108', 'part': 'conclusion'}
    it['notes']['text'] += (f' Since {WHEN} (D070) a side effect that ranks nothing against the goal is negative '
                            'consequences, a rebuttal at the conclusion of the argument it answers (guidelines §3.1, '
                            "OC2), not an answer at the exception, which is now the claim that the tax's bad "
                            'consequences outweigh the goal (bad_outweighs(tax_meat,environmental_harm_from_meat_reduced)).')
    record(cid, ['role', 'target', 'notes'], "a plain side effect: a rebuttal at 13292.108's conclusion, no longer "
           "an answer at its side-effect exception (D070)")

# (4) better_means for the claims that another course would do better.
BETTER = {
    '13292.123': dict(
        a2='abolish_meat_subsidies', comparison='more surely', comp_status='explicit',
        comp_note=('"instead": abolishing the subsidies removes at its source the distortion the tax would only '
                   'offset'),
        brings_status='inferred', brings_note='"distortions caused by subsidies should be removed by abolishing those subsidies"',
        challenges=[
            ('efficacy', 'possible', '13292.171: removing subsidies is politically easier than taxing (it backs the premise)'),
            ('efficacy', 'brings', 'removing subsidies does not price in externalities (13292.218)'),
            ('better means', 'more', 'the comparison: more surely; a reply that the tax gets there sooner attacks it')]),
    '13292.104': dict(
        a2='regulate_animal_farming_industry', comparison='more surely', comp_status='explicit',
        comp_note=('"what is problematic is the animal farming industry": regulating the industry addresses what is '
                   'said to be the problem, so it reaches the goal more surely; "rather than penalize those who wish '
                   'to consume meat" prefers one value to another, which is not this comparison (D089)'),
        brings_status='inferred', brings_note='the industry, not meat, is said to be the problem',
        challenges=[
            ('efficacy', 'possible', 'feasibility of regulating the industry'),
            ('efficacy', 'brings', ('attacked by 13292.208 (non-cultured meat production is inherently problematic: '
                                    'regulation changes the harm only in degree); can regulation remove harms that '
                                    'scale with consumption?')),
            ('better means', 'more', 'the comparison: more surely')]),
}
for cid, spec in BETTER.items():
    it = I[cid]
    f = nows(it['target']['actual'].get('formula'))
    m = re.fullmatch(r'realize_more_effectively\(([^,]+),([^)]+)\)', f)
    if it['role'].get('cq') != 'CQ:other_means' or not m or any(x['scheme']['walton_id'] == 'better_means'
                                                                 for x in it['instances']):
        raise SystemExit(f'{cid}: not the other-means answer it was, or already recoded')
    g, a = m.group(1), m.group(2)
    tgt = it['target']['actual']['claim']
    a2 = spec['a2']
    for x in it['instances']:
        x['primary'] = False
    ins = {'primary': True,
           'scheme': {'walton_id': 'better_means', 'compendium': None, 'wag_id': None, 'fit': 'reconstructed',
                      'confidence': it['role']['confidence'],
                      'alternatives': [{'walton_id': [x for x in it['instances']][0]['scheme']['walton_id'],
                                        'why_not': ('it argues for the other course on its own; the claim is that the '
                                                    'other course does better than the tax, the exception it answers '
                                                    '(D075)')}]},
           'bindings': {'A': a, 'A2': a2, 'G': g, 'P': AG},
           'comparison': spec['comparison'],
           'premises': [], 'conclusion': {}}
    rebuild(ins, 'better_means', {
        f'goal_of_agent({AG},{g})': {'status': 'inferred', 'cite': tgt, 'note': (
            "the answered argument's own goal premise (D075, D089)")},
        f'possible_bring_about({AG},{a2})': {'status': 'supplied', 'note': (
            'a premise of the answer (D089): that governments can take the other course; the claim does not argue it')},
        f'brings_about({a2},{g})': {'status': spec['brings_status'], 'note': spec['brings_note']},
        f'more_effective_means({AG},{a2},{g},{a})': {'status': spec['comp_status'], 'note': spec['comp_note']},
    }, {})
    it['instances'].insert(0, ins)
    forms = {'possible': f'possible_bring_about({AG},{a2})', 'brings': f'brings_about({a2},{g})',
             'more': f'more_effective_means({AG},{a2},{g},{a})'}
    # The records of the argument for the other course, now secondary, stay
    # (renamed below with that instance's formulas); the better_means
    # instance's come first, one record where the two name one formula.
    def renamed(c):
        f = nows(c.get('formula'))
        new = {m[f] for m in OWN.get(cid, []) if f in m}
        return new.pop() if len(new) == 1 else f
    fresh = []
    for t, k, n in spec['challenges']:
        same = [c for c in it['challenges'] if renamed(c) == nows(forms[k])]
        if same:
            shape = {'possible': 'possible_bring_about(P,A2)', 'brings': 'brings_about(A2,G)',
                     'more': 'more_effective_means(P,A2,G,A)'}[k]
            same[0]['note'] = since(same[0].get('note'), f'also a premise of the primary better_means instance, {shape} '
                                                         '(D089)')
        else:
            fresh.append({'cq': None, 'wag_cq': None, 'type': t, 'formula': forms[k], 'note': n})
    it['challenges'] = fresh + it['challenges']
    it['notes']['text'] += (f' Since {WHEN} (D075, D089) its primary instance is better_means, the answer scheme for '
                            'another means that does better; the argument for the other course is secondary.')
    record(cid, ['instances', 'target.actual.formula', 'challenges', 'notes'],
           'a primary better_means instance at the exception it answers, realize_more_effectively(P,G,tax_meat), '
           'with its comparison (D075, D089); the argument for the other course is secondary')

# 13292.147 offers another course and claims nothing better (its children
# compare it the other way, 13292.149, or by who bears the cost, 13292.151):
# under D006 it is no reason against the tax, so no answer at the exception.
it = I['13292.147']
if it['role'].get('cq') != 'CQ:other_means' or primary(it)['scheme']['walton_id'] != 'positive_consequences':
    raise SystemExit('13292.147: not the other-means answer it was')
it['role'] = {'kind': 'action_amendment', 'confidence': it['role']['confidence'], 'cq': None, 'polarity': 'against',
              'move': 'alternative_means'}
it['target']['actual'] = {'claim': '13292.108', 'part': 'conclusion'}
it['target']['re_parented'] = it['target']['actual']['claim'] != it['target']['kialo_parent']
it['notes']['text'] += (f' Since {WHEN} (D006, D089): it claims no better means (its children compare it the other '
                        'way, 13292.149, or by who bears the cost, 13292.151), so it answers nothing at the exception; '
                        'it proposes another course for the same goal, a separate alternative (D066, C017), until the '
                        'consequences proposal settles how alternatives are argued.')
record('13292.147', ['role', 'target', 'notes'], 'another course not claimed better: no answer at '
       'realize_more_effectively (D006); a proposal of another course, at 13292.108\'s conclusion')

# 13292.208 follows 13292.104's new primary instance.
it = I['13292.208']
if nows(it['target']['actual'].get('formula')) != 'current_circumstances(farming_industry_not_meat_is_problem)':
    raise SystemExit('13292.208: not the target it had')
it['target']['actual'] = {'claim': '13292.104', 'part': 'premise',
                          'formula': 'brings_about(regulate_animal_farming_industry,environmental_harm_from_meat_reduced)',
                          'instance': 'primary'}
it['notes']['text'] += (f' Since {WHEN} 13292.104\'s primary instance is better_means (D075): the claim disputes that '
                        'regulating the industry would reduce the harm, the premise brings_about(A2,G) of that '
                        'instance, as its note above already read it.')
record('13292.208', ['target', 'notes'], "13292.104's primary instance is better_means: the target is its premise "
       "that regulating the industry brings about the goal")
for c in I['13292.104']['challenges']:
    if nows(c.get('formula')) == 'current_circumstances(farming_industry_not_meat_is_problem)':
        c['note'] = since(c.get('note'), "13292.208 attacks the better_means instance's premise "
                                         "brings_about(A2,G) now (D075)")

# The item notes the recode makes stale get a dated addendum.
I['13292.108']['notes']['text'] += (
    f' Since {WHEN} (proposal D004-D075): 13292.104 answers the other-means question with better_means; 13292.147 '
    'claims no better means, so it is a proposal of another course at the conclusion (D006); 13292.174 and 13292.5 '
    'are rebuttals at the conclusion (D070); 13292.127 weighs a side effect against the goal at the exception '
    'bad_outweighs(A,G).')
for c in I['13292.108']['challenges']:
    if nows(c.get('formula')) == 'realize_more_effectively(environmental_harm_from_meat_reduced,tax_meat)':
        c['note'] = since(c.get('note'), '13292.147 claims no better means: a proposal of another course, not an answer '
                                         'here (D006)')
I['13292.108']['adjudication']['recoded'][-1]['fields'].append('notes')
I['13292.45']['notes']['text'] += (
    f' Since {WHEN} (D030, D090): the scheme states P in each premise and has no A2; G is the goal (the atom bound to '
    'A before) and S the barriers to cheaper meat imports, of which higher standards are one.')
I['13292.45']['adjudication']['recoded'][-1]['fields'].append('notes')

# (5) the weighings coded Other become instances of C003's pair.
WEIGH = {
    '13292.43': dict(wid='good_outweighs_bad', h='sustainability_and_animal_welfare', g='tax_revenue_reduced',
                     sign=('good', 'inferred', "'other, more pressing issues': what the tax would serve is good"),
                     comp=('explicit', "'more pressing'"), old_comp='more_pressing'),
    '13292.127': dict(wid='bad_outweighs_good', h='starvation_from_reduced_meat_access',
                      g='environmental_harm_from_meat_reduced',
                      sign=('bad', 'inferred', "'a more severe threat'"),
                      comp=('explicit', ("'a more severe threat than any consequence of high meat consumption': ranked "
                                         "against 13292.108's goal, the earliest argument of the class on the board "
                                         "(OC7)")), old_comp='more_serious_consequences'),
}
for cid, w in WEIGH.items():
    it = I[cid]
    old = primary(it)
    oldp = {parse(p['formula'])[1] + '/' + parse(p['formula'])[2][1]: p for p in old['premises']
            if parse(p['formula'])[1] == 'brings_about'}
    brings = oldp.get(f"brings_about/{w['h']}")
    if brings is None:
        raise SystemExit(f'{cid}: no brings_about premise for {w["h"]}')
    ins = {'primary': True,
           'scheme': {'walton_id': w['wid'], 'compendium': None, 'wag_id': None, 'fit': 'reconstructed',
                      'confidence': old['scheme'].get('confidence'),
                      'alternatives': old['scheme'].get('alternatives') or []},
           'bindings': {'A': 'tax_meat', 'H': w['h'], 'G': w['g']},
           'premises': [], 'conclusion': {}}
    sign, sstatus, snote = w['sign']
    rebuild(ins, w['wid'], {
        f"brings_about(tax_meat,{w['h']})": {k: brings[k] for k in ('status', 'cite', 'note') if k in brings},
        f"{sign}({w['h']})": {'status': sstatus, 'note': snote},
        f"matters_more({w['h']},{w['g']})": {'status': w['comp'][0], 'note': w['comp'][1]},
    }, {})
    it['instances'] = [ins] + [x for x in it['instances'] if x is not old]
    for c in it['challenges']:
        if parse(c['formula'])[1] == w['old_comp']:
            c['formula'] = f"matters_more({w['h']},{w['g']})"
    it['notes']['text'] += (f' Since {WHEN} (D064, D075) the weighing is an instance of {w["wid"]}, C003\'s answer '
                            'scheme, in place of Other.')
    record(cid, ['instances', 'challenges', 'notes'], f'the weighing is an instance of {w["wid"]} (D064, D075)')
it = I['13292.127']
it['target']['actual'] = {'claim': '13292.108', 'part': 'exception',
                          'formula': 'bad_outweighs(tax_meat,environmental_harm_from_meat_reduced)', 'instance': 'primary'}
it['target']['re_parented'] = it['target']['actual']['claim'] != it['target']['kialo_parent']
it['role']['cq'] = '22.1 CQ5'
it['notes']['text'] += (f' Since {WHEN} practical reasoning is in the class (D070, D091), so the earliest argument of '
                        'the class on the board is 13292.108, not 13292.115 (OC7): the answer is at its exception, '
                        '22.1 CQ5.')
I['13292.127']['adjudication']['recoded'][-1]['fields'] += ['role.cq', 'target']

# Every target and challenge record that named a renamed formula follows. A
# record names a formula of its own claim's instances or of its target's
# primary instance (rule R10; 13292.137 names 13292.108's), so it is renamed
# by the instance it names, never by another claim's: two instances can share
# A and S1 and differ in G (13292.108 and 13292.6), and side_effects(A,S1)
# becomes a different bad_outweighs(A,G) in each.
SIDE_NOTES = {
    '13292.108': ("no claim weighs a side effect against this goal but 13292.127 (starvation, bad_outweighs_good); "
                  "13292.174 and 13292.5 cite side effects that rank nothing, rebuttals at the conclusion (D070)"),
    '13292.118': ("13292.7's affordability is a plain side effect, a rebuttal at the conclusion; it would answer here "
                  "only if claimed to outweigh the price goal (D070)"),
}
BEFORE = {it['kialo_claim_id']: {f: json.dumps(it[f], sort_keys=True) for f in ('target', 'challenges')}
          for it in gold['items']}
for it in gold['items']:
    cid = it['kialo_claim_id']
    a = it['target'].get('actual') or {}
    m = REMAP.get(a.get('claim'), {})
    if a.get('formula') and nows(a['formula']) in m:
        a['formula'] = m[nows(a['formula'])]
    for o in it['target'].get('other_kialo_parents') or []:
        m = REMAP.get(o.get('claim_id'), {})
        if o.get('formula') and nows(o['formula']) in m:
            o['formula'] = m[nows(o['formula'])]
    maps = OWN.get(cid, []) + ([REMAP[a['claim']]] if a.get('claim') in REMAP else [])
    for c in it['challenges']:
        f = nows(c.get('formula'))
        new = {m[f] for m in maps if f in m}
        if len(new) > 1:
            raise SystemExit(f'{cid}: the record {f} names a formula that two instances rename differently: {sorted(new)}')
        if new:
            c['formula'] = new.pop()
            if f.startswith('side_effects('):
                c['type'] = 'side effects that outweigh'
                c['note'] = since(c.get('note'), SIDE_NOTES.get(cid, 'the exception is the claim that they outweigh '
                                                 'the goal (D070); a plain side effect is a rebuttal at the conclusion'))

for it in gold['items']:
    cid = it['kialo_claim_id']
    changed = [f for f in ('target', 'challenges') if json.dumps(it[f], sort_keys=True) != BEFORE[cid][f]]
    if not changed:
        continue
    mine = [r for r in it['adjudication'].get('recoded') or [] if r.get('by') == BY]
    if mine:
        for f in changed:
            if f not in mine[-1]['fields'] and not any(x.startswith(f + '.') for x in mine[-1]['fields']):
                mine[-1]['fields'].append(f)
    else:
        record(cid, changed, "the formulas it names follow walton.yml's renamed ones (proposal D004-D075)")

# Written as the gold and the lexicon are (indent 1, no newline at the end).
for path, data in ((GOLD, gold), (LEX, lex)):
    with open(path, 'w', encoding='utf-8') as fh:
        fh.write(json.dumps(data, indent=1, ensure_ascii=False))
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    for name in ('gold.json', 'lexicon.json'):
        shutil.copyfile(DOC + name, COPY + name)
print('recoded: 8 practical_reasoning2 instances; 13292.45; 13292.174 and 13292.5; 13292.123 and 13292.104; '
      '13292.147; 13292.208; 13292.43 and 13292.127; the challenge records')
