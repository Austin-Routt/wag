"""The rounds of the first gold (D086; docs/d086-first-gold.md §2), computed from three codings.

usage: python3 -I d086_rounds.py REFERENCE.json [--init]

It measures, on the ten core fields, where annotator A, annotator B and the reference (the gold
before the draft, REFERENCE.json) disagree, by agreement.py: A with B, A with the reference and B
with the reference, each with the case record. A field disagrees for a claim when any two of the
three differ on it (an unresolved value is left out, as agreement.py leaves it out). Round 1: the
claims where three or more fields disagree, the items the identity contract left open for the
draft (docs/annotation-records.md: the four fragments and the four targets forced onto a primary
instance), and three controls drawn from the claims where no field disagrees, less those already in
round 1: random.Random(86).sample(population, 3), the population in items.json's order. Round 2:
the other claims where one or two fields disagree. Round 3: the rest.

Without --init it computes the rounds and compares them with d086-rounds.json's, and says whether
they are the same. With --init it writes d086-rounds.json, every claim not drafted yet; it refuses
to overwrite a file in which a claim has moved on.
"""
import hashlib
import json
import os
import random
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
ROUNDS = os.path.join(HERE, 'd086-rounds.json')
FORMAT = 'wag-d086-rounds/1'
CORE = ['scheme', 'fit', 'role', 'polarity', 'target_claim', 'target_instance', 'target_part',
        'target_formula', 'evidence', 'cq']
# The identity contract's items left open for the first gold's draft (docs/annotation-records.md).
OPEN_FRAGMENTS = ['13292.6', '13292.11', '13292.108', '13292.218']
OPEN_TARGETS = ['13292.12', '13292.141', '13292.182', '13292.198']
THRESHOLD, SEED, CONTROLS = 3, 86, 3


def fail(msg):
    print('d086_rounds: ' + msg, file=sys.stderr)
    sys.exit(1)


def load_json(path):
    with open(path, encoding='utf-8') as fh:
        return json.load(fh)


def disagreements(a, b, tmp):
    """The core fields on which two codings differ, per claim (agreement.py's own lists)."""
    out = os.path.join(tmp, 'agreement.json')
    r = subprocess.run([sys.executable, '-I', os.path.join(HERE, 'agreement.py'), a, b,
                        '--case', os.path.join(HERE, 'case.json'), '--json', out],
                       capture_output=True, text=True)
    if r.returncode != 0:
        fail(f'agreement.py {os.path.basename(a)} {os.path.basename(b)}: {r.stderr.strip()}')
    rep = load_json(out)
    per = {}
    for f in CORE:
        for d in rep['disagreements'].get(f, []):
            per.setdefault(d['claim'], set()).add(f)
    return per


def compute(reference):
    items = [i['kialo_claim_id'] for i in load_json(os.path.join(HERE, 'items.json'))['items']]
    a, b = os.path.join(HERE, 'annotator-A.json'), os.path.join(HERE, 'annotator-B.json')
    with tempfile.TemporaryDirectory() as tmp:
        pairs = [disagreements(a, b, tmp), disagreements(a, reference, tmp), disagreements(b, reference, tmp)]
    fields = {c: sorted(set().union(*(p.get(c, set()) for p in pairs))) for c in items}
    claims, round1 = {}, []
    for c in items:
        n = len(fields[c])
        if n >= THRESHOLD:
            why = f'the three codings differ on {n} of the ten core fields'
        elif c in OPEN_FRAGMENTS:
            why = 'a fragment the identity contract left open for the draft'
        elif c in OPEN_TARGETS:
            why = 'a target the identity contract left open for the draft'
        else:
            continue
        round1.append(c)
        claims[c] = {'round': 1, 'why': why}
    population = [c for c in items if not fields[c] and c not in claims]
    drawn = random.Random(SEED).sample(population, CONTROLS)
    for c in items:
        if c in drawn:
            round1.append(c)
            claims[c] = {'round': 1, 'why': 'a control: the three codings agree on every core field'}
    round1 = [c for c in items if c in claims]
    round2 = [c for c in items if c not in claims and fields[c]]
    for c in round2:
        n = len(fields[c])
        claims[c] = {'round': 2, 'why': f'the three codings differ on {n} of the ten core fields'}
    round3 = [c for c in items if c not in claims]
    for c in round3:
        claims[c] = {'round': 3, 'why': 'the three codings agree on every core field'}
    for c in items:
        claims[c]['fields'] = fields[c]
    with open(reference, 'rb') as fh:
        ref_sha = hashlib.sha256(fh.read()).hexdigest()
    measure = {
        'fields': CORE,
        'codings': ['annotator-A.json', 'annotator-B.json', 'the reference, gold.json before the draft'],
        'reference_sha256': ref_sha,
        'threshold': THRESHOLD,
        'open_items': OPEN_FRAGMENTS + OPEN_TARGETS,
        'controls': {'seed': SEED, 'k': CONTROLS, 'population': population, 'drawn': drawn},
    }
    rounds = [{'round': 1, 'claims': round1}, {'round': 2, 'claims': round2}, {'round': 3, 'claims': round3}]
    return measure, rounds, {c: claims[c] for c in items}


def fresh(entry):
    return dict(entry, draft=0, review='not drafted', agreed_rev=None, agreed_by=None, agreed_date=None,
                objections=[])


def main(argv):
    args = list(argv)
    init = '--init' in args
    if init:
        args.remove('--init')
    if len(args) != 1:
        fail('usage: python3 -I d086_rounds.py REFERENCE.json [--init]')
    measure, rounds, claims = compute(args[0])
    for r in rounds:
        print(f'round {r["round"]}: {len(r["claims"])} claims: {" ".join(r["claims"])}')
    print(f'controls: {" ".join(measure["controls"]["drawn"])} (from {len(measure["controls"]["population"])})')
    if init:
        if os.path.exists(ROUNDS):
            old = load_json(ROUNDS)
            if any(e.get('review') != 'not drafted' for e in old.get('claims', {}).values()):
                fail(f'{ROUNDS}: a claim has moved on since the rounds were chosen: not overwritten')
        doc = {
            'format': FORMAT,
            'note': ('The rounds of the first gold (D086; docs/d086-first-gold.md): which claims each round '
                     'takes and why (d086_rounds.py, from the reference named by its hash), and where each '
                     'claim stands: its latest draft, the review\'s state (not drafted, not reviewed, '
                     'objections, agreed, to the owner) and the revision of the item agreed on '
                     '(verify_gold.py writes a verification only at that revision).'),
            'measure': measure,
            'rounds': rounds,
            'claims': {c: fresh(e) for c, e in claims.items()},
        }
        with open(ROUNDS, 'wb') as fh:
            fh.write(json.dumps(doc, indent=1, ensure_ascii=False).encode('utf-8') + b'\n')
        print(f'written: {ROUNDS}')
        return
    doc = load_json(ROUNDS)
    same = doc['rounds'] == rounds and all(
        {k: doc['claims'][c].get(k) for k in ('round', 'why', 'fields')} == {k: e[k] for k in ('round', 'why', 'fields')}
        for c, e in claims.items()) and doc['measure'] == measure
    print('the same as d086-rounds.json' if same else 'NOT the same as d086-rounds.json')
    if not same:
        sys.exit(1)


if __name__ == '__main__':
    main(sys.argv[1:])
