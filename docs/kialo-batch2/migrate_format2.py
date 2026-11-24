"""Migrate the batch-2 gold to format 2 of the annotation records
(docs/annotation-records.md, decision D092), with decision D039's recoding of
the hypothetical causal steps, in one new version of the gold (2026-11-24).

    python3 -I migrate_format2.py

1. The gold as it was is kept, byte for byte, as gold-format1.json: D039's
   "the version before is kept", and format 1's last gold.
2. D039: the 11 causal steps whose cause has not occurred (cause_to_effect
   with a supplied has_occurred premise, coded under the N8 note) become
   cause_to_effect_modal: may_occur(E1) and causes(E1,E2), concluding
   may_occur(E2). Their status, fit and confidence are kept: the confidence
   is the one the N8 note capped, until the first gold's draft (D086)
   revisits it. Each change is recorded in the item's adjudication.recoded.
3. Format 2: every instance gets its id, <claim>#<n>, n its place in the
   item; every target names its instance (format 1's targets are their
   claim's primary instance) and its formula (a conclusion target's is its
   instance's conclusion); the entries of other_kialo_parents likewise; every
   challenge names the instance it is about; re_parented, which is derived,
   is dropped; the file pins its inputs and records the migration in its
   history. No coding changes in this step.

One-off, kept as the record: it refuses to run on a gold already in format 2.
Paths are relative to this file."""
import hashlib
import json
import os
import re
import sys

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
WHEN = '2026-11-24'
GOLD, KEPT = DOC + 'gold.json', DOC + 'gold-format1.json'
INPUTS = {'items.json': DOC + 'items.json', 'lexicon.json': DOC + 'lexicon.json', 'case.json': DOC + 'case.json',
          'walton.yml': REPO + 'docs/scheme-registry/walton-clean/walton.yml'}
N8 = 'hypothetical cause (needs a modal causal scheme, audit N8)'

raw = open(GOLD, 'rb').read()
gold = json.loads(raw)
if gold.get('format') == 'wag-annotation/2':
    raise SystemExit('gold.json is in format 2 already: this migration has run')
if os.path.exists(KEPT) and open(KEPT, 'rb').read() != raw:
    raise SystemExit('gold-format1.json exists and is not the gold being migrated')
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
case = json.load(open(DOC + 'case.json', encoding='utf-8'))
items = json.load(open(DOC + 'items.json', encoding='utf-8'))['items']
G = {it['kialo_claim_id']: it for it in gold['items']}


def nows(f):
    return re.sub(r'\s+', '', f or '')


def en_event(template, atom):
    return template % ("'" + lex[atom] + "'")


def primary_index(it):
    ps = [n for n, ins in enumerate(it['instances']) if ins['primary']]
    if len(ps) != 1:
        raise SystemExit(f"{it['kialo_claim_id']}: {len(ps)} primary instances")
    return ps[0]


# --- 2. D039 -----------------------------------------------------------------
MAY = 'An event %s has happened, or could happen.'  # walton.yml's may_occur/1
steps = []  # (claim, instance index, its has_occurred premise)
for cid, it in G.items():
    for n, ins in enumerate(it['instances']):
        if ins['scheme']['walton_id'] != 'cause_to_effect':
            continue
        ho = [p for p in ins['premises'] if p['formula'].startswith('has_occurred(')]
        if len(ho) != 1 or ho[0]['status'] != 'supplied':
            continue  # an explicit occurrence: the cause has occurred, the scheme stays
        e1, e2 = ins['bindings']['E1'], ins['bindings']['E2']
        if nows(ho[0]['formula']) != f'has_occurred({e1})' or nows(ins['conclusion']['formula']) != f'will_occur({e2})':
            raise SystemExit(f'{cid}: instance {n} is not the causal step this was written for')
        if N8 not in (ins.get('note') or '') or ins['scheme']['fit'] != 'reconstructed' or ins['scheme']['confidence'] > 0.5:
            raise SystemExit(f'{cid}: instance {n} was not coded under the N8 note')
        steps.append((cid, n, ho[0]))
if len(steps) != 11:
    raise SystemExit(f'D039 names 11 hypothetical causal steps, not {len(steps)}: {[(c, n) for c, n, _ in steps]}')
# The formulas that change (has_occurred(E1), will_occur(E2) of each step) may
# be named only by the steps themselves: no target, challenge or other
# instance may name one.
H = {(c, n) for c, n, _ in steps}
OLD = set()
for cid, n, ho in steps:
    OLD |= {nows(ho['formula']), nows(G[cid]['instances'][n]['conclusion']['formula'])}
for other in G.values():
    oid = other['kialo_claim_id']
    names = [nows((other['target'].get('actual') or {}).get('formula'))]
    names += [nows(o.get('formula')) for o in other['target'].get('other_kialo_parents') or []]
    names += [nows(c.get('formula')) for c in other['challenges']]
    for m, oi in enumerate(other['instances']):
        if (oid, m) not in H:
            names += [nows(p['formula']) for p in oi['premises']] + [nows((oi.get('conclusion') or {}).get('formula'))]
    if OLD & set(names):
        raise SystemExit(f'{oid} names {OLD & set(names)}, which D039 changes')
