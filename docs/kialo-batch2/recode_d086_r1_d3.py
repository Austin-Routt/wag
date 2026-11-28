"""Round 1 of the first gold, draft 3, part 4: exchange 2's objections and verdicts, and D102.

    python3 -I recode_d086_r1_d3.py

A one-off migration, kept as the record of what changed and why. Exchange 2 (the other reviewer's item
review of draft 2, wag-workspace verify/chatgpt-d086-x2-2026-11-27/REVIEW.md) agreed with 11 of the 13
items it reviewed and objected to two; the developer accepted both objections, and the owner decided
one of them as a rule (D102, 2026-11-27). So:

- 13292.8 (D102): it states a condition 13292.115's premise presupposes, excess consumption for the tax
  to reduce, not the premise; its relation to the premise is unresolved: premise_backing, its target's
  formula {"unresolved": why}, no question. It keeps its content and its Kialo relation; it is not
  played, not scored as a relation and not verified until a connecting argument is recorded and
  reviewed (guidelines §3);
- 13292.104#2 (guidelines §5 item 4): the only S2 practical reasoning could name, a regulated industry,
  restates the action, and a proposal with a purpose is positive_consequences: brings_about(A,G)
  explicit ('the appropriate solution'), good(G) supplied. Its challenge records follow: likelihood
  (33.1 CQ1) and the opposite consequences (33.1 CQ3) at the new formulas; the feasibility record goes
  to the better means, whose premise possible_bring_about(P,A2) it is (D089); the record of the
  circumstances, which no instance states now, goes (13292.104#1's record of its link names 13292.208's
  attack);
- 13292.192#2: the source's 'may' is carried by the modal form (may_occur), and causes(E1,E2) is a
  tendency, strong or weak (D036), whose strength 28 CQ1 asks: its premise note and the item's note say
  so exactly;
- 13292.1#2: its scheme's alternatives were copied from 13292.108's record ("the claim states no
  consequence of the tax, only the circumstances") and are false of the discussion's argument (the
  independent review of D096's application, W5): they are written for it;
- d086-rounds.json: exchange 2's verdicts are recorded for its 13 items; the four agreed items unchanged
  since (13292.25, .218, .43, .204) are agreed at that revision; the items this recode changes are at
  draft 3, not reviewed.

It refuses to run on a gold other than the one it was written for (its SHA-256), writes gold.json and
d086-rounds.json, checks the gold with validate.py and copies it to the annotation engine's copy.
"""
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys

import yaml

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
COPY = REPO + 'wag/data/annotation/kialo-13292-batch2/'
GOLD = DOC + 'gold.json'
ROUNDS = DOC + 'd086-rounds.json'
# The gold this recode was written for: after recode_d098.py (2534911).
BEFORE = '3196b3d2b55d52c63568cdb4653f039254fe92a13646eeeb60296605973c8734'
WHEN = '2026-11-28'
BY = "round 1's draft 3, part 4: exchange 2's objections, and D102 (recode_d086_r1_d3.py)"
RECORD = ('wag-workspace verify/chatgpt-d086-x2-2026-11-27/REVIEW.md; verify/chatgpt-rule-2026-11-27/REVIEW.md '
          '(the second review)')

raw = open(GOLD, 'rb').read()
if hashlib.sha256(raw).hexdigest() != BEFORE:
    raise SystemExit(f'{GOLD} is not the gold this recode was written for (already recoded?)')
gold = json.loads(raw)
G = {it['kialo_claim_id']: it for it in gold['items']}
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
LANG = yaml.safe_load(open(REPO + 'docs/scheme-registry/walton-clean/walton.yml', encoding='utf-8'))['language']


def refuse(why):
    raise SystemExit(f'{GOLD}: {why}: not the gold this recode was written for')


def en(f):
    """walton.yml's English for a formula over lexicon atoms, as validate.py's R21 reads it."""
    m = re.fullmatch(r'([a-z_][a-z0-9_]*)\(([^()]*)\)', f)
    args = ["'" + lex[a] + "'" for a in m.group(2).split(',')]
    seq = iter(args)
    return re.sub(r'%(?:\[(\d+)\])?s', lambda x: args[int(x.group(1)) - 1] if x.group(1) else next(seq),
                  LANG[f'{m.group(1)}/{len(args)}'])


