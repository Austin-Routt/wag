"""Recode the batch-2 gold for proposal D008-D074, version 5 (stage 2, item C3).

    python3 -I recode_c3.py

A one-off migration, kept as the record of what changed and why. It applies
the gold effects of proposal D008-D074 (docs/proposals/) to gold.json here
and copies it to the annotation engine's copy
(wag/data/annotation/kialo-13292-batch2/). It refuses to run on a gold that
no longer has the formulas it was written for. English renderings come from
walton.yml's templates (with the proposal applied) and lexicon.json, as the
gold's were made.

(1) The 16 consequence instances list the new exceptions, status supplied:
    too_slight and the bearer exception of their side (D065:
    gain_discounted_by_bearer for A, harm_discounted_by_bearer against it).
    D074's objection is an undercutter of the specific argument, not an
    exception (the owner, 2026-11-25), so no instance lists it.
(2) The mood rule (D008): 13292.174 ('might'), 13292.121 ('could also result
    in') and 13292.9 ('could actually') claim their consequence only
    possible. Each instance keeps its id and becomes the modal variant,
    secondary; each claim gains a bridge instance, primary, whose conclusion
    is the old one, so its role and target stand; the bridge's exception is
    worth_finding_out_first (D008, as the owner refined it).
(3) The claims that answer them follow: at the causal premise, the modal
    premise and the book's question (13292.12, 13292.233, 13292.180's second
    parent); 13292.235 at the bridge's conclusion. Their notes, and 13292.33's,
    get a dated addendum where the recode makes them stale.
(4) OC9 (D065): 13292.25 argues that the burden 13292.7 cites is too slight,
    not that it is not bad: an undercutter at too_slight.
(5) Challenge records follow: on the recoded instances, the likelihood and
    evidence records name may_bring_about with the book's question only, the
    opposite-consequences record the bridge's conclusion. The records of how
    serious a consequence is or on whom it falls are judged one by one by
    what they ask (OC9, OC16): two that ask how much name too_slight, one that
    asks on whom the bearer exception; the others, among them a value
    objection (13292.9#1's), stay as they are.
(6) 13292.127's note says why its starvation link stays categorical.
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
I = {i['kialo_claim_id']: i for i in gold['items']}
WHEN = '2026-11-25'
BY = 'proposal D008-D074, version 5 (recode_c3.py)'
for k in ('may_bring_about/2', 'matters_to_decision/2', 'too_slight/2', 'harm_discounted_by_bearer/2'):
    if k not in LANG:
        raise SystemExit(f'{WALTON} has no template {k}: apply the proposal to walton.yml first')


def nows(f):
    return re.sub(r'\s+', '', f or '')


def parse(f):
    m = re.fullmatch(r'(¬?)([a-z_0-9]+)\((.*)\)', nows(f))
    if not m:
        raise SystemExit(f'cannot parse {f!r}')
    return m.group(1), m.group(2), m.group(3).split(',')


def en(f):
    neg, pred, args = parse(f)
    text = LANG[f'{pred}/{len(args)}'] % tuple("'" + lex[a] + "'" for a in args)
    if neg:
        raise SystemExit(f'negated formula {f!r}: not expected here')
    return text


def subst(f, b):
    return re.sub(r'\b([A-Z][A-Z0-9]*)\b', lambda m: b.get(m.group(1), m.group(1)), f)


def since(note, text):
    """An earlier note, with what changed on WHEN after it."""
    base = (note or '').rstrip('.')
    sep = ' ' if base.endswith(('?', '!')) else '. '
    return (base + sep if note else '') + f'Since {WHEN}: {text}'


def record(cid, fields, what):
    I[cid]['adjudication'].setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': fields, 'what': what})


EXC_NOTE = ('exception: presumed not to hold unless shown. Since 2026-11-25 (decision D065, proposal D008-D074): '
            'how serious the consequence is and on whom it falls are exceptions of the consequence schemes, the '
            'critic\'s to support.')
NEW_EXC = {'positive_consequences': ['too_slight(A,G)', 'gain_discounted_by_bearer(A,G)'],
           'negative_consequences': ['too_slight(A,G)', 'harm_discounted_by_bearer(A,G)']}
MODAL = {'13292.174': ("'might force': a possibility", 'the bridge to 13292.108\'s conclusion'),
         '13292.121': ("'could also result in': a possibility", 'the bridge to 13292.115\'s conclusion'),
         '13292.9': ("'could actually': a possibility", 'the bridge to the thesis')}

# Guards: the gold is the one this recode was written for.
CONSEQ = [(it['kialo_claim_id'], ins) for it in gold['items'] for ins in it['instances']
          if ins['scheme']['walton_id'] in NEW_EXC]
if len(CONSEQ) != 16:
    raise SystemExit(f'expected 16 consequence instances, found {len(CONSEQ)}: already recoded?')
for cid, ins in CONSEQ:
    if any(p['formula'].startswith('too_slight(') for p in ins['premises']):
        raise SystemExit(f'{cid}: already recoded')
for cid in MODAL:
    first = I[cid]['instances'][0]
    if first['id'] != cid + '#1' or first['scheme']['walton_id'] != 'negative_consequences' or not first['primary']:
        raise SystemExit(f'{cid}: not the categorical instance it was, or already recoded')

# (1) The new exceptions, after each instance's own; one record per item.
done = []
for cid, ins in CONSEQ:
    sch = ins['scheme']['walton_id']
    for f in NEW_EXC[sch]:
        g = nows(subst(f, ins['bindings']))
        ins['premises'].append({'role': 'exception', 'formula': g, 'en': en(g), 'status': 'supplied', 'note': EXC_NOTE})
    if cid not in done:
        done.append(cid)
        record(cid, ['instances'], 'the new exceptions of its consequence scheme (D065), supplied')

# (2) The modal variant and its bridge.
for cid, (hedge, label) in MODAL.items():
    it = I[cid]
    ins = it['instances'][0]
    b = ins['bindings']
    old_concl = ins['conclusion']['formula']
    ins['scheme']['walton_id'] = 'negative_consequences_modal'
    ins['scheme']['wag_id'] = None
    ins['scheme']['alternatives'] = [{'walton_id': 'negative_consequences', 'why_not': (
        hedge + ': the categorical premise brings_about(A,G) would claim more than the claim does (D008)')}]
    for p in ins['premises']:
        if p['formula'].startswith('brings_about('):
            p['formula'] = nows(subst('may_bring_about(A,G)', b))
            p['en'] = en(p['formula'])
            p['note'] = since(p.get('note'), 'the modal premise of negative_consequences_modal (D008)')
    c = nows(subst('may_be_unwise_to_perform(A,G)', b))
    ins['conclusion'] = {'formula': c, 'en': en(c)}
    ins['primary'] = False
    ins['label'] = 'the possible consequence, which the bridge carries to the target (D008)'
    # The next number, past the retired ids too (an id is not given again).
    n = 1 + max(int(x.split('#')[1]) for x in [y['id'] for y in it['instances']] + list(it.get('retired') or []))
    bid = f'{cid}#{n}'
    bridge = {
        'id': bid, 'primary': True,
        'scheme': {'walton_id': 'bridge_from_possible_harm', 'compendium': None, 'wag_id': None, 'fit': 'reconstructed',
                   'confidence': ins['scheme']['confidence'], 'alternatives': []},
        'label': label + ' (D008)',
        'bindings': dict(b),
        'premises': [
            {'role': 'premise', 'formula': c, 'en': en(c), 'status': 'supplied',
             'note': f'the conclusion of this claim\'s own modal argument, {ins["id"]}'},
            {'role': 'premise', 'formula': nows(subst('matters_to_decision(A,G)', b)),
             'en': en(nows(subst('matters_to_decision(A,G)', b))), 'status': 'supplied',
             'note': 'the claim offers the possibility as a reason against the tax'},
            {'role': 'exception', 'formula': nows(subst('worth_finding_out_first(A,G)', b)),
             'en': en(nows(subst('worth_finding_out_first(A,G)', b))), 'status': 'supplied',
             'note': 'exception: presumed not to hold unless shown'},
        ],
        'conclusion': {'formula': old_concl, 'en': "'governments put a tax on meat' should not be done."},
    }
    if old_concl != '¬should_be_performed(tax_meat)':
        raise SystemExit(f'{cid}: unexpected conclusion {old_concl}')
    it['instances'].append(bridge)
    for ch in it['challenges']:
        if ch.get('instance') != ins['id']:
            continue
        f = nows(ch.get('formula'))
        if f.startswith('brings_about('):
            ch['formula'] = nows(subst('may_bring_about(A,G)', b))
            ch['wag_cq'] = None
            ch['note'] = since(ch.get('note'), 'the modal premise, with the book\'s question; the game asks nothing '
                                                'of the modal variant until step C5')
        elif f == old_concl:
            ch['instance'] = bid
            ch['wag_cq'] = None
            ch['note'] = since(ch.get('note'), 'at the bridge\'s conclusion, which meets the target (D008); the game '
                                                'asks nothing of the bridge until step C5, so the record has no game '
                                                'question')
        elif ch.get('type') == 'other means' and ch.get('formula') is None:
            ch['instance'] = bid
            ch['note'] = since(ch.get('note'), 'about the bridge, whose conclusion the amended action answers (D008)')
    it['notes']['text'] = since(it['notes'].get('text'), (
        f'{hedge.split(":")[0]} claims the consequence only possible (the mood rule, D008): {ins["id"]} is the modal '
        f'variant, secondary, and the bridge {bid}, primary, carries its conclusion to the target, so the role and '
        'the target stand.'))
    record(cid, ['instances', 'challenges', 'notes'],
           'the modal variant and its bridge (D008): the consequence is claimed only possible')

# (3) The claims that answer them.
for cid, tc in (('13292.12', '13292.9'), ('13292.233', '13292.121')):
    a = I[cid]['target']['actual']
    if a.get('instance') != tc + '#1' or not a['formula'].startswith('brings_about('):
        raise SystemExit(f'{cid}: not the target it was')
    a['formula'] = 'may_bring_about' + a['formula'][len('brings_about'):]
    r = I[cid]['role']
    r['cq'] = {'CQ:likelihood': '33.2 CQ1', 'CQ:evidence': '33.2 CQ2'}[r['cq']]
    record(cid, ['role.cq', 'target'], 'its target is the modal premise of ' + tc + '#1, with the book\'s question (D008)')
o = [x for x in I['13292.180']['target']['other_kialo_parents'] if x.get('instance') == '13292.121#1'][0]
o['formula'] = 'may_bring_about' + o['formula'][len('brings_about'):]
o['cq'] = {'CQ:likelihood': '33.2 CQ1', 'CQ:evidence': '33.2 CQ2'}[o['cq']]
record('13292.180', ['target'], 'its second parent\'s target is the modal premise of 13292.121#1, with the book\'s '
       'question (D008)')
a = I['13292.235']['target']['actual']
if a.get('instance') != '13292.121#1' or a.get('part') != 'conclusion':
    raise SystemExit('13292.235: not the target it was')
a['instance'] = [x['id'] for x in I['13292.121']['instances'] if x['primary']][0]
record('13292.235', ['target'], 'its target is the conclusion of 13292.121\'s bridge, which meets 13292.115 (D008)')
ADDENDA = {
    '13292.12': ('13292.9 is now coded as claiming its consequence only possible (D008): the target is 13292.9#1, its modal '
                 'argument, now secondary (the bridge 13292.9#4 is primary), and its premise may_bring_about(A,G), '
                 'which the game does not ask until step C5: the question is the book\'s, 33.2 CQ1.'),
    '13292.233': ('13292.121 is now coded as claiming its consequence only possible (D008): the target is the premise '
                  'may_bring_about(tax_meat,unhealthy_meat_underconsumption) of 13292.121#1, its modal argument, and '
                  'the question is the book\'s, 33.2 CQ1 (the game asks nothing of the modal variant until step C5).'),
    '13292.180': ('toward 13292.121 it backs the premise may_bring_about(tax_meat,unhealthy_meat_underconsumption) of '
                  '13292.121#1, the modal argument (D008), with the book\'s question 33.2 CQ1.'),
    '13292.235': ('its target is the conclusion of 13292.121\'s bridge, 13292.121#2 (D008); the alternative coding '
                  'would answer 33.2 CQ1 at the modal premise of 13292.121#1, not CQ:likelihood.'),
    '13292.33': ('a claim that the consequence would be mild argues too_slight(A,G), not bad(G) (OC9, D065), as its '
                 'sibling 13292.25 now does; this claim still answers the likelihood question.'),
}
for cid, text in ADDENDA.items():
    I[cid]['notes']['text'] = since(I[cid]['notes'].get('text'), text)
    mine = [r for r in I[cid]['adjudication'].get('recoded', []) if r.get('by') == BY]
    if mine:
        mine[-1]['fields'].append('notes')
    else:
        record(cid, ['notes'], 'a dated addendum where the recode makes its note stale (D008, D065)')

# (4) 13292.25: too slight, not denied.
it = I['13292.25']
ins = [x for x in it['instances'] if x['primary']][0]
if ins['conclusion']['formula'] != '¬bad(food_less_affordable_for_poor)' or it['role']['kind'] != 'premise_attack':
    raise SystemExit('13292.25: not the coding it was')
c = 'too_slight(tax_meat,food_less_affordable_for_poor)'
ins['conclusion'] = {'formula': c, 'en': en(c)}
ins['scheme']['description'] = (
    'Mitigation by adaptation: those affected can avoid most of a cited cost by a cheap change of behaviour (their '
    'diet), so the cited consequence is too slight to count, not denied (D065). No walton.yml scheme concludes '
    'too_slight(A,G) from a low-cost way of avoiding G; the parts use walton.yml predicates (means_to_end/2, '
    'easy_to_achieve/1, too_slight/2).')
it['role'] = {'kind': 'undercutter', 'confidence': it['role']['confidence'], 'cq': None, 'polarity': 'against'}
it['target']['actual'] = {'claim': '13292.7', 'instance': '13292.7#1', 'part': 'exception', 'formula': c}
it['notes']['text'] = since(it['notes'].get('text'), (
    'it says the burden is small, not that it is not bad: an undercutter at 13292.7#1\'s exception too_slight (OC9, '
    'D065). Its own words hedge ("could be as small as"); fit and confidence say so.'))
record('13292.25', ['instances', 'role', 'target', 'notes'], 'too slight, not denied: an undercutter at too_slight (D065)')

# (5) Records of how serious a consequence is, or on whom it falls, judged
# one by one (proposal version 4, from the check of version 3, E1): two ask
# how much, one on whom. The others stay, among them 13292.9#1's ("revenue is
# not the purpose of the tax", a value objection, at bad(G)) and the backings
# of 13292.115#1 and 13292.12#3.
# What makes the amount asked about bear on the decision (the owner's scope of
# too_slight, D053: not merely few or small in number).
SCOPE = {'13292.5#1': ('; the amount asked about bears on the decision as the share of the tax\'s environmental '
                       'aim that deterring regenerative farms would cost, not as how few farms it reaches '
                       '(the owner\'s scope, D053)')}
SEVERITY = {('13292.7', '13292.7#1'): ('too_slight', 'severity'), ('13292.5', '13292.5#1'): ('too_slight', 'severity'),
            ('13292.174', '13292.174#1'): ('harm_discounted_by_bearer', 'bearer')}
for (cid, iid), (pred, typ) in SEVERITY.items():
    ins = [x for x in I[cid]['instances'] if x['id'] == iid][0]
    recs = [ch for ch in I[cid]['challenges'] if ch.get('instance') == iid and ch.get('type') == 'severity or value']
    if len(recs) != 1:
        raise SystemExit(f'{cid}: {len(recs)} severity records')
    ch = recs[0]
    ch['formula'] = nows(subst(pred + '(A,G)', ins['bindings']))
    ch['type'] = typ
    ch['note'] = since(ch.get('note'), (
        'how serious the consequence is has its own exception, too_slight (D065)' if typ == 'severity' else
        'on whom the harm falls has its own exception, harm_discounted_by_bearer (D065)') + SCOPE.get(iid, ''))
    mine = [r for r in I[cid]['adjudication']['recoded'] if r.get('by') == BY]
    if 'challenges' not in mine[-1]['fields']:
        mine[-1]['fields'].append('challenges')

# (6) 13292.127: why its starvation link stays categorical (the check of
# version 3, N5).
n127 = I['13292.127']['notes']
n127['text'] = since(n127.get('text'), (
    "the mood rule (D008) leaves its link categorical: 'a more severe threat than' ranks how bad starvation is, and "
    "its link from the tax to starvation is its own, inferred from 13292.7's firm 'would make food more expensive' "
    "(cited)."))
mine = [r for r in I['13292.127']['adjudication'].get('recoded', []) if r.get('by') == BY]
if not mine:
    raise SystemExit('13292.127: its new exceptions were not recorded')
mine[-1]['fields'].append('notes')
mine[-1]['what'] += '; its note says why its link stays categorical (D008)'

# Written as the gold is (indent 1, no newline at the end).
with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
print('recoded: the new exceptions of 16 consequence instances; 13292.174, .121 and .9 (modal, bridged); '
      '13292.12, .233, .180 and .235; 13292.25; the severity and bearer records')
