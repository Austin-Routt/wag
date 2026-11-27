"""Recode the batch-2 gold for the argument the discussion needs (decision D096; round 1, draft 3, part 2).

    python3 -I recode_d096.py

A one-off migration, kept as the record of what changed and why. 13292.108 says only that meat
production and consumption harm the environment; the thesis has no argument of its own; draft 1 bound
the whole environmental argument for the tax in 13292.108's record so that the claims answering it had
somewhere to attach. D096 (the owner, 2026-11-27) keeps two layers apart, and
docs/annotation-records.md §13 says how (R20):

- the environmental argument becomes the discussion's, an instance of the thesis, 13292.1#2, marked
  "by": "the discussion": its circumstances cited to 13292.108, which states them, its efficacy premise
  to 13292.14, which states the chain from the tax through the price to less consumption against
  13292.13's denial, the rest supplied; its challenge records move with it;
- 13292.108 becomes what it states: a fragment of that argument backing its circumstances
  (premise_backing), as 13292.11 backs 13292.118's; the claims at the circumstances (13292.2, .4,
  .135, .137) keep their targets;
- the seven claims that answer the argument's other parts target 13292.1#2: 13292.5, .147 and .174 at
  its conclusion, 13292.13 at its efficacy premise, 13292.6 at its goal link, 13292.104 and .127 at its
  exceptions;
- 13292.104's goal premise, cited to 13292.108, which does not state the goal, is supplied (the
  reconstruction is not a claim, R20); 13292.137's conditional other-means challenge, which named an
  exception 13292.108's own record no longer has, is the thesis's record's now (R10);
- every note that named 13292.108's argument names the discussion's.

It refuses to run on a gold other than the one it was written for (its SHA-256), writes gold.json, and
in d086-rounds.json sets the round-1 claims it changes to draft 3, not reviewed, and releases the claims
that waited on Q1 (D096). It then checks the gold with validate.py and copies it to the annotation
engine's copy.
"""
import copy
import hashlib
import json
import os
import shutil
import subprocess
import sys

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
COPY = REPO + 'wag/data/annotation/kialo-13292-batch2/'
GOLD = DOC + 'gold.json'
ROUNDS = DOC + 'd086-rounds.json'
# The gold this recode was written for: after recode_d097.py (8b883c0).
BEFORE = '8b541d6064610beaf5141d5afdbfd4af4a1c6a876db808c009ff22852ce8472b'
WHEN = '2026-11-27'
BY = 'decision D096 (docs/annotation-records.md §13; recode_d096.py)'
ENV = 'the environmental argument for the tax (the discussion\'s, 13292.1#2; D096)'

raw = open(GOLD, 'rb').read()
if hashlib.sha256(raw).hexdigest() != BEFORE:
    raise SystemExit(f'{GOLD} is not the gold this recode was written for (already recoded?)')
gold = json.loads(raw)
G = {it['kialo_claim_id']: it for it in gold['items']}


def refuse(why):
    raise SystemExit(f'{GOLD}: {why}: not the gold this recode was written for')


changes = {}  # claim -> (fields, [what])


def note(cid, field, what=None):
    fields, whats = changes.setdefault(cid, (set(), []))
    fields.add(field)
    if what and what not in whats:
        whats.append(what)


def swap(cid, path_desc, text, old, new):
    if text.count(old) != 1:
        refuse(f'{cid}\'s {path_desc} does not read as this recode expects ({old[:60]!r})')
    return text.replace(old, new)


# --- 1. The discussion's argument, at the thesis ---------------------------------------------
thesis, frag = G['13292.1'], G['13292.108']
if len(thesis['instances']) != 1 or len(frag['instances']) != 1:
    refuse('the thesis or 13292.108 has more than one instance')
old = frag['instances'][0]
if old['id'] != '13292.108#1' or old['scheme']['walton_id'] != 'practical_reasoning2' \
        or old['conclusion']['formula'] != 'should_be_performed(tax_meat)':
    refuse('13292.108#1 is not the environmental argument')
