"""Bring the batch-2 gold in line with guidelines §3.1 where the independent
review of the D055 application (2026-10-23) found it differs.

    python3 -I recode_d055_review.py

A one-off, kept as the record, like recode_d055.py: it changes gold.json here
and copies it to the annotation engine's copy, and refuses to run twice.

- 13292.29 (eating meat is a personal choice, the state should keep out):
  a liberty set against the consequence whatever its size, which OC5 codes
  with the scheme that fits (here the argument from rules, as the gold has
  it) and names candidate C007 in the notes. The gold named no candidate.
"""
import json
import os
import shutil

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
COPY = os.path.normpath(DOC + '../../wag/data/annotation/kialo-13292-batch2/gold.json')
GOLD = DOC + 'gold.json'
gold = json.load(open(GOLD, encoding='utf-8'))
I = {i['kialo_claim_id']: i for i in gold['items']}

it = I['13292.29']
if it['notes'].get('candidate_pattern'):
    raise SystemExit('13292.29 already names a candidate: already recoded?')
it['notes']['candidate_pattern'] = 'a liberty set against a consequence, whatever its size (C007; guidelines §3.1, OC5)'
it['adjudication'].setdefault('recoded', []).append({
    'date': '2026-10-23',
    'by': 'the independent review of the D055 application (recode_d055_review.py)',
    'fields': ['notes.candidate_pattern'],
    'what': 'Names C007, as OC5 asks of a right or value set against consequences; the coding is unchanged.',
})

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
shutil.copyfile(GOLD, COPY)
print('13292.29 names C007; copied to', os.path.relpath(COPY, DOC))
