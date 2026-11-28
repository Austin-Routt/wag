"""Recode the batch-2 gold for the two independent reviews of 28 November: of D098 and R21, and of the
D096 fixes and round 1's draft 3, part 4.

    python3 -I recode_reviews_1128.py

A one-off, kept as the record of what changed and why. Text only: no instance, target, role or
formula changes, but for the English of two nested premises. From the review of the D096 fixes and
draft 3, part 4 (wag-workspace verify/d096/FIXCHECK.md):

- W1: 13292.104's likelihood record names 33.1 CQ1, so its note no longer opens "no printed CQ" (the
  recode of part 4 copied it from the record it replaced); its record at the better means'
  possible_bring_about premise is typed efficacy, as the gold's three other records on that premise
  (13292.118, .123, .45), not "feasibility", a type no other record has.

From the review of D098 and R21 (wag-workspace verify/d098/REVIEW-batch.md):

- W3: 13292.45's note says that how its action is coded is its round's (OC6 codes the claim for the
  amended action, the tax with higher import standards; the coding concludes the added measure
  alone), as rule case OC6-f now says; 13292.194's note no longer says 13292.45 is "coded the same
  way", only that it is at the same place.
- N2: 13292.61#1's two premises that embed its conclusion's formula render it as walton.yml's
  template does, as its conclusion has since recode_d098.py (they had the game's phrasing).
- N4: 13292.204's and 13292.202's notes no longer leave to the review what D098 settled; the
  candidate pattern "mitigation by substitution" of 13292.25, .33, .182 and .198 is "substitution by
  those affected": since D098, a mitigation is the agent's measure (OC6, C006).

It refuses to run on a gold other than the one it was written for (its SHA-256), writes gold.json,
sets the round-1 claims it changes that were agreed to draft 3, not reviewed (d086-rounds.json),
checks the gold with validate.py and copies it to the annotation engine's copy.
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
# The gold this recode was written for: after recode_d086_r1_d3.py (1e2dd3b).
BEFORE = '9ab67d80f842f559a3273ebfa9f4122cb4267d8853ac23bc99d0bf57a0cbe9f7'
WHEN = '2026-11-28'
BY = 'the independent reviews of 28 November (D098 and R21; the D096 fixes and draft 3, part 4) (recode_reviews_1128.py)'

raw = open(GOLD, 'rb').read()
if hashlib.sha256(raw).hexdigest() != BEFORE:
    raise SystemExit(f'{GOLD} is not the gold this recode was written for (already recoded?)')
gold = json.loads(raw)
G = {it['kialo_claim_id']: it for it in gold['items']}
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
LANG = yaml.safe_load(open(REPO + 'docs/scheme-registry/walton-clean/walton.yml', encoding='utf-8'))['language']


def refuse(why):
    raise SystemExit(f'{GOLD}: {why}: not the gold this recode was written for')


def swap(text, old, new, where):
    if text.count(old) != 1:
        refuse(f'{where} does not say {old[:60]!r} once')
    return text.replace(old, new)


changes = {}  # claim -> ([fields], what)


def recoded(cid, fields, what):
    G[cid]['adjudication'].setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': fields, 'what': what})
    changes[cid] = (fields, what)


# 1. 13292.104's two challenge records (the D096 fixes' check, W1).
ch = G['13292.104']['challenges']
lk = [c for c in ch if c.get('cq') == '33.1 CQ1' and c.get('type') == 'likelihood']
if len(lk) != 1 or not lk[0]['note'].startswith('no printed CQ; '):
    refuse("13292.104 has not one likelihood record at 33.1 CQ1 whose note opens 'no printed CQ'")
lk[0]['note'] = lk[0]['note'][len('no printed CQ; '):]
fe = [c for c in ch if c.get('type') == 'feasibility']
if len(fe) != 1 or fe[0]['formula'] != 'possible_bring_about(governments,regulate_animal_farming_industry)':
    refuse('13292.104 has not one feasibility record at its better means\' possible_bring_about premise')
fe[0]['type'] = 'efficacy'
recoded('13292.104', ['challenges'], "its likelihood record's note no longer says no CQ is printed (it names 33.1 CQ1); "
        "its record at possible_bring_about is typed efficacy, as the gold's others on that premise")

# 2. 13292.45 and 13292.194 (the D098 review, W3).
it = G['13292.45']
it['notes']['text'] = it['notes']['text'].rstrip() + (
    ' How its action is coded is its round\'s: OC6 codes the claim for the amended action, the tax with higher import '
    'standards, where this coding concludes the added measure alone (the independent review of D098, W3; rule case OC6-f).')
recoded('13292.45', ['notes'], 'its note leaves the coding of its action, the added measure or the amended action (OC6), '
        'to its round')
it = G['13292.194']
it['notes']['text'] = swap(it['notes']['text'], 'Its sibling 13292.45 is coded the same way (D098).',
                           'Its sibling 13292.45 is at the same place (D098); its action is coded in its round.',
                           "13292.194's note")
recoded('13292.194', ['notes'], 'its note says that 13292.45 is at the same place, not coded the same way')

# 3. 13292.61#1's nested premises in walton.yml's English (the D098 review, N2).
ins = G['13292.61']['instances'][0]
concl = ins['conclusion']
args = ["'" + lex[a] + "'" for a in re.fullmatch(r'would_bring_about\(([^()]*)\)', concl['formula']).group(1).split(',')]
seq = iter(args)
whole = re.sub(r'%s', lambda _: next(seq), LANG['would_bring_about/3'])
if concl['en'] != whole:
    refuse("13292.61#1's conclusion is not in walton.yml's English")
nested = "'" + whole[0].lower() + whole[1:].rstrip('.') + "'"
old = ("'doing 'governments put a tax on meat' in the situation 'meat production and consumption are bad for the "
       "environment' would lead to 'less meat is demanded and consumed (and so less is produced)''")
n = 0
for pm in ins['premises']:
    if concl['formula'] in pm['formula'] and pm['formula'] != concl['formula']:
        pm['en'] = swap(pm['en'], old, nested, f"13292.61#1's premise {pm['formula'][:30]}")
        n += 1
if n != 2:
    refuse(f"13292.61#1 has {n} premises that embed its conclusion, not 2")
recoded('13292.61', ['instances.premises'], "the two premises that embed 13292.61#1's conclusion render it as "
        "walton.yml's template does, as the conclusion itself")

# 4. 13292.204's and 13292.202's notes, and the candidate pattern (the D098 review, N4).
it = G['13292.204']
it['notes']['text'] = swap(it['notes']['text'],
                           "D055's recoding moved it to the weighing's premise, which OC10 now supports: a question for "
                           "the review (D086 round 1, with 13292.202's).",
                           "D055's recoding moved it to the weighing's premise, where OC10 puts it since D098.",
                           "13292.204's note")
recoded('13292.204', ['notes'], 'its note says D098 settled where it is, no longer that it is a question for the review')
it = G['13292.202']
it['notes']['text'] = swap(it['notes']['text'],
                           'once the weighing became primary, D055: a question for the review, D086 round 1).',
                           'once the weighing became primary, D055; D098 settled it, OC6).',
                           "13292.202's note")
recoded('13292.202', ['notes'], 'its note says D098 settled where it is, no longer that it is a question for the review')
PATTERNS = {
    '13292.25': ('mitigation by substitution or adaptation (also 13292.33, 13292.182, 13292.198)',
                 'substitution or adaptation by those affected (also 13292.33, 13292.182, 13292.198)'),
    '13292.33': ('mitigation by substitution (also 13292.182, 13292.198; by adaptation 13292.25)',
                 'substitution by those affected (also 13292.182, 13292.198; adaptation, 13292.25)'),
    '13292.182': ('mitigation by substitution (also 13292.33, 13292.198; by adaptation 13292.25)',
                  'substitution by those affected (also 13292.33, 13292.198; adaptation, 13292.25)'),
    '13292.198': ('mitigation by substitution (also 13292.33, 13292.182; by adaptation 13292.25)',
                  'substitution by those affected (also 13292.33, 13292.182; adaptation, 13292.25)'),
}
for cid, (old, new) in PATTERNS.items():
    if G[cid]['notes'].get('candidate_pattern') != old:
        refuse(f"{cid}'s candidate pattern is not {old!r}")
    G[cid]['notes']['candidate_pattern'] = new
    recoded(cid, ['notes'], "its candidate pattern says substitution by those affected: since D098 a mitigation is the "
            "agent's measure (OC6, C006)")

if set(changes) != {'13292.104', '13292.45', '13292.194', '13292.61', '13292.204', '13292.202', '13292.25', '13292.33',
                    '13292.182', '13292.198'}:
    refuse(f'the items changed are {sorted(changes)}')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Text only, from the independent reviews of 28 November: 13292.104's likelihood note and its "
             "possible_bring_about record's type; 13292.45's note leaves its action's coding to its round (OC6-f) and "
             "13292.194's says it is at the same place; 13292.61#1's two nested premises in walton.yml's English; "
             "13292.204's and 13292.202's notes cite D098; the candidate pattern of 13292.25, .33, .182 and .198 is "
             "substitution by those affected.")})

# d086-rounds.json: the round-1 claims changed go to draft 3, not reviewed.
rounds = json.load(open(ROUNDS, encoding='utf-8'))
for cid in sorted(changes):
    e = rounds['claims'][cid]
    if e['round'] == 1 or e.get('joined') == 1:
        e.update(draft=3, review='not reviewed', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])
    elif e['review'] not in ('not drafted', 'not reviewed'):
        refuse(f'{cid} is not of round 1 but has been reviewed')

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
