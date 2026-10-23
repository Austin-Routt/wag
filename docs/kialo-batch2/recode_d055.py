"""Recode the batch-2 gold for decisions D055 and D059 (2026-10-23).

    python3 -I recode_d055.py

A one-off migration, kept as the record of what changed and why. It applies
the gold effects of proposal D055-D059, version 5 (docs/proposals/), items
(1)-(4), to gold.json here and copies the result to the annotation engine's
copy (wag/data/annotation/kialo-13292-batch2/gold.json). It refuses to run
on a gold that no longer has the old exception. The rules it applies are
guidelines.md §3.1 (OC1-OC14).

(1) The 16 instances whose exception was other_consequences(A) get the
    sided exception of their scheme, with their own A and G.
(2) 13292.127: the weighing is primary (Other concluding bad_outweighs at
    13292.115's exception, re-parented, cq_answer to 33.1 CQ3); the
    starvation consequence is secondary; 13292.204 becomes a premise attack
    and 13292.202 is retargeted, both to the Other instance's premise.
(3) 13292.43: Other concluding good_outweighs at 13292.9's exception.
(4) The eight 33.x CQ3 challenges at a conclusion are split: the book's
    question at the exception, the game's at the conclusion.
"""
import json
import os
import shutil

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
COPY = os.path.normpath(DOC + '../../wag/data/annotation/kialo-13292-batch2/gold.json')
GOLD = DOC + 'gold.json'
gold = json.load(open(GOLD, encoding='utf-8'))
lex = json.load(open(DOC + 'lexicon.json', encoding='utf-8'))
I = {i['kialo_claim_id']: i for i in gold['items']}
RECODED = {'date': '2026-10-23', 'by': 'decision D055 (proposal D055-D059, version 5; recode_d055.py)'}
EXC_NOTE = ("exception: presumed not to hold unless shown. Since 2026-10-23 the sided exception of decision D055 "
            "(walton-clean/CHANGELOG.md), naming the consequence this argument cites.")


def q(atom):
    return "'" + lex[atom] + "'"


def outweighs(side, a, g):
    word = {'bad': 'bad', 'good': 'good'}[side]
    return (f"{side}_outweighs({a},{g})", f"The action {q(a)} has {word} consequences that outweigh {q(g)}.")


def record(cid, fields, what):
    I[cid]['adjudication'].setdefault('recoded', []).append(dict(RECODED, fields=fields, what=what))


n = sum(p['formula'].startswith('other_consequences(') for it in gold['items'] for ins in it['instances']
        for p in ins['premises'])
if n != 16:
    raise SystemExit(f'expected the 16 instances with other_consequences, found {n}: already recoded?')

# (1) the sided exceptions
for it in gold['items']:
    for ins in it['instances']:
        wid, b = ins['scheme']['walton_id'], ins['bindings']
        for p in ins['premises']:
            if p['role'] == 'exception' and p['formula'].startswith('other_consequences('):
                side = {'positive_consequences': 'bad', 'negative_consequences': 'good'}[wid]
                p['formula'], p['en'] = outweighs(side, b['A'], b['G'])
                p['note'] = EXC_NOTE

# (2) 13292.127: the weighing is the primary instance
it = I['13292.127']
neg, oth = it['instances']
assert neg['scheme']['walton_id'] == 'negative_consequences' and oth['scheme']['walton_id'] == 'other'
neg['primary'], oth['primary'] = False, True
f, en = outweighs('bad', 'tax_meat', 'reduced_unhealthy_meat_consumption')
oth['conclusion'] = {'formula': f, 'en': en}
oth['scheme']['description'] = (
    "Weighing of consequences: a bad consequence of the tax (starvation) is more serious than the harms of high meat "
    "consumption that the pros cite, so the tax's bad consequences outweigh the good one 13292.115 argues: the "
    "exception of that positive consequences argument (D055; guidelines §3.1, OC3, OC4). No scheme concludes the "
    "exception (candidate C003). walton.yml has more_serious_consequences/2 only inside negative_practical (26.2), "
    "which compares the costs of acting on a presumption under ignorance and does not fit.")
