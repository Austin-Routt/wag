"""Writes the owner's verifications into batch 2's gold (D086; docs/d086-first-gold.md §5).

It reads a run's review/human.json, the verdicts the owner gave in the dashboard and sends from the
data folder. For each verdict it finds the gold the verdict was given on, by the hash the verdict
recorded (of the file with LF or with CRLF line endings), in the working tree or in the
repository's history, and compares the item's revision then with its revision now: the verdict
holds of the item as it is now when the two are equal (docs/annotation-records.md §9).

- A verdict "The gold is right" (choice "gold") that holds of its item now writes the item's
  `verified` (by "owner", the verdict's UTC date, the item's revision and its claim's text
  revision), when the item is agreed in d086-rounds.json at the revision it has and has no
  unresolved value (R17).
- A verdict of another choice that holds of a verified item now takes the verification back.
- A verdict of another choice on an item not verified is recorded in the history entry only (a
  receipt), and so is a same-day confirmation of a verification: the item does not change, but the
  verdict's time is kept.
- Every remaining verdict is listed with the reason it is not written.

Each verdict applied or recorded keeps its time in the history entry (`verification_verdicts`), and
each new verdict is compared with the latest so recorded for its item (docs/d086-first-gold.md §5):
an older one is not written; verdicts of different choices at one time are not ordered, so the item
is left unverified, whichever comes first (and of two recorded at one time, the one other than
"gold" counts as the latest); and a verification whose time is not recorded has only its date, so a
conflicting verdict on that date, or any before it, is not written.

Nothing is written when a required input hash is missing, an input differs from the repository's
now, or an input is not one of the built-in dataset's; when the two copies of the gold differ, or the
gold or items.json is not valid JSON; when the file is not a list of verdicts, a JSON object repeats
a member, a verdict lacks its claim, choice, RFC 3339 time or inputs, or two verdicts name one claim;
or when d086-rounds.json cannot be read. The file is written as the recoding scripts write it,
gold.json and the built-in copy (wag/data/annotation/kialo-13292-batch2/gold.json), changing only the
`verified` fields, the file's `status` and one `history` entry.

usage: python3 -I verify_gold.py HUMAN.json [--gold GOLD.json] [--builtin GOLD.json]
                                 [--rounds ROUNDS.json] [--also GOLD.json ...]

--gold, --builtin and --rounds name other copies to read and write (tests use them); --also adds a
version of the gold to look for a verdict's gold in, besides the working tree and the history
(tests). The claims' texts and the other inputs are always the repository's.
"""
import datetime
import hashlib
import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.normpath(os.path.join(HERE, '..', '..'))
BUILTIN = os.path.join(ROOT, 'wag', 'data', 'annotation', 'kialo-13292-batch2')
GOLD_NAME = 'gold.json'
# The gold's two copies, as the repository's history holds them.
GOLD_IN_GIT = ('docs/kialo-batch2/gold.json', 'wag/data/annotation/kialo-13292-batch2/gold.json')
# Each input the built-in dataset records, by the name its verdicts give it, and its file
# (wag/internal/annotate/dataset.go: the manifest's files, walton.yml and the game's NASCQ).
INPUTS = {
    'items.json': os.path.join(BUILTIN, 'items.json'),
    'guidelines.md': os.path.join(BUILTIN, 'guidelines.md'),
    'lexicon.json': os.path.join(BUILTIN, 'lexicon.json'),
    'case.json': os.path.join(BUILTIN, 'case.json'),
    'walton.yml': os.path.join(ROOT, 'wag', 'data', 'annotation', 'walton.yml'),
    'schemes.json': os.path.join(ROOT, 'wag', 'data', 'nascq', 'schemes.json'),
    'critical_questions.json': os.path.join(ROOT, 'wag', 'data', 'nascq', 'critical_questions.json'),
    'edges.json': os.path.join(ROOT, 'wag', 'data', 'nascq', 'edges.json'),
}
ROUNDS_FORMAT = 'wag-d086-rounds/1'
REVIEW_STATES = ('not drafted', 'not reviewed', 'objections', 'agreed', 'to the owner')
AGREED_BY = ('the reviewer and the developer', "the owner's choice")
CHOICE_WORDS = {'gold': 'The gold is right', 'none': 'None is right', 'unsure': 'Unsure',
                'reviewer': 'The reviewer is right', '': 'no choice, a note only'}