recoded = {}
for cid, n, ho in steps:
    ins = G[cid]['instances'][n]
    e1, e2 = ins['bindings']['E1'], ins['bindings']['E2']
    ins['scheme']['walton_id'] = 'cause_to_effect_modal'
    ho['formula'], ho['en'] = f'may_occur({e1})', en_event(MAY, e1)
    note = ho.get('note') or ''
    ho['note'] = ('the cause is proposed or possible, not occurred (D039)' + note[len(N8):]) if note.startswith(N8) \
        else 'the cause is proposed or possible, not occurred (D039)'
    ins['conclusion'] = {'formula': f'may_occur({e2})', 'en': en_event(MAY, e2)}
    rest = ins['note'].replace(N8 + '.', '').replace(N8, '').strip().replace('Kept under R4: ', 'A causal scheme because ')
    ins['note'] = ('Recoded with cause_to_effect_modal (D039, ' + WHEN + '): the cause is only proposed or possible. '
                   'The confidence is the one the N8 note capped, until the first gold\'s draft (D086).'
                   + (' ' + rest if rest else ''))
    recoded.setdefault(cid, []).append(n)
# An item's own note that describes the N8 coding gets a dated addendum; the
# adjudicator's text is kept as written.
ADDENDUM = (' Note of ' + WHEN + ' (D039): the hypothetical causal steps are coded with cause_to_effect_modal now; '
            'what this note says of the N8 coding and of R4 is the earlier coding (gold-format1.json).')
noted = set()
for cid in recoded:
    if 'N8' in G[cid]['notes']['text'] or 'R4' in G[cid]['notes']['text']:
        G[cid]['notes']['text'] += ADDENDUM
        noted.add(cid)
# A format-1 target at a conclusion names no formula: a claim that targets the
# conclusion of a recoded primary instance now targets may_occur(E2). Its own
# coding must name neither formula (checked above, by name); its record says
# that its target changed.
for other in G.values():
    a = other['target'].get('actual') or {}
    if a.get('part') != 'conclusion' or a.get('claim') not in recoded:
        continue
    k = primary_index(G[a['claim']])
    if k not in recoded[a['claim']]:
        continue
    e2 = G[a['claim']]['instances'][k]['bindings']['E2']
    other['adjudication'].setdefault('recoded', []).append({
        'date': WHEN, 'by': 'decision D039 (migrate_format2.py)', 'fields': ['target'],
        'what': (f"Its target, the conclusion of {a['claim']}'s causal step, is may_occur({e2}) since that step's recoding "
                 f"(was will_occur({e2})); its own coding names neither formula.")})
for other in G.values():
    for o in other['target'].get('other_kialo_parents') or []:
        if o.get('part') == 'conclusion' and o.get('claim_id') in recoded:
            raise SystemExit(f"{other['kialo_claim_id']}: another Kialo parent's conclusion is a recoded step: not written for")
for cid, ns in recoded.items():
    G[cid]['adjudication'].setdefault('recoded', []).append({
        'date': WHEN, 'by': 'decision D039 (migrate_format2.py)', 'fields': ['instances'] + (['notes'] if cid in noted else []),
        'what': ('The hypothetical causal step' + ('s' if len(ns) > 1 else '') + ' (instance' + ('s ' if len(ns) > 1 else ' ')
                 + ', '.join(f'{cid}#{n + 1}' for n in ns) + ') recoded with cause_to_effect_modal: may_occur(E1) and '
                 'causes(E1,E2), concluding may_occur(E2), for has_occurred(E1) and will_occur(E2) under the N8 note.')})

# --- 3. Format 2 -------------------------------------------------------------


def ids(it):
    return [f"{it['kialo_claim_id']}#{n + 1}" for n in range(len(it['instances']))]


def forms(ins, part):
    if part == 'conclusion':
        f = (ins.get('conclusion') or {}).get('formula')
        return [nows(f)] if f else []
    roles = {'premise': ('premise', 'assumption'), 'exception': ('exception',)}[part]
    return [nows(p['formula']) for p in ins['premises'] if p['role'] in roles]


def exact(claim, part, formula, said):
    """The instance and formula a format-1 target names: its claim's primary
    instance (format 1 has no other), and the formula given, or for a
    conclusion that instance's conclusion."""
    if said not in (None, 'primary'):
        raise SystemExit(f'a target names the instance {said!r}, not the primary one')
    t = G[claim]
    k = primary_index(t)
    ins = t['instances'][k]
    if part == 'inference':
        return ids(t)[k], None
    if part == 'conclusion' and not formula:
        formula = ins['conclusion']['formula']
    if nows(formula) not in forms(ins, part):
        raise SystemExit(f'{claim}: {formula} is not a {part} of its primary instance')
    return ids(t)[k], formula


