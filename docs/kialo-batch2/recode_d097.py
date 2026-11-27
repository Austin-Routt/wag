"""Recode the batch-2 gold for the causal exception that names its effect (proposal D040-D071-D097).

    python3 -I recode_d097.py

A one-off migration, kept as the record of what changed and why. Since 2026-11-27 (decisions D071
and D097, proposal D040-D071-D097, version 2) cause_to_effect and cause_to_effect_modal have the
exception interference(E1,E2), that another causal factor in this case keeps, or would keep, E1 from
bringing about E2, where they had interference(E1), naming only the cause: an answer that blocked one
effect blocked every causal step from that cause (D086 round 1, G2). So:

- the exception of the gold's 17 causal instances names their own cause and effect, its English
  rendered by the new template;
- the 6 challenges at the book's 28 CQ3 name the new formula; 13292.36's no longer cites 13292.24
  (tax credits or food stamps added to the policy) as interference: a safeguard the agent would add
  is an action amendment (OC6), not a factor that keeps the step from its effect;
- 13292.182 and .198, which answer 13292.180#2's step at its exception, name
  interference(reduced_meat_consumption,more_anemia_cases) in their targets and in their Other
  instances' conclusions (R8), and their notes no longer say that the old English misread the step.

It refuses to run on a gold other than the one it was written for (its SHA-256), writes gold.json, and
in d086-rounds.json sets the round-1 claims it changes to draft 3, not reviewed, and releases the claims
that waited on G2 (D097). It then checks the gold with validate.py --repin, which renews the pins of
walton.yml and links.json only when nothing else fails, and copies it to the annotation engine's copy.
"""
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
# The gold this recode was written for: after step 1 after exchange 2 (a514fea).
BEFORE = '851b1c71be1743eb261d99eb9f5a47ba4e27c0e0b406362c1ed8259c29f6213f'
WHEN = '2026-11-27'
BY = 'proposal D040-D071-D097, version 2 (decisions D071 and D097; recode_d097.py)'
TEMPLATE = "Another causal factor in this case keeps, or would keep, {} from bringing about {}."
EXC_NOTE = ("exception: presumed not to hold unless shown. Since 2026-11-27 it names the effect as well as the cause "
            "(D071, D097; walton-clean/CHANGELOG.md), so it blocks this step only")

raw = open(GOLD, 'rb').read()
if hashlib.sha256(raw).hexdigest() != BEFORE:
    raise SystemExit(f'{GOLD} is not the gold this recode was written for (already recoded?)')
gold = json.loads(raw)
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
G = {it['kialo_claim_id']: it for it in gold['items']}


def refuse(why):
    raise SystemExit(f'{GOLD}: {why}: not the gold this recode was written for')


def q(atom):
    return "'" + lex[atom] + "'"


changes = {}  # claim -> (fields, [what])


def note(cid, field, what):
    fields, whats = changes.setdefault(cid, (set(), []))
    fields.add(field)
    if what and what not in whats:
        whats.append(what)


# 1. The exception of every causal instance names its effect.
CAUSAL = ('cause_to_effect', 'cause_to_effect_modal')
steps = {}  # instance id -> (E1, E2)
for it in gold['items']:
    for ins in it['instances']:
        if ins['scheme']['walton_id'] not in CAUSAL:
            continue
        e1, e2 = ins['bindings']['E1'], ins['bindings']['E2']
        ex = [p for p in ins['premises'] if p['role'] == 'exception']
        if len(ex) != 1 or ex[0]['formula'] != f'interference({e1})':
            refuse(f'{ins["id"]} has the exceptions {[p["formula"] for p in ex]}, not interference({e1})')
        p = ex[0]
        p['formula'] = f'interference({e1},{e2})'
        p['en'] = TEMPLATE.format(q(e1), q(e2))
        p['note'] = EXC_NOTE
        steps[ins['id']] = (e1, e2)
        note(it['kialo_claim_id'], 'instances.premises',
             f'{ins["id"]}\'s exception names its effect: interference({e1},{e2}) (D071, D097)')