S1 = 'current_circumstances(meat_production_and_consumption_harm_environment)'
EFF = 'would_bring_about(tax_meat,meat_production_and_consumption_harm_environment,reduced_meat_consumption)'
LINK = 'would_be_realized(environmental_harm_from_meat_reduced,reduced_meat_consumption)'
GOAL = 'goal_of_agent(governments,environmental_harm_from_meat_reduced)'
disc = copy.deepcopy(old)
disc['id'], disc['primary'], disc['by'] = '13292.1#2', False, 'the discussion'
disc['scheme']['fit'] = 'reconstructed'
disc['label'] = 'the environmental argument for the tax, as the discussion gives it (D096)'
for pm in disc['premises']:
    if pm['formula'] == S1:
        pm['status'], pm['cite'] = 'inferred', '13292.108'
        pm['note'] = 'stated by 13292.108, the fragment that backs it; 13292.2 and .4 back it, 13292.135 and .137 contest it'
    elif pm['formula'] == EFF:
        pm['status'], pm['cite'] = 'inferred', '13292.14'
        pm['note'] = ("stated by 13292.14 (the tax raises the price of meat, and a higher price lowers consumption) "
                      "against 13292.13's denial; 13292.61 backs it from the tobacco tax")
    elif pm['formula'] == LINK:
        pm['note'] = 'supplied: no claim states it; 13292.6 states a case of it (less demand slows deforestation) and backs it'
    elif pm['formula'] == GOAL:
        pm['note'] = ("the agent's goal (D005): the discussion's environmental pros presuppose it and no claim states it; "
                      "the agent is the thesis's, governments")
    if pm['status'] == 'explicit':
        refuse(f'13292.1#2 would keep the explicit premise {pm["formula"]}')
thesis['instances'].append(disc)
moved = [c for c in frag['challenges'] if c['formula'] != S1]
kept = [c for c in frag['challenges'] if c['formula'] == S1]
if len(moved) != 4 or len(kept) != 1:
    refuse('13292.108\'s challenges are not the five this recode moves and keeps')
for c in moved:
    c['instance'] = '13292.1#2'
thesis['challenges'] = moved
frag['challenges'] = kept
thesis['notes']['text'] += (
    ' 13292.1#2 is the environmental argument for the tax as the discussion gives it (D096, since 2026-11-27; '
    'docs/annotation-records.md §13): no claim states it as a whole, so it is recorded here, marked by the '
    'discussion, not as the thesis\'s own argument. 13292.108 states its circumstances (a fragment that backs them), '
    '13292.14 its efficacy premise; 13292.5, .147 and .174 answer it at its conclusion, 13292.13 and .6 at its '
    'premises, 13292.104 and .127 at its exceptions. Its challenge records are here.')
note('13292.1', 'instances', 'the environmental argument for the tax is the discussion\'s instance 13292.1#2 (D096): '
     'circumstances cited to 13292.108, efficacy to 13292.14, the rest supplied')
note('13292.1', 'challenges', 'the four challenge records of the environmental argument moved here from 13292.108')
note('13292.1', 'notes')

# --- 2. 13292.108: what it states -----------------------------------------------------------
ins = old
ins['scheme']['fit'] = 'fragment'
ins['bindings'] = {
    'A': {'unstated': "the action: the discussion's environmental argument (13292.1#2) has it (tax_meat); this claim does not state it"},
    'S1': 'meat_production_and_consumption_harm_environment',
    'S2': {'unstated': "the resulting state: the discussion's environmental argument has it (less meat is demanded and consumed); this claim does not state it"},
    'G': {'unstated': "the goal: the discussion's environmental argument has it (the environmental harm is reduced); this claim does not state it"},
    'P': {'unstated': "the agent: the discussion's environmental argument has it (governments); this claim does not state it"},
}
ins['premises'] = [p for p in ins['premises'] if p['formula'] == S1]
if len(ins['premises']) != 1 or ins['premises'][0]['status'] != 'explicit':
    refuse('13292.108 does not state its circumstances explicitly')
ins['conclusion'] = {'formula': None, 'en': ("Not stated: the conclusion names the action, which this fragment leaves to the "
                                             "discussion's environmental argument (13292.1#2; should_be_performed(tax_meat) there).")}