DATE = re.compile(r'\d{4}-\d{2}-\d{2}$')
VERDICT_TIME = re.compile(r'^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})(\.\d{1,9})?(Z|[+-]\d{2}:\d{2})$')
SHA = re.compile(r'[0-9a-f]{64}$')


def read(path):
    with open(path, 'rb') as fh:
        return fh.read()


def unique_members(pairs):
    out = {}
    for key, value in pairs:
        if key in out:
            raise ValueError(f'duplicate JSON member {key!r}')
        out[key] = value
    return out


def reject_constant(value):
    raise ValueError(f'non-JSON number {value}')


def loads_json(raw):
    return json.loads(raw, object_pairs_hook=unique_members, parse_constant=reject_constant)


def load_json(path):
    with open(path, encoding='utf-8') as fh:
        return loads_json(fh.read())


def valid_date(value):
    try:
        return isinstance(value, str) and bool(DATE.fullmatch(value)) and bool(datetime.date.fromisoformat(value))
    except ValueError:
        return False


def verdict_time(value):
    """A dashboard RFC3339/RFC3339Nano timestamp, normalized to UTC without losing nanoseconds.
    Return its order key, UTC date and canonical timestamp; reject a mere date or invalid calendar."""
    m = VERDICT_TIME.fullmatch(value) if isinstance(value, str) else None
    if m is None:
        raise ValueError('at is not an RFC3339 timestamp with a timezone')
    dt = datetime.datetime.fromisoformat(m[1] + 'T' + m[2] + m[4].replace('Z', '+00:00'))
    dt = dt.astimezone(datetime.timezone.utc)
    fraction = (m[3] or '')[1:]
    ns = int(fraction.ljust(9, '0') or '0')
    key = (dt, ns)
    stamp = dt.strftime('%Y-%m-%dT%H:%M:%S') + ('.' + fraction.rstrip('0') if ns else '') + 'Z'
    return key, dt.date().isoformat(), stamp


def sha(b):
    return hashlib.sha256(b).hexdigest()


def hashes(b):
    """The SHA-256 of a file's bytes, and of the same text with LF and with CRLF line endings: a
    checkout made before .gitattributes kept these files LF may have CRLF, and the dataset's
    hashes read both (dataset.go, lineHashes)."""
    lf = b.replace(b'\r\n', b'\n')
    return {sha(b), sha(lf), sha(lf.replace(b'\n', b'\r\n'))}


