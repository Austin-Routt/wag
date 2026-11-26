"""Validate gold.json and lexicon.json (or another annotation file).

    python3 -I validate.py [ANNOTATIONS.json] [--case CASE.json] [--items ITEMS.json]

Everything it reads is in the repository: walton.yml
(docs/scheme-registry/walton-clean/), the compendium links
(docs/scheme-registry/links.json: which walton.yml formula each of the
book's critical questions is about; labels only, no book text), the
game's schemes.json, critical_questions.json and edges.json (wag/data/nascq/),
and, for format 2, the case record (case.json, beside this file unless
--case names another) and the typed places of the predicates
(wag/data/annotation/positions.json). Paths are relative to this file, so it
runs from any directory.

ANNOTATIONS defaults to gold.json. A file that is not the adjudicated gold
(an annotator's, or a model's from wag-llm annotate) is checked the same way,
except for the adjudication records and the pinned inputs.

A file is format 2 when it says so ("format": "wag-annotation/2";
docs/annotation-records.md, decision D092), else format 1: format 1 is
checked by the rules it was made under (R1-R11), format 2 by R2-R17."""
import hashlib, json, os, re, sys
import yaml

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
ARGS = sys.argv[1:]
REPIN = '--repin' in ARGS
if REPIN:
    ARGS.remove('--repin')
CASE = DOC + 'case.json'
if '--case' in ARGS:
    i = ARGS.index('--case')
    CASE = ARGS[i + 1]
    del ARGS[i:i + 2]
# Another items.json (its claims' Kialo parents) for the planted-defect tests,
# which need a claim with more parents than batch 2 has; pinned like the rest.
ITEMS = DOC + 'items.json'
if '--items' in ARGS:
    i = ARGS.index('--items')
    ITEMS = ARGS[i + 1]
    del ARGS[i:i + 2]
ANN = ARGS[0] if ARGS else DOC + 'gold.json'
gold = json.load(open(ANN, encoding='utf-8'))
IS_GOLD = gold.get('annotator') == 'gold'
FORMAT = gold.get('format')
if FORMAT not in (None, 'wag-annotation/1', 'wag-annotation/2'):
    print(f'ERRORS (1):\nthe file is in format {FORMAT!r}: this reads format 1 (no "format", or wag-annotation/1) and wag-annotation/2')
    sys.exit(1)
F2 = FORMAT == 'wag-annotation/2'
WALTON_PATH = REPO + 'docs/scheme-registry/walton-clean/walton.yml'
# A format-1 file is a kept version (D039: the gold before format 2 is kept as
# it was): it is checked against the walton.yml it was made with, which
# format1/ keeps, so that a later change to a scheme does not fail it.
if not F2 and os.path.exists(DOC + 'format1/walton.yml'):
    WALTON_PATH = DOC + 'format1/walton.yml'
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
items = json.load(open(ITEMS, encoding='utf-8'))['items']
W = {s['id']: s for s in yaml.safe_load(open(WALTON_PATH, encoding='utf-8'))['argument_schemes']}
# Upstream IDs renamed in the cleaned file read as their new ones.
for old, new in json.load(open(REPO + 'docs/scheme-registry/walton-clean/id-migrations.json'))['schemes'].items():
    W[old] = W[new]
comp = json.load(open(REPO + 'docs/scheme-registry/links.json'))['entries']
NASCQ = REPO + 'wag/data/nascq/'
edges = json.load(open(NASCQ + 'edges.json'))['ask_edges']
cqs = json.load(open(NASCQ + 'critical_questions.json'))
WAG2W = {k: v['carneades_id'] for k, v in json.load(open(NASCQ + 'schemes.json')).items()}
# The answer schemes (D075): each concludes an exception of another scheme.
ANSWER_SCHEMES = ('bad_outweighs_good', 'good_outweighs_bad', 'better_means')
errors, unresolved_values = [], []
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
    return re.sub(r'\b[A-Z][A-Za-z0-9_]*\b', lambda m: b[m.group(0)] if isinstance(b.get(m.group(0)), str) else m.group(0), f)

def marker(v, kind):
    """{"unresolved": why} or {"unstated": why} (docs/annotation-records.md §8)."""
    return isinstance(v, dict) and set(v) == {kind} and isinstance(v[kind], str) and v[kind].strip() != ''