frag['role']['kind'] = 'premise_backing'
frag['target']['actual'] = {'claim': '13292.1', 'instance': '13292.1#2', 'part': 'premise', 'formula': S1}
frag['notes']['text'] = (
    "States the circumstances S1 of the environmental argument for the tax: meat production and consumption harm the "
    "environment. The thesis has no argument of its own, and no claim states this one as a whole; it is recorded as "
    "the discussion's, at the thesis (13292.1#2; D096, docs/annotation-records.md §13), and this claim is a fragment of "
    "it (guidelines §3; §5 item 6) that backs its circumstances (premise_backing), as 13292.11 backs 13292.118's. It "
    "binds the variable of the one formula it states; the action, the resulting state, the goal and the agent are the "
    "discussion's argument's, and so is the conclusion. 13292.2 and .4 back its circumstances; 13292.135 and .137 "
    "contest them. Until draft 3 its record bound the whole argument, the tax and the goal supplied (D086 round 1, Q1).")
note('13292.108', 'instances', 'a fragment of the discussion\'s argument 13292.1#2 stating its circumstances only (D096)')
note('13292.108', 'role', 'premise_backing of the circumstances it states')
note('13292.108', 'target', 'the circumstances of 13292.1#2')
note('13292.108', 'challenges', 'its four challenge records of the whole argument moved to the thesis\'s record')
note('13292.108', 'notes')

# --- 3. The seven claims that answer the argument's other parts --------------------------------
SEVEN = {'13292.5': 'conclusion', '13292.147': 'conclusion', '13292.174': 'conclusion', '13292.13': 'premise',
         '13292.6': 'premise', '13292.104': 'exception', '13292.127': 'exception'}
for cid, part in SEVEN.items():
    a = G[cid]['target']['actual']
    if a.get('claim') != '13292.108' or a.get('instance') != '13292.108#1' or a.get('part') != part:
        refuse(f'{cid} does not target 13292.108#1\'s {part}')
    a['claim'], a['instance'] = '13292.1', '13292.1#2'
    note(cid, 'target', f'its target is the discussion\'s argument 13292.1#2 (D096), at the same {part}')

# --- 4. The notes, and the two records R20 and R10 refuse -----------------------------------------
it = G['13292.13']
it['notes']['text'] = swap('13292.13', 'notes', it['notes']['text'],
    "so the actual target is the would_bring_about premise of 108's reconstructed practical reasoning (re-parented, R1);",
    f"so the actual target is the would_bring_about premise of {ENV}, its Kialo parent's argument;")
note('13292.13', 'notes')

it = G['13292.127']
pm = it['instances'][0]['premises'][2]
pm['note'] = swap('13292.127', 'premise note', pm['note'], "ranked against 13292.108's goal,",
                  "ranked against the goal of the environmental argument for the tax (13292.1#2),")
it['notes']['text'] = swap('13292.127', 'notes', it['notes']['text'],
    "the earliest argument of the class on the board, 13292.108's practical reasoning, at its exception",
    f"the earliest argument of the class on the board, {ENV}, at its exception")
note('13292.127', 'instances.premises')
note('13292.127', 'notes')

it = G['13292.61']
pm = it['instances'][0]['premises'][2]
pm['note'] = swap('13292.61', 'premise note', pm['note'], 'the efficacy premise of 13292.108 that 13292.13 denies',
                  'the efficacy premise of the environmental argument for the tax (13292.1#2) that 13292.13 denies')
it['notes']['text'] = swap('13292.61', 'notes', it['notes']['text'], 'the efficacy premise of 13292.108 that 13292.13 denies',
                           'the efficacy premise of the environmental argument for the tax (13292.1#2) that 13292.13 denies')
note('13292.61', 'instances.premises')
note('13292.61', 'notes', 'its notes name the discussion\'s argument (D096)')

it = G['13292.174']
lab = it['instances'][1]
lab['label'] = swap('13292.174', 'label', lab['label'], "the bridge to 13292.108's conclusion (D008)",
                    "the bridge to the environmental argument's conclusion (13292.1#2; D008)")
