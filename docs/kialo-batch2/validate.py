"""Validate gold.json and lexicon.json (or another annotation file).

    python3 -I validate.py [ANNOTATIONS.json]

Everything it reads is in the repository: walton.yml
(docs/scheme-registry/walton-clean/), the compendium links
(docs/scheme-registry/links.json: which walton.yml formula each of the
book's critical questions is about; labels only, no book text), and the
game's schemes.json, critical_questions.json and edges.json (wag/data/nascq/).
Paths are relative to this file, so it runs from any directory.

ANNOTATIONS defaults to gold.json. A file that is not the adjudicated gold
(an annotator's, or a model's from wag-llm annotate) is checked the same way,
except for the adjudication records."""
import json, os, re, sys
import yaml

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
ANN = sys.argv[1] if len(sys.argv) > 1 else DOC + 'gold.json'
gold = json.load(open(ANN))
IS_GOLD = gold.get('annotator') == 'gold'
lex = json.load(open(DOC + 'lexicon.json'))
items = json.load(open(DOC + 'items.json'))['items']
W = {s['id']: s for s in yaml.safe_load(open(REPO + 'docs/scheme-registry/walton-clean/walton.yml'))['argument_schemes']}
# Upstream IDs renamed in the cleaned file read as their new ones.
for old, new in json.load(open(REPO + 'docs/scheme-registry/walton-clean/id-migrations.json'))['schemes'].items():
    W[old] = W[new]
comp = json.load(open(REPO + 'docs/scheme-registry/links.json'))['entries']
NASCQ = REPO + 'wag/data/nascq/'
edges = json.load(open(NASCQ + 'edges.json'))['ask_edges']
cqs = json.load(open(NASCQ + 'critical_questions.json'))
WAG2W = {k: v['carneades_id'] for k, v in json.load(open(NASCQ + 'schemes.json')).items()}
errors, warns = [], []
def err(m): errors.append(m)

def split_args(s):
    out, d, cur = [], 0, ''
    for ch in s:
        if ch == ',' and d == 0:
            out.append(cur); cur = ''; continue
        d += ch == '('; d -= ch == ')'; cur += ch
    if cur: out.append(cur)
    return [a.strip() for a in out]

def parse(f):
    f = f.strip(); neg = f.startswith('¬'); f = f[1:] if neg else f
    m = re.match(r'^([a-z_][A-Za-z0-9_\-]*)\((.*)\)$', f)
    return (neg, m.group(1), split_args(m.group(2))) if m else (neg, None, [f])

def atoms(f):
    neg, p, args = parse(f)
    if p is None: return [args[0]]
    out = []
    for a in args:
        out += atoms(a) if ('(' in a or a.startswith('¬')) else [a]
    return out

def nows(f):
    """A formula without spaces: 'f(a, b)' and 'f(a,b)' are the same formula."""
    return re.sub(r'\s+', '', f or '')

def subst(f, b):
    return re.sub(r'\b[A-Z][A-Z0-9]*\b', lambda m: b.get(m.group(0), m.group(0)), f)

# 1. 50 items, in items.json order
ids = [i['kialo_claim_id'] for i in items]
gids = [i['kialo_claim_id'] for i in gold['items']]
if gids != ids: err('items not in items.json order or count differs: %d' % len(gids))
G = {i['kialo_claim_id']: i for i in gold['items']}
kp = {i['kialo_claim_id']: i['kialo_parents'] for i in items}

def primary(it):
    p = [x for x in it['instances'] if x['primary']]
    if len(p) != 1: err(f"{it['kialo_claim_id']}: {len(p)} primary instances")
    return p[0] if p else it['instances'][0]

# CQ tables: walton id -> {predicate: [(cq label or game id)]}
cq_pred = {}
for e in comp:
    for l in e.get('walton_links') or []:
        for m in l.get('cq_map') or []:
            for w in m['walton']:
                pred = parse(w['formula'])[1] or w['formula']
                cq_pred.setdefault(l['id'], {}).setdefault(pred, set()).add(('comp', f"{m.get('printed_under') or e['number']} {m['cq']}", w['locus']))
for e in edges:
    if 'pattern' in e:
        wid = WAG2W[e['scheme']]
        pred = parse(e['pattern'])[1]
        cq_pred.setdefault(wid, {}).setdefault(pred, set()).add(('game', e['cq'], e['locus']))