def swap(text, old, new, where):
    if text.count(old) != 1:
        refuse(f'{where} does not say {old[:60]!r} once')
    return text.replace(old, new)


changes = {}


def recoded(cid, fields, what):
    G[cid]['adjudication'].setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': fields, 'what': what})
    changes[cid] = (fields, what)


# 1. 13292.8 (D102).
it = G['13292.8']
a = it['target']['actual']
if (a.get('claim'), a.get('instance'), a.get('part'), a.get('formula')) != (
        '13292.115', '13292.115#1', 'premise', 'brings_about(tax_meat,reduced_unhealthy_meat_consumption)'):
    refuse(f'13292.8 targets {a}')
if it['role'].get('kind') != 'cq_answer' or it['role'].get('cq') != 'CQ:likelihood':
    refuse("13292.8's role is not the cq_answer to CQ:likelihood")
a['formula'] = {'unresolved': "the claim states a condition 13292.115's premise presupposes (excess consumption for "
                              "the tax to reduce), not the premise; no recorded argument connects them (D102)"}
it['role'].update(kind='premise_backing', cq=None)
old_head = ("Backs a background condition of 13292.115's premise brings_about(tax_meat,reduced_unhealthy_meat_consumption): "
            "that there is unhealthy (excess) meat consumption for the tax to reduce. That condition is not the premise, "
            "which says that the tax would reduce such consumption: the claim backs it only in part, and the step from "
            "excess consumption to the tax's reducing it is not argued. The target is the nearest formula the format offers "
            "(a target names no component of a premise), and the relation is partial. positive_consequences has "
            "CQ:likelihood (33.1 CQ1) on that premise, so the role is cq_answer with polarity for (R2).")
new_head = ("It states a background condition of 13292.115's premise brings_about(tax_meat,reduced_unhealthy_meat_consumption): "
            "that there is unhealthy (excess) meat consumption for the tax to reduce. That condition is not the premise, "
            "which says that the tax would reduce such consumption, and no claim argues the step from the one to the other. "
            "So its relation to the premise is unresolved (D102; guidelines §3): premise_backing of 13292.115#1's premise, "
            "its formula unresolved, no question. It is not played, not scored as a relation and not verified until an "
            "argument that connects them is recorded and reviewed.")
it['notes']['text'] = swap(it['notes']['text'], old_head, new_head, "13292.8's note")
recoded('13292.8', ['notes', 'role', 'target'],
        "D102: its relation to 13292.115's premise is unresolved (premise_backing, the formula {\"unresolved\": why}, no "
        "question), where it was a cq_answer to CQ:likelihood at that premise; exchange 2's objection, accepted")

# 2. 13292.104#2 (guidelines §5 item 4).
it = G['13292.104']
ins = it['instances'][1]
if ins['id'] != '13292.104#2' or ins['scheme']['walton_id'] != 'practical_reasoning2' or \
        ins['bindings'].get('S2') != 'animal_farming_industry_regulated':
    refuse('13292.104#2 is not the practical reasoning with S2 animal_farming_industry_regulated')
A, Gl = 'regulate_animal_farming_industry', 'meat_problems_resolved'
ref = G['13292.115']['instances'][0]  # positive_consequences, for the exceptions' notes and the scheme's fields
if ref['scheme']['walton_id'] != 'positive_consequences':
    refuse('13292.115#1 is not positive_consequences')
ex_note = {p['formula'].split('(')[0]: p['note'] for p in ref['premises'] if p['role'] == 'exception'}


def pm(role, f, status, note=None):
    x = {'role': role, 'formula': f, 'en': en(f), 'status': status}
    if note:
        x['note'] = note
    return x


ins['scheme'] = {'walton_id': 'positive_consequences', 'compendium': ref['scheme']['compendium'],
                 'wag_id': ref['scheme']['wag_id'], 'fit': 'reconstructed', 'confidence': ins['scheme']['confidence'],
                 'alternatives': [{'walton_id': 'practical_reasoning2',
                                   'why_not': "the only state S2 it could name, a regulated industry, restates the action "
                                              "(guidelines §5 item 4)"}]}
