"""Recode the batch-2 gold for step C5.2's switch (docs/c5-game.md).

    python3 -I recode_c52.py

A one-off migration, kept as the record of what changed and why. Since step
C5.2 the game's data is generated from the scheme catalogue with the entries'
game parts: the game plays the modal variants of the consequence schemes, the
bridges of possible consequences and the answer schemes (D075, D095), and asks
questions at formulas where it asked none. Where it now asks one at a formula
the gold names, the gold names the game's question, as R2 asks (validate.py):

- 13292.25 argues the exception too_slight of 13292.7#1. No question asked it,
  so it was an undercutter (D074); it answers CQ:too_slight now.
- 13292.12 and 13292.233 answer the modal premise may_bring_about of 13292.9#1
  and 13292.121#1, and 13292.180's second function backs 13292.121#1's: the
  book's 33.2 CQ1 is the game's CQ:modal_likelihood there.
- 13292.127 answers the outweighing exception bad_outweighs of 13292.108#1: the
  book's 22.1 CQ5 is the game's CQ:bad_outweighs there.
- 13292.43 answers the exception good_outweighs of 13292.9#1: the book's 33.2
  CQ3 is the game's CQ:good_outweighs there.
- The challenges at bad_outweighs(...) of 13292.108#1, 13292.118#1, 13292.104#2
  and 13292.190#1 named CQ:side_effect, which the game asked at practical
  reasoning's side_effects(A,S1) before; it asks CQ:bad_outweighs at the
  exception now, and CQ:side_effect, the plain side effect, at the conclusion.
- Seventeen challenges named no game question (wag_cq null) at formulas where
  the game asks one now: the book's 33.1 CQ3 and 33.2 CQ3 at the outweighing
  exceptions (CQ:bad_outweighs, CQ:good_outweighs; the game's question at the
  conclusion, CQ:side_effect or CQ:opposite_consequences, keeps its own
  record, D055), 33.2 CQ1 and CQ2 at the modal premises (CQ:modal_likelihood,
  CQ:modal_evidence), and the severity and bearer challenges at too_slight and
  harm_discounted_by_bearer (CQ:too_slight, CQ:harm_discounted). validate.py
  checks a challenge's wag_cq as it checks a role's cq (R2).
- The instances of the schemes the game now plays name its scheme (wag_id,
  R16): bad_outweighs_good, good_outweighs_bad, better_means,
  negative_consequences_modal and bridge_from_possible_harm.

It refuses to run on a gold that is not the one it was written for. It then
checks the gold with validate.py, which renews the pins of the inputs that
changed (the game's NASCQ files) only when nothing else fails, and copies it
to the annotation engine's copy (wag/data/annotation/kialo-13292-batch2/).
"""
import json
import os
import shutil
import subprocess
import sys

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
COPY = REPO + 'wag/data/annotation/kialo-13292-batch2/'
GOLD = DOC + 'gold.json'
gold = json.load(open(GOLD, encoding='utf-8'))
game = json.load(open(REPO + 'wag/data/nascq/schemes.json', encoding='utf-8'))
GAME_OF = {v['carneades_id']: k for k, v in game.items()}
WHEN = '2026-11-27'
BY = "step C5.2's switch (docs/c5-game.md; recode_c52.py)"
G = {it['kialo_claim_id']: it for it in gold['items']}
SINCE = 'Since 2026-11-27 (step C5.2)'

# The roles: claim -> (where, the book's question, the game's question).
ROLES = {
    '13292.12': ('role', '33.2 CQ1', 'CQ:modal_likelihood'),
    '13292.233': ('role', '33.2 CQ1', 'CQ:modal_likelihood'),
    '13292.127': ('role', '22.1 CQ5', 'CQ:bad_outweighs'),
    '13292.43': ('role', '33.2 CQ3', 'CQ:good_outweighs'),
    '13292.180': ('other parent 13292.121', '33.2 CQ1', 'CQ:modal_likelihood'),
}
# The challenges at the outweighing exception of practical reasoning.
CHALLENGES = {'13292.108': '13292.108#1', '13292.118': '13292.118#1', '13292.104': '13292.104#2',
              '13292.190': '13292.190#1'}
