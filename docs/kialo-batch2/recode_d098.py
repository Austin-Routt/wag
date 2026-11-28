"""Recode the batch-2 gold for decision D098 and the independent review of D040-D071-D097's application.

    python3 -I recode_d098.py

A one-off migration, kept as the record of what changed and why. D098 (the owner, 2026-11-27):
guidelines §3.1 states the convention both reviewers of the first gold applied, that an amendment that
mitigates a harm is at the conclusion of the argument that cites the harm against the action (OC6),
and that a denial of the link a weighing relies on is at the weighing (OC10). So:

- 13292.45 (raise import standards along with the tax), an amended action at 13292.12#1's conclusion,
  the causal step that predicts the harm, moves to 13292.12#3's conclusion, the argument that cites the
  harm against the tax, where its sibling 13292.194 is; 13292.12's and 13292.194's records, which said
  it waited for its round, say where it is;

and from the independent review of the application of proposal D040-D071-D097 (wag-workspace
verify/d097/REVIEW-application.md):

- 13292.61#1's conclusion reads as walton.yml's template gives it (R21, W6): it had the game's phrasing;
- 13292.9#3's challenge at 28 CQ3 says the meat tax's own yield could make up for the lost industry
  taxes, a factor that keeps the harm from reducing revenue, not one that only lessens the loss (N5).

It refuses to run on a gold other than the one it was written for (its SHA-256), writes gold.json, sets
the round-1 claims it changes to draft 3, not reviewed (d086-rounds.json), checks the gold with
validate.py and copies it to the annotation engine's copy.
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
# The gold this recode was written for: after recode_d096.py (e1d6aa6).
BEFORE = '4562d203d43d0a76fa90be4d98a2c1e6c77fe6b2d227033981f42ce249953713'
WHEN = '2026-11-28'
BY = 'decision D098 and the independent review of D040-D071-D097\'s application (recode_d098.py)'

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


changes = {}  # claim -> ([fields], what)


def recoded(cid, fields, what):
    G[cid]['adjudication'].setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': fields, 'what': what})
    changes[cid] = (fields, what)


# 1. 13292.45 at the conclusion of the argument that cites the harm (OC6, D098).
it = G['13292.45']
a = it['target']['actual']
if a != {'claim': '13292.12', 'instance': '13292.12#1', 'part': 'conclusion',
         'formula': 'may_occur(domestic_meat_industry_pushed_out)'}:
    refuse(f'13292.45 targets {a}')
third = next((x for x in G['13292.12']['instances'] if x['id'] == '13292.12#3'), None)
if not third or third['conclusion']['formula'] != '¬should_be_performed(tax_meat)':
    refuse('13292.12#3 does not conclude ¬should_be_performed(tax_meat)')
it['target']['actual'] = {'claim': '13292.12', 'instance': '13292.12#3', 'part': 'conclusion',
                          'formula': '¬should_be_performed(tax_meat)'}
it['notes']['text'] = it['notes']['text'].rstrip() + (
    ' Since 2026-11-28 (D098; OC6) it is at the conclusion of the argument that cites the harm against the tax, '
    '13292.12#3, as its sibling 13292.194 is: an amended policy answers the argument against the tax, not the '
    'causal step 13292.12#1 that predicts the harm.')
recoded('13292.45', ['notes', 'target'],
        "at 13292.12#3's conclusion, the argument that cites the harm against the tax, as 13292.194 (OC6, D098); "
        "it was at 13292.12#1's, the causal step that predicts the harm")

# 13292.12's records and 13292.194's note no longer say that it waits for its round.
it = G['13292.12']
ch = [c for c in it['challenges'] if '13292.45' in (c.get('note') or '')]
if len(ch) != 1 or ch[0]['instance'] != '13292.12#1':
    refuse('13292.12 has not one challenge record that names 13292.45, at 13292.12#1')
ch[0]['note'] = swap(ch[0]['note'],
                     "13292.194 at 13292.12#3's conclusion, the argument that cites the harm against the tax, and "
                     "13292.45, until its round, at this argument's conclusion, the predicted harm",
                     "both at 13292.12#3's conclusion, the argument that cites the harm against the tax (D098)",
                     "13292.12's challenge at 13292.12#1")
it['notes']['text'] = swap(it['notes']['text'],
                           "amended actions: 13292.194 at 13292.12#3's conclusion, 13292.45, until its round, at "
                           "13292.12#1's, the predicted harm)",
                           "amended actions, both at 13292.12#3's conclusion, the argument that cites the harm (OC6, D098))",
                           "13292.12's note")
recoded('13292.12', ['challenges', 'notes'], "its records say that 13292.45, like 13292.194, is at 13292.12#3's "
        "conclusion (D098)")

it = G['13292.194']
t = swap(it['notes']['text'], 'As for 13292.202 (OC6),', 'As for 13292.202 (OC6, D098),', "13292.194's note")
it['notes']['text'] = swap(t, 'Its sibling 13292.45 (round 2) is coded the same way when its round comes.',
                           'Its sibling 13292.45 is coded the same way (D098).', "13292.194's note")
recoded('13292.194', ['notes'], 'its note cites D098, and says that 13292.45 is coded the same way')

# 2. 13292.61#1's conclusion in walton.yml's English (R21).
c = G['13292.61']['instances'][0]['conclusion']
if not c['en'].startswith("Doing 'governments put a tax on meat' in the situation"):
    refuse("13292.61#1's conclusion is not in the game's phrasing")
c['en'] = en(c['formula'])
recoded('13292.61', ['instances.conclusion'],
        "13292.61#1's conclusion reads as walton.yml's template gives it (R21); it had the game's phrasing")

# 3. 13292.9#3's challenge at 28 CQ3: a factor that keeps the harm from reducing revenue.
ch = [c for c in G['13292.9']['challenges'] if c.get('instance') == '13292.9#3' and c.get('cq') == '28 CQ3']
if len(ch) != 1 or ch[0]['note'] != "the meat tax's own yield may offset the lost industry taxes":
    refuse("13292.9's challenge at 13292.9#3's 28 CQ3 is not the one this recode rewrites")
ch[0]['note'] = "the meat tax's own yield could make up for the lost industry taxes, so that revenue does not fall"
recoded('13292.9', ['challenges'], "its challenge at 13292.9#3's 28 CQ3 names a factor that would keep the harm from "
        "reducing revenue, not one that only lessens the loss (§5 item 11)")

if set(changes) != {'13292.45', '13292.12', '13292.194', '13292.61', '13292.9'}:
    refuse(f'the items changed are {sorted(changes)}')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Decision D098 (guidelines §3.1, OC6 and OC10): 13292.45, an amended action, is at 13292.12#3's "
             "conclusion, the argument that cites the harm against the tax, as 13292.194, where it was at the causal "
             "step 13292.12#1; 13292.12's and 13292.194's records say so. From the independent review of D040-D071-"
             "D097's application: 13292.61#1's conclusion reads as walton.yml's template gives it (R21), and "
             "13292.9#3's challenge at 28 CQ3 names a factor that keeps the harm from reducing revenue.")})

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
