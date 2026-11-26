"""Recode the batch-2 gold for proposal D022-D093 (stage 2, item C4).

    python3 -I recode_c4.py

A one-off migration, kept as the record of what changed and why. It applies
the gold effects of proposal D022-D093 (docs/proposals/) to gold.json here
and copies it to the annotation engine's copy
(wag/data/annotation/kialo-13292-batch2/). It refuses to run on a gold that
no longer has the instance it was written for. English renderings come from
walton.yml's templates (with the proposal applied) and lexicon.json, as the
gold's were made.

The proposal changes one scheme the gold uses: end_and_means_1, which borrows
the consequence schemes' question of the opposite value as the exception
means_bad_outweighs(X,Y) (D023). 13292.3#1, the gold's only instance of it,
lists the exception, status supplied, as every instance lists its scheme's
exceptions. No other instance uses a scheme the proposal changes; 13292.25
names end_and_means_3, which does not change, among its alternatives.
"""
import json
import os
import re
import shutil

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
WHEN = '2026-11-26'
BY = 'proposal D022-D093 (recode_c4.py)'
if SCHEMES['end_and_means_1'].get('exceptions') != ['means_bad_outweighs(X,Y)']:
    raise SystemExit(f'{WALTON}: end_and_means_1 has not the exception of proposal D022-D093: apply it to walton.yml first')

CHANGED = {'fear', 'danger', 'threat', 'ad_baculum', 'pragmatic_alternative', 'appeal_to_vanity_ad_populum',
           'practical_goal', 'two_person_practical', 'waste', 'end_and_means_1'}
uses = [(it['kialo_claim_id'], ins['id']) for it in gold['items'] for ins in it['instances']
        if ins['scheme']['walton_id'] in CHANGED]
if uses != [('13292.3', '13292.3#1')]:
    raise SystemExit(f'the gold uses the changed schemes in {uses}, not in 13292.3#1 alone: already recoded, or another gold')
it = next(i for i in gold['items'] if i['kialo_claim_id'] == '13292.3')
ins = it['instances'][0]
if any(p['role'] == 'exception' for p in ins['premises']):
    raise SystemExit('13292.3#1 lists an exception already: recoded?')


def en(f):
    """The formula in English, its atoms quoted, as WAG's renderer reads the
    template: %s in order, %[n]s the nth argument (an indexed template)."""
    m = re.fullmatch(r'([a-z_0-9]+)\((.*)\)', f)
    args = ["'" + lex[a] + "'" for a in m.group(2).split(',')]
    seq = iter(args)
    return re.sub(r'%(?:\[(\d+)\])?s', lambda x: args[int(x.group(1)) - 1] if x.group(1) else next(seq),
                  LANG[f'{m.group(1)}/{len(args)}'])


b = ins['bindings']
f = 'means_bad_outweighs({},{})'.format(b['X'], b['Y'])
ins['premises'].append({
    'role': 'exception', 'formula': f, 'en': en(f), 'status': 'supplied',
    'note': ('exception: presumed not to hold unless shown. Since 2026-11-26 (decision D023, proposal D022-D093): '
             'end_and_means_1 borrows the consequence schemes\' question of the opposite value, the critic\'s to '
             'support.')})
it['adjudication'].setdefault('recoded', []).append(
    {'date': WHEN, 'by': BY, 'fields': ['instances'],
     'what': "end_and_means_1's new exception, means_bad_outweighs (D023), supplied"})
with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
print('13292.3#1: means_bad_outweighs listed; gold.json written')
