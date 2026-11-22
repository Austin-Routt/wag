"""Apply the decisions of 2026-11-22 to a WAG checkout.

    python3 -I apply_decisions.py REPO

Text edits to docs/decisions.yml (its layout is kept); data edits to the
catalogue entries the decided decisions list (wag-cli --catalogue-gen then
writes them in its own layout). Refuses to run twice.
"""
import os
import re
import sys
import textwrap

import yaml

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from decisions_2026_11_22 import APPLY, CANDIDATES, D, KEEP_OPEN, NEW, REASSIGN  # noqa: E402

repo = sys.argv[1]
REG = os.path.join(repo, 'docs', 'decisions.yml')
CAT = os.path.join(repo, 'wag', 'data', 'catalogue', 'schemes')
DATE = '2026-11-22'
BY = 'owner (standing rule of 2026-11-22)'

text = open(REG, encoding='utf-8').read()
if 'standing rule of 2026-11-22' in text:
    sys.exit('already applied')
reg = yaml.safe_load(text)
byid = {d['id']: d for d in reg['decisions']}

lines = text.split('\n')


def block(i0):
    """The line range of the decision starting at line i0."""
    j = i0 + 1
    while j < len(lines) and not lines[j].startswith('- id: ') and not lines[j].startswith('# ---'):
        j += 1
    while lines[j - 1].strip() == '':
        j -= 1
    return i0, j


def wrap(s, indent):
    return [indent + w for w in textwrap.wrap(' '.join(s.split()), 78 - len(indent), break_on_hyphens=False)]


def quote(s):
    return '"' + s.replace('\\', '\\\\').replace('"', '\\"') + '"'


def decided_lines(decision, recorded):
    return (['  decided:', f'    date: "{DATE}"', f'    by: {BY}', '    decision: >-'] + wrap(decision, '      ')
            + [f'    recorded: {quote(recorded)}'])


# 1. Decide the open decisions of D.
for did, decision in D.items():
    d = byid[did]
    assert d['status'] == 'open', did
    i0 = lines.index(f'- id: {did}')
    a, b = block(i0)
    seg = lines[a:b]
    k = seg.index('  status: open')
    seg[k] = '  status: decided'
    rec = 'docs/decisions.yml; to apply: ' + APPLY[d['category']]
    seg += decided_lines(decision, rec)
    lines[a:b] = seg

# 2. Add what models found to the decisions a person must still verify.
for did, note in KEEP_OPEN.items():
    i0 = lines.index(f'- id: {did}')
    a, b = block(i0)
    seg = lines[a:b]
    q = seg.index('  question: >-')
    e = q + 1
    while seg[e].startswith('    '):
        e += 1
    old = ' '.join(x.strip() for x in seg[q + 1:e])
    seg[q + 1:e] = wrap(old + ' ' + note, '    ')
    lines[a:b] = seg

# 3. New decisions, at the end of their priority's section.
heads = {1: '# --- Priority 1', 2: '# --- Priority 2', 3: '# --- Priority 3'}
order = ['# --- Priority 1', '# --- Priority 2', '# --- Priority 3', '# --- Decided']
for n in NEW:
    assert n['id'] not in byid, n['id']
    head = heads[n['priority']]
    nxt = order[order.index(head) + 1]
    j = next(i for i, l in enumerate(lines) if l.startswith(nxt))
    while lines[j - 1].strip() == '':
        j -= 1
    out = ['', f"- id: {n['id']}", f"  title: {n['title']}", '  question: >-'] + wrap(n['question'], '    ')
    out += [f"  category: {n['category']}"]
    if n.get('blocks'):
        out += ['  blocks: [' + ', '.join(n['blocks']) + ']']
    out += ['  decider: owner', f"  priority: {n['priority']}", '  status: ' + ('decided' if 'decided' in n else 'open')]
    out += ['  docs: [' + ', '.join(quote(x) for x in n['docs']) + ']']
    if 'decided' in n:
        out += decided_lines(n['decided'], n['recorded'])
    lines[j:j] = out

open(REG, 'w', encoding='utf-8').write('\n'.join(lines))
yaml.safe_load(open(REG, encoding='utf-8'))  # still YAML

# 4. The catalogue discrepancies the decided decisions list no longer say Open.
open_items = set()
for d in reg['decisions']:
    if d['status'] == 'open' and d['id'] not in D:
        for it in d.get('items') or []:
            open_items.add((it['entry'], it['kind'], str(it['about'])))
OPEN = re.compile(r'(^|\. )Open\b')
changed = 0
for did, decision in D.items():
    for it in byid[did].get('items') or []:
        key = (it['entry'], it['kind'], str(it['about']))
        if key in open_items:
            print(f'{did}: {key} is also listed by an open decision; left open')
            continue
        p = os.path.join(CAT, it['entry'] + '.yml')
        e = yaml.safe_load(open(p, encoding='utf-8'))
        hit = [x for x in e['discrepancies'] if x['kind'] == it['kind'] and str(x['about']) == str(it['about'])]
        assert len(hit) == 1, (did, key, len(hit))
        x = hit[0]
        old = OPEN.sub(lambda m: m.group(1) + 'open', x['decision'])
        x['decision'] = f'Decided {DATE} ({did}): {decision} Not yet applied. Before: {old}'
        assert not OPEN.search(x['decision']), (did, key)
        yaml.safe_dump(e, open(p, 'w', encoding='utf-8'), allow_unicode=True, sort_keys=False, width=120)
        changed += 1
# 5. The candidates under the decided decisions: an outcome, or the new D083.
CANDS = os.path.join(repo, 'docs', 'candidates.yml')
cl = open(CANDS, encoding='utf-8').read().split('\n')
for cid in list(CANDIDATES) + list(REASSIGN):
    i0 = cl.index(f'  - id: {cid}')
    j = i0 + 1
    while j < len(cl) and not cl[j].startswith('  - id: '):
        j += 1
    k = next(i for i in range(i0, j) if cl[i].startswith('    decision: D'))
    if cid in REASSIGN:
        cl[k] = f'    decision: {REASSIGN[cid]}'
        continue
    result, into, rec = CANDIDATES[cid]
    out = ['    outcome:', f'      date: "{DATE}"', f'      by: {BY}', f'      result: {result}']
    if into:
        out.append(f'      into: {quote(into)}')
    out.append(f'      recorded: {quote(rec)}')
    cl[k + 1:k + 1] = out
open(CANDS, 'w', encoding='utf-8').write('\n'.join(cl))
yaml.safe_load(open(CANDS, encoding='utf-8'))
print(f'{len(D)} decided, {len(KEEP_OPEN)} annotated, {len(NEW)} new; {changed} catalogue discrepancies updated')