# The instances of the schemes the game now plays.
PLAYED = {'13292.9#1', '13292.9#4', '13292.127#1', '13292.174#1', '13292.174#2', '13292.104#1', '13292.43#1',
          '13292.121#1', '13292.121#2', '13292.123#1'}


def refuse(why):
    raise SystemExit(f'{GOLD}: {why}: not the gold this recode was written for (already recoded?)')


changes = {}  # claim -> (fields, [what])


def note(cid, field, what):
    fields, whats = changes.setdefault(cid, (set(), []))
    fields.add(field)
    whats.append(what)


# 13292.25: the undercutter at too_slight answers CQ:too_slight.
it = G['13292.25']
a = it['target']['actual']
if it['role']['kind'] != 'undercutter' or it['role']['cq'] is not None or a.get('formula') != 'too_slight(tax_meat,food_less_affordable_for_poor)':
    refuse('13292.25 is not the undercutter at too_slight')
it['role']['kind'], it['role']['cq'] = 'cq_answer', 'CQ:too_slight'
it['notes']['text'] += (f' {SINCE}: the game asks CQ:too_slight at the exception, so the claim answers that question: '
                        'cq_answer, CQ:too_slight (R2).')
note('13292.25', 'role', 'the game asks CQ:too_slight at too_slight: the undercutter is its answer (R2)')
note('13292.25', 'notes', None)

for cid, (where, book, wag) in ROLES.items():
    it = G[cid]
    if where == 'role':
        r = it['role']
        if r['kind'] != 'cq_answer' or r['cq'] != book:
            refuse(f'{cid} does not answer {book}')
        r['cq'] = wag
        field, whose = 'role', 'the answer'
    else:
        others = [o for o in it['target'].get('other_kialo_parents') or [] if o.get('claim_id') == '13292.121']
        if len(others) != 1 or others[0].get('cq') != book:
            refuse(f'{cid} does not answer {book} toward 13292.121')
        others[0]['cq'] = wag
        field, whose = 'target', 'its function toward 13292.121'
    it['notes']['text'] += f' {SINCE}: the game asks {book} as {wag}, the question {whose} names now (R2).'
    note(cid, field, f'the game asks {book} as {wag}: {whose} names it (R2)')
    note(cid, 'notes', None)

def add_note(c, text):
    c['note'] = (c['note'].rstrip('.') + '. ' if c.get('note') else '') + f'{SINCE}: {text}'


for cid, iid in CHALLENGES.items():
    it = G[cid]
    cs = [c for c in it['challenges'] if c.get('instance') == iid and c.get('wag_cq') == 'CQ:side_effect'
          and (c.get('formula') or '').startswith('bad_outweighs(')]
    if len(cs) != 1:
        refuse(f'{cid} has no challenge CQ:side_effect at the bad_outweighs of {iid}')
    c = cs[0]
    c['wag_cq'] = 'CQ:bad_outweighs'
    add_note(c, 'the game asks the exception as CQ:bad_outweighs; CQ:side_effect is the plain side effect, at the conclusion')
    note(cid, 'challenges', f'the challenge at the bad_outweighs of {iid} names CQ:bad_outweighs, the game\'s question there (R2)')

# The challenges that named no game question where the game asks one now: by
# the formula's predicate and the book's question (or, with none, the type).
FILL = {('may_bring_about', '33.2 CQ1'): 'CQ:modal_likelihood', ('may_bring_about', '33.2 CQ2'): 'CQ:modal_evidence',
        ('good_outweighs', '33.2 CQ3'): 'CQ:good_outweighs', ('bad_outweighs', '33.1 CQ3'): 'CQ:bad_outweighs',
        ('too_slight', 'severity'): 'CQ:too_slight', ('harm_discounted_by_bearer', 'bearer'): 'CQ:harm_discounted'}