ins['bindings'] = {'A': A, 'G': Gl}
ins['premises'] = [
    pm('premise', f'brings_about({A},{Gl})', 'explicit',
       "'the appropriate solution': regulating the industry resolves the problems 13292.13 names"),
    pm('premise', f'good({Gl})', 'supplied', 'the claim presupposes that resolving the problems is good'),
    pm('exception', f'bad_outweighs({A},{Gl})', 'supplied', ex_note['bad_outweighs']),
    pm('exception', f'too_slight({A},{Gl})', 'supplied', ex_note['too_slight']),
    pm('exception', f'gain_discounted_by_bearer({A},{Gl})', 'supplied', ex_note['gain_discounted_by_bearer']),
]
ins['conclusion'] = {'formula': f'should_be_performed({A})', 'en': en(f'should_be_performed({A})')}
old = [c for c in it['challenges'] if c.get('instance') == '13292.104#2']
if [(c['type'], c['formula'].split('(')[0]) for c in old] != [
        ('circumstances', 'current_circumstances'), ('efficacy', 'would_be_realized'),
        ('efficacy', 'possible_bring_about'), ('side effects that outweigh', 'bad_outweighs')]:
    refuse("13292.104#2's challenge records are not the four this recode rewrites")
efficacy_note = old[1]['note']
it['challenges'] = [c for c in it['challenges'] if c.get('instance') != '13292.104#2'] + [
    {'cq': '33.1 CQ1', 'wag_cq': 'CQ:likelihood', 'type': 'likelihood', 'instance': '13292.104#2',
     'formula': f'brings_about({A},{Gl})', 'note': efficacy_note},
    {'cq': '33.1 CQ3', 'wag_cq': 'CQ:bad_outweighs', 'type': 'opposite consequences', 'instance': '13292.104#2',
     'formula': f'bad_outweighs({A},{Gl})', 'note': old[3]['note']},
    {'cq': None, 'wag_cq': None, 'type': 'feasibility', 'instance': '13292.104#1',
     'formula': f'possible_bring_about(governments,{A})',
     'note': "that governments can regulate the industry, a premise of the better means (D089); 22.1 CQ4 asked it of "
             "the practical reasoning this claim's second instance was until draft 3"},
]
it['notes']['text'] = swap(it['notes']['text'],
                           "The argument for the other course is a secondary instance (practical reasoning, reconstructed: "
                           "its goal premise is supplied);",
                           "The argument for the other course is a secondary instance: positive consequences of regulating "
                           "the industry, a proposal with a purpose (guidelines §5 item 4: the only state S2 practical "
                           "reasoning could name, a regulated industry, restates the action), its link explicit ('the "
                           "appropriate solution'), its evaluation supplied;", "13292.104's note")
recoded('13292.104', ['challenges', 'instances', 'notes'],
        "13292.104#2 is positive_consequences (§5 item 4: its S2 restated the action), where it was practical reasoning; "
        "its challenge records follow (the feasibility record at the better means, the circumstances' dropped); "
        "exchange 2's objection, accepted")

# 3. 13292.192: the source's 'may' and the tendency.
it = G['13292.192']
p = [x for x in it['instances'][1]['premises'] if x['formula'] == 'causes(crop_production_increased,deforestation_for_cropland)']
if len(p) != 1 or p[0].get('note') != '"may still occur": a possibility':
    refuse("13292.192#2's causal premise note is not the one this recode rewrites")
p[0]['note'] = ('"may still occur": the possibility is carried by the modal form (may_occur); causes(E1,E2) is a '
                'tendency, strong or weak (D036), whose strength 28 CQ1 asks')
it['notes']['text'] = swap(it['notes']['text'], "Each step's cause is only possible, so cause_to_effect_modal (§5 item 11):",
                           "Each step's cause is only possible, so cause_to_effect_modal (§5 item 11; the source's 'may' is "
                           "carried by may_occur, and causes(E1,E2) is a tendency, strong or weak, D036):", "13292.192's note")
recoded('13292.192', ['instances.premises', 'notes'],
        "the source's 'may' is carried by the modal form, and causes(E1,E2) is a tendency of any strength (D036): "
        "13292.192#2's premise note and the item's note say so exactly (exchange 2)")