it['instances'] = [oth, neg]
it['role'] = {'kind': 'cq_answer', 'confidence': 0.5, 'cq': '33.1 CQ3', 'polarity': 'against'}
it['target']['actual'] = {'claim': '13292.115', 'part': 'exception', 'formula': f, 'instance': 'primary'}
it['target']['re_parented'] = True
for c in it['challenges']:
    if c.get('cq') == '33.2 CQ1':
        c['note'] = 'of the secondary instance, the starvation consequence'
    if c.get('type') == 'severity or value':
        c['note'] = 'the comparison is asserted, not argued; a reply that denies it attacks this premise (OC10)'
it['notes']['text'] = (
    "The claim ranks two consequences: starvation from reduced access to meat is worse than any harm of high meat "
    "consumption, the harms the pros cite. A ranking argues the exception of the argument it weighs against "
    "(guidelines §3.1, OC3, OC4). It names a class, so it is coded against one argument of the class (OC7): its Kialo "
    "parent, 13292.7, is not in it, so the earliest on the board, 13292.115's positive consequences argument; "
    "13292.108, practical reasoning, has no outweighing exception and is left out (D070). The weighing is primary "
    "because the words state the comparison (OC8): Other concluding bad_outweighs(tax_meat,reduced_unhealthy_meat_"
    "consumption), a cq_answer to 33.1 CQ3 at 13292.115's exception, re-parented from 13292.7. The starvation "
    "consequence, which the text presupposes, is a secondary negative consequences instance. Its children engage "
    "the starvation link, a premise of the primary instance: 13292.202 amends the action, 13292.204 denies the link.")
it['notes']['candidate_pattern'] = 'weighing of consequences (C003); a comparison with a class (OC7)'
record('13292.127', ['instances.primary', 'instances.conclusion', 'role', 'target', 'challenges', 'notes'],
       'The weighing became the primary instance, an answer to 33.1 CQ3 at 13292.115\'s new exception; the starvation '
       'consequence is secondary. Before, negative consequences was primary, an independent pro of 13292.7.')

it = I['13292.204']
it['role'] = {'kind': 'premise_attack', 'confidence': 0.6, 'cq': None, 'polarity': 'against'}
it['notes']['text'] = (
    "Denies the starvation link that 13292.127 presupposes: the people at risk do not eat meat, so a meat tax cannot "
    "starve them by reducing their meat access. Since D055 that link is a premise of 13292.127's primary instance, "
    "the weighing (Other), which no question covers, so this is a premise attack (R2). The premise is 127's own "
    "(inferred from 13292.7), so this is not re-parenting (R1).")
record('13292.204', ['role'], "cq_answer CQ:likelihood -> premise_attack: its target premise is now in 13292.127's "
       "Other instance, which no question covers (R2).")

it = I['13292.202']
it['target']['actual'] = {'claim': '13292.127', 'part': 'premise',
                          'formula': 'brings_about(tax_meat,starvation_from_reduced_meat_access)', 'instance': 'primary'}
it['notes']['text'] = (
    "'A only for X': answers 13292.127's starvation objection by restricting the tax's scope, so it targets the "
    "starvation link, a premise of 127's primary instance since D055. Polarity against 127; it still supports a "
    "(restricted) meat tax. Stated as a possibility ('there can be', 'could'). Its sibling 13292.204 instead denies "
    "the starvation link.")
record('13292.202', ['target'], "Retargeted from 13292.127's conclusion to its starvation premise, since 127's "
       "primary instance is now the weighing.")