counts = {'instances': 0, 'checked_formulas': 0}
for it in gold['items']:
    cid = it['kialo_claim_id']
    for k in ('literal', 'instances', 'role', 'target', 'evidence', 'challenges', 'notes') + (('adjudication',) if IS_GOLD else ()):
        if k not in it: err(f'{cid}: missing {k}')
    if IS_GOLD:
        ad = it['adjudication']
        if not isinstance(ad.get('changed_fields'), list) or not ad.get('reason'): err(f'{cid}: bad adjudication')
        if ad['changed_fields'] == [] and ad['reason'] != 'agreed': err(f'{cid}: empty changes but reason != agreed')
    # 2. schemes and bindings; formulas follow walton.yml
    for ins in it['instances']:
        counts['instances'] += 1
        s = ins['scheme']; wid = s['walton_id']
        if s['fit'] not in ('explicit', 'reconstructed', 'fragment', 'none'): err(f'{cid}: fit {s["fit"]}')
        if wid in ('other', 'none'):
            if ins['bindings']: err(f'{cid}: {wid} with bindings')
            if wid == 'none' and s['fit'] != 'none': err(f'{cid}: none scheme with fit {s["fit"]}')
            continue
        if wid not in W: err(f'{cid}: unknown walton_id {wid}'); continue
        sch = W[wid]
        if set(ins['bindings']) != set(sch['variables']): err(f'{cid}: {wid} binds {sorted(ins["bindings"])} not {sch["variables"]}')
        exp = []
        for role, key in (('premise', 'premises'), ('assumption', 'assumptions'), ('exception', 'exceptions')):
            exp += [(role, nows(subst(f, ins['bindings']))) for f in sch.get(key) or []]
        got = [(p['role'], nows(p['formula'])) for p in ins['premises']]
        if sorted(exp) != sorted(got): err(f'{cid}: {wid} premises differ from walton.yml: {set(exp) ^ set(got)}')
        if nows(ins['conclusion'].get('formula')) != nows(subst(sch['conclusions'][0], ins['bindings'])): err(f'{cid}: {wid} conclusion differs')
        counts['checked_formulas'] += len(got) + 1
        # R4: hypothetical cause
        if wid == 'cause_to_effect':
            ho = [p for p in ins['premises'] if p['formula'].startswith('has_occurred(')][0]
            if ho['status'] == 'supplied':
                if s['fit'] != 'reconstructed' or s['confidence'] > 0.5 or 'hypothetical cause (needs a modal causal scheme, audit N8)' not in (ins.get('note') or ''):
                    err(f'{cid}: R4 not applied to a hypothetical cause_to_effect')
    # 3. target formula in the target's primary instance; R1; R2
    t = it['target']; a = t['actual']; r = it['role']
    parents = kp[cid]
    if t['kialo_parent'] != (parents[0]['claim_id'] if parents else None): err(f'{cid}: kialo_parent mismatch')
    if len(parents) > 1:
        others = {o.get('claim_id') for o in t.get('other_kialo_parents') or []}
        if others != {p['claim_id'] for p in parents[1:]}: err(f'{cid}: other_kialo_parents incomplete (R7)')
    if a:
        if t['re_parented'] != (a['claim'] != t['kialo_parent']): err(f'{cid}: R1 re_parented inconsistent')
        if a['claim'] not in G: err(f'{cid}: target claim {a["claim"]} not in gold')
        tgt_list = [(a['part'], a.get('formula'), r['kind'], r.get('cq'), a['claim'])]
        for o in t.get('other_kialo_parents') or []:
            if o.get('part') in ('premise', 'exception'):
                tgt_list.append((o['part'], o['formula'], o.get('role'), o.get('cq'), o['claim_id']))
        for part, f, kind, cq, tc in tgt_list:
            if part in ('premise', 'exception'):
                tp = primary(G[tc])
                roles = {'premise': ('premise', 'assumption'), 'exception': ('exception',)}[part]
                forms = [nows(p['formula']) for p in tp['premises'] if p['role'] in roles]
                if nows(f) not in forms: err(f'{cid}: target {part} {f} is not a {part} of {tc} primary instance'); continue
                twid = tp['scheme']['walton_id']
                pred = parse(f)[1]
                avail = cq_pred.get(twid, {}).get(pred, set())
                game = sorted(x[1] for x in avail if x[0] == 'game')
                compl = sorted(x[1] for x in avail if x[0] == 'comp')
                if avail and kind != 'cq_answer': err(f'{cid}: R2 target {f} of {twid} has CQs {game or compl} but kind={kind}')
                if not avail and kind == 'cq_answer': err(f'{cid}: R2 cq_answer but {twid}.{pred} has no CQ')
                if kind == 'cq_answer':
                    want = game if game else compl
                    if cq not in want: err(f'{cid}: R2 cq {cq} not in {want}')
            elif part not in ('conclusion', 'inference'):
                err(f'{cid}: bad target part {part}')
    if r['kind'] == 'cq_answer' and not r.get('cq'): err(f'{cid}: cq_answer without cq')
    if r['kind'] != 'cq_answer' and r.get('cq'): err(f'{cid}: cq set on non-cq_answer')
    if r['polarity'] not in ('for', 'against', None): err(f'{cid}: polarity {r["polarity"]}')
    # 4. atoms
    fs = []
    for ins in it['instances']:
        fs += [v for v in ins['bindings'].values() if v] + [p['formula'] for p in ins['premises']]
        if (ins.get('conclusion') or {}).get('formula'): fs.append(ins['conclusion']['formula'])  # none/other may have no formula
    if a and a.get('formula'): fs.append(a['formula'])
    fs += [o['formula'] for o in t.get('other_kialo_parents') or [] if o.get('formula')]
    fs += [c['formula'] for c in it['challenges'] if c.get('formula')]
    for f in fs:
        for x in atoms(f):
            if x not in lex: err(f'{cid}: atom {x} not in lexicon')
    for c in it['challenges']:
        if c.get('wag_cq') and c['wag_cq'] not in cqs: err(f'{cid}: unknown game CQ {c["wag_cq"]}')

print('items', len(gold['items']), '| instances', counts['instances'], '| walton formulas checked', counts['checked_formulas'], '| lexicon atoms', len(lex))
if errors:
    print('ERRORS (%d):' % len(errors)); print('\n'.join(errors)); sys.exit(1)
print('OK: order, schemes and bindings, premises/conclusions = walton.yml instantiations, R1, R2, R4, R7, target formulas in primary instances, atoms in lexicon')