def item_rev(it):
    """An item's revision: the SHA-256 of the item without its adjudication and
    verification, in canonical JSON (docs/annotation-records.md §9)."""
    x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
    return hashlib.sha256(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8')).hexdigest()

def sha256_file(p):
    return hashlib.sha256(open(p, 'rb').read()).hexdigest()

def valid_date(d):
    import datetime
    try:
        return isinstance(d, str) and re.match(r'^\d{4}-\d{2}-\d{2}$', d) and bool(datetime.date.fromisoformat(d))
    except ValueError:
        return False

# 1. 50 items, in items.json order
if not isinstance(gold.get('items'), list) or not all(isinstance(i, dict) for i in gold['items']):
    print('ERRORS (1):\nthe file\'s items are not a list of objects'); sys.exit(1)
ids = [i['kialo_claim_id'] for i in items]
gids = [i.get('kialo_claim_id') for i in gold['items']]
if gids != ids: err('items not in items.json order or count differs: %d' % len(gids))
G = {i.get('kialo_claim_id'): i for i in gold['items']}
kp = {i['kialo_claim_id']: i['kialo_parents'] for i in items}
# A claim's text revision (docs/annotation-records.md §9): the SHA-256 of its text in items.json.
TEXT_REV = {i['kialo_claim_id']: hashlib.sha256(i['text'].encode('utf-8')).hexdigest() for i in items}

def primary(it):
    insts = [x for x in it.get('instances') or [] if isinstance(x, dict)]
    p = [x for x in insts if x.get('primary') is True]
    if len(p) != 1: err(f"{it.get('kialo_claim_id')}: {len(p)} primary instances")
    return p[0] if p else (insts[0] if insts else None)

def formulas(ins, part):
    """An instance's formulas in a part of a target: premise (its premises and
    assumptions), exception, or conclusion."""
    if part == 'conclusion':
        f = (ins.get('conclusion') or {}).get('formula')
        return [nows(f)] if isinstance(f, str) and f else []
    roles = {'premise': ('premise', 'assumption'), 'exception': ('exception',)}.get(part, ())
    return [nows(p.get('formula')) for p in ins.get('premises') or []
            if isinstance(p, dict) and p.get('role') in roles and isinstance(p.get('formula'), str)]

def all_formulas(ins):
    return formulas(ins, 'premise') + formulas(ins, 'exception') + formulas(ins, 'conclusion')

# --- Format 2: the case record, the typed places, the instances by id -------
POS = {'agent': {}, 'action': {}}
ACTORS, ACTOR, CANON, CONTEXTS, DEFAULT_CTX, THESIS = set(), {}, {}, set(), None, {}
INST = {}  # instance id -> (claim, instance)
MALFORMED = set()  # format 2: items that break the field contract (R16)
ID_RE = re.compile(r'^(.+)#([1-9][0-9]*)$')
# Every input this check reads, which a gold pins (R16; docs/annotation-records.md §2).
PINNED = {'items.json': ITEMS, 'lexicon.json': DOC + 'lexicon.json', 'case.json': CASE,
          'walton.yml': WALTON_PATH, 'id-migrations.json': REPO + 'docs/scheme-registry/walton-clean/id-migrations.json',
          'links.json': REPO + 'docs/scheme-registry/links.json', 'positions.json': REPO + 'wag/data/annotation/positions.json',
          'nascq/schemes.json': REPO + 'wag/data/nascq/schemes.json',
          'nascq/critical_questions.json': REPO + 'wag/data/nascq/critical_questions.json',
          'nascq/edges.json': REPO + 'wag/data/nascq/edges.json'}
if F2:
    POS = json.load(open(REPO + 'wag/data/annotation/positions.json', encoding='utf-8'))
    case = None
    if not os.path.exists(CASE):
        err(f'case.json: R14 there is no case record ({CASE})')
    else:
        case = json.load(open(CASE, encoding='utf-8'))
        if case.get('format') != 'wag-case/1':
            err(f'case.json: R14 its format is {case.get("format")!r}, not wag-case/1')
        if gold.get('case') != case.get('id'):
            err(f'case.json: R16 the file names the case {gold.get("case")!r}, not {case.get("id")!r}')
        for c in case.get('contexts') or []:
            if not isinstance(c, dict) or not isinstance(c.get('id'), str) or not c['id'] or c['id'] in CONTEXTS:
                err(f'case.json: R15 a context with no id, or one already used: {c!r}'); continue
            CONTEXTS.add(c['id'])
        DEFAULT_CTX = case.get('default_context')
        if DEFAULT_CTX not in CONTEXTS:
            err(f'case.json: R15 its default context {DEFAULT_CTX!r} is not one of its contexts')
        for a in case.get('actors') or []:
            aid = a.get('id') if isinstance(a, dict) else None
            if not isinstance(aid, str) or aid in ACTORS:
                err(f'case.json: R14 an actor with no id, or one already used: {a!r}'); continue
            if aid not in lex: err(f'case.json: R14 the actor {aid} is not an atom of the lexicon')
            ACTORS.add(aid)
        for a in case.get('actions') or []:
            aid = a.get('id') if isinstance(a, dict) else None
            if not isinstance(aid, str) or aid in CANON:
                err(f'case.json: R14 an action with no id, or an id already used: {a!r}'); continue
            if aid in ACTORS:
                err(f'case.json: R14 {aid!r} is an actor, so it cannot be an action'); continue
            if aid not in lex: err(f'case.json: R14 the action {aid} is not an atom of the lexicon')
            if a.get('actor') not in ACTORS: err(f'case.json: R14 the action {aid} is performed by {a.get("actor")!r}, not an actor of the case')
            CANON[aid], ACTOR[aid] = aid, a.get('actor')
        for a in case.get('actions') or []:
            if not isinstance(a, dict) or not isinstance(a.get('aliases'), list):
                err(f'case.json: R14 an action with no list of aliases: {a!r}'); continue
            for al in a['aliases']:
                if not isinstance(al, str) or al in CANON:
                    err(f'case.json: R14 {al!r} is an action or an alias already, so it cannot be an alias of {a.get("id")}'); continue
                if al in ACTORS:
                    err(f'case.json: R14 {al!r} is an actor, so it cannot be an alias of {a.get("id")}'); continue
                if al not in lex: err(f'case.json: R14 the alias {al} is not an atom of the lexicon')
                CANON[al] = a['id']
        THESIS = case.get('thesis') or {}
        if THESIS.get('action') not in ACTOR:
            err(f'case.json: R14 the thesis\'s action {THESIS.get("action")!r} is not an action of the case')
        if THESIS.get('claim') not in G:
            err(f'case.json: R14 the thesis {THESIS.get("claim")!r} is not a claim of the file')
    for it in gold['items']:
        for ins in it.get('instances') if isinstance(it.get('instances'), list) else []:
            if isinstance(ins, dict) and isinstance(ins.get('id'), str) and ins['id'] not in INST:
                INST[ins['id']] = (it.get('kialo_claim_id'), ins)
    if IS_GOLD:
        pins = gold.get('inputs')
        if not isinstance(pins, dict) or set(pins) != set(PINNED):
            err(f'inputs: R16 a gold pins exactly {sorted(PINNED)}, not {sorted(pins) if isinstance(pins, dict) else pins!r}')
        else:
            for name, p in PINNED.items():
                if os.path.exists(p) and pins[name] != sha256_file(p):
                    err(f'inputs: R16 the gold was checked against another {name} (pinned {pins[name][:12]}…, it is '
                        f'{sha256_file(p)[:12]}…): check it again, and pin the new one (validate.py --repin)')

def target_instance(claim, ref):
    """The instance a target names: in format 1, its claim's primary instance;
    in format 2, the instance with that id, if it is one of that claim's
    (check_target_entry does not look in a malformed item)."""
    if claim not in G:
        return None
    if not F2:
        return primary(G[claim])
    x = INST.get(ref) if isinstance(ref, str) else None
    return x[1] if x and x[0] == claim else None

def typed(sch):
    """A scheme's variables in an agent's or an action's place (positions.json)."""
    out = {}
    for key in ('premises', 'assumptions', 'exceptions', 'conclusions'):
        for f in sch.get(key) or []:
            neg, p, args = parse(f)
            for kind in ('agent', 'action'):
                for i in POS[kind].get(f'{p}/{len(args)}', []):
                    if i - 1 < len(args):
                        out.setdefault(args[i - 1], kind)
    return out

def canon(a):
    return CANON.get(a, a)

def context_of(ins):
    return ins.get('context', DEFAULT_CTX)

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

def check_target_entry(cid, where, claim, ref, part, formula):
    """R13 (and R16's shape) for a format-2 target: its instance, part and
    formula. Returns the target instance, or None when it cannot be read."""
    for name, v in (('instance', ref), ('part', part), ('formula', formula)):
        if marker(v, 'unresolved'):
            unresolved_values.append(f'{cid}: {where} {name} ({v["unresolved"]})')
    if marker(ref, 'unresolved'):
        return None
    if not isinstance(ref, str):
        err(f'{cid}: R16 {where} names no instance (an instance id, or {{"unresolved": why}})'); return None
    if claim in MALFORMED:
        return None  # its own errors are said; its instances are not read
    tp = target_instance(claim, ref)
    if tp is None:
        err(f'{cid}: R13 {where} names the instance {ref}, which is not an instance of {claim}'); return None
    if marker(part, 'unresolved') or marker(formula, 'unresolved'):
        return tp
    if part == 'inference':
        if formula is not None: err(f'{cid}: R16 {where} is the inference, which has no formula, but names {formula}')
    elif part in ('premise', 'exception', 'conclusion'):
        if not isinstance(formula, str) or not formula:
            err(f'{cid}: R16 {where} names no formula of its {part}')
        elif nows(formula) not in formulas(tp, part):
            err(f'{cid}: R13 {where} names {formula}, which is not a {part} of {ref}')
    else:
        err(f'{cid}: bad target part {part}')
    return tp


ROLES = ('independent_pro', 'independent_con', 'rebuttal', 'undercutter', 'premise_attack', 'premise_backing',
         'cq_answer', 'side_effect', 'alternative_means', 'action_amendment', 'clarification', 'none')
FITS = ('explicit', 'reconstructed', 'fragment', 'none')
PARTS = ('premise', 'exception', 'conclusion', 'inference')
EVIDENCE = ('source', 'definition', 'example', 'assertion')

def text(v):
    return isinstance(v, str) and v.strip() != ''

def target_shape(a_, claim_key='claim'):
    """R16 for one exact target: its claim, instance, part and formula
    (docs/annotation-records.md §6)."""
    out = []
    if not text(a_.get(claim_key)): out.append(f'.{claim_key} is not a claim')
    ref, part, f = a_.get('instance'), a_.get('part'), a_.get('formula')
    if not (text(ref) or marker(ref, 'unresolved')): out.append('.instance is not an instance id, or {"unresolved": why}')
    if not (part in PARTS or marker(part, 'unresolved')): out.append(f'.part is {part!r}: premise, exception, conclusion, inference, or {{"unresolved": why}}')
    if 'formula' not in a_:
        out.append('.formula is missing (null for an inference, §8)')
    elif part == 'inference':
        if f is not None: out.append('.formula is not null: an inference target has no formula')
    elif not (text(f) or marker(f, 'unresolved')):
        out.append('.formula is not a formula, or {"unresolved": why}')
    return out

def shape_errors(it):
    """R16 (format 2): the field contract of docs/annotation-records.md §4-§8,
    before any rule reads the record: every required field, each of its type;
    null only where §8 lists it; a marker only where it is allowed. A record
    with any of these is reported and not checked further, so that a
    malformed record is said, not a traceback."""
    out = []
    def bad(path, what):
        out.append(f'R16 {path} {what}')
    def num01(v):
        return isinstance(v, (int, float)) and not isinstance(v, bool) and 0 <= v <= 1
    cid = it.get('kialo_claim_id')
    parents = kp.get(cid) or []
    thesis = not parents
    lit = it.get('literal')
    if not isinstance(lit, list) or not lit or not all(text(x) for x in lit):
        bad('literal', "is not a list of the claim's propositions")
    if 'retired' in it:
        ok = isinstance(it['retired'], list) and all(isinstance(x, str) and ID_RE.match(x) and ID_RE.match(x).group(1) == cid
                                                       for x in it['retired'])
        if not ok: bad('retired', f"is not a list of ids of {cid}'s instances")
    insts = it.get('instances')
    if not isinstance(insts, list) or not insts or not all(isinstance(x, dict) for x in insts):
        bad('instances', 'is not a list of instances (at least one)')
        insts = []
    for n, ins in enumerate(insts):
        where = ins['id'] if text(ins.get('id')) else f'instances[{n}]'
        if not text(ins.get('id')): bad(f'{where}.id', 'is not an instance id')
        if not isinstance(ins.get('primary'), bool): bad(f'{where}.primary', 'is not true or false')
        if 'context' in ins and not text(ins['context']): bad(f'{where}.context', 'is not a context id')
        sc = ins.get('scheme')
        if not isinstance(sc, dict):
            bad(f'{where}.scheme', 'is not an object')
            continue
        if not text(sc.get('walton_id')): bad(f'{where}.scheme.walton_id', 'is not a scheme id, other or none')
        for k in ('compendium', 'wag_id'):
            if k not in sc or not (sc[k] is None or text(sc[k])): bad(f'{where}.scheme.{k}', 'is not text, or null when there is none')
        if sc.get('fit') not in FITS: bad(f'{where}.scheme.fit', f'is {sc.get("fit")!r}: explicit, reconstructed, fragment or none')
        if not num01(sc.get('confidence')): bad(f'{where}.scheme.confidence', 'is not a number from 0 to 1')
        if not isinstance(sc.get('alternatives'), list): bad(f'{where}.scheme.alternatives', 'is not a list')
        bm = ins.get('bindings')
        if not isinstance(bm, dict):
            bad(f'{where}.bindings', 'is not an object')
        else:
            for v, x in bm.items():
                if not (text(x) or (marker(x, 'unstated') and sc.get('fit') == 'fragment')):
                    bad(f'{where}.bindings.{v}', f'is {x!r}: an atom, or in a fragment {{"unstated": why}}')
        ps = ins.get('premises')
        if not isinstance(ps, list) or not all(isinstance(x, dict) for x in ps):
            bad(f'{where}.premises', 'is not a list of premises')
        else:
            for k, pm in enumerate(ps):
                if pm.get('role') not in ('premise', 'assumption', 'exception'): bad(f'{where}.premises[{k}].role', f'is {pm.get("role")!r}')
                if not text(pm.get('formula')): bad(f'{where}.premises[{k}].formula', 'is not a formula')
                if not isinstance(pm.get('en'), str): bad(f'{where}.premises[{k}].en', 'is not text')
                if pm.get('status') not in ('explicit', 'inferred', 'supplied'):
                    bad(f'{where}.premises[{k}].status', f'is {pm.get("status")!r}: explicit, inferred or supplied')
        cm = ins.get('conclusion')
        if not isinstance(cm, dict) or 'formula' not in cm or not isinstance(cm.get('en'), str):
            bad(f'{where}.conclusion', 'is not an object with a formula and its English')
        elif not (text(cm['formula']) or cm['formula'] is None):
            bad(f'{where}.conclusion.formula', 'is not a formula, or null')
        elif cm['formula'] is None and sc.get('walton_id') not in ('other', 'none') and sc.get('fit') != 'fragment':
            bad(f'{where}.conclusion.formula', 'is null, which only other, none and a fragment whose conclusion is unstated may be')
        for k in ('label', 'note', 'comparison'):
            if k in ins and not isinstance(ins[k], str): bad(f'{where}.{k}', 'is not text')
    r_ = it.get('role')
    if not isinstance(r_, dict):
        bad('role', 'is not an object')
    else:
        kind = r_.get('kind')
        if kind not in ROLES: bad('role.kind', f"is {kind!r}, not one of the guidelines' roles")
        if 'move' in r_ and r_['move'] not in ('side_effect', 'alternative_means'): bad('role.move', f'is {r_["move"]!r}: side_effect or alternative_means')
        if 'cq' not in r_: bad('role.cq', 'is missing (null when the role is not cq_answer)')
        elif kind == 'cq_answer' and not text(r_['cq']): bad('role.cq', 'names no question, but the role is cq_answer')
        elif kind != 'cq_answer' and r_['cq'] is not None: bad('role.cq', f'is {r_["cq"]!r}: null, as the role is not cq_answer')
        if 'polarity' not in r_: bad('role.polarity', 'is missing')
        elif thesis and r_['polarity'] is not None: bad('role.polarity', 'is not null: the thesis has no polarity')
        elif not thesis and r_['polarity'] not in ('for', 'against'): bad('role.polarity', f'is {r_["polarity"]!r}: for or against, null only for the thesis')
        if not num01(r_.get('confidence')): bad('role.confidence', 'is not a number from 0 to 1')
    t_ = it.get('target')
    if not isinstance(t_, dict):
        bad('target', 'is not an object')
    else:
        if 're_parented' in t_: bad('target.re_parented', 'is recorded: it is derived from the target, not recorded')
        for k in ('kialo_parent', 'kialo_relation', 'actual'):
            if k not in t_: bad(f'target.{k}', 'is missing (null only for the thesis)')
        if thesis:
            if any(t_.get(k) is not None for k in ('kialo_parent', 'kialo_relation', 'actual')):
                bad('target', 'is not null: the thesis has no target (its kialo_parent, kialo_relation and actual are null)')
        else:
            if t_.get('kialo_parent') != parents[0].get('claim_id'):
                bad('target.kialo_parent', f'is {t_.get("kialo_parent")!r}, not {parents[0].get("claim_id")!r}')
            if t_.get('kialo_relation') != parents[0].get('relation'):
                bad('target.kialo_relation', f'is {t_.get("kialo_relation")!r}, not {parents[0].get("relation")!r}')
            a_ = t_.get('actual')
            if not isinstance(a_, dict) or not a_:
                bad('target.actual', f'is {a_!r}: only the thesis has no actual target')
            else:
                out += [f'R16 target.actual{e}' for e in target_shape(a_)]
        others = t_.get('other_kialo_parents')
        if len(parents) > 1 and not isinstance(others, list):
            bad('target.other_kialo_parents', 'is missing: the claim has more than one Kialo parent (R7)')
        if len(parents) <= 1 and 'other_kialo_parents' in t_:
            bad('target.other_kialo_parents', 'is recorded, but the claim has one Kialo parent or none (R7)')
        for k, o in enumerate(others if isinstance(others, list) else []):
            if not isinstance(o, dict) or not text(o.get('claim_id')) or o.get('relation') not in ('pro', 'con'):
                bad(f'target.other_kialo_parents[{k}]', 'needs claim_id and relation (pro or con)')
            elif 'part' in o:
                out += [f'R16 target.other_kialo_parents[{k}]{e}' for e in target_shape(o, claim_key='claim_id')]
    ev = it.get('evidence')
    if not (isinstance(ev, dict) and ev.get('status') in EVIDENCE and isinstance(ev.get('urls'), list)
            and all(isinstance(u, str) for u in ev['urls'])):
        bad('evidence', 'needs status (source, definition, example or assertion) and a list of urls (empty when it cites none)')
    ch = it.get('challenges')
    if not isinstance(ch, list) or not all(isinstance(x, dict) for x in ch):
        bad('challenges', 'is not a list of challenges')
    else:
        for k, c in enumerate(ch):
            for key in ('cq', 'wag_cq'):
                if key not in c or not (c[key] is None or text(c[key])):
                    bad(f'challenges[{k}].{key}', "is not a question's label, or null when there is none")
            if not text(c.get('type')): bad(f'challenges[{k}].type', 'is not a type')
            if not text(c.get('instance')): bad(f'challenges[{k}].instance', 'names no instance')
            if 'formula' not in c or not (c['formula'] is None or text(c['formula'])):
                bad(f'challenges[{k}].formula', 'is not a formula, or null for the whole instance')
    no = it.get('notes')
    if not (isinstance(no, dict) and isinstance(no.get('text'), str) and 'candidate_pattern' in no
            and (no['candidate_pattern'] is None or text(no['candidate_pattern']))):
        bad('notes', 'needs text and candidate_pattern (null when it names none)')
    return out

counts = {'instances': 0, 'checked_formulas': 0}
REQUIRED = ('literal', 'instances', 'role', 'target', 'evidence', 'challenges', 'notes')
# The field contract first, for every item: a malformed item is reported once,
# and no other item's rule reads it as a target.
SHAPE = {it.get('kialo_claim_id'): shape_errors(it) for it in gold['items']} if F2 else {}
MALFORMED = {cid for cid, e in SHAPE.items() if e}
for it in gold['items']:
    cid = it.get('kialo_claim_id')
    unstated_conclusions = set()  # instances whose conclusion R16 refused (not checked for atoms again)
    if cid not in kp:
        err(f'{cid}: not a claim of items.json'); continue
    if F2 and SHAPE.get(cid):
        for e in SHAPE[cid]:
            err(f'{cid}: {e}')
        continue
    for k in REQUIRED + (('adjudication',) if IS_GOLD else ()):
        if k not in it: err(f'{cid}: missing {k}')
    if IS_GOLD:
        ad = it.get('adjudication')
        if not isinstance(ad, dict) or not isinstance(ad.get('changed_fields'), list) or not ad.get('reason'):
            err(f'{cid}: bad adjudication')
        elif ad['changed_fields'] == [] and ad['reason'] != 'agreed':
            err(f'{cid}: empty changes but reason != agreed')
    thesis = not kp.get(cid)
    # R12 (format 2): every instance has an id of its own claim, used once and
    # never retired.
    if F2:
        retired = it.get('retired', [])  # its shape is the schema pass's
        # An annotator's record (one with no adjudication: not a gold's,
        # whose ids stay as its history made them) lists its primary
        # instance first, as <claim>#1, so that a target can name another
        # claim's primary instance before it is shown (§5).
        first = it['instances'][0]
        if 'adjudication' not in it and not (first.get('primary') is True and first.get('id') == f'{cid}#1'):
            err(f'{cid}: R12 its primary instance is not listed first as {cid}#1, which a new record numbers it')
        seen = set()
        for ins in it['instances']:
            iid = ins.get('id')
            m = ID_RE.match(iid) if isinstance(iid, str) else None
            if not m or m.group(1) != cid:
                err(f'{cid}: R12 an instance has the id {iid!r}, not {cid}#<n>')
            elif iid in seen:
                err(f'{cid}: R12 two instances have the id {iid}')
            elif iid in retired:
                err(f'{cid}: R12 the instance {iid} has a retired id')
            seen.add(iid)
            if not isinstance(ins.get('primary'), bool): err(f'{cid}: R16 the instance {iid} says nothing of primary (true or false)')
            if 'context' in ins and ins['context'] not in CONTEXTS:
                err(f'{cid}: R15 the instance {iid} is in the context {ins["context"]!r}, which is not one of the case\'s')
    # 2. schemes and bindings; formulas follow walton.yml
    for ins in it['instances']:
        counts['instances'] += 1
        s = ins['scheme']; wid = s['walton_id']
        iid = ins.get('id', '?')
        if s['fit'] not in ('explicit', 'reconstructed', 'fragment', 'none'): err(f'{cid}: fit {s["fit"]}')
        bad_binding = False
        if F2:
            for k in ('compendium', 'wag_id', 'confidence', 'alternatives'):
                if k not in s: err(f'{cid}: R16 the scheme of {iid} says nothing of {k} (null when there is none)')
            for v, b in ins['bindings'].items():
                if isinstance(b, str) and b: continue
                if marker(b, 'unstated') and s['fit'] == 'fragment': continue
                err(f'{cid}: R16 {iid} binds {v} to {b!r}: an atom, or in a fragment {{"unstated": why}}')
                bad_binding = True
        if wid in ('other', 'none'):
            if ins['bindings']: err(f'{cid}: {wid} with bindings')
            if wid == 'none' and s['fit'] != 'none': err(f'{cid}: none scheme with fit {s["fit"]}')
            continue
        if wid not in W: err(f'{cid}: unknown walton_id {wid}'); continue
        if bad_binding: continue  # its formulas cannot be compared: said once, above
        sch = W[wid]
        if set(ins['bindings']) != set(sch['variables']): err(f'{cid}: {wid} binds {sorted(ins["bindings"])} not {sch["variables"]}')
        unstated = {v for v, b in ins['bindings'].items() if not isinstance(b, str)}
        exp = []
        for role, key in (('premise', 'premises'), ('assumption', 'assumptions'), ('exception', 'exceptions')):
            exp += [(role, nows(subst(f, ins['bindings']))) for f in sch.get(key) or []
                    if not (set(re.findall(r'\b[A-Z][A-Za-z0-9_]*\b', f)) & unstated)]
        got = [(p['role'], nows(p['formula'])) for p in ins['premises']]
        if unstated:
            # A fragment states some of its scheme's formulas, and leaves the
            # variables of the others unstated (docs/annotation-records.md §5).
            extra = [x for x in got if x not in exp]
            if extra: err(f'{cid}: {wid} premises differ from walton.yml: {set(extra)}')
        elif sorted(exp) != sorted(got): err(f'{cid}: {wid} premises differ from walton.yml: {set(exp) ^ set(got)}')
        concl = (ins.get('conclusion') or {}).get('formula')
        open_concl = bool(set(re.findall(r'\b[A-Z][A-Za-z0-9_]*\b', sch['conclusions'][0])) & unstated)
        if F2 and open_concl and concl is not None:
            err(f'{cid}: R16 {iid} is a fragment whose conclusion names a variable it leaves unstated: its formula is null')
            unstated_conclusions.add(id(ins))
        elif not (open_concl and concl is None) and nows(concl) != nows(subst(sch['conclusions'][0], ins['bindings'])):
            err(f'{cid}: {wid} conclusion differs')
        counts['checked_formulas'] += len(got) + 1
        # R4: hypothetical cause. Format 1: under the N8 note; format 2: a
        # cause that has not occurred is cause_to_effect_modal's (D039).
        if wid == 'cause_to_effect':
            occurrences = [p for p in ins['premises'] if p['formula'].startswith('has_occurred(')]
            if not occurrences:
                err(f'{cid}: cause_to_effect lacks has_occurred premise'); continue
            ho = occurrences[0]
            if F2:
                # Format 2 (D039): by the record, not by a premise's status. A
                # cause that is an action the case deliberates has not
                # occurred, nor has one that a modal step of the same claim
                # concludes may occur: both are cause_to_effect_modal's.
                e1 = ins['bindings'].get('E1')
                modal = {x['bindings'].get('E2') for x in it['instances']
                         if x['scheme']['walton_id'] == 'cause_to_effect_modal' and isinstance(x['bindings'].get('E2'), str)}
                if isinstance(e1, str) and canon(e1) in ACTOR:
                    err(f'{cid}: R4 {iid} has the cause {e1}, an action the case deliberates, which has not occurred: cause_to_effect_modal (D039)')
                elif isinstance(e1, str) and e1 in modal:
                    err(f'{cid}: R4 {iid} has the cause {e1}, which a modal step of the claim says only may occur: cause_to_effect_modal (D039)')
            elif ho['status'] == 'supplied':
                if s['fit'] != 'reconstructed' or s['confidence'] > 0.5 or 'hypothetical cause (needs a modal causal scheme, audit N8)' not in (ins.get('note') or ''):
                    err(f'{cid}: R4 not applied to a hypothetical cause_to_effect')
        # R14 (format 2): an agent's place holds an actor of the case, an
        # action's place an action (or an alias of one), and an instance with
        # an agent binds only its agent's actions (the case's mapping, not
        # the action's name: D063, D091).
        if F2:
            t = typed(sch)
            agents = {ins['bindings'][v] for v, k in t.items() if k == 'agent' and isinstance(ins['bindings'].get(v), str)}
            for v, k in sorted(t.items()):
                b = ins['bindings'].get(v)
                if not isinstance(b, str): continue
                if k == 'agent' and b not in ACTORS:
                    err(f'{cid}: R14 {iid} binds the agent {v} to {b}, which is not an actor of the case')
                elif k == 'action':
                    if b not in CANON:
                        err(f'{cid}: R14 {iid} binds the action {v} to {b}, which is not an action of the case')
                    elif agents and ACTOR[canon(b)] not in agents:
                        err(f'{cid}: R14 {iid} binds {v} to {b}, which is {ACTOR[canon(b)]}\'s action, not {" or ".join(sorted(agents))}\'s')
    # R16's field contract (format 2) is the schema pass's (shape_errors):
    # a malformed item does not reach this point.
    t = it['target']; a = t.get('actual'); r = it['role']
    # 3. target formula in the target instance; R1; R2
    parents = kp[cid]
    if t['kialo_parent'] != (parents[0]['claim_id'] if parents else None): err(f'{cid}: kialo_parent mismatch')
    if len(parents) > 1:
        got = [o for o in t.get('other_kialo_parents') or [] if isinstance(o, dict)]
        if {o.get('claim_id') for o in got} != {p['claim_id'] for p in parents[1:]}:
            err(f'{cid}: other_kialo_parents incomplete (R7)')
        else:
            # Each other parent with its relation, as items.json has them, each
            # as often as it has it: the source's pro or con, not the
            # interpreted polarity (role.polarity), which may differ.
            pairs = lambda xs: sorted((str(x.get('claim_id')), str(x.get('relation'))) for x in xs)
            if pairs(got) != pairs(parents[1:]):
                err(f'{cid}: R7 other_kialo_parents record ' + ', '.join(' '.join(x) for x in pairs(got)) +
                    '; items.json has ' + ', '.join(' '.join(x) for x in pairs(parents[1:])))
    tp = None
    if a:
        if not F2 and t['re_parented'] != (a['claim'] != t['kialo_parent']): err(f'{cid}: R1 re_parented inconsistent')
        if a['claim'] not in G: err(f'{cid}: target claim {a["claim"]} not in gold')
        if F2:
            tp = check_target_entry(cid, 'its target', a['claim'], a.get('instance'), a.get('part'), a.get('formula'))
        else:
            tp = target_instance(a['claim'], None)
        tgt_list = [(a['part'], a.get('formula'), r['kind'], r.get('cq'), a['claim'], tp)]
        for o in t.get('other_kialo_parents') or []:
            if F2:
                # Any entry that names a part is an exact target (its shape is
                # the schema pass's), the inference included.
                if 'part' in o:
                    otp = check_target_entry(cid, f'its target in {o["claim_id"]}', o['claim_id'], o.get('instance'), o['part'], o.get('formula'))
                    if o['part'] in ('premise', 'exception', 'conclusion'):
                        tgt_list.append((o['part'], o['formula'], o.get('role'), o.get('cq'), o['claim_id'], otp))
            elif o.get('part') in ('premise', 'exception'):
                if not o.get('claim_id') or not o.get('formula'):
                    err(f'{cid}: other_kialo_parents target requires claim_id and formula'); continue
                tgt_list.append((o['part'], o['formula'], o.get('role'), o.get('cq'), o['claim_id'], target_instance(o['claim_id'], None)))
        for part, f, kind, cq, tc, ins_t in tgt_list:
            if part in ('premise', 'exception'):
                if tc not in G:
                    err(f'{cid}: target claim {tc} absent'); continue
                if ins_t is None:
                    if not F2: err(f'{cid}: target {tc} has no argument instance')
                    continue
                if marker(f, 'unresolved'): continue
                if not F2:
                    if nows(f) not in formulas(ins_t, part): err(f'{cid}: target {part} {f} is not a {part} of {tc} primary instance'); continue
                elif nows(f) not in formulas(ins_t, part): continue  # R13 said so
                twid = ins_t['scheme']['walton_id']
                pred = parse(f)[1]
                avail = cq_pred.get(twid, {}).get(pred, set())
                game = sorted(x[1] for x in avail if x[0] == 'game')
                compl = sorted(x[1] for x in avail if x[0] == 'comp')
                if avail and kind != 'cq_answer': err(f'{cid}: R2 target {f} of {twid} has CQs {game or compl} but kind={kind}')
                if not avail and kind == 'cq_answer': err(f'{cid}: R2 cq_answer but {twid}.{pred} has no CQ')
                if kind == 'cq_answer':
                    want = game if game else compl
                    if cq not in want: err(f'{cid}: R2 cq {cq} not in {want}')
            elif part not in ('conclusion', 'inference') and not marker(part, 'unresolved'):
                if not F2: err(f'{cid}: bad target part {part}')
    # R8 (D075): a claim at an exception (an answer, R2) is an argument that
    # the exception holds, so its primary instance concludes the exception;
    # with polarity for, a defence of the argument, it concludes the
    # exception's negation (as to_wag.py plays it).
    own = primary(it)
    if a and a.get('part') == 'exception' and isinstance(a.get('formula'), str):
        got = nows(((own or {}).get('conclusion') or {}).get('formula'))
        want = ('¬' if r.get('polarity') == 'for' else '') + nows(a.get('formula'))
        if got != want: err(f'{cid}: R8 an answer at the exception {a.get("formula")} concludes {got or "nothing"}, not {want} (D075)')
    # R15 (format 2): the claim's own argument is in its target instance's
    # context, so that an answer for another situation does not answer it.
    if F2 and own is not None and tp is not None and context_of(own) != context_of(tp):
        err(f'{cid}: R15 its primary instance is in the context {context_of(own)!r}, its target {a.get("instance")} in {context_of(tp)!r}')
    # R9 (D075, D089): a better means is another action than the one it
    # answers (in format 2, as the case maps them: an alias of A is A),
    # records which comparison it makes, and, as the claim's primary
    # instance, has its target's goal premise, whatever the claim's role. A
    # fragment, which states one premise of its parent's argument (§2, §3),
    # is that argument's, and is checked there.
    for ins in it['instances']:
        if ins['scheme']['walton_id'] != 'better_means' or ins['scheme'].get('fit') == 'fragment': continue
        b = ins['bindings']
        if canon(b.get('A2')) == canon(b.get('A')):
            alias = f' ({b.get("A2")} is its alias)' if b.get('A2') != b.get('A') else ''
            err(f'{cid}: R9 better_means compares {canon(b.get("A"))} with itself{alias}')
        if ins.get('comparison') not in ('more surely', 'sooner', 'at lower cost'):
            err(f'{cid}: R9 better_means records no comparison (more surely, sooner or at lower cost)')
        if ins.get('primary') and tp is not None:
            goal = lambda x: sorted(nows(p['formula']) for p in (x or {}).get('premises', []) if parse(p['formula'])[1] == 'goal_of_agent')
            want, got = goal(tp), goal(ins)
            if got != want: err(f'{cid}: R9 better_means goal premise {got} is not the answered argument\'s {want}')
    # R11 (D075): an answer scheme argues the exception it concludes, so a
    # claim whose primary instance is one is at an exception (R8 then makes
    # it the one it concludes), unless it is a fragment that backs a premise
    # of its parent's argument (§2, §3).
    fragment_backing = own and own['scheme'].get('fit') == 'fragment' and a and a.get('part') == 'premise'
    if own and own['scheme']['walton_id'] in ANSWER_SCHEMES and not (a and a.get('part') == 'exception') and not fragment_backing:
        err(f'{cid}: R11 its primary instance is the answer scheme {own["scheme"]["walton_id"]}, but it is not at an exception')
    # R10: a challenge record names a formula of the claim's own instances or
    # of its target's instance, so that a formula renamed in walton.yml
    # cannot stay behind in a record. In format 2 it names the instance, one
    # of its own or its target's, and a formula of that instance.
    if F2:
        mine = {ins.get('id'): ins for ins in it['instances']}
        for c in it['challenges']:
            ref = c.get('instance')
            ci = mine.get(ref) if isinstance(ref, str) else None
            if ci is None and a and ref == a.get('instance'):
                if a.get('claim') in MALFORMED:
                    continue  # the target claim's own errors are said; its instances are not read
                ci = tp
            if ci is None:
                err(f'{cid}: R10 the challenge {c.get("cq") or c.get("type")} names the instance {ref!r}, neither one of its own nor its target\'s'); continue
            if c['formula'] is not None and nows(c['formula']) not in all_formulas(ci):
                err(f'{cid}: R10 the challenge {c.get("cq") or c.get("type")} names {c["formula"]}, which is not a formula of {ref}')
    else:
        names = set()
        for ins in it['instances']:
            names |= {nows(p['formula']) for p in ins['premises']}
            if (ins.get('conclusion') or {}).get('formula'): names.add(nows(ins['conclusion']['formula']))
        if tp:
            names |= {nows(p['formula']) for p in tp['premises']}
        for c in it['challenges']:
            if c.get('formula') and nows(c['formula']) not in names:
                err(f'{cid}: R10 the challenge {c.get("cq") or c.get("type")} names {c["formula"]}, a formula of neither its own instances nor its target\'s primary instance')
    if r['kind'] == 'cq_answer' and not r.get('cq'): err(f'{cid}: cq_answer without cq')
    if r['kind'] != 'cq_answer' and r.get('cq'): err(f'{cid}: cq set on non-cq_answer')
    if r['polarity'] not in ('for', 'against', None): err(f'{cid}: polarity {r["polarity"]}')
    # R17 (format 2): a verified item is the revision the person verified,
    # with nothing left unresolved.
    if F2 and 'verified' in it:
        v = it['verified']
        if not (isinstance(v, dict) and v.get('by') == 'owner' and valid_date(v.get('date'))
                and isinstance(v.get('item_rev'), str) and isinstance(v.get('text_sha256'), str)):
            err(f'{cid}: R17 its verification needs by "owner" (D086), a date (YYYY-MM-DD), the item_rev and the text_sha256 verified')
        elif v['item_rev'] != item_rev(it):
            err(f'{cid}: R17 it was verified at revision {v["item_rev"][:12]}…, but it is {item_rev(it)[:12]}… now: verify it again')
        elif v['text_sha256'] != TEXT_REV.get(cid):
            err(f"{cid}: R17 it was verified on another text of the claim (items.json's text has changed): verify it again")
        if any(u.startswith(cid + ':') for u in unresolved_values):
            err(f'{cid}: R17 it is verified, but some of its values are unresolved')
    # 4. atoms
    fs = []
    for ins in it['instances']:
        fs += [v for v in ins['bindings'].values() if isinstance(v, str) and v] + [p['formula'] for p in ins['premises']]
        if (ins.get('conclusion') or {}).get('formula') and id(ins) not in unstated_conclusions:
            fs.append(ins['conclusion']['formula'])  # none/other may have no formula
    if a and isinstance(a.get('formula'), str) and a.get('formula'): fs.append(a['formula'])
    fs += [o['formula'] for o in t.get('other_kialo_parents') or [] if isinstance(o.get('formula'), str) and o.get('formula')]
    fs += [c['formula'] for c in it['challenges'] if c.get('formula')]
    for f in fs:
        for x in atoms(f):
            if x not in lex: err(f'{cid}: atom {x} not in lexicon')
    for c in it['challenges']:
        if c.get('wag_cq') and c['wag_cq'] not in cqs: err(f'{cid}: unknown game CQ {c["wag_cq"]}')
# R14 (format 2): the thesis concludes that the case's thesis action be done.
if F2 and THESIS.get('claim') in G and THESIS.get('claim') not in MALFORMED and THESIS.get('action') in ACTOR:
    tconc = nows(((primary(G[THESIS['claim']]) or {}).get('conclusion') or {}).get('formula'))
    if tconc != f'should_be_performed({THESIS["action"]})':
        err(f'{THESIS["claim"]}: R14 the thesis concludes {tconc}, not should_be_performed({THESIS["action"]}), the case\'s thesis action')

print('items', len(gold['items']), '| instances', counts['instances'], '| walton formulas checked', counts['checked_formulas'],
      '| lexicon atoms', len(lex), '| format', 2 if F2 else 1)
if unresolved_values:
    print('UNRESOLVED (%d), kept as development material, not scored (docs/annotation-records.md §6):' % len(unresolved_values))
    print('\n'.join('  ' + u for u in unresolved_values))
if REPIN and F2 and IS_GOLD and errors and all(e.startswith('inputs: R16 ') for e in errors):
    # Only the pins are stale: everything else holds on the current inputs,
    # so pin them (written as the gold is: indent 1, no newline at the end).
    gold['inputs'] = {name: sha256_file(p) for name, p in PINNED.items()}
    with open(ANN, 'w', encoding='utf-8') as fh:
        fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
    print('re-pinned the inputs of', ANN, 'after checking everything else on them')
    errors = []
if errors:
    print('ERRORS (%d):' % len(errors)); print('\n'.join(errors)); sys.exit(1)
if F2:
    print('OK: order, schemes and bindings, premises/conclusions = walton.yml instantiations, R2, R4, R7, R8, R9, R10, R11, '
          'R12 ids, R13 exact targets, R14 actors and actions, R15 contexts, R16 the field contract and the pinned inputs, '
          'R17 verifications, atoms in lexicon')
else:
    print('OK: order, schemes and bindings, premises/conclusions = walton.yml instantiations, R1, R2, R4, R7, R8, R9, R10, R11, target formulas in primary instances, atoms in lexicon')
