"""Tests of verify_gold.py: the cases V1 to V51 of docs/d086-first-gold.md §6, whose outcomes were
written before the tests ran (V17-V32 at 8c64ee9; V33 added with the code, V34 with draft 2; V35-V42 at
3cabca9, after the independent review of draft 2; V43-V46 at a3b9141, after the planted defects; V47-V49
at cdfd2de, after ChatGPT's review of exchange 2; V50-V51 at 8e900ec, after the independent review of
step 1), and three more (V3 through the history, the rounds file it reads, and the repository's own).
Each case works on temporary copies of the gold and of the rounds file; the claims' texts and the
other inputs are the repository's.

usage: python3 -I test_verify_gold.py
"""
import hashlib
import json
from pathlib import Path
import os
import shutil
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.normpath(os.path.join(HERE, '..', '..'))
BUILTIN = os.path.join(ROOT, 'wag', 'data', 'annotation', 'kialo-13292-batch2')
SCRIPT = os.path.join(HERE, 'verify_gold.py')
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


def read(path):
    with open(path, 'rb') as fh:
        return fh.read()


def load_json(path):
    with open(path, encoding='utf-8') as fh:
        return json.load(fh)


def dump(obj):
    return json.dumps(obj, indent=1, ensure_ascii=False).encode('utf-8')


def sha(b):
    return hashlib.sha256(b).hexdigest()