def owned(it, c):
    """The instance a challenge is about: the claim's own instance with its
    formula (where two have it, the one of the scheme its book label names),
    else its target's instance; a challenge with no formula is about the
    claim's primary instance as a whole."""
    f = nows(c.get('formula'))
    k = primary_index(it)
    if not f:
        return ids(it)[k]
    have = [n for n, ins in enumerate(it['instances'])
            if f in [nows(p['formula']) for p in ins['premises']] + [nows((ins.get('conclusion') or {}).get('formula'))]]
    if len(have) > 1:
        book = (c.get('cq') or '').split(' ')[0]
        have = [n for n in have if book and it['instances'][n]['scheme'].get('compendium') == book]
        if len(have) != 1:
            raise SystemExit(f"{it['kialo_claim_id']}: the challenge {c} is about more than one instance")
    if have:
        return ids(it)[have[0]]
    a = it['target'].get('actual') or {}
    if a.get('claim') in G:
        t = G[a['claim']]
        k = primary_index(t)
        ins = t['instances'][k]
        if f in [nows(p['formula']) for p in ins['premises']] + [nows(ins['conclusion'].get('formula'))]:
            return ids(t)[k]
    raise SystemExit(f"{it['kialo_claim_id']}: no instance has the challenge's formula {c.get('formula')}")


def ordered(d, first):
    return {**{k: d[k] for k in first if k in d}, **{k: v for k, v in d.items() if k not in first}}


out = []
for it in gold['items']:
    cid = it['kialo_claim_id']
    insts = []
    for n, ins in enumerate(it['instances']):
        insts.append({'id': f'{cid}#{n + 1}', **ins})
    t = it['target']
    if t.get('actual'):
        a = t['actual']
        inst, formula = exact(a['claim'], a['part'], a.get('formula'), a.get('instance'))
        actual = {'claim': a['claim'], 'instance': inst, 'part': a['part'], 'formula': formula}
        extra = {k: v for k, v in a.items() if k not in ('claim', 'instance', 'part', 'formula')}
        if extra:
            raise SystemExit(f'{cid}: the target has fields this was not written for: {sorted(extra)}')
    else:
        if t.get('kialo_parent') is not None:
            raise SystemExit(f'{cid}: no actual target, but a Kialo parent')
        actual = None
    target = {'kialo_parent': t['kialo_parent'], 'kialo_relation': t['kialo_relation'], 'actual': actual}
    if t.get('other_kialo_parents'):
        others = []
        for o in t['other_kialo_parents']:
            o = dict(o)
            if o.get('part') in ('premise', 'exception', 'conclusion'):
                o['instance'], o['formula'] = exact(o['claim_id'], o['part'], o.get('formula'), o.get('instance'))
                o = ordered(o, ['claim_id', 'relation', 'role', 'cq', 'polarity', 'part', 'instance', 'formula'])
            others.append(o)
        target['other_kialo_parents'] = others
    extra = set(t) - {'kialo_parent', 'kialo_relation', 'actual', 're_parented', 'other_kialo_parents'}
    if extra:
        raise SystemExit(f'{cid}: the target has fields this was not written for: {sorted(extra)}')
    challenges = [ordered({**c, 'instance': owned(it, c)}, ['cq', 'wag_cq', 'type', 'instance', 'formula']) for c in it['challenges']]
    new = {}
    for k, v in it.items():
        new[k] = {'instances': insts, 'target': target, 'challenges': challenges}.get(k, v)
    out.append(new)

sha = {name: hashlib.sha256(open(p, 'rb').read()).hexdigest() for name, p in INPUTS.items()}
migrated = {
    'format': 'wag-annotation/2',
    'annotator': gold['annotator'],
    'case': case['id'],
    'inputs': sha,
    'status': ('Development material (D086): adjudicated by a language model, no item verified by a person. '
               'The first gold is drafted from it.'),
    'history': [{'date': WHEN, 'by': 'decision D092 (migrate_format2.py)', 'what': (
        'Format 2 (docs/annotation-records.md): instance ids, exact targets (their instance, and a conclusion target\'s '
        'formula), the instance each challenge is about, re_parented dropped (derived), the inputs pinned. No coding '
        'changed. The format-1 gold is kept as gold-format1.json. Decision D039\'s recoding of the 11 hypothetical '
        'causal steps is recorded in their items.')}],
    'items': out,
}
if [i['kialo_claim_id'] for i in out] != [i['kialo_claim_id'] for i in items]:
    raise SystemExit('the gold is not in items.json order')
with open(KEPT, 'wb') as fh:
    fh.write(raw)
# Written as the gold and the lexicon are (indent 1, no newline at the end).
with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(migrated, indent=1, ensure_ascii=False))
print(f'gold.json: format 2, {len(out)} items, {sum(len(i["instances"]) for i in out)} instances; '
      f'D039: {sum(len(v) for v in recoded.values())} causal steps in {len(recoded)} items; gold-format1.json kept')