it['notes']['text'] = swap('13292.174', 'notes', it['notes']['text'],
    "offered against the thesis's environmental argument 13292.108 (a sibling of the Kialo parent, hence re-parented)",
    f"offered against {ENV}, re-parented from 13292.13")
note('13292.174', 'instances')
note('13292.174', 'notes')

it = G['13292.104']
goal = it['instances'][0]['premises'][0]
if goal['formula'] != GOAL or goal.get('cite') != '13292.108' or goal['status'] != 'inferred':
    refuse('13292.104#1\'s goal premise is not cited to 13292.108')
goal['status'] = 'supplied'
del goal['cite']
goal['note'] = ("the goal of the argument it answers, the discussion's environmental argument (13292.1#2; D075, D089): no "
                "claim states it, so supplied; it was cited to 13292.108, which states only the circumstances (D096, R20)")
pm = it['instances'][0]['premises'][2]
pm['note'] = swap('13292.104', 'premise note', pm['note'], "the formula's S2 is that harm, 13292.108's goal,",
                  "the formula's S2 is that harm, the goal of the environmental argument (13292.1#2),")
it['notes']['text'] = swap('13292.104', 'notes', it['notes']['text'],
    "Read as a better means, it answers 13292.108's exception realize_more_effectively",
    "Read as a better means, it answers the exception of the discussion's environmental argument (13292.1#2; D096) "
    "realize_more_effectively")
it['notes']['text'] = swap('13292.104', 'notes', it['notes']['text'], 're-parented from 13292.13, a sibling of 13292.108.',
                           're-parented from 13292.13.')
note('13292.104', 'instances.premises', 'its goal premise is supplied: 13292.108 does not state the goal, and the '
     'discussion\'s argument is not a claim to cite (R20)')
note('13292.104', 'notes')

it = G['13292.137']
other = [c for c in it['challenges'] if c.get('cq') == '22.1 CQ2']
if len(other) != 1 or other[0]['instance'] != '13292.108#1':
    refuse('13292.137 has no other-means challenge at 13292.108#1')
it['challenges'].remove(other[0])
it['notes']['text'] += (" Read so, it would raise the environmental argument's other-means question (22.1 CQ2), whose "
                        "challenge record is the thesis's, at 13292.1#2 (D096); this claim's target is the circumstances "
                        "13292.108 states.")
note('13292.137', 'challenges', 'its conditional other-means challenge is the thesis\'s record\'s (13292.1#2; R10)')
note('13292.137', 'notes')

it = G['13292.2']
it['notes']['text'] = swap('13292.2', 'notes', it['notes']['text'],
                           "Backs the circumstances S1 of 13292.108's environmental argument, for production:",
                           "Backs the circumstances S1 that 13292.108 states, for production:")
note('13292.2', 'notes')

it = G['13292.6']
x = it['instances'][0]
alt = x['scheme']['alternatives'][2]
alt['why_not'] = swap('13292.6', 'alternative', alt['why_not'], "so the fragment of 13292.108's practical reasoning",
                      "so the fragment of the environmental argument for the tax (13292.1#2)")
x['bindings']['A'] = {'unstated': "the action: the discussion's environmental argument (13292.1#2) has it (tax_meat); this claim does not state it"}
x['bindings']['S1'] = {'unstated': "the circumstances: the discussion's environmental argument (13292.1#2) has them, as 13292.108 states them; this claim presupposes them, it does not state them"}
x['bindings']['P'] = {'unstated': "the agent: the discussion's environmental argument (13292.1#2) has it (governments); this claim does not state it"}
x['conclusion']['en'] = ("Not stated: the conclusion names the action, which this fragment leaves to the discussion's "
                         "environmental argument (13292.1#2; should_be_performed(tax_meat) there).")
t = it['notes']['text']
t = swap('13292.6', 'notes', t, "States the S2 -> G link of the thesis's environmental argument (13292.108)",
         f"States the S2 -> G link of {ENV}")
t = swap('13292.6', 'notes', t, "the action, the circumstances and the agent are 13292.108's and are left unstated here",
         "the action, the circumstances and the agent are that argument's and are left unstated here")