def item_rev(it):
    x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
    return sha(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8'))


def verdict(item, choice, gold_bytes, at='2026-11-28T10:00:00Z', note=None, **inputs):
    """A verdict as the dashboard records it: every input's hash now, the gold's given."""
    h = {name: sha(read(p)) for name, p in INPUTS.items()}
    if gold_bytes is not None:
        h['gold.json'] = sha(gold_bytes)
    h.update(inputs)
    v = {'item': item, 'choice': choice, 'at': at, 'inputs_sha256': h}
    if note:
        v['note'] = note
    return v


class VerifyGold(unittest.TestCase):
    def setUp(self):
        self.dir = tempfile.mkdtemp()
        self.gold = os.path.join(self.dir, 'gold.json')
        self.builtin = os.path.join(self.dir, 'builtin.json')
        self.rounds = os.path.join(self.dir, 'rounds.json')
        shutil.copy(os.path.join(HERE, 'gold.json'), self.gold)
        shutil.copy(os.path.join(HERE, 'gold.json'), self.builtin)
        shutil.copy(os.path.join(HERE, 'd086-rounds.json'), self.rounds)
        self.bytes = read(self.gold)
        self.original = os.path.join(self.dir, 'original.json')
        shutil.copy(self.gold, self.original)

    def tearDown(self):
        shutil.rmtree(self.dir)

    def run_script(self, verdicts, also=()):
        human = os.path.join(self.dir, 'human.json')
        with open(human, 'w', encoding='utf-8') as fh:
            json.dump(verdicts, fh)
        cmd = [sys.executable, '-I', SCRIPT, human, '--gold', self.gold, '--builtin', self.builtin, '--rounds', self.rounds]
        for a in also:
            cmd += ['--also', a]
        return subprocess.run(cmd, capture_output=True, text=True)

    def items(self, path=None):
        return {it['kialo_claim_id']: it for it in load_json(path or self.gold)['items']}

    def agree(self, *cids, rev=None):
        """The claims agreed in their round at the revision they have in the temporary gold, waiting for
        no decision of the owner (whatever the repository's rounds file says of them now)."""
        doc = load_json(self.rounds)
        now = self.items()
        for cid in cids:
            doc['claims'][cid].update(draft=1, review='agreed', agreed_rev=rev or item_rev(now[cid]),
                                      agreed_by='the reviewer and the developer', agreed_date='2026-11-28', objections=[])
            doc['claims'][cid].pop('pending', None)
        with open(self.rounds, 'wb') as fh:
            fh.write(dump(doc))

    def write_gold(self, gold):
        for p in (self.gold, self.builtin):
            with open(p, 'wb') as fh:
                fh.write(dump(gold))

    def validate(self):
        """validate.py on the written gold (it checks the file it is given against the repository's inputs)."""
        return subprocess.run([sys.executable, '-I', os.path.join(HERE, 'validate.py'), self.gold], capture_output=True, text=True)

    def unchanged(self):
        self.assertEqual(read(self.gold), self.bytes)
        self.assertEqual(read(self.builtin), self.bytes)

    def refused(self, r, says):
        """Refused with the script's own message, naming what is wrong, not with a trace of the program."""
        self.assertNotEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertNotIn('Traceback', r.stderr)
        self.assertTrue(r.stderr.startswith('verify_gold: '), r.stderr)
        self.assertIn(says, r.stderr)

    def hand_verify(self, cid, date):
        """A verification made by hand, with no receipt: only its date (§5)."""
        g = load_json(self.gold)
        text = {i['kialo_claim_id']: i['text'] for i in load_json(INPUTS['items.json'])['items']}[cid]
        for it in g['items']:
            if it['kialo_claim_id'] == cid:
                it['verified'] = {'by': 'owner', 'date': date, 'item_rev': item_rev(it),
                                  'text_sha256': sha(text.encode('utf-8'))}
        self.write_gold(g)
        return read(self.gold)

    def receipts(self, cid):
        return [x for h in load_json(self.gold)['history'] for x in h.get('verification_verdicts', [])
                if x['item'] == cid]

    def test_v1_two_verdicts_on_the_gold_as_it_is(self):
        self.agree('13292.6', '13292.33')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'gold', self.bytes, at='2026-11-29T23:59:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        before = {it['kialo_claim_id']: it for it in json.loads(self.bytes)['items']}
        after = self.items()
        texts = {i['kialo_claim_id']: i['text'] for i in load_json(INPUTS['items.json'])['items']}
        for cid, date in (('13292.6', '2026-11-28'), ('13292.33', '2026-11-29')):
            self.assertEqual(after[cid]['verified'], {'by': 'owner', 'date': date, 'item_rev': item_rev(before[cid]),
                                                      'text_sha256': sha(texts[cid].encode('utf-8'))})
        # Byte for byte the file before but for the two fields, the status and one history entry.
        g = load_json(self.gold)
        old = json.loads(self.bytes)
        self.assertIn('2 of its 50 items are verified', g['status'])
        self.assertEqual(len(g['history']), len(old['history']) + 1)
        self.assertIn('13292.6', g['history'][-1]['what'])
        for it in g['items']:
            it.pop('verified', None)
        g['status'] = old['status']
        g['history'].pop()
        self.assertEqual(dump(g), self.bytes)
        self.assertEqual(read(self.builtin), read(self.gold))
        v = self.validate()
        self.assertEqual(v.returncode, 0, v.stdout + v.stderr)
        self.assertIn('R17', v.stdout)

    def test_v2_an_earlier_gold_in_which_the_item_differs(self):
        self.agree('13292.6')
        old = json.loads(self.bytes)
        for it in old['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['notes']['text'] += ' An earlier wording.'
        earlier = os.path.join(self.dir, 'earlier.json')
        with open(earlier, 'wb') as fh:
            fh.write(dump(old))
        r = self.run_script([verdict('13292.6', 'gold', read(earlier))], also=[earlier])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: the item has changed since the verdict', r.stdout)
        self.unchanged()

    def test_v3_an_earlier_gold_in_which_the_item_is_the_same(self):
        self.agree('13292.6')
        old = json.loads(self.bytes)
        for it in old['items']:
            if it['kialo_claim_id'] == '13292.33':
                it['notes']['text'] += ' Another item differs.'
        earlier = os.path.join(self.dir, 'earlier.json')
        with open(earlier, 'wb') as fh:
            fh.write(dump(old))
        r = self.run_script([verdict('13292.6', 'gold', read(earlier))], also=[earlier])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('verified: 13292.6', r.stdout)
        self.assertIn('verified', self.items()['13292.6'])
        v = self.validate()
        self.assertEqual(v.returncode, 0, v.stdout + v.stderr)

    def test_v4_a_gold_the_repository_does_not_hold(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', None, **{'gold.json': sha(b'another gold')})])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: given on a gold this repository does not hold', r.stdout)
        self.unchanged()

    def test_v5_other_choices_on_items_not_verified(self):
        vs = [verdict('13292.6', 'none', self.bytes), verdict('13292.33', 'unsure', self.bytes),
              verdict('13292.1', 'A', self.bytes), verdict('13292.30', 'reviewer', self.bytes),
              verdict('13292.36', '', self.bytes, note='a note only')]
        self.agree('13292.6', '13292.33', '13292.1', '13292.30', '13292.36')
        r = self.run_script(vs)
        self.assertEqual(r.returncode, 0, r.stderr)
        for cid, word in (('13292.6', 'None is right'), ('13292.33', 'Unsure'), ('13292.1', 'A is right'),
                          ('13292.30', 'The reviewer is right'), ('13292.36', 'no choice, a note only')):
            self.assertIn(f'receipt recorded: {cid}: its verdict is "{word}"; the item remains unverified', r.stdout)
        self.assertIn('5 receipts recorded, 0 not written', r.stdout)
        # Non-gold decisions do not verify or recode items, but their
        # timestamps must survive so an older positive import cannot win.
        after = load_json(self.gold)
        before = json.loads(self.bytes)
        self.assertEqual(after['items'], before['items'])
        self.assertEqual(len(after['history']), len(before['history']) + 1)
        self.assertEqual({r['item'] for r in after['history'][-1]['verification_verdicts']},
                         {v['item'] for v in vs})
        # Each receipt holds the verdict's time in UTC and the item's revision.
        now = self.items()
        self.assertEqual(sorted((r['item'], r['choice'], r['at'], r['item_rev'])
                                for r in after['history'][-1]['verification_verdicts']),
                         sorted((v['item'], v['choice'], '2026-11-28T10:00:00Z', item_rev(now[v['item']])) for v in vs))
        for cid, word in (('13292.6', 'None is right'), ('13292.33', 'Unsure'), ('13292.1', 'A is right'),
                          ('13292.30', 'The reviewer is right'), ('13292.36', 'no choice, a note only')):
            self.assertIn(f'{cid} ({word})', after['history'][-1]['what'])
        self.assertTrue(all('verified' not in it for it in after['items']))
        self.assertEqual(read(self.builtin), read(self.gold))

    def test_v6_a_later_verdict_takes_a_verification_back(self):
        self.agree('13292.6', '13292.33')
        self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'gold', self.bytes)])
        written = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', written, at='2026-11-30T08:00:00Z'), verdict('13292.33', 'gold', written)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("taken back: 13292.6 (the owner's later verdict: None is right)", r.stdout)
        after = self.items()
        self.assertNotIn('verified', after['13292.6'])
        self.assertIn('verified', after['13292.33'])
        g = load_json(self.gold)
        self.assertIn('1 of its 50 items are verified', g['status'])
        self.assertIn('taken back', g['history'][-1]['what'])
        v = self.validate()
        self.assertEqual(v.returncode, 0, v.stdout + v.stderr)

    def test_v6b_every_verification_taken_back(self):
        # With none left, the status is the drafts' own again, word for word (who made the material).
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes)])
        g = load_json(self.gold)
        self.assertIn('1 of its 50 items are verified', g['status'])
        self.assertIn('made by language models', g['status'])
        r = self.run_script([verdict('13292.6', 'unsure', read(self.gold), at='2026-11-30T08:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        g = load_json(self.gold)
        self.assertNotIn('verified', self.items()['13292.6'])
        self.assertEqual(g['status'], json.loads(self.bytes)['status'])

    def test_v7_a_claim_not_in_the_gold(self):
        r = self.run_script([verdict('13292.99999', 'gold', self.bytes)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.99999: no such claim in the gold', r.stdout)
        self.unchanged()

    def test_v8_another_input(self):
        self.agree('13292.6', '13292.33')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'gold', self.bytes, **{'items.json': sha(b'other items')})])
        self.assertNotEqual(r.returncode, 0)
        self.assertIn('items.json', r.stderr)
        self.assertIn('nothing written', r.stderr)
        self.unchanged()

    def test_v9_crlf_line_endings(self):
        self.agree('13292.6', '13292.33')
        crlf = self.bytes.replace(b'\n', b'\r\n')
        guidelines = read(INPUTS['guidelines.md']).replace(b'\r\n', b'\n').replace(b'\n', b'\r\n')
        r = self.run_script([verdict('13292.6', 'gold', crlf), verdict('13292.33', 'gold', self.bytes, **{'guidelines.md': sha(guidelines)})])
        self.assertEqual(r.returncode, 0, r.stderr)
        after = self.items()
        self.assertIn('verified', after['13292.6'])
        self.assertIn('verified', after['13292.33'])

    def test_v10_a_verified_item_recoded_is_stale(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes)])
        g = load_json(self.gold)
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['notes']['text'] += ' Recoded.'
        with open(self.gold, 'wb') as fh:
            fh.write(dump(g))
        v = self.validate()
        self.assertNotEqual(v.returncode, 0)
        self.assertIn('13292.6: R17 it was verified at revision', v.stdout + v.stderr)

    def test_v11_twice(self):
        self.agree('13292.6')
        vs = [verdict('13292.6', 'gold', self.bytes)]
        self.run_script(vs)
        once = read(self.gold)
        r = self.run_script(vs, also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: the gold already records this verdict', r.stdout)
        self.assertEqual(read(self.gold), once)
        self.assertEqual(read(self.builtin), once)

    def test_v12_not_verdicts(self):
        self.agree('13292.6')
        good = verdict('13292.6', 'gold', self.bytes)
        no_inputs = {k: v for k, v in good.items() if k != 'inputs_sha256'}
        for bad, says in (({'item': '13292.6'}, 'not a list'), ([no_inputs], 'inputs_sha256'), ([good, dict(good)], 'two verdicts on 13292.6')):
            r = self.run_script(bad)
            self.assertNotEqual(r.returncode, 0)
            self.assertIn('human.json', r.stderr)
            self.assertIn(says, r.stderr)
            self.unchanged()

    def test_v13_the_copies_differ(self):
        self.agree('13292.6')
        with open(self.builtin, 'ab') as fh:
            fh.write(b' ')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes)])
        self.assertNotEqual(r.returncode, 0)
        self.assertIn('differs', r.stderr)
        self.assertEqual(read(self.gold), self.bytes)
        self.assertEqual(read(self.builtin), self.bytes + b' ')

    def test_v14_an_unresolved_value(self):
        g = json.loads(self.bytes)
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['target']['actual']['formula'] = {'unresolved': 'which premise is not clear'}
        self.write_gold(g)
        b = read(self.gold)
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', b)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: it has an unresolved value', r.stdout)
        self.assertEqual(read(self.gold), b)

    def test_v15_not_agreed(self):
        doc = load_json(self.rounds)
        del doc['claims']['13292.1']
        doc['claims']['13292.6'].update(draft=1, review='objections', agreed_rev=None, agreed_by=None, agreed_date=None,
                                        objections=['the reviewer reads the link otherwise'])
        with open(self.rounds, 'wb') as fh:
            fh.write(dump(doc))
        self.agree('13292.33', rev='0' * 64)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'gold', self.bytes), verdict('13292.1', 'gold', self.bytes)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: not agreed in its round (the review: objections)', r.stdout)
        self.assertIn('13292.33: agreed at another revision than it has now', r.stdout)
        self.assertIn('13292.1: not in a round of the first gold', r.stdout)
        self.unchanged()

    def test_v16_inputs_recorded(self):
        self.agree('13292.6')
        no_gold = verdict('13292.6', 'gold', None)
        r = self.run_script([no_gold])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: the verdict recorded no gold', r.stdout)
        self.unchanged()
        folder = verdict('13292.6', 'gold', self.bytes)
        folder['inputs_sha256']['guidelines.md (built in)'] = folder['inputs_sha256'].pop('guidelines.md')
        r = self.run_script([folder])
        self.assertNotEqual(r.returncode, 0)
        self.assertIn('guidelines.md (built in)', r.stderr)
        self.unchanged()

    # Beyond the plan's cases.

    def test_v3_through_the_history(self):
        """V3 with the earlier gold found in the repository's history, not given: the verdict was
        given on the gold of the commit checked out, and the gold now differs in another item."""
        r = subprocess.run(['git', '-C', ROOT, 'show', 'HEAD:docs/kialo-batch2/gold.json'], capture_output=True)
        if r.returncode != 0:
            self.skipTest('no repository history here')
        head = r.stdout
        g = json.loads(head)
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.33':
                it['notes']['text'] += ' Another item differs.'
        self.write_gold(g)
        self.agree('13292.6')
        out = self.run_script([verdict('13292.6', 'gold', head)])
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn('verified: 13292.6', out.stdout)

    def test_a_rounds_file_it_cannot_read(self):
        doc = load_json(self.rounds)
        doc['claims']['13292.6'].update(draft=1, review='agreed')  # agreed, but no revision agreed on
        with open(self.rounds, 'wb') as fh:
            fh.write(dump(doc))
        r = self.run_script([verdict('13292.6', 'gold', self.bytes)])
        self.assertNotEqual(r.returncode, 0)
        self.assertIn('13292.6: not a claim', r.stderr)
        self.unchanged()

    def test_the_repository_rounds_file(self):
        """d086-rounds.json is readable, and every claim of items.json is in exactly one round."""
        doc = load_json(os.path.join(HERE, 'd086-rounds.json'))
        items = [i['kialo_claim_id'] for i in load_json(INPUTS['items.json'])['items']]
        listed = [c for r in doc['rounds'] for c in r['claims']]
        self.assertEqual(sorted(listed), sorted(items))
        self.assertEqual(sorted(doc['claims']), sorted(items))
        for r in doc['rounds']:
            for c in r['claims']:
                self.assertEqual(doc['claims'][c]['round'], r['round'])
        r = self.run_script([])  # the temporary rounds file is the repository's, read by the script
        self.assertEqual(r.returncode, 0, r.stderr)


    # V17-V32: the order of the verdicts (§5, "Order: the verdicts' times"), written before these tests ran.

    def test_v17_a_missing_input_hash(self):
        self.agree('13292.6')
        v = verdict('13292.6', 'gold', self.bytes)
        del v['inputs_sha256']['items.json']
        r = self.run_script([v])
        self.assertNotEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn('missing required input hashes', r.stderr)
        self.unchanged()
        v = verdict('13292.6', 'gold', self.bytes)
        v['inputs_sha256'] = {'gold.json': v['inputs_sha256']['gold.json']}
        r = self.run_script([v])
        self.assertNotEqual(r.returncode, 0, r.stdout + r.stderr)
        self.unchanged()

    def test_v18_a_time_that_is_not_rfc3339(self):
        self.agree('13292.6')
        for at in ('2026-99-99garbage', '2026-11-28', '2026-02-30T10:00:00Z', '2026-11-28T10:00:00'):
            r = self.run_script([verdict('13292.6', 'gold', self.bytes, at=at)])
            self.refused(r, 'verdict 1')
            self.unchanged()

    def test_v19_the_utc_date(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T23:30:00-06:00')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertEqual(self.items()['13292.6']['verified']['date'], '2026-11-29')
        self.assertEqual(load_json(self.gold)['history'][-1]['verification_verdicts'][0]['at'], '2026-11-29T05:30:00Z')

    def test_v20_a_repeated_member(self):
        self.agree('13292.6')
        v = verdict('13292.6', 'none', self.bytes)
        body = json.dumps([v]).replace('"choice": "none"', '"choice": "none", "choice": "gold"')
        path = Path(self.dir) / 'human.json'
        path.write_text(body)
        r = subprocess.run([sys.executable, '-I', SCRIPT, str(path), '--gold', self.gold, '--builtin', self.builtin,
                            '--rounds', self.rounds], capture_output=True, text=True)
        self.assertNotEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn('duplicate JSON member', r.stderr)
        self.unchanged()

    def test_v21_an_older_none_after_a_newer_verification(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-30T10:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at='2026-11-28T10:00:00Z')], also=[self.original])
        self.assertIn('older than the latest', r.stdout)
        self.assertEqual(read(self.gold), now, r.stdout + r.stderr)

    def test_v22_an_older_gold_after_a_newer_withdrawal(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        self.run_script([verdict('13292.6', 'none', read(self.gold), at='2026-11-30T10:00:00Z')])
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')], also=[self.original])
        self.assertEqual(read(self.gold), now, r.stdout + r.stderr)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v23_a_newer_none_on_an_item_never_verified(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at='2026-11-30T10:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('receipt recorded: 13292.6', r.stdout)
        self.assertEqual(self.items(), self.items(self.original))
        current = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertEqual(read(self.gold), current, r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v24_a_gold_between_two_nones(self):
        self.agree('13292.6')
        for choice, at in [('gold', '2026-11-27T10:00:00Z'), ('none', '2026-11-28T10:00:00Z'),
                           ('none', '2026-11-30T10:00:00Z')]:
            r = self.run_script([verdict('13292.6', choice, self.bytes, at=at)], also=[self.original])
            self.assertEqual(r.returncode, 0, r.stderr)
        current = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-29T10:00:00Z')], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertEqual(read(self.gold), current, r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v25_a_same_day_confirmation_keeps_its_time(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        r = self.run_script([verdict('13292.6', 'gold', read(self.gold), at='2026-11-28T15:00:00Z')])
        self.assertIn('receipt recorded: 13292.6: a later confirmation', r.stdout)
        self.assertIn('A later confirmation recorded', load_json(self.gold)['history'][-1]['what'])
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T14:00:00Z')])
        self.assertEqual(read(self.gold), now, r.stdout + r.stderr)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v26_verdicts_of_different_choices_at_one_time(self):
        stamp = '2026-11-28T10:00:00.123456789Z'
        # gold, then none: the verification taken back.
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at=stamp)])
        self.assertEqual(r.returncode, 0, r.stderr)
        r = self.run_script([verdict('13292.6', 'none', read(self.gold), at=stamp)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('taken back: 13292.6 (a verdict "None is right" at the same time', r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])
        self.assertIn('not ordered', load_json(self.gold)['history'][-1]['what'])
        settled = read(self.gold)
        for choice in ('gold', 'none'):
            r = self.run_script([verdict('13292.6', choice, self.bytes, at=stamp)], also=[self.original])
            self.assertEqual(r.returncode, 0, r.stderr)
            self.assertEqual(read(self.gold), settled, choice + ': ' + r.stdout)
        # none, then gold: the gold not written.
        self.write_gold(json.loads(self.bytes))
        self.unchanged()
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at=stamp)])
        self.assertEqual(r.returncode, 0, r.stderr)
        settled = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at=stamp)], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: a verdict of another choice has the same time', r.stdout)
        self.assertEqual(read(self.gold), settled, r.stdout)
        for choice in ('gold', 'none'):
            r = self.run_script([verdict('13292.6', choice, self.bytes, at=stamp)], also=[self.original])
            self.assertEqual(read(self.gold), settled, choice + ': ' + r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v27_one_nanosecond_later(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00.123456788Z')])
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T10:00:00.123456789Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("taken back: 13292.6 (the owner's later verdict: None is right)", r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v28_a_later_gold_after_a_none(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at='2026-11-28T10:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-30T10:00:00Z')], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('verified', self.items()['13292.6'])
        v = self.validate()
        self.assertEqual(v.returncode, 0, v.stdout + v.stderr)

    def test_v29_the_same_none_twice(self):
        self.agree('13292.6')
        stamp = '2026-11-30T10:00:00Z'
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at=stamp)])
        self.assertEqual(r.returncode, 0, r.stderr)
        current = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at=stamp)], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertEqual(read(self.gold), current, r.stdout)

    def test_v30_a_verification_with_only_its_date(self):
        # A verification whose receipt is gone (as before the times were kept): a none earlier that day.
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        g = load_json(self.gold)
        for h in g['history']:
            h.pop('verification_verdicts', None)
        self.write_gold(g)
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T09:00:00Z')])
        self.assertIn('records only a date', r.stdout)
        self.assertEqual(read(self.gold), now, r.stdout + r.stderr)
        # A none of the 28th recorded, a verification of the 30th made by hand, then a none of the 29th:
        # the receipt of the 28th does not order the 29th after a verification it does not record.
        self.write_gold(json.loads(self.bytes))
        self.unchanged()
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at='2026-11-28T10:00:00Z')])
        self.assertIn('receipt recorded: 13292.6', r.stdout)
        g = load_json(self.gold)
        text = {i['kialo_claim_id']: i['text'] for i in load_json(INPUTS['items.json'])['items']}['13292.6']
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['verified'] = {'by': 'owner', 'date': '2026-11-30', 'item_rev': item_rev(it),
                                  'text_sha256': sha(text.encode('utf-8'))}
        self.write_gold(g)
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-29T10:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('records only a date', r.stdout)
        self.assertEqual(read(self.gold), now, r.stdout)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v31_the_history_names_every_verdict(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'none', self.bytes),
                             verdict('13292.30', 'unsure', self.bytes)])
        self.assertEqual(r.returncode, 0, r.stderr)
        entry = load_json(self.gold)['history'][-1]
        self.assertIn('Verified by the owner: 13292.6', entry['what'])
        self.assertIn('13292.33 (None is right)', entry['what'])
        self.assertIn('13292.30 (Unsure)', entry['what'])
        self.assertEqual(sorted(x['item'] for x in entry['verification_verdicts']), ['13292.30', '13292.33', '13292.6'])
        self.assertTrue(all(x['at'] == '2026-11-28T10:00:00Z' for x in entry['verification_verdicts']))

    def test_v32_a_gold_that_is_not_valid_json(self):
        bad = self.bytes.replace(b'"format": "wag-annotation/2"',
                                 b'"format": "wag-annotation/2", "format": "wag-annotation/2"', 1)
        self.assertNotEqual(bad, self.bytes)
        # The gold itself: refused, nothing written, no traceback.
        for p in (self.gold, self.builtin):
            with open(p, 'wb') as fh:
                fh.write(bad)
        r = self.run_script([verdict('13292.6', 'gold', bad)])
        self.assertNotEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn('is not valid JSON', r.stderr)
        self.assertNotIn('Traceback', r.stderr)
        self.assertEqual(read(self.gold), bad)
        self.assertEqual(read(self.builtin), bad)
        # An earlier gold given with --also: that verdict listed, not written.
        self.write_gold(json.loads(self.bytes))
        self.unchanged()
        earlier = os.path.join(self.dir, 'earlier.json')
        with open(earlier, 'wb') as fh:
            fh.write(bad)
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', bad)], also=[earlier])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('13292.6: the gold it was given on is not valid JSON', r.stdout)
        self.assertEqual(read(self.gold), self.bytes)

    def test_v33_a_tie_on_an_item_no_longer_verified(self):
        stamp = '2026-11-28T10:00:00Z'
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at=stamp)])
        g = load_json(self.gold)
        for it in g['items']:
            it.pop('verified', None)
        self.write_gold(g)
        r = self.run_script([verdict('13292.6', 'none', read(self.gold), at=stamp)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('receipt recorded: 13292.6', r.stdout)
        settled = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at=stamp)], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertEqual(read(self.gold), settled, r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v34_a_claim_that_waits_for_a_decision(self):
        self.agree('13292.6', '13292.33')
        doc = load_json(self.rounds)
        doc['claims']['13292.6']['pending'] = ['Q1']
        doc['claims']['13292.33']['pending'] = ['G2']
        with open(self.rounds, 'wb') as fh:
            fh.write(dump(doc))
        r = self.run_script([verdict('13292.6', 'gold', self.bytes), verdict('13292.33', 'none', self.bytes)])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("not written: 13292.6: its coding waits for the owner's decisions Q1", r.stdout)
        self.assertIn('receipt recorded: 13292.33', r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])
        for bad in ([], [''], 'Q1', [1]):
            doc['claims']['13292.6']['pending'] = bad
            with open(self.rounds, 'wb') as fh:
                fh.write(dump(doc))
            before = read(self.gold)
            r = self.run_script([verdict('13292.6', 'gold', self.bytes)])
            self.assertNotEqual(r.returncode, 0, repr(bad) + ': ' + r.stdout)
            self.assertIn('13292.6: not a claim', r.stderr)
            self.assertEqual(read(self.gold), before)

    def test_v35_a_same_day_gold_on_a_verification_by_hand(self):
        self.agree('13292.6')
        now = self.hand_verify('13292.6', '2026-11-28')
        r = self.run_script([verdict('13292.6', 'gold', now, at='2026-11-28T08:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: already verified so; the verification records only a date (2026-11-28)',
                      r.stdout)
        self.assertEqual(read(self.gold), now)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T12:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verification records only a date (2026-11-28), and a verdict of '
                      'another choice that day', r.stdout)
        self.assertEqual(read(self.gold), now)
        self.assertEqual(read(self.builtin), now)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v36_a_receipt_that_does_not_record_the_verification(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T08:00:00Z')])
        self.assertIn('verified: 13292.6', r.stdout)
        r = self.run_script([verdict('13292.6', 'none', read(self.gold), at='2026-11-28T12:00:00Z')])
        self.assertIn("taken back: 13292.6 (the owner's later verdict: None is right)", r.stdout)
        now = self.hand_verify('13292.6', '2026-11-28')
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T13:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verification records only a date (2026-11-28)', r.stdout)
        self.assertEqual(read(self.gold), now)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v37_a_later_day_after_a_verification_by_hand(self):
        self.agree('13292.6')
        now = self.hand_verify('13292.6', '2026-11-28')
        r = self.run_script([verdict('13292.6', 'gold', now, at='2026-11-29T09:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('verified: 13292.6', r.stdout)
        self.assertEqual(self.items()['13292.6']['verified']['date'], '2026-11-29')
        self.assertEqual([(x['choice'], x['at']) for x in self.receipts('13292.6')], [('gold', '2026-11-29T09:00:00Z')])
        r = self.run_script([verdict('13292.6', 'none', read(self.gold), at='2026-11-29T15:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("taken back: 13292.6 (the owner's later verdict: None is right)", r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v38_a_newer_none_whose_gold_is_not_found(self):
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'none', None, at='2026-11-30T10:00:00Z',
                                     **{'gold.json': sha(b'a gold built elsewhere')})])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('receipt recorded: 13292.6: its verdict is "None is right"; the item remains unverified (given '
                      'on a gold this repository does not hold (not in the working tree or its history): verify it '
                      'again)', r.stdout)
        self.assertEqual(self.receipts('13292.6'),
                         [{'item': '13292.6', 'choice': 'none', 'at': '2026-11-30T10:00:00Z', 'item_rev': None}])
        self.assertEqual(self.validate().returncode, 0)
        settled = read(self.gold)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')], also=[self.original])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verdict is older than the latest verdict the gold records', r.stdout)
        self.assertEqual(read(self.gold), settled)
        self.assertNotIn('verified', self.items()['13292.6'])

    def test_v39_verdicts_whose_gold_cannot_be_placed(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        self.assertIn('verified', self.items()['13292.6'])
        r = self.run_script([verdict('13292.6', 'none', None, at='2026-11-30T10:00:00Z',
                                     **{'gold.json': sha(b'a gold built elsewhere')})])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("taken back: 13292.6 (the owner's later verdict: None is right; given on a gold this repository "
                      'does not hold', r.stdout)
        self.assertNotIn('verified', self.items()['13292.6'])
        entry = load_json(self.gold)['history'][-1]
        self.assertEqual(entry['verification_verdicts'][0]['item_rev'], None)
        self.assertIn('13292.6 (None is right; given on a gold not found, the revision unknown)', entry['what'])
        # On items not verified: a gold given on a gold not held, a none given on a gold that is not valid
        # JSON (found with --also), and a none that recorded no gold.
        bad = self.bytes.replace(b'"format": "wag-annotation/2"',
                                 b'"format": "wag-annotation/2", "format": "wag-annotation/2"', 1)
        earlier = os.path.join(self.dir, 'earlier.json')
        with open(earlier, 'wb') as fh:
            fh.write(bad)
        self.agree('13292.33')
        before = load_json(self.gold)
        r = self.run_script([verdict('13292.33', 'gold', None, **{'gold.json': sha(b'a gold built elsewhere')}),
                             verdict('13292.1', 'none', bad), verdict('13292.30', 'none', None)], also=[earlier])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.33: given on a gold this repository does not hold (not in the working tree '
                      'or its history): verify it again', r.stdout)
        self.assertIn('receipt recorded: 13292.1: its verdict is "None is right"; the item remains unverified (the '
                      'gold it was given on is not valid JSON: verify it again)', r.stdout)
        self.assertIn('receipt recorded: 13292.30: its verdict is "None is right"; the item remains unverified (the '
                      'verdict recorded no gold: verify it again)', r.stdout)
        after = load_json(self.gold)
        self.assertEqual(after['items'], before['items'])
        self.assertEqual(sorted((x['item'], x['item_rev']) for x in after['history'][-1]['verification_verdicts']),
                         [('13292.1', None), ('13292.30', None)])
        self.assertEqual(self.receipts('13292.33'), [])

    def test_v40_two_choices_at_one_time_imported_again(self):
        stamp = '2026-11-28T10:00:00Z'
        self.agree('13292.6')
        r = self.run_script([verdict('13292.6', 'none', self.bytes, at=stamp)])
        self.assertIn('receipt recorded: 13292.6', r.stdout)
        r = self.run_script([verdict('13292.6', 'unsure', self.bytes, at=stamp)], also=[self.original])
        self.assertIn('receipt recorded: 13292.6', r.stdout)
        settled = read(self.gold)
        for choice in ('none', 'unsure', 'none', 'unsure'):
            r = self.run_script([verdict('13292.6', choice, self.bytes, at=stamp)], also=[self.original])
            self.assertEqual(r.returncode, 0, r.stderr)
            self.assertIn('not written: 13292.6: the gold already records this verdict', r.stdout)
            self.assertEqual(read(self.gold), settled)
        self.assertEqual(sorted(x['choice'] for x in self.receipts('13292.6')), ['none', 'unsure'])

    def test_v41_a_malformed_gold(self):
        self.agree('13292.6')
        receipt = {'item': '13292.6', 'choice': 'none', 'at': '2026-11-28T10:00:00Z', 'item_rev': '0' * 64}
        def history(vv):
            return lambda g: g['history'].append({'date': '2026-11-28', 'by': 'x', 'what': 'x',
                                                  'verification_verdicts': vv})
        def verified_as(v):
            def change(g):
                for it in g['items']:
                    if it['kialo_claim_id'] == '13292.6':
                        it['verified'] = v
            return change
        cases = [
            (history(None), 'verification_verdicts is not a list of receipts'),
            (history(5), 'verification_verdicts is not a list of receipts'),
            (history([dict(receipt, item_rev=['x'])]), 'receipt 1, is not a verification receipt'),
            (history([dict(receipt, choice='gold', item_rev=None)]), 'receipt 1, is not a verification receipt'),
            (history([dict(receipt, at='0001-01-01T00:00:00+23:59')]), 'outside the calendar'),
            (verified_as({'by': 'owner', 'date': '2026-11-28', 'item_rev': ['x'], 'text_sha256': '0' * 64}),
             "13292.6: its verification does not have R17's fields"),
            (verified_as('x'), "13292.6: its verification does not have R17's fields"),
            (lambda g: g['items'].append(dict(g['items'][0])), 'has two items for 13292.1'),
        ]
        for change, says in cases:
            g = json.loads(self.bytes)
            change(g)
            self.write_gold(g)
            before = read(self.gold)
            r = self.run_script([verdict('13292.6', 'none', before, at='2026-11-29T10:00:00Z')])
            self.refused(r, says)
            self.assertEqual(read(self.gold), before, says)
            self.assertEqual(read(self.builtin), before, says)

    def test_v42_a_time_outside_the_calendar_and_a_deep_rounds_file(self):
        self.agree('13292.6')
        for at in ('0001-01-01T00:00:00+23:59', '9999-12-31T23:59:59-23:59'):
            r = self.run_script([verdict('13292.6', 'gold', self.bytes, at=at)])
            self.refused(r, 'outside the calendar')
            self.unchanged()
        with open(self.rounds, 'w', encoding='utf-8') as fh:
            fh.write('[' * 100000 + ']' * 100000)
        r = self.run_script([verdict('13292.6', 'gold', self.bytes)])
        self.refused(r, self.rounds)
        self.unchanged()

    def test_v43_a_receipt_of_another_date(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        g = load_json(self.gold)
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['verified']['date'] = '2026-11-30'
        self.write_gold(g)
        now = read(self.gold)
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-30T10:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verification records only a date (2026-11-30)', r.stdout)
        self.assertEqual(read(self.gold), now)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v44_a_receipt_of_another_revision(self):
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at='2026-11-28T10:00:00Z')])
        g = load_json(self.gold)
        for it in g['items']:
            if it['kialo_claim_id'] == '13292.6':
                it['notes']['text'] += ' (recoded)'
                del it['verified']
        self.write_gold(g)
        self.agree('13292.6')  # at its new revision
        now = self.hand_verify('13292.6', '2026-11-28')
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T12:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verification records only a date (2026-11-28)', r.stdout)
        self.assertEqual(read(self.gold), now)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v45_the_none_of_two_at_one_time_is_the_latest(self):
        stamp = '2026-11-28T10:00:00Z'
        self.agree('13292.6')
        self.run_script([verdict('13292.6', 'gold', self.bytes, at=stamp)])
        r = self.run_script([verdict('13292.6', 'none', read(self.gold), at=stamp)])
        self.assertIn('taken back: 13292.6 (a verdict "None is right" at the same time', r.stdout)
        now = self.hand_verify('13292.6', '2026-11-28')
        r = self.run_script([verdict('13292.6', 'none', now, at='2026-11-28T15:00:00Z')])
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn('not written: 13292.6: the verification records only a date (2026-11-28)', r.stdout)
        self.assertEqual(read(self.gold), now)
        self.assertIn('verified', self.items()['13292.6'])

    def test_v46_a_receipt_without_a_field_or_with_one_more(self):
        self.agree('13292.6')
        receipt = {'item': '13292.6', 'choice': 'none', 'at': '2026-11-28T10:00:00Z', 'item_rev': '0' * 64}
        for bad in ({k: x for k, x in receipt.items() if k != 'at'}, dict(receipt, by='owner')):
            g = json.loads(self.bytes)
            g['history'].append({'date': '2026-11-28', 'by': 'x', 'what': 'x', 'verification_verdicts': [bad]})
            self.write_gold(g)
            before = read(self.gold)
            r = self.run_script([verdict('13292.6', 'none', before, at='2026-11-29T10:00:00Z')])
            self.refused(r, f'history entry {len(g["history"])}, receipt 1, is not a verification receipt')
            self.assertEqual(read(self.gold), before)
            self.assertEqual(read(self.builtin), before)

    # V47-V49 (d086-first-gold.md §6): strict input and receipts (ChatGPT's review of exchange 2, patch 01).
    def test_v47_offset_outside_rfc3339_refused(self):
        c = self
        c.agree('13292.33')
        for offset in ('+00:60', '-00:60', '+01:99'):
            with self.subTest(offset=offset):
                result = c.run_script([verdict('13292.33', 'gold', c.bytes,
                                                   at='2026-11-28T12:00:00' + offset)])
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertNotIn('Traceback', result.stderr)
                self.assertIn('timezone offset', result.stderr)
                c.unchanged()

    def test_v48_early_year_receipt_read_back(self):
        c = self
        c.agree('13292.33')
        result = c.run_script([verdict('13292.33', 'none', c.bytes, at='0001-01-01T00:00:00Z')])
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(c.receipts('13292.33')[-1]['at'], '0001-01-01T00:00:00Z')
        result = c.run_script([verdict('13292.33', 'gold', read(c.gold),
                                           at='2026-11-28T10:00:00Z')])
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('verified', c.items()['13292.33'])

    def test_v49_non_finite_number_refused(self):
        c = self
        c.agree('13292.33')
        gold = json.loads(c.bytes)
        for item in gold['items']:
            if item['kialo_claim_id'] == '13292.33':
                item['instances'][0]['scheme']['confidence'] = 1.0
        raw = dump(gold).replace(b'"confidence": 1.0', b'"confidence": 1e999', 1)
        for path in (c.gold, c.builtin):
            Path(path).write_bytes(raw)
        c.agree('13292.33')
        result = c.run_script([verdict('13292.33', 'gold', raw)])
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn('finite', result.stderr)
        self.assertNotIn('Traceback', result.stderr)
        self.assertEqual(read(c.gold), raw)
        self.assertEqual(read(c.builtin), raw)

    def test_v50_integer_beyond_float_refused(self):
        # Go reads every number of the gold as a float64, and refuses one beyond its range.
        c = self
        c.agree('13292.33')
        gold = json.loads(c.bytes)
        for item in gold['items']:
            if item['kialo_claim_id'] == '13292.33':
                item['instances'][0]['scheme']['confidence'] = 1.0
        raw = dump(gold).replace(b'"confidence": 1.0', b'"confidence": 1' + b'0' * 400, 1)
        for path in (c.gold, c.builtin):
            Path(path).write_bytes(raw)
        c.agree('13292.33')
        result = c.run_script([verdict('13292.33', 'gold', raw)])
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn('finite', result.stderr)
        self.assertIn('10000000000', result.stderr)
        self.assertNotIn('Traceback', result.stderr)
        self.assertEqual(read(c.gold), raw)
        self.assertEqual(read(c.builtin), raw)

    def test_v51_non_finite_number_in_verdicts_or_rounds_refused(self):
        c = self
        c.agree('13292.33')
        before = read(c.gold)
        # The verdicts file: a number field of no meaning, read before anything is checked.
        human = os.path.join(c.dir, 'human.json')
        v = verdict('13292.33', 'gold', before)
        v['note'] = 'PLACEHOLDER'
        Path(human).write_bytes(json.dumps([v]).encode('utf-8').replace(b'"PLACEHOLDER"', b'1e999', 1))
        cmd = [sys.executable, '-I', SCRIPT, human, '--gold', c.gold, '--builtin', c.builtin, '--rounds', c.rounds]
        result = subprocess.run(cmd, capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn(human, result.stderr)
        self.assertIn('finite', result.stderr)
        self.assertNotIn('Traceback', result.stderr)
        self.assertEqual(read(c.gold), before)
        # The rounds file: a claim's draft number.
        rounds = read(c.rounds)
        bad = rounds.replace(b'"draft": 1', b'"draft": 1e999', 1)
        self.assertNotEqual(bad, rounds)
        Path(c.rounds).write_bytes(bad)
        result = c.run_script([verdict('13292.33', 'gold', before)])
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn(c.rounds, result.stderr)
        self.assertIn('finite', result.stderr)
        self.assertNotIn('Traceback', result.stderr)
        self.assertEqual(read(c.gold), before)
        self.assertEqual(read(c.builtin), before)


if __name__ == '__main__':
    unittest.main(verbosity=2)