if len(steps) != 17:
    refuse(f'{len(steps)} causal instances, not 17')

# 2. The challenges at 28 CQ3 name the new formula.
SAFEGUARD = ("no reply in the batch argues such a factor: 13292.24 (tax credits or food stamps for those who cannot "
             "afford food) proposes a safeguard added to the policy, an action amendment (OC6), not a factor that keeps "
             "this step from its effect (since 2026-11-27, proposal D040-D071-D097)")
challenged = set()
for it in gold['items']:
    for c in it['challenges']:
        if not (c.get('formula') or '').startswith('interference('):
            continue
        if c['instance'] not in steps or c.get('cq') != '28 CQ3':
            refuse(f'{it["kialo_claim_id"]} has a challenge at {c["formula"]} of {c["instance"]}, not at a causal step\'s 28 CQ3')
        e1, e2 = steps[c['instance']]
        if c['formula'] != f'interference({e1})':
            refuse(f'{it["kialo_claim_id"]}\'s challenge names {c["formula"]}')
        c['formula'] = f'interference({e1},{e2})'
        if c['instance'] == '13292.36#3':
            if c.get('note') != '13292.24: tax credits or food stamps for those who cannot afford food':
                refuse("13292.36's challenge does not cite 13292.24 as before")
            c['note'] = SAFEGUARD
            note('13292.36', 'challenges', "its challenge at 28 CQ3 no longer cites 13292.24, a safeguard added to the "
                                           "policy (an action amendment, OC6), as interference")
        challenged.add(c['instance'])
        note(it['kialo_claim_id'], 'challenges', f'the challenge at 28 CQ3 of {c["instance"]} names interference({e1},{e2})')
if challenged != {'13292.9#3', '13292.180#2', '13292.36#3', '13292.14#1', '13292.12#1', '13292.141#1'}:
    refuse(f'the challenges at 28 CQ3 are at {sorted(challenged)}')

# 3. The two answers at 13292.180#2's exception.
OLD, NEW = 'interference(reduced_meat_consumption)', 'interference(reduced_meat_consumption,more_anemia_cases)'
NEW_EN = TEMPLATE.format(q('reduced_meat_consumption'), q('more_anemia_cases'))
NOTES = {
    '13292.182': (
        "which is what the book's 28 CQ3 asks (other factors that would interfere with or counteract the production "
        "of the effect; the game asks no question there). So it is a cq_answer at that step's exception "
        "interference(reduced_meat_consumption), against, and its own argument (Other: no walton.yml scheme concludes "
        "it from a substitute) concludes the exception (R8). It does not deny that the tax reduces meat eating; "
        "walton.yml's English for the exception ('an event has occurred which interferes with event ...') reads as if "
        "it did, but the exception is the book's CQ3.",
        "which is what the book's 28 CQ3 asks about: a factor in the case that keeps the cause from bringing about the "
        "effect (the game asks no question there). So it is a cq_answer at that step's exception "
        "interference(reduced_meat_consumption,more_anemia_cases), against, and its own argument (Other: no walton.yml "
        "scheme concludes it from a substitute) concludes the exception (R8). It does not deny that the tax reduces meat "
        "eating: since D071 and D097 the exception names the effect it blocks, so 13292.192#1's step from the same "
        "cause to more crop production is not touched."),
    '13292.198': (
        "a cq_answer at that step's exception interference(reduced_meat_consumption), against; its Other instance "
        "concludes the exception (R8), whose walton.yml English reads as if the reduction itself were interfered with.",
        "a cq_answer at that step's exception interference(reduced_meat_consumption,more_anemia_cases), against; its "
        "Other instance concludes the exception (R8), which names the effect it blocks since D071 and D097."),
}
for cid, (old_note, new_note) in NOTES.items():
    it = G[cid]
    a = it['target']['actual']
    if a.get('instance') != '13292.180#2' or a.get('part') != 'exception' or a.get('formula') != OLD:
        refuse(f'{cid} does not answer 13292.180#2 at {OLD}')
    a['formula'] = NEW
    ins = it['instances'][0]
    if ins['scheme']['walton_id'] != 'other' or ins['conclusion']['formula'] != OLD:
        refuse(f'{cid}\'s Other instance does not conclude {OLD}')
    ins['conclusion'] = {'formula': NEW, 'en': NEW_EN}
    if it['notes']['text'].count(old_note) != 1:
        refuse(f'{cid}\'s note is not the one this recode rewrites')
    it['notes']['text'] = it['notes']['text'].replace(old_note, new_note)
    note(cid, 'target', f'its target, 13292.180#2\'s exception, is {NEW}')
    note(cid, 'instances.conclusion', f'its Other instance concludes {NEW} (R8), in the new template\'s English')
    note(cid, 'notes', 'the note names the exception that names its effect')

