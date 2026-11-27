"""Tests of verify_gold.py: the cases V1 to V16 of docs/d086-first-gold.md §6, whose outcomes were
written before the tests ran, and two more (the rounds file it reads, and the repository's own).
Each case works on temporary copies of the gold and of the rounds file; the claims' texts and the
other inputs are the repository's.

usage: python3 -I test_verify_gold.py
"""
import hashlib
import json
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
        """The claims agreed in their round at the revision they have in the temporary gold."""
        doc = load_json(self.rounds)
        now = self.items()
        for cid in cids:
            doc['claims'][cid].update(draft=1, review='agreed', agreed_rev=rev or item_rev(now[cid]),
                                      agreed_by='the reviewer and the developer', agreed_date='2026-11-28', objections=[])
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
            self.assertIn(f'not written: {cid}: its verdict is "{word}", not "The gold is right"', r.stdout)
        self.unchanged()

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
        self.assertIn('13292.6: already verified so', r.stdout)
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


if __name__ == '__main__':
    unittest.main(verbosity=2)