t = swap('13292.6', 'notes', t, "It backs 13292.108's premise would_be_realized", "It backs that argument's premise would_be_realized")
t = swap('13292.6', 'notes', t, "which also supports 13292.108's circumstances.", "which also supports the circumstances 13292.108 states.")
it['notes']['text'] = t
note('13292.6', 'instances', 'its unstated variables and conclusion name the discussion\'s argument (D096)')
note('13292.6', 'notes')

it = G['13292.147']
it['notes']['text'] = swap('13292.147', 'notes', it['notes']['text'],
    "coded at the conclusion of the argument it replies to, 13292.108's should_be_performed(tax_meat) (an ancestor, hence re-parented)",
    f"coded at the conclusion of the argument it replies to, {ENV}, should_be_performed(tax_meat) (re-parented from 13292.2)")
note('13292.147', 'notes')

it = G['13292.5']
it['notes']['text'] = swap('13292.5', 'notes', it['notes']['text'],
    "offered against the thesis's environmental argument 13292.108 (an ancestor, hence re-parented)",
    f"offered against {ENV}, re-parented from 13292.4")
note('13292.5', 'notes')

it = G['13292.123']
alt = it['instances'][0]['scheme']['alternatives'][0]
alt['why_not'] = swap('13292.123', 'alternative', alt['why_not'], "as 13292.104 does for 13292.108's,",
                      "as 13292.104 does for the environmental argument's (13292.1#2),")
it['notes']['text'] = swap('13292.123', 'notes', it['notes']['text'], "as 13292.104 does 13292.108's:",
                           "as 13292.104 does the environmental argument's (13292.1#2):")
note('13292.123', 'instances')
note('13292.123', 'notes')

CHANGED = {'13292.1', '13292.108', '13292.5', '13292.147', '13292.174', '13292.13', '13292.6', '13292.104', '13292.127',
           '13292.61', '13292.137', '13292.2', '13292.123'}
if set(changes) != CHANGED:
    refuse(f'the items changed are {sorted(changes)}')
for cid, (fields, whats) in sorted(changes.items()):
    G[cid]['adjudication'].setdefault('recoded', []).append(
        {'date': WHEN, 'by': BY, 'fields': sorted(fields),
         'what': '; '.join(whats) if whats else 'its notes name the discussion\'s environmental argument, 13292.1#2 (D096)'})
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Round 1, draft 3, part 2: the argument the discussion needs (D096; docs/annotation-records.md §13, R20). "
             "The environmental argument for the tax, which no claim states as a whole, is the discussion's instance of "
             "the thesis, 13292.1#2: its circumstances cited to 13292.108, its efficacy to 13292.14, the rest supplied, "
             "its four challenge records moved with it. 13292.108 is a fragment of it backing its circumstances. "
             "13292.5, .147, .174, .13, .6, .104 and .127 target 13292.1#2; 13292.104's goal premise is supplied (R20); "
             "13292.137's other-means challenge is the thesis's record's (R10); notes that named 13292.108's argument "
             "name the discussion's.")})

# d086-rounds.json: the round-1 claims changed go to draft 3, not reviewed; Q1 is decided and applied.
rounds = json.load(open(ROUNDS, encoding='utf-8'))
ROUND1 = {c for c, e in rounds['claims'].items() if e['round'] == 1 or e.get('joined') == 1}
WAITED = {c for c, e in rounds['claims'].items() if 'Q1' in (e.get('pending') or [])}
if not WAITED <= CHANGED & ROUND1 or len(WAITED) != 7:
    refuse(f'the claims pending Q1 are {sorted(WAITED)}')
for c in sorted(CHANGED & ROUND1):
    e = rounds['claims'][c]
    e.update(draft=3, review='not reviewed', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])
    left = [p for p in e.get('pending') or [] if p != 'Q1']
    if left:
        e['pending'] = left
    else:
        e.pop('pending', None)
for c in CHANGED - ROUND1:
    if rounds['claims'][c]['review'] != 'not drafted':
        refuse(f'{c} is not of round 1 but has been drafted')

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