# 4. 13292.1#2's alternatives.
it = G['13292.1']
d = next(x for x in it['instances'] if x.get('by') == 'the discussion')
if [x['walton_id'] for x in d['scheme']['alternatives']] != ['practical_reasoning1', 'positive_consequences']:
    refuse("13292.1#2's alternatives are not the two copied from 13292.108")
d['scheme']['alternatives'] = [
    {'walton_id': 'practical_reasoning1',
     'why_not': 'no claim of the discussion states a value (environmental protection is implicit at most)'},
    {'walton_id': 'positive_consequences',
     'why_not': 'the discussion states the circumstances the tax would address (13292.108) and how the tax would change '
                'them (13292.14), not only a good outcome'}]
recoded('13292.1', ['instances.scheme'],
        "13292.1#2's alternatives are written for the discussion's argument; they were copied from 13292.108's record "
        "(the independent review of D096's application, W5)")

if set(changes) != {'13292.8', '13292.104', '13292.192', '13292.1'}:
    refuse(f'the items changed are {sorted(changes)}')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Round 1, draft 3, part 4: exchange 2's two objections, accepted: 13292.8's relation to 13292.115's premise "
             "is unresolved (D102: premise_backing, its formula unresolved, no question); 13292.104#2 is positive "
             "consequences, since its S2 restated the action (§5 item 4), its challenge records following. 13292.192's "
             "notes say that the modal form carries the source's 'may' and that causes(E1,E2) is a tendency (D036); "
             "13292.1#2's alternatives are written for the discussion's argument.")})

# d086-rounds.json: exchange 2's verdicts; the four agreed items unchanged since are agreed at that revision.
rounds = json.load(open(ROUNDS, encoding='utf-8'))
VERDICTS = {'13292.25': 'agreed', '13292.127': 'agreed', '13292.104': 'objection', '13292.218': 'agreed',
            '13292.11': 'agreed', '13292.43': 'agreed', '13292.6': 'agreed', '13292.12': 'agreed',
            '13292.8': 'objection', '13292.192': 'agreed', '13292.194': 'agreed', '13292.204': 'agreed',
            '13292.123': 'agreed'}
DEVELOPER = {'13292.104': "accepted: 13292.104#2 is positive_consequences (draft 3)",
             '13292.8': "accepted: its relation is unresolved, by the owner's rule D102 (draft 3)"}
UNCHANGED = {'13292.25', '13292.218', '13292.43', '13292.204'}


def item_rev(it):
    x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
    return hashlib.sha256(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8')).hexdigest()


for cid, verdict in VERDICTS.items():
    e = rounds['claims'][cid]
    if any(x.get('exchange') == 2 for x in e.get('exchanges', [])):
        refuse(f'{cid} has exchange 2 recorded already')
    e.setdefault('exchanges', []).append({'exchange': 2, 'date': '2026-11-27', 'draft': 2, 'reviewer': verdict,
                                          'developer': DEVELOPER.get(cid, 'agreed'), 'record': RECORD})
    if cid in UNCHANGED:
        if e['draft'] != 2 or e['review'] != 'not reviewed' or cid in changes:
            refuse(f'{cid} is not at draft 2, not reviewed, unchanged')
        e.update(review='agreed', agreed_rev=item_rev(G[cid]), agreed_by='the reviewer and the developer',
                 agreed_date='2026-11-27')
for cid in sorted(changes):
    e = rounds['claims'][cid]
    if e['round'] != 1 and e.get('joined') != 1:
        refuse(f'{cid} is not of round 1')
    e.update(draft=3, review='not reviewed', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
with open(ROUNDS, 'wb') as fh:
    fh.write(json.dumps(rounds, indent=1, ensure_ascii=False).encode('utf-8') + b'\n')
print(f'{len(changes)} items recoded; gold.json and d086-rounds.json written; checking the gold (validate.py)')
run = subprocess.run([sys.executable, '-I', DOC + 'validate.py'], capture_output=True, text=True)
print(run.stdout.strip()[-3000:])
if run.returncode != 0:
    raise SystemExit('validate.py refuses the recoded gold: fix it, then copy it to ' + COPY)
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
    print('copied to', COPY)