FILLED = {('13292.7#1', 'good_outweighs'), ('13292.7#1', 'too_slight'), ('13292.9#1', 'may_bring_about', '33.2 CQ1'),
          ('13292.9#1', 'may_bring_about', '33.2 CQ2'), ('13292.9#1', 'good_outweighs'), ('13292.115#1', 'bad_outweighs'),
          ('13292.180#1', 'good_outweighs'), ('13292.127#2', 'good_outweighs'),
          ('13292.174#1', 'may_bring_about', '33.2 CQ1'), ('13292.174#1', 'may_bring_about', '33.2 CQ2'),
          ('13292.174#1', 'harm_discounted_by_bearer'), ('13292.121#1', 'may_bring_about', '33.2 CQ1'),
          ('13292.121#1', 'may_bring_about', '33.2 CQ2'), ('13292.121#1', 'good_outweighs'),
          ('13292.147#1', 'bad_outweighs'), ('13292.194#1', 'bad_outweighs'), ('13292.5#1', 'too_slight')}
filled = set()
for it in gold['items']:
    for c in it['challenges']:
        if c.get('wag_cq') is not None or not c.get('formula'):
            continue
        pred = c['formula'].lstrip('¬').split('(')[0]
        wag = FILL.get((pred, c.get('cq') or c.get('type')))
        if wag is None:
            continue
        c['wag_cq'] = wag
        add_note(c, f'the game asks it as {wag}')
        filled.add((c['instance'], pred, c['cq']) if pred == 'may_bring_about' else (c['instance'], pred))
        note(it['kialo_claim_id'], 'challenges',
             f'the challenge {c.get("cq") or c.get("type")} at the {pred} of {c["instance"]} names {wag}, the game\'s question there (R2)')
if filled != FILLED:
    refuse(f'the challenges with no game question where the game asks one are {sorted(filled)}, not {sorted(FILLED)}')

seen = set()
for it in gold['items']:
    for ins in it['instances']:
        s = ins['scheme']
        want = GAME_OF.get(s['walton_id'])
        if s.get('wag_id') == want:
            continue
        if ins['id'] not in PLAYED or s.get('wag_id') is not None:
            refuse(f'{ins["id"]} names {s.get("wag_id")!r} as its game scheme, not null for a scheme the game now plays')
        s['wag_id'] = want
        seen.add(ins['id'])
        note(it['kialo_claim_id'], 'instances', f'the game plays {s["walton_id"]} as {want}: {ins["id"]} names it (R16)')
if seen != PLAYED:
    refuse(f'the instances of the newly played schemes are {sorted(seen)}, not {sorted(PLAYED)}')

for cid, (fields, whats) in changes.items():
    G[cid]['adjudication'].setdefault('recoded', []).append(
        {'date': WHEN, 'by': BY, 'fields': sorted(fields), 'what': '; '.join(w for w in whats if w)})
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("The game's data is generated from the catalogue with the entries' game parts since step C5.2: it plays "
             "the modal variants, the bridges of possible consequences and the answer schemes, and asks questions at "
             "formulas where it asked none. The gold names the game's question where it now asks one (R2): 13292.25 "
             "answers CQ:too_slight (an undercutter before); 13292.12, 13292.233 and 13292.180's function toward "
             "13292.121 name CQ:modal_likelihood, 13292.127 CQ:bad_outweighs and 13292.43 CQ:good_outweighs, for the "
             "book's questions; four challenges at practical reasoning's outweighing exception name CQ:bad_outweighs "
             "for CQ:side_effect, and seventeen that named no game question name the one the game asks at their "
             "formula now (the outweighing exceptions, the modal premises, too_slight and the bearer exception). "
             "The ten instances of the schemes the game now plays name its scheme (wag_id, R16). "
             "The pins of the game's NASCQ files renewed. No claim's reading changed.")})
with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
print(f'{len(changes)} items recoded; gold.json written; checking it (validate.py --repin)')
run = subprocess.run([sys.executable, '-I', DOC + 'validate.py', '--repin'], capture_output=True, text=True)
print(run.stdout.strip())
if run.returncode != 0:
    raise SystemExit('validate.py refuses the recoded gold: fix it, then copy it to ' + COPY)
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
    print('copied to', COPY)