# (3) 13292.43: a weighing against 13292.9
it = I['13292.43']
f, en = outweighs('good', 'tax_meat', 'tax_revenue_reduced')
it['instances'] = [{
    'primary': True,
    'scheme': {'walton_id': 'other', 'compendium': None, 'wag_id': None, 'fit': 'reconstructed', 'confidence': 0.5,
               'alternatives': [{'walton_id': 'positive_value',
                                 'why_not': 'it ranks two consequences of the tax, and the values it names are things '
                                            'the tax would promote, so they count as consequences (OC3, OC5)'}],
               'description': "Weighing of consequences: what the tax would promote (sustainability, animal welfare) "
                              "presses harder than the lost revenue 13292.9 cites, so the tax's good consequences "
                              "outweigh it: the exception of that negative consequences argument (D055; guidelines "
                              "§3.1, OC4, OC5). No scheme concludes the exception (candidate C003)."},
    'bindings': {},
    'premises': [
        {'role': 'premise', 'formula': 'brings_about(tax_meat,sustainability_and_animal_welfare)',
         'en': f"The action {q('tax_meat')} brings about {q('sustainability_and_animal_welfare')}.", 'status': 'explicit',
         'note': 'these issues, not revenue, are the reasons to tax meat'},
        {'role': 'premise', 'formula': 'more_pressing(sustainability_and_animal_welfare,tax_revenue_reduced)',
         'en': f"{q('sustainability_and_animal_welfare')} presses harder than {q('tax_revenue_reduced')}.",
         'status': 'explicit', 'note': '"other, more pressing issues"'},
    ],
    'conclusion': {'formula': f, 'en': en},
}]
it['role'] = {'kind': 'cq_answer', 'confidence': 0.5, 'cq': '33.2 CQ3', 'polarity': 'against'}
it['target']['actual'] = {'claim': '13292.9', 'part': 'exception', 'formula': f, 'instance': 'primary'}
it['target']['re_parented'] = False
it['challenges'] = [
    {'cq': None, 'wag_cq': None, 'type': 'efficacy', 'formula': 'brings_about(tax_meat,sustainability_and_animal_welfare)',
     'note': 'does the tax actually serve sustainability and animal welfare? (cf. 13292.13)'},
    {'cq': None, 'wag_cq': None, 'type': 'severity or value',
     'formula': 'more_pressing(sustainability_and_animal_welfare,tax_revenue_reduced)',
     'note': "'more pressing' than revenue is asserted, not argued"},
]
it['notes']['text'] = (
    "Answers 33.2 CQ3 of 13292.9 against it: the tax serves sustainability and animal welfare, which press harder than "
    "the lost revenue. It ranks two consequences, so it argues the exception of 13292.9's argument, "
    "good_outweighs(tax_meat,tax_revenue_reduced), with Other (guidelines §3.1, OC3, OC4). The values it names are "
    "things the tax would promote, so they count as consequences (OC5). Its Kialo parent cites the outweighed "
    "consequence, so it is not re-parented.")
it['notes']['candidate_pattern'] = 'weighing of consequences (C003)'
record('13292.43', ['instances', 'role.cq', 'target', 'challenges', 'notes'],
       'Was an argument from positive value answering CQ:opposite_consequences at 13292.9\'s conclusion; now Other '
       'concluding good_outweighs at 13292.9\'s exception, an answer to 33.2 CQ3.')

# (4) the 33.x CQ3 challenges at a conclusion: the book's question moves to the exception
EXC_SIDE = {  # notes that describe a weighing go with the book's question
    '13292.7': 'the thesis\'s pros may outweigh it',
    '13292.9': '13292.43: the tax serves more pressing goods (sustainability, animal welfare)',
}
n = 0
for it in gold['items']:
    out = []
    for c in it['challenges']:
        if c.get('cq') in ('33.1 CQ3', '33.2 CQ3') and (c.get('formula') or '').lstrip('¬').startswith('should_be_performed('):
            ins = [x for x in it['instances'] if x['conclusion'].get('formula') == c['formula']
                   and x['scheme']['walton_id'] in ('positive_consequences', 'negative_consequences')][0]
            exc = [p['formula'] for p in ins['premises'] if p['role'] == 'exception'][0]
            book = {'cq': c['cq'], 'wag_cq': None, 'type': 'opposite consequences', 'formula': exc}
            game = {'cq': None, 'wag_cq': c['wag_cq'], 'type': c['type'], 'formula': c['formula']}
            cid = it['kialo_claim_id']
            if cid in EXC_SIDE:
                book['note'] = EXC_SIDE[cid]
            elif c.get('note'):
                game['note'] = c['note']
            game['note'] = (game.get('note', '') + '; ' if game.get('note') else '') + \
                'the game asks it at the conclusion, as a rebuttal, until workstream E'
            out += [book, game]
            n += 1
        else:
            out.append(c)
    it['challenges'] = out
if n != 8:
    raise SystemExit(f'expected 8 challenges to split, found {n}')
for cid in ('13292.7', '13292.9', '13292.115', '13292.121', '13292.147', '13292.180', '13292.194'):  # .127 above
    record(cid, ['challenges'], "33/CQ3's challenge split: the book's question at the sided exception (D055), the "
           "game's at the conclusion.")

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
shutil.copyfile(GOLD, COPY)
print('recoded: 16 exceptions, 13292.127, .204, .202, .43, 8 challenges split; copied to', os.path.relpath(COPY, DOC))