CHANGED = {'13292.9', '13292.180', '13292.36', '13292.14', '13292.11', '13292.2', '13292.12', '13292.4', '13292.192',
           '13292.141', '13292.182', '13292.198'}
if set(changes) != CHANGED:
    refuse(f'the items changed are {sorted(changes)}')
for cid, (fields, whats) in sorted(changes.items()):
    G[cid]['adjudication'].setdefault('recoded', []).append(
        {'date': WHEN, 'by': BY, 'fields': sorted(fields), 'what': '; '.join(whats)})
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("The causal exception names its effect (decisions D071 and D097; walton.yml through edits.py, "
             "CAUSAL_D097): the 17 causal instances have interference(E1,E2), their own cause and effect, in the new "
             "template's English; the 6 challenges at 28 CQ3 name it, and 13292.36's no longer cites 13292.24, a "
             "safeguard added to the policy (an action amendment, OC6), as interference; 13292.182 and .198 answer "
             "13292.180#2 at interference(reduced_meat_consumption,more_anemia_cases), which their Other instances "
             "conclude (R8). An answer about anemia no longer undercuts 13292.192#1's step from the same cause to more "
             "crop production. The pins of walton.yml and links.json renewed.")})

# d086-rounds.json: the round-1 claims changed go to draft 3, not reviewed; G2 is decided and applied.
rounds = json.load(open(ROUNDS, encoding='utf-8'))
ROUND1 = {c for c, e in rounds['claims'].items() if e['round'] == 1 or e.get('joined') == 1}
WAITED = {c for c, e in rounds['claims'].items() if 'G2' in (e.get('pending') or [])}
if WAITED != CHANGED & ROUND1 or len(WAITED) != 10:
    refuse(f'the claims pending G2 are {sorted(WAITED)}, not the 10 round-1 claims this recode changes')
for c in sorted(CHANGED & ROUND1):
    e = rounds['claims'][c]
    e.update(draft=3, review='not reviewed', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])
    left = [p for p in e['pending'] if p != 'G2']
    if left:
        e['pending'] = left
    else:
        del e['pending']
for c in CHANGED - ROUND1:
    if rounds['claims'][c]['review'] != 'not drafted':
        refuse(f'{c} is not of round 1 but has been drafted')

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
with open(ROUNDS, 'wb') as fh:
    fh.write(json.dumps(rounds, indent=1, ensure_ascii=False).encode('utf-8') + b'\n')
print(f'{len(changes)} items recoded; gold.json and d086-rounds.json written; checking the gold (validate.py --repin)')
run = subprocess.run([sys.executable, '-I', DOC + 'validate.py', '--repin'], capture_output=True, text=True)
print(run.stdout.strip()[-3000:])
if run.returncode != 0:
    raise SystemExit('validate.py refuses the recoded gold: fix it, then copy it to ' + COPY)
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
    print('copied to', COPY)