def item_rev(it):
    """An item's revision: the SHA-256 of the item without its adjudication and verification, in
    canonical JSON (docs/annotation-records.md §9). A copy of validate.py's, which runs its whole
    check when imported; the tests run validate.py on what this script writes, and R17 compares."""
    x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
    return sha(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8'))


def unresolved(v):
    """Whether a value holds a marker {"unresolved": why} anywhere (§8)."""
    if isinstance(v, dict):
        if set(v) == {'unresolved'}:
            return True
        return any(unresolved(x) for x in v.values())
    if isinstance(v, list):
        return any(unresolved(x) for x in v)
    return False


def fail(msg):
    print('verify_gold: ' + msg, file=sys.stderr)
    sys.exit(1)


def read_or_fail(path):
    try:
        return read(path)
    except OSError as e:
        fail(f'{path}: {e}')


def read_rounds(path):
    """d086-rounds.json's claims, checked: each claim's review state, and the revision agreed on
    exactly when it is agreed."""
    try:
        doc = load_json(path)
    except (OSError, ValueError) as e:
        fail(f'{path}: {e}')
    if not (isinstance(doc, dict) and doc.get('format') == ROUNDS_FORMAT and isinstance(doc.get('claims'), dict)):
        fail(f'{path} is not the rounds of the first gold ({ROUNDS_FORMAT})')
    for cid, e in doc['claims'].items():
        ok = (isinstance(e, dict) and e.get('review') in REVIEW_STATES and isinstance(e.get('draft'), int)
              and not isinstance(e.get('draft'), bool) and e['draft'] >= 0
              and (e['draft'] == 0) == (e['review'] == 'not drafted')
              and isinstance(e.get('objections'), list) and all(isinstance(o, str) and o for o in e['objections'])
              and bool(e['objections']) == (e['review'] == 'objections'))
        if ok and e['review'] == 'agreed':
            ok = (isinstance(e.get('agreed_rev'), str) and bool(SHA.match(e['agreed_rev']))
                  and e.get('agreed_by') in AGREED_BY and isinstance(e.get('agreed_date'), str)
                  and valid_date(e['agreed_date']))
        elif ok:
            ok = e.get('agreed_rev') is None and e.get('agreed_by') is None and e.get('agreed_date') is None
        # The owner's decisions a claim's coding waits for (docs/d086-first-gold.md §11), when there are any.
        if ok and 'pending' in e:
            ok = isinstance(e['pending'], list) and bool(e['pending']) and all(isinstance(p, str) and p for p in e['pending'])
        if not ok:
            fail(f'{path}: {cid}: not a claim\'s state (draft, review, objections, the revision agreed on '
                 'exactly when it is agreed, and the decisions it waits for, if any, by name)')
    return doc['claims']


def history_versions(wanted):
    """The gold's versions in the repository's history whose hashes are among `wanted`, by hash."""
    found = {}
    try:
        log = subprocess.run(['git', '-C', ROOT, 'log', '--format=%H', '--', *GOLD_IN_GIT],
                             capture_output=True, text=True, timeout=120)
    except (OSError, subprocess.SubprocessError):
        return found
    if log.returncode != 0:
        return found
    for commit in log.stdout.split():
        for path in GOLD_IN_GIT:
            r = subprocess.run(['git', '-C', ROOT, 'show', f'{commit}:{path}'], capture_output=True, timeout=120)
            if r.returncode != 0:
                continue
            for h in hashes(r.stdout) & wanted:
                found.setdefault(h, r.stdout)
        if wanted <= set(found):
            break
    return found


def status_text(n, total):
    if n == 0:
        return ('Development material (D086): made by language models (the batch\'s adjudication, and the '
                'developer\'s drafts of the first gold, round by round: docs/d086-first-gold.md); no item is '
                'verified by a person yet.')
    return (f'The first gold, in part (D086): {n} of its {total} items are verified by the owner (each '
            'such item\'s `verified`), and each is gold while it is the revision verified (R17). The '
            'others are development material, made by language models.')


def main(argv):
    args = list(argv)
    paths = {'--gold': os.path.join(HERE, GOLD_NAME), '--builtin': os.path.join(BUILTIN, GOLD_NAME),
             '--rounds': os.path.join(HERE, 'd086-rounds.json')}
    also = []
    i = 0
    rest = []
    while i < len(args):
        a = args[i]
        if a in paths or a == '--also':
            if i + 1 >= len(args):
                fail(f'{a} needs a file')
            if a == '--also':
                also.append(args[i + 1])
            else:
                paths[a] = args[i + 1]
            i += 2
            continue
        rest.append(a)
        i += 1
    if len(rest) != 1:
        fail('usage: python3 -I verify_gold.py HUMAN.json [--gold GOLD.json] [--builtin GOLD.json] '
             '[--rounds ROUNDS.json] [--also GOLD.json ...]')
    human = rest[0]
    gold_path, builtin_path = paths['--gold'], paths['--builtin']

    # The verdicts: a list, each with its claim, choice, time and inputs, one per claim.
    try:
        verdicts = load_json(human)
    except (OSError, ValueError, RecursionError) as e:
        fail(f'{human}: {e}')
    if not isinstance(verdicts, list):
        fail(f'{human} is not a list of verdicts (a run\'s review/human.json)')
    seen = set()
    times = {}
    for n, v in enumerate(verdicts):
        if not (isinstance(v, dict) and isinstance(v.get('item'), str) and v['item']
                and isinstance(v.get('choice'), str) and isinstance(v.get('at'), str)
                and isinstance(v.get('inputs_sha256'), dict)
                and all(isinstance(k, str) and isinstance(h, str) for k, h in v['inputs_sha256'].items())):
            fail(f'{human}: verdict {n + 1} is not a verdict with its claim, choice, time and inputs_sha256 '
                 '(a verdict given before the dashboard recorded its inputs cannot verify an item)')
        try:
            times[v['item']] = verdict_time(v['at'])
        except ValueError as e:
            fail(f'{human}: verdict {n + 1}: {e}')
        unknown = set(v['inputs_sha256']) - set(INPUTS) - {GOLD_NAME}
        if unknown:
            fail(f'{v["item"]}: the verdict records inputs the built-in dataset does not have, {sorted(unknown)}: nothing written')
        missing = set(INPUTS) - set(v['inputs_sha256'])
        if missing:
            fail(f'{v["item"]}: the verdict is missing required input hashes {sorted(missing)}: nothing written')
        if any(not SHA.fullmatch(h) for h in v['inputs_sha256'].values()):
            fail(f'{v["item"]}: an input hash is not a SHA-256: nothing written')
        if v['item'] in seen:
            fail(f'{human}: two verdicts on {v["item"]}: a run\'s file holds one per claim')
        seen.add(v['item'])

    # Every input but the gold, as the repository has it now.
    for v in verdicts:
        for name, h in sorted(v['inputs_sha256'].items()):
            if name == GOLD_NAME:
                continue
            if name not in INPUTS:
                fail(f'{v["item"]}: the verdict records an input the built-in dataset does not have, {name}: '
                     'nothing written')
            if h not in hashes(read_or_fail(INPUTS[name])):
                fail(f'{v["item"]}: the verdict was given on another {name} than the repository\'s now: nothing written')

    gold_bytes = read_or_fail(gold_path)
    if read_or_fail(builtin_path) != gold_bytes:
        fail(f'the built-in copy {builtin_path} differs from {gold_path}: make them equal first')
    rounds = read_rounds(paths['--rounds'])
    try:
        gold = loads_json(gold_bytes)
    except (ValueError, RecursionError) as e:
        fail(f'{gold_path} is not valid JSON: {e}: nothing written')
    if not (isinstance(gold, dict) and isinstance(gold.get('items'), list)
            and all(isinstance(it, dict) and isinstance(it.get('kialo_claim_id'), str) for it in gold['items'])
            and isinstance(gold.get('history', []), list)
            and all(isinstance(e, dict) for e in gold.get('history', []))):
        fail(f'{gold_path} is not an annotation file with its items and history: nothing written')
    gold.setdefault('history', [])
    by_id = {it['kialo_claim_id']: it for it in gold['items']}
    try:
        texts = {i['kialo_claim_id']: i['text'] for i in load_json(INPUTS['items.json'])['items']}
    except (OSError, ValueError, RecursionError, KeyError, TypeError) as e:
        fail(f'{INPUTS["items.json"]} is not the claims of the dataset: {e}: nothing written')

    # The gold each verdict was given on: now, a version given with --also, or the history's.
    versions = {h: gold_bytes for h in hashes(gold_bytes)}
    for p in also:
        b = read_or_fail(p)
        for h in hashes(b):
            versions.setdefault(h, b)
    wanted = {v['inputs_sha256'][GOLD_NAME] for v in verdicts if GOLD_NAME in v['inputs_sha256']} - set(versions)
    if wanted:
        versions.update(history_versions(wanted))
    parsed = {}
    invalid = object()

    def item_then(h, cid):
        """The item as the gold the verdict was given on has it; `invalid` when that gold is not
        valid JSON with its items (then the verdict is listed, not written)."""
        if h not in parsed:
            try:
                doc = loads_json(versions[h])
                parsed[h] = {it['kialo_claim_id']: it for it in doc['items']}
            except (ValueError, RecursionError, KeyError, TypeError):
                parsed[h] = invalid
        return invalid if parsed[h] is invalid else parsed[h].get(cid)

    # The verdicts already applied or recorded, from the history entries' receipts: for each claim,
    # the latest (of two at one time, one other than "gold": verdicts at one time are not ordered, and
    # the item is left unverified), and the date and revision of each applied verdict "gold", which
    # tell a verification whose time is recorded from one that has only its date (§5).
    latest, timed = {}, {}
    for entry in gold['history']:
        for receipt in entry.get('verification_verdicts', []):
            try:
                key, rdate, _ = verdict_time(receipt['at'])
                cid = receipt['item']
                if not isinstance(cid, str) or not isinstance(receipt['choice'], str):
                    raise ValueError('receipt has no claim or choice')
            except (KeyError, TypeError, ValueError) as e:
                fail(f'the gold has an invalid verification receipt: {e}')
            if receipt['choice'] == 'gold':
                timed.setdefault(cid, set()).add((rdate, receipt.get('item_rev')))
            if (cid not in latest or key > latest[cid][0]
                    or (key == latest[cid][0] and receipt['choice'] != 'gold')):
                latest[cid] = (key, receipt)
    verified, withdrawn, noted, confirmed, recorded, skipped, receipts = [], [], [], [], [], [], []
    for v in verdicts:
        cid, choice = v['item'], v['choice']
        word = CHOICE_WORDS.get(choice, f'{choice} is right')
        h = v['inputs_sha256'].get(GOLD_NAME)
        it = by_id.get(cid)
        if it is None:
            skipped.append((cid, 'no such claim in the gold'))
            continue
        if h is None:
            skipped.append((cid, 'the verdict recorded no gold'))
            continue
        if h not in versions:
            skipped.append((cid, 'given on a gold this repository does not hold (not in the working tree or its '
                                 'history): verify it again'))
            continue
        then = item_then(h, cid)
        if then is invalid:
            skipped.append((cid, 'the gold it was given on is not valid JSON: verify it again'))
            continue
        if then is None or item_rev(then) != item_rev(it):
            skipped.append((cid, 'the item has changed since the verdict: verify it again'))
            continue
        key, utc_date, utc_at = times[cid]
        prior = latest.get(cid)
        if prior and key < prior[0]:
            skipped.append((cid, 'the verdict is older than the latest applied owner verdict'))
            continue
        # Verdicts of different choices at one time are not ordered: the item is left unverified.
        tie = prior is not None and key == prior[0] and choice != prior[1]['choice']
        if tie and choice == 'gold':
            skipped.append((cid, 'a verdict of another choice has the same time: the two are not ordered, and the '
                                 'item stays unverified until a newer verdict'))
            continue
        # A verification whose time no receipt records has only its date: a conflicting verdict on that
        # date, or any verdict before it, cannot be ordered with it, whatever else the history records.
        old_verification = it.get('verified')
        if (isinstance(old_verification, dict) and valid_date(old_verification.get('date'))
                and (old_verification['date'], old_verification.get('item_rev')) not in timed.get(cid, set())):
            if utc_date < old_verification['date'] or (utc_date == old_verification['date'] and choice != 'gold'):
                skipped.append((cid, 'the earlier verification records only a date; this conflicting verdict is not provably later'))
                continue
        receipt = {'item': cid, 'choice': choice, 'at': utc_at, 'item_rev': item_rev(it)}
        if choice != 'gold':
            if 'verified' in it:
                del it['verified']
                withdrawn.append((cid, word, tie))
                receipts.append(receipt)
            elif not prior or key > prior[0] or tie:
                # The item does not change, but the verdict's time is kept: otherwise an older verdict
                # "gold" imported later could verify the item.
                receipts.append(receipt)
                noted.append((cid, word))
                recorded.append((cid, f'its verdict is "{word}"; the item remains unverified'))
            else:
                skipped.append((cid, f'its verdict is "{word}", not "The gold is right"'))
            continue
        state = rounds.get(cid)
        if state is None:
            skipped.append((cid, 'not in a round of the first gold'))
            continue
        if state['review'] != 'agreed':
            skipped.append((cid, f'not agreed in its round (the review: {state["review"]})'))
            continue
        if state['agreed_rev'] != item_rev(it):
            skipped.append((cid, 'agreed at another revision than it has now: review it again'))
            continue
        if state.get('pending'):
            skipped.append((cid, 'its coding waits for the owner\'s decisions ' + ', '.join(state['pending']) +
                                 ' (d086-rounds.json): verify it once they are made'))
            continue
        if unresolved({k: x for k, x in it.items() if k not in ('adjudication', 'verified')}):
            skipped.append((cid, 'it has an unresolved value, and a verified item has none (R17)'))
            continue
        rec = {'by': 'owner', 'date': utc_date, 'item_rev': item_rev(it),
               'text_sha256': sha(texts[cid].encode('utf-8'))}
        if it.get('verified') == rec:
            if not prior or key > prior[0]:
                # A later confirmation the same day: its time is kept, so that a verdict between the two
                # cannot take the verification back.
                receipts.append(receipt)
                confirmed.append(cid)
                recorded.append((cid, 'a later confirmation; the verification itself is unchanged'))
            else:
                skipped.append((cid, 'already verified so'))
            continue
        it['verified'] = rec
        verified.append((cid, rec))
        receipts.append(receipt)

    if receipts:
        n = sum('verified' in it for it in gold['items'])
        gold['status'] = status_text(n, len(gold['items']))
        what = []
        if verified:
            what.append('Verified by the owner: ' + ', '.join(f'{c} (the verdict of {r["date"]})' for c, r in verified) + '.')
        later = [(c, w) for c, w, tie in withdrawn if not tie]
        tied = [(c, w) for c, w, tie in withdrawn if tie]
        if later:
            what.append('Verification taken back by the owner\'s later verdict: ' +
                        ', '.join(f'{c} ({w})' for c, w in later) + '.')
        if tied:
            what.append('Verification taken back by a verdict of another choice at the same time as the one that '
                        'verified it (not ordered): ' + ', '.join(f'{c} ({w})' for c, w in tied) + '.')
        if noted:
            what.append('The owner\'s verdicts recorded, on items not verified: ' +
                        ', '.join(f'{c} ({w})' for c, w in noted) + '.')
        if confirmed:
            what.append('A later confirmation recorded, the verification unchanged: ' + ', '.join(confirmed) + '.')
        gold['history'].append({'date': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%d'),
                                'by': "the owner's verification (verify_gold.py; D086)", 'what': ' '.join(what),
                                'verification_verdicts': receipts})
        out = json.dumps(gold, indent=1, ensure_ascii=False).encode('utf-8')
        for path in (gold_path, builtin_path):
            with open(path, 'wb') as fh:
                fh.write(out)
    for cid, rec in verified:
        print(f'verified: {cid} (revision {rec["item_rev"][:12]}, {rec["date"]})')
    for cid, word, tie in withdrawn:
        if tie:
            print(f'taken back: {cid} (a verdict "{word}" at the same time as the one that verified it: '
                  'the two are not ordered, and the item is left unverified)')
        else:
            print(f'taken back: {cid} (the owner\'s later verdict: {word})')
    for cid, why in recorded:
        print(f'receipt recorded: {cid}: {why}')
    for cid, why in skipped:
        print(f'not written: {cid}: {why}')
    print(f'{len(verified)} verified, {len(withdrawn)} taken back, '
          f'{len(recorded)} receipts recorded, {len(skipped)} not written')


if __name__ == '__main__':
    main(sys.argv[1:])
