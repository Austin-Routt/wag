"""Round 1 of the first gold, draft 1 (D086; docs/d086-first-gold.md §3 and §10).

    python3 -I recode_d086_r1.py

A one-off, kept as the record. It writes the developer's draft of round 1 (d086-rounds.json) into the
gold, each item with an adjudication.recoded entry (by "developer (D086 draft 1)") that names what
changed and why, or says the reference's coding is kept and why. Corrected on 2026-11-27, before
any review, after an independent check of its first version (6e9b28c; the workspace's
verify/d086/DRAFT1-CHECK.md) and an independent check of those corrections (b888983;
verify/d086/FIXCHECK-draft1.md): it is run again from the reference, so that draft 1 is one record.

- Changed. The targets that format 1's rule of a primary instance's formulas (adjudication.md R5)
  had kept from what they answer name it: 13292.12 at 13292.9#2's causes(tax_meat,
  meat_industry_harmed) and 13292.141 at 13292.2#2's causes(meat_production,
  significant_ghg_emissions), both the book's 28 CQ1 (a mechanism offered for the link); 13292.182
  and 13292.198 at 13292.180#2's exception interference(reduced_meat_consumption), 28 CQ3, their
  own arguments concluding it (R8); 13292.202, a mitigation, at the conclusion of the argument that
  cites the harm it mitigates, 13292.127#2 (OC6), where both annotators aimed it. The true fragments
  13292.6, 13292.11 and 13292.218 follow guidelines §5 item 6: they bind only the variables of the
  formulas they state, the others {"unstated": why}, list those formulas only, and their
  conclusions, which name the action, are null; the challenge records at formulas they no longer
  list are left to the arguments they are fragments of, which record the same challenges.
- Joined (plan §3: an item a draft changes joins the round). 13292.9, 13292.2 and 13292.180 named
  the retargeted claims at their old places in their challenge records: those records are put
  right, the missing ones at the causal steps added, and the items are drafted with the round.
- Kept. 13292.108 keeps every variable bound: the thesis has no argument, so the fragment carries
  the environmental argument, and §5 item 6 would leave seven of the eleven targets into it nothing
  to point at (a question for the review and the owner). The other items keep the reference's
  coding, which the developer's reading of each matched or, where it did not, the guidelines decide
  for the reference (13292.8, .36, .46) or the question goes to the review (13292.194's target)
  (workspace verify/d086/round1-readings.md, round2-readings.md).
- The nine causal steps of these items whose confidence the N8 note capped "until the first gold's
  draft" have it set on their merits: both annotators coded the same nine steps (with cause_to_effect),
  and §5 item 11 prescribes the modal form, so 0.55, but 0.5 for 13292.192#2, whose link the text
  states only as a possibility. Their notes say whether the cause is the proposed tax or an effect
  predicted to follow from it.
- Notes. Where a note, or a challenge record's note, told the coding's history in dated additions,
  described a coding the item no longer has, or said that the game asks a question at the conclusion
  only "until workstream E" (it asks it there by design since step C5.2, the rebuttal route), it is
  rewritten as one current account; the history is in the item's recoded entries and in
  gold-format1.json. So are three premise notes (13292.127#2's, 13292.174#1's, 13292.108#1's); the
  boilerplate of exception and assumption notes, which names the decision a premise comes from
  ("Since 2026-11-24: ..."), is kept, as everywhere in the gold. 13292.147 names candidate C017.

It refuses any gold but the one the rounds were measured on (d086-rounds.json's reference), then
writes gold.json, sets each claim's draft to 1 and its review to "not reviewed" in d086-rounds.json
(the joined ones marked so), checks the gold with validate.py, and copies it to the built-in copy
(wag/data/annotation/kialo-13292-batch2/gold.json).
"""
import hashlib
import json
import os
import shutil
import subprocess
import sys

DOC = os.path.dirname(os.path.abspath(__file__)) + '/'
REPO = os.path.normpath(DOC + '../..') + '/'
COPY = REPO + 'wag/data/annotation/kialo-13292-batch2/'
GOLD = DOC + 'gold.json'
ROUNDS = DOC + 'd086-rounds.json'
WHEN = '2026-11-27'
BY = 'developer (D086 draft 1)'


def refuse(why):
    raise SystemExit('recode_d086_r1: ' + why + ': nothing written')


raw = open(GOLD, 'rb').read()
rounds = json.load(open(ROUNDS, encoding='utf-8'))
if hashlib.sha256(raw).hexdigest() != rounds['measure']['reference_sha256']:
    refuse('gold.json is not the reference the rounds were measured on (run it once, on that gold)')
gold = json.loads(raw)
G = {it['kialo_claim_id']: it for it in gold['items']}
ROUND1 = rounds['rounds'][0]['claims']
JOINED = {
    '13292.9': 'its challenge record at the modal premise named 13292.12, which backs its causal step 13292.9#2 now',
    '13292.2': 'its challenge record at the classification premise named 13292.141, which backs its causal step 13292.2#2 now',
    '13292.180': 'its challenge record at the consequence premise named 13292.182 and .198, which answer its causal step 13292.180#2 now',
}
if any(rounds['claims'][c]['review'] != 'not drafted' for c in list(ROUND1) + list(JOINED)):
    refuse('a claim of round 1, or one joining it, is drafted already')

done = {}


def record(cid, fields, what):
    adj = G[cid].setdefault('adjudication', {'changed_fields': [], 'reason': 'agreed'})
    adj.setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': sorted(fields), 'what': what})
    done[cid] = True


def instance(cid, iid):
    for ins in G[cid]['instances']:
        if ins['id'] == iid:
            return ins
    refuse(f'{cid} has no instance {iid}')


def set_target(cid, claim, iid, part, formula):
    tgt = G[cid]['target']['actual']
    tgt.update({'claim': claim, 'instance': iid, 'part': part, 'formula': formula})
    want = instance(claim, iid)
    if part == 'conclusion':
        if (want.get('conclusion') or {}).get('formula') != formula:
            refuse(f'{formula} is not the conclusion of {iid}')
    elif formula not in [p['formula'] for p in want['premises']]:
        refuse(f'{formula} is not a formula of {iid}')


def set_note(cid, iid, formula, cq, note):
    """Rewrite the note of the one challenge record with this instance, formula and question."""
    hits = [c for c in G[cid]['challenges'] if c.get('instance') == iid and c.get('formula') == formula and c.get('cq') == cq]
    if len(hits) != 1:
        refuse(f'{cid} has {len(hits)} challenge records at {iid} {formula} {cq}')
    hits[0]['note'] = note


def add_challenge(cid, rec):
    ins = instance(cid, rec['instance'])
    if rec['formula'] not in [p['formula'] for p in ins['premises']]:
        refuse(f'{rec["formula"]} is not a formula of {rec["instance"]}')
    if any(c.get('instance') == rec['instance'] and c.get('formula') == rec['formula'] and c.get('cq') == rec['cq'] for c in G[cid]['challenges']):
        refuse(f'{cid} has a record at {rec["instance"]} {rec["formula"]} {rec["cq"]} already')
    G[cid]['challenges'].append(rec)


def lift_cap(cid, iid, note, conf=0.55):
    """The N8 note capped a hypothetical causal step's confidence "until the first gold's draft"."""
    ins = instance(cid, iid)
    if 'until the first gold' not in (ins.get('note') or '') or ins['scheme']['confidence'] > 0.5:
        refuse(f'{iid} is not a capped causal step')
    stated = [p for p in ins['premises'] if p['formula'].startswith('causes(')]
    if not stated or stated[0]['status'] != 'explicit':
        refuse(f'{iid} does not state its causal link')
    ins['scheme']['confidence'] = conf
    ins['note'] = note


def set_premise_note(cid, iid, formula, note):
    hits = [p for p in instance(cid, iid)['premises'] if p['formula'] == formula]
    if len(hits) != 1:
        refuse(f'{iid} has {len(hits)} premises {formula}')
    hits[0]['note'] = note


MODAL_STEP = 'A causal step whose cause is only proposed, the un-enacted tax (cause_to_effect_modal, §5 item 11).'
MODAL_EFFECT = ('A causal step whose cause is only possible, itself predicted to follow from the un-enacted tax '
                '(cause_to_effect_modal, §5 item 11).')
WORKSTREAM_E = ("without the remark that the game asks its question at the conclusion only 'until workstream E': it "
                "asks it there by design, the rebuttal route beside the exception's question (D055; D095 (1))")
MERITS = 'both annotators coded the same steps (with cause_to_effect), and §5 item 11 prescribes the modal form'
LIFTED = ('the confidence of its causal steps set on their merits, 0.55: ' + MERITS + '; this lifts the cap the N8 note '
          'put on them until this draft, and their notes now say why the scheme is modal')
LIFTED_ONE = ('the confidence of its causal step set on its merits, 0.55: ' + MERITS.replace('steps', 'step') + '; this '
              'lifts the cap the N8 note put on it until this draft, and its note now says why the scheme is modal')
LIFTED_192 = ("the confidence of its causal steps set on their merits: 0.55 for 13292.192#1, whose link the text states "
              "firmly ('would require'), and 0.5 for 13292.192#2, whose link it states only as a possibility ('may still "
              "occur'), which the scheme's firm link does not express; " + MERITS + "; this lifts the cap the N8 note put "
              "on them until this draft, and their notes now say why the scheme is modal")
NO_REPLY = 'no claim of the batch argues it'


# --- Changed: the targets format 1 kept from what they answer -------------------------------------

# 13292.12: the mechanism of 13292.9's first step.
set_target('13292.12', '13292.9', '13292.9#2', 'premise', 'causes(tax_meat,meat_industry_harmed)')
G['13292.12']['role'].update({'kind': 'cq_answer', 'cq': '28 CQ1', 'polarity': 'for'})
lift_cap('13292.12', '13292.12#1', MODAL_EFFECT + ' A causal scheme because the claim\'s function is the causal mechanism, which no other walton.yml scheme expresses.')
lift_cap('13292.12', '13292.12#2', MODAL_STEP)
set_note('13292.12', '13292.12#1', 'interference(imported_meat_relatively_cheaper)', '28 CQ3',
         'the import measures 13292.45 and 13292.194 propose would block it; they are coded as amended actions at '
         '13292.12\'s predicted harm (OC6), not as answers here')
record('13292.12', ['challenges', 'instances.note', 'instances.scheme.confidence', 'notes', 'role', 'target'],
       "Retargeted from the modal premise of 13292.9#1 to the causal step it backs, 13292.9#2's "
       "causes(tax_meat,meat_industry_harmed), as both annotators aimed it: format 1 required a formula of the primary "
       "instance (adjudication.md R5), which put the target at 13292.9's consequence premise, and the recoding for D008 "
       "kept that choice, at the modal argument. A cq_answer to the book's 28 CQ1 there (a mechanism offered for the "
       "link; the game asks no question at the formula; guidelines §5 item 3), polarity for. " + LIFTED[0].upper() + LIFTED[1:] +
       ". Its record at 13292.12#1's interference no longer says that 13292.45 and .194 answer there (they are amended "
       "actions at 13292.12's predicted harm). A cascade in the replay now. Notes rewritten as one current account.")

# 13292.141: a mechanism for 13292.2's causal step.
set_target('13292.141', '13292.2', '13292.2#2', 'premise', 'causes(meat_production,significant_ghg_emissions)')
G['13292.141']['role'].update({'kind': 'cq_answer', 'cq': '28 CQ1', 'polarity': 'for'})
record('13292.141', ['notes', 'role', 'target'],
       "Retargeted from 13292.2#1's classification premise instance(meat_production,significant_ghg_source) (14 CQ1; "
       "format 1 required a formula of the primary instance) to the causal step it backs, 13292.2#2's "
       "causes(meat_production,significant_ghg_emissions), as both annotators aimed it; a cq_answer to the book's 28 CQ1 "
       "there (a mechanism, energy use, offered for the link; it cites no data), polarity for, as 13292.2's own record of "
       "the question has it. A cascade in the replay now. Notes rewritten.")

# 13292.182 and 13292.198: a substitute counteracts 13292.180's causal step.
step = instance('13292.180', '13292.180#2')
interference = next(p for p in step['premises'] if p['role'] == 'exception')
if interference['formula'] != 'interference(reduced_meat_consumption)':
    refuse('13292.180#2 has another exception than interference(reduced_meat_consumption)')
DESCRIPTION = ("Mitigation by substitution: an available substitute gives people enough iron without meat, so eating "
               "less meat need not bring more anemia: it counteracts the causal step from reduced meat consumption to "
               "anemia. No walton.yml scheme concludes the interference from an available substitute; the parts use "
               "walton.yml predicates.")
for cid in ('13292.182', '13292.198'):
    own = instance(cid, cid + '#1')
    own['scheme']['description'] = DESCRIPTION
    own['conclusion'] = {'formula': interference['formula'], 'en': interference['en']}
    set_target(cid, '13292.180', '13292.180#2', 'exception', interference['formula'])
    G[cid]['role'].update({'kind': 'cq_answer', 'cq': '28 CQ3', 'polarity': 'against'})
    record(cid, ['instances.conclusion', 'instances.scheme.description', 'notes', 'role', 'target'],
           "Retargeted from 13292.180#1's premise brings_about(tax_meat,more_anemia_cases) (CQ:likelihood; format 1 "
           "required a formula of the primary instance) to the causal step it answers, the exception "
           "interference(reduced_meat_consumption) of 13292.180#2, the book's 28 CQ3 (other factors that counteract the "
           "production of the effect; the game asks no question there), as both annotators aimed it (A an undercutter, "
           "B a cq_answer to 28 CQ3); its Other instance concludes that exception now (R8), and its description says how "
           "a substitute counteracts the step. A cascade in the replay now. Notes rewritten.")

# 13292.202: a mitigation, at the conclusion of the argument that cites the harm.
set_target('13292.202', '13292.127', '13292.127#2', 'conclusion', '¬should_be_performed(tax_meat)')
record('13292.202', ['notes', 'target'],
       "Retargeted from the starvation link of 13292.127#1 (format 1 required a formula of the primary instance once the "
       "weighing became 13292.127's primary, D055) to the conclusion of the argument that cites the starvation, "
       "13292.127#2's ¬should_be_performed(tax_meat), as OC6 puts a mitigation and as both annotators aimed it (at "
       "13292.127's conclusion, when that argument was its primary); scheme, role and polarity unchanged. A cascade in "
       "the replay now. Notes rewritten.")


# --- Changed: the true fragments follow guidelines §5 item 6 --------------------------------------

def fragment(cid, iid, keep_vars, why, keep_formulas):
    ins = instance(cid, iid)
    for v in list(ins['bindings']):
        if v not in keep_vars:
            ins['bindings'][v] = {'unstated': why[v]}
    before = [p['formula'] for p in ins['premises']]
    ins['premises'] = [p for p in ins['premises'] if p['formula'] in keep_formulas]
    if [p['formula'] for p in ins['premises']] != keep_formulas:
        refuse(f'{iid} does not list {keep_formulas} (it lists {before})')
    for p in ins['premises']:
        if p['status'] != 'explicit':
            refuse(f'{iid} lists {p["formula"]} as {p["status"]}, not stated')
    return ins


fr = fragment('13292.6', '13292.6#1', ('G', 'S2'),
              {'A': "the action: 13292.108's environmental argument has it (tax_meat); this claim does not state it",
               'S1': "the circumstances: 13292.108's environmental argument has them; this claim presupposes them, it does not state them",
               'P': "the agent: 13292.108's environmental argument has it (governments); this claim does not state it"},
              ['would_be_realized(deforestation_slowed,reduced_meat_consumption)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.108's argument (should_be_performed(tax_meat) there)."}
G['13292.6']['challenges'] = [c for c in G['13292.6']['challenges'] if c['formula'] == 'would_be_realized(deforestation_slowed,reduced_meat_consumption)']
set_note('13292.6', '13292.6#1', 'would_be_realized(deforestation_slowed,reduced_meat_consumption)', None,
         'no printed question; attacked by 13292.192 (crop expansion may still deforest), defended by 13292.200, a reply to 13292.192')
record('13292.6', ['challenges', 'instances.bindings', 'instances.conclusion', 'instances.premises', 'notes'],
       "Guidelines §5 item 6 applied (the identity contract left it open for this draft): the fragment binds only the "
       "variables of the formula it states, G and S2 of would_be_realized(deforestation_slowed,reduced_meat_consumption); "
       "A, S1 and P are unstated, its premises are that formula alone, and its conclusion, which names A, is null. The "
       "challenge at would_bring_about(...) is left to 13292.108, whose own record has it; the remaining record's note "
       "says that 13292.200 replies to 13292.192. 13292.192's target, the stated link, is unchanged. Notes rewritten.")

PARENT_118 = "13292.118's argument has it"
fr = fragment('13292.11', '13292.11#1', ('S1',),
              {'A': f'the action: {PARENT_118} (tax_meat); this claim does not state it',
               'S2': f'the resulting state: {PARENT_118} (higher_meat_price); this claim does not state it',
               'G': f'the goal: {PARENT_118} (meat_price_reflects_real_costs); this claim does not state it',
               'P': f'the agent: {PARENT_118} (governments); this claim does not state it'},
              ['current_circumstances(meat_price_below_real_costs)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.118's argument (should_be_performed(tax_meat) there)."}
fr['note'] = "The claim states the circumstances S1 of 13292.118's goal-based argument, and nothing else of it."
G['13292.11']['challenges'] = [c for c in G['13292.11']['challenges']
                               if c['instance'] == '13292.11#2' or c['formula'] == 'current_circumstances(meat_price_below_real_costs)']
record('13292.11', ['challenges', 'instances.bindings', 'instances.conclusion', 'instances.note', 'instances.premises', 'notes'],
       "Guidelines §5 item 6 applied: 13292.11#1 binds S1 alone, the variable of the formula it states "
       "(current_circumstances(meat_price_below_real_costs)); A, S2, G and P are unstated, its premises are that formula "
       "alone, its conclusion null, and its note says it states S1 and nothing else of the argument. The challenge at "
       "realize_more_effectively(...) is left to 13292.118, whose own record has it. 13292.11#2, the causal explanation, "
       "unchanged. Notes rewritten.")

fr = fragment('13292.218', '13292.218#1', ('S1', 'S2', 'G'),
              {'A': ("the action: 13292.118's argument has it (tax_meat); the claim refers to it as 'it' but states no "
                     "formula with it (the transition from the tax, would_bring_about, is inferred from 13292.118)"),
               'P': f'the agent: {PARENT_118} (governments); this claim does not state it'},
              ['current_circumstances(meat_price_below_real_costs)',
               'would_be_realized(meat_price_reflects_real_costs,higher_meat_price)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.118's argument (should_be_performed(tax_meat) there)."}
G['13292.218']['challenges'] = [c for c in G['13292.218']['challenges']
                                if c['formula'] in ('current_circumstances(meat_price_below_real_costs)',
                                                    'would_be_realized(meat_price_reflects_real_costs,higher_meat_price)')]
record('13292.218', ['challenges', 'instances.bindings', 'instances.conclusion', 'instances.premises', 'notes'],
       "Guidelines §5 item 6 applied: the fragment binds S1, S2 and G, the variables of the formulas it states "
       "(current_circumstances(meat_price_below_real_costs), would_be_realized(meat_price_reflects_real_costs,"
       "higher_meat_price)); A and P are unstated, its premises are those two formulas, its conclusion null. The "
       "challenge at realize_more_effectively(...) is left to 13292.118, whose own record has it. 13292.224's target, the "
       "stated circumstances, is unchanged. Notes rewritten.")


# --- Joined: the items whose records named the retargeted claims (plan §3) ------------------------

lift_cap('13292.9', '13292.9#2', MODAL_STEP + ' First link of the chain tax -> harm to the meat industry -> lower tax revenue.')
lift_cap('13292.9', '13292.9#3', MODAL_EFFECT + ' Second link: harm to an industry that pays substantial taxes lowers tax revenue.')
modal = instance('13292.9', '13292.9#1')
mp = next(p for p in modal['premises'] if p['formula'] == 'may_bring_about(tax_meat,tax_revenue_reduced)')
mp['note'] = '"could actually" marks a possibility, not a firm prediction: the modal premise of negative_consequences_modal (D008)'
set_note('13292.9', '13292.9#1', 'may_bring_about(tax_meat,tax_revenue_reduced)', '33.2 CQ1',
         "'could'; 13292.12 backs the link with a mechanism at its first causal step, 13292.9#2 (28 CQ1)")
set_note('13292.9', '13292.9#1', 'may_bring_about(tax_meat,tax_revenue_reduced)', '33.2 CQ2', 'the claim cites nothing for the possibility')
set_note('13292.9', '13292.9#1', 'good_outweighs(tax_meat,tax_revenue_reduced)', '33.2 CQ3',
         '13292.43: the tax serves more pressing goods (sustainability, animal welfare)')
set_note('13292.9', '13292.9#4', '¬should_be_performed(tax_meat)', None,
         'at the bridge\'s conclusion, which meets the thesis (D008); a further consequence that ranks nothing would be a rebuttal here')
add_challenge('13292.9', {'cq': '28 CQ1', 'wag_cq': None, 'type': 'likelihood', 'instance': '13292.9#2',
                          'formula': 'causes(tax_meat,meat_industry_harmed)',
                          'note': 'backed by 13292.12 (a mechanism: cheaper dumped imports would push domestic producers out)'})
record('13292.9', ['challenges', 'instances.note', 'instances.premises', 'instances.scheme.confidence', 'notes'],
       "Joined round 1 (plan §3): " + JOINED['13292.9'] + "; that record put right, and a 28 CQ1 record added at "
       "13292.9#2 for 13292.12's mechanism. Otherwise the reference's coding kept, to be reviewed with round 1 (the "
       "developer's reading, round2-readings.md, matched; not blind, as that file says): the notes now name 33.3, as OC15 asks of a possible "
       "consequence to the agent's own goal, and tell one current account; " + LIFTED + "; the notes of its modal "
       "premise and of its other challenge records rewritten without their history, and " + WORKSTREAM_E + ".")

set_note('13292.2', '13292.2#1', 'instance(meat_production,significant_ghg_source)', '14 CQ1',
         "how significant? 13292.135 (less than other sources), 13292.190 (mostly beef); 13292.141's energy comparison "
         "backs the causal step 13292.2#2 instead (28 CQ1)")
set_note('13292.2', '13292.2#2', 'causes(meat_production,significant_ghg_emissions)', '28 CQ1',
         'backed by 13292.141 (a mechanism: the energy meat production takes)')
set_note('13292.2', '13292.2#2', 'causes(meat_production,significant_ghg_emissions)', '28 CQ2',
         'the claim links a source for it (evidence); 13292.46\'s beef figures also bear on it, though 13292.46 itself '
         'targets the thesis (a tax graded by impact)')
record('13292.2', ['challenges', 'notes'],
       "Joined round 1 (plan §3): " + JOINED['13292.2'] + "; that record put right, and the records at 13292.2#2 sort "
       "what bears on the link by the draft's reading of the book's questions (a mechanism, 13292.141, at 28 CQ1; data, "
       "13292.46's beef figures, at 28 CQ2). Otherwise the reference's coding kept, to be reviewed with round 1: the "
       "developer's reading (round2-readings.md; written after round 1's draft had used its causal step, so not blind) "
       "matched. Notes rewritten as one current account, naming what bears on the emissions and their significance.")

lift_cap('13292.180', '13292.180#2', MODAL_EFFECT + ' The step that backs the primary brings_about premise; the step tax -> less meat eaten is left implicit.')
set_note('13292.180', '13292.180#1', 'brings_about(tax_meat,more_anemia_cases)', '33.2 CQ1',
         'the substitutes of 13292.182 and 13292.198 bear on it, but they answer at the causal step (13292.180#2\'s exception, 28 CQ3)')
set_note('13292.180', '13292.180#1', 'good_outweighs(tax_meat,more_anemia_cases)', '33.2 CQ3', NO_REPLY)
add_challenge('13292.180', {'cq': '28 CQ3', 'wag_cq': None, 'type': 'causal interference', 'instance': '13292.180#2',
                            'formula': 'interference(reduced_meat_consumption)',
                            'note': 'answered against by 13292.182 (iron supplements) and 13292.198 (iron-rich plant foods)'})
record('13292.180', ['challenges', 'instances.note', 'instances.scheme.confidence', 'notes'],
       "Joined round 1 (plan §3): " + JOINED['13292.180'] + " (28 CQ3, at its exception); that record put right and a "
       "28 CQ3 record added at the exception. Otherwise the reference's coding kept, to be reviewed with round 1 (A, B "
       "and the reference agree on every core field; the developer read it with 13292.182 and .198, not blind): notes "
       "rewritten as one current account; " +
       LIFTED_ONE +
       "; the notes of its records at the outweighing exception and at the conclusion rewritten as current: without "
       "their history, and " + WORKSTREAM_E + ".")


# --- Challenge notes of kept items, rewritten without their history --------------------------------

set_note('13292.108', '13292.108#1', 'realize_more_effectively(governments,environmental_harm_from_meat_reduced,tax_meat)', '22.1 CQ2',
         '13292.104 (regulate the industry) answers it with a better means; 13292.147 (carbon rationing) claims no better '
         'means, so it is a proposal of another course at the conclusion, not an answer here (D006)')
set_note('13292.108', '13292.108#1', 'bad_outweighs(tax_meat,environmental_harm_from_meat_reduced)', '22.1 CQ5',
         '13292.127 weighs a side effect, starvation, against the goal here (bad_outweighs_good); 13292.174 (more dangerous '
         'production) and 13292.5 (regenerative farming deterred) cite side effects that rank nothing, rebuttals at the '
         'conclusion (D070), where the game asks CQ:side_effect')
set_note('13292.127', '13292.127#2', 'good_outweighs(tax_meat,starvation_from_reduced_meat_access)', '33.2 CQ3', NO_REPLY)
# The game asks these at the conclusion by design, the rebuttal route beside the exception's question (D055; D070;
# D095 (1), since step C5.2), not "until workstream E", as the notes said before that step.
REBUTTAL_ROUTE = 'the game asks it at the conclusion, the rebuttal route of {} (D055), beside the exception\'s question'
set_note('13292.127', '13292.127#2', '¬should_be_performed(tax_meat)', None,
         'no claim of the batch argues a good consequence against this argument (13292.202, at the same conclusion, '
         'amends the action instead); ' + REBUTTAL_ROUTE.format('33.2 CQ3'))
set_note('13292.180', '13292.180#1', '¬should_be_performed(tax_meat)', None,
         '13292.233\'s point (less meat, better health) would answer it; 13292.233 itself answers 13292.121, at its modal '
         'premise; ' + REBUTTAL_ROUTE.format('33.2 CQ3'))
set_note('13292.147', '13292.147#1', 'should_be_performed(ration_meat_via_carbon_budgets)', None,
         '13292.149 (a child outside the batch): harder and costlier to implement than a plain meat tax; '
         + REBUTTAL_ROUTE.format('33.1 CQ3'))
set_note('13292.194', '13292.194#1', 'should_be_performed(tax_meat_with_import_safeguards)', None,
         'possible side effects (trade disputes, verification costs); the claim has no replies; '
         + REBUTTAL_ROUTE.format('33.1 CQ3'))
set_note('13292.174', '13292.174#1', 'may_bring_about(tax_meat,more_dangerous_meat_production)', '33.2 CQ1',
         "'might'; why would a tax push producers to cheaper methods rather than raise prices?")
set_note('13292.174', '13292.174#1', 'may_bring_about(tax_meat,more_dangerous_meat_production)', '33.2 CQ2', 'the claim cites nothing for the possibility')
set_note('13292.174', '13292.174#1', 'harm_discounted_by_bearer(tax_meat,more_dangerous_meat_production)', None,
         'no printed question; dangerous to whom is unstated: on whom the harm falls is the exception harm_discounted_by_bearer (D065)')
set_note('13292.104', '13292.104#2', 'current_circumstances(farming_industry_not_meat_is_problem)', None,
         'no printed question; 13292.208 (non-cultured meat production is inherently problematic) contradicts it; the gold '
         'has 13292.208 at the better means\' premise brings_about(regulate_animal_farming_industry,G) of 13292.104#1, '
         'where format 1 moved it (round 2 of the first gold takes up its target)')
set_note('13292.104', '13292.104#2', 'possible_bring_about(governments,regulate_animal_farming_industry)', '22.1 CQ4',
         'feasibility of regulating the industry; also a premise of the primary better_means instance, possible_bring_about(P,A2) (D089)')
set_note('13292.104', '13292.104#2', 'bad_outweighs(regulate_animal_farming_industry,meat_problems_resolved)', '22.1 CQ5',
         'no claim of the batch argues that the other course\'s bad consequences outweigh its goal (D070); a plain side '
         'effect would be a rebuttal at the conclusion')
set_note('13292.147', '13292.147#1', 'bad_outweighs(ration_meat_via_carbon_budgets,carbon_emissions_reduced)', '33.1 CQ3', NO_REPLY)
set_note('13292.194', '13292.194#1', 'bad_outweighs(tax_meat_with_import_safeguards,domestic_producers_protected_from_imports)', '33.1 CQ3', NO_REPLY)
set_note('13292.5', '13292.5#1', 'too_slight(tax_meat,regenerative_agriculture_deterred)', None,
         'no printed question; how much meat comes from regenerative farms? How serious the consequence is, is the '
         'exception too_slight (D065): the amount bears on the decision as the share of the tax\'s environmental aim that '
         'deterring regenerative farms would cost, not as how few farms it reaches (the owner\'s scope, D053)')
set_note('13292.123', '13292.123#2', 'possible_bring_about(governments,abolish_meat_subsidies)', '22.1 CQ4',
         '13292.171: removing subsidies is politically easier than taxing; also a premise of the primary better_means '
         'instance, possible_bring_about(P,A2) (D089)')
# Premise notes of kept items that told a history or said something no longer true.
set_premise_note('13292.127', '13292.127#2', 'brings_about(tax_meat,starvation_from_reduced_meat_access)',
                 "presupposed: the tax decreases access to meat by making it dearer (13292.7); 13292.204 denies this link at "
                 "the weighing's own premise (13292.127#1, OC10), and 13292.202 amends the action at this argument's conclusion")
set_premise_note('13292.174', '13292.174#1', 'may_bring_about(tax_meat,more_dangerous_meat_production)',
                 '"might force" marks a possibility, not a firm prediction: the modal premise of negative_consequences_modal (D008)')
set_premise_note('13292.108', '13292.108#1',
                 'would_bring_about(tax_meat,meat_production_and_consumption_harm_environment,reduced_meat_consumption)',
                 'denied by 13292.13; 13292.14 and 13292.61 rebut that denial')
lift_cap('13292.192', '13292.192#1', MODAL_EFFECT + ' A causal scheme because the attack consists of an offsetting causal chain, which no other walton.yml scheme expresses.')
lift_cap('13292.192', '13292.192#2', MODAL_EFFECT + " The link itself is stated only as a possibility ('may still "
         "occur'), which the scheme's firm link does not express.", conf=0.5)
lift_cap('13292.36', '13292.36#2', MODAL_STEP)
lift_cap('13292.36', '13292.36#3', MODAL_EFFECT)


# --- Notes: one current account -------------------------------------------------------------------

NOTES = {
    '13292.108': (
        "A pro of the thesis that states only the circumstances S1 (meat production and consumption harm the "
        "environment) of a goal-based practical reasoning for the tax: a fragment (guidelines §3). The thesis has no "
        "argument of its own (scheme none), so no other claim states the argument this fragment belongs to; 13292.108 "
        "carries it, binding every variable, with the goal (less environmental harm from meat), the resulting state "
        "(less meat consumed), the transition and the agent supplied, so that the replies to the environmental argument "
        "can name its parts: 13292.13 denies the efficacy premise; 13292.6 states a case of the link to the goal; "
        "13292.104 answers the other-means question with a better means; 13292.147 proposes another course without "
        "saying it does better, at the conclusion (D006); 13292.174 and 13292.5 cite side effects that rank nothing, "
        "rebuttals at the conclusion (D070); 13292.127 weighs a side effect against the goal at the exception "
        "bad_outweighs(A,G); 13292.135, .137, .2 and .4 bear on the circumstances it states. Guidelines §5 item 6 would "
        "leave A, S2, G and P unstated and the conclusion null: the four targets at the stated circumstances would keep "
        "their formula, and the seven others would point at nothing. Whether a fragment whose parent has no argument "
        "carries that argument is a question the guidelines leave open (D086 round 1, for the review and the owner). "
        "Goal-based because no value is stated."),
    '13292.6': (
        "States the S2 -> G link of the thesis's environmental argument (13292.108) with a specific goal: less demand "
        "for meat would slow or reverse deforestation. A fragment (guidelines §3; §5 item 6): it binds the variables of "
        "the one formula it states, would_be_realized(deforestation_slowed,reduced_meat_consumption); the action, the "
        "circumstances and the agent are 13292.108's and are left unstated here, and so is the conclusion, which names "
        "the action. It backs 13292.108's premise would_be_realized(environmental_harm_from_meat_reduced,"
        "reduced_meat_consumption) with a case of it; no critical question is about that premise, so premise_backing "
        "(§5 item 3). It presupposes that meat demand drives deforestation, which also supports 13292.108's "
        "circumstances. 'Reduced demand' is the shared atom reduced_meat_consumption. 13292.192 attacks the link it "
        "states."),
    '13292.11': (
        "States the circumstances S1 of 13292.118's goal-based argument, which 13292.118 only presupposes: the price of "
        "meat is below its real costs, because of subsidies. A fragment (guidelines §3; §5 item 6): 13292.11#1 binds the "
        "variable of the one formula it states, current_circumstances(meat_price_below_real_costs); the action, the "
        "resulting state, the goal and the agent are 13292.118's and are left unstated, and so is the conclusion. "
        "13292.11#2 explains the circumstances causally (subsidies -> a price below the real costs); the subsidies "
        "exist, so the cause has occurred and cause_to_effect applies. It backs 13292.118's premise "
        "current_circumstances(meat_price_below_real_costs), about which no critical question is asked, so "
        "premise_backing (§5 item 3). The explanation opens the other means its child 13292.123 takes up (abolishing "
        "the subsidies), an answer at 13292.118's exception realize_more_effectively(...)."),
    '13292.218': (
        "'It' is the meat tax of 13292.118. The claim restates 13292.118's link from the higher price to its goal with "
        "content (the price would then include the external costs: climate change, water pollution, disease risks) and "
        "names the circumstances (the price leaves those costs out). A fragment of 13292.118's goal-based argument "
        "(guidelines §3; §5 item 6): it binds the variables of the formulas it states, "
        "current_circumstances(meat_price_below_real_costs) and would_be_realized(meat_price_reflects_real_costs,"
        "higher_meat_price); the action and the agent are 13292.118's and are left unstated (the claim names the tax as "
        "'it', but the transition from the tax, would_bring_about, is inferred from 13292.118, not stated), and so is "
        "the conclusion. It backs 13292.118's premise would_be_realized(...), about which no critical question is "
        "asked, so premise_backing (§5 item 3); 'clarification' is a possible alternative role. 13292.224 backs the "
        "circumstances it states."),
    '13292.12': (
        "Gives the mechanism of 13292.9's 'by harming the meat industry': the tax would make imported meat cheaper "
        "(dumping, predatory pricing), and cheaper imports would push domestic producers out of the market. Two causal "
        "steps whose cause is only proposed (the tax) or possible (the cheaper imports it would bring), so "
        "cause_to_effect_modal (§5 item 11): 13292.12#2 (the tax -> cheaper "
        "imports) and 13292.12#1 (cheaper imports -> domestic producers pushed out), primary as the step that ends at "
        "the harm to the industry. Together they back the causal premise of 13292.9's first step, "
        "causes(tax_meat,meat_industry_harmed) (13292.9#2), which the book's 28 CQ1 asks about (how strong the causal "
        "generalization is; the game asks no question there): cq_answer with polarity for (§5 item 3). A mechanism "
        "offered for a link answers 28 CQ1; data offered for it would answer 28 CQ2 (a question for the review, D086 "
        "round 1). 13292.12#3 codes the same content as a con of the tax in its own right, as its children treat it "
        "(13292.131 backs the badness; 13292.45 and 13292.194 propose remedies, amended actions at 13292.12's predicted "
        "harm)."),
    '13292.141': (
        "Backs 13292.2's claim that meat production releases significant greenhouse gas emissions, through its energy "
        "use: producing meat takes far more energy than growing the same nutrients from plants. Its argument is a causal "
        "step whose cause is an actual state (cause_to_effect: the energy-intensive production causes significant "
        "emissions); the link from energy use to emissions is supplied, and the comparison cites nothing. It offers a "
        "mechanism for the causal premise of 13292.2's own step, causes(meat_production,significant_ghg_emissions) "
        "(13292.2#2), which the book's 28 CQ1 asks about (how strong the causal generalization is; the game asks no "
        "question there): cq_answer with polarity for (§5 item 3), as 13292.2's own record of the question has it. It "
        "could also be read as evidence for 13292.2's classification premise instance(meat_production,"
        "significant_ghg_source) (14 CQ1). The comparison with plant sources also implies that they are more efficient "
        "(not coded)."),
    '13292.182': (
        "A substitute (iron supplements) lets people get enough iron without meat, so eating less meat need not bring "
        "more anemia: it counteracts the causal step 13292.180#2, from reduced meat consumption to more anemia, which is "
        "what the book's 28 CQ3 asks (other factors that would interfere with or counteract the production of the "
        "effect; the game asks no question there). So it is a cq_answer at that step's exception "
        "interference(reduced_meat_consumption), against, and its own argument (Other: no walton.yml scheme concludes "
        "it from a substitute) concludes the exception (R8). It does not deny that the tax reduces meat eating; "
        "walton.yml's English for the exception ('an event has occurred which interferes with event ...') reads as if it "
        "did, but the exception is the book's CQ3. The substitute is what people would do, not an amended action of the "
        "agent's (OC6). Its sibling 13292.198 makes the same move with plant foods; 13292.184 replies that supplements are "
        "not always a good replacement."),
    '13292.198': (
        "Plant foods rich in iron (beans, lentils) can replace meat as a source of iron, so eating less meat need not "
        "bring more anemia: like its sibling 13292.182, it counteracts the causal step 13292.180#2 (reduced meat "
        "consumption -> more anemia), the book's 28 CQ3 (the game asks no question there): a cq_answer at that step's "
        "exception interference(reduced_meat_consumption), against; its Other instance concludes the exception (R8), "
        "whose walton.yml English reads as if the reduction itself were interfered with. The substitute is what people "
        "would do, not an amended action of the agent's (OC6). It also implicitly contests 13292.180's premise that meat "
        "is the easiest source of iron. 13292.210 backs the substitution."),
    '13292.25': (
        "Concedes the burden ('the additional burden') and says it is small: at most a change in one's eating pattern, "
        "which is easy (its premises: changing one's eating pattern avoids the added cost, and is easy to do). It argues "
        "that the consequence 13292.7 cites is too slight to count, the exception too_slight(tax_meat,"
        "food_less_affordable_for_poor) of 13292.7#1 (OC9), not that the tax does not bring it about: its sibling "
        "13292.33 says that, pointing to cheaper protein. No scheme concludes too_slight, so its own argument is Other, "
        "concluding the exception (R8); the game asks CQ:too_slight there, so the role is cq_answer (§5 item 3), "
        "against. Its own words hedge ('could be as small as'); the fit and the confidence say so."),
    '13292.33': (
        "With cheaper protein available, a meat tax need not make food unaffordable for those who cannot afford it: the "
        "claim offers a way to avoid the consequence 13292.7 cites, so it answers the likelihood question at 13292.7#1's "
        "premise brings_about(tax_meat,food_less_affordable_for_poor), the game's CQ:likelihood (33.2 CQ1), against. It "
        "does not say the burden is slight (its sibling 13292.25 does, at too_slight). 13292.7 has no causal step of its "
        "own in the gold, so the answer is at its consequence premise (compare 13292.182 and 13292.198, at the causal "
        "step of 13292.180)."),
    '13292.36': (
        "A further bad consequence for the same conclusion as 13292.7 (food security falls where the poor need easy "
        "access to all food), so an independent argument for it; backing of 13292.7's bad(G) is the alternative role, "
        "since the need sentence backs it. 'Taxes' generalizes from the meat tax; A is bound to the thesis's action. The "
        "chain tax -> production costs -> food security is coded as two causal steps, whose causes, the tax and the "
        "higher costs it would bring, are only proposed or possible (cause_to_effect_modal, §5 item 11)."),
    '13292.127': (
        "The claim ranks two consequences: starvation from reduced access to meat is a more severe threat than any "
        "consequence of high meat consumption, the harms the pros cite. A ranking argues the exception of the argument "
        "it weighs against (OC3, OC4). It names a class, so it is coded against one argument of the class (OC7): its "
        "Kialo parent, 13292.7, argues against the tax and is not in it, so the earliest argument of the class on the "
        "board, 13292.108's practical reasoning, at its exception bad_outweighs(tax_meat,"
        "environmental_harm_from_meat_reduced) (D070, D091); re-parented from 13292.7. The weighing is primary because "
        "the words state the comparison (OC8): an instance of bad_outweighs_good, C003's answer scheme, a cq_answer to "
        "the game's CQ:bad_outweighs (22.1 CQ5), against. The starvation consequence, which the text presupposes, is a "
        "secondary negative-consequences instance (13292.127#2). Its link is categorical: 'a more severe threat than' "
        "ranks how bad starvation is, and the link from the tax to starvation is inferred from 13292.7's firm 'would "
        "make food more expensive' (the mood rule, D008). Its children engage the starvation consequence: 13292.202 "
        "amends the action, at the conclusion of the starvation argument 13292.127#2 (OC6), and 13292.204 denies the "
        "link the weighing relies on (OC10)."),
    '13292.174': (
        "Kialo files it as a pro of 13292.13 (the tax would not resolve, and might worsen, the problems of production). "
        "Its content is a further bad consequence of the tax that ranks nothing, offered against the thesis's "
        "environmental argument 13292.108 (a sibling of the Kialo parent, hence re-parented): a rebuttal at that "
        "argument's conclusion (OC2; D070), with role.move side_effect, against. 'Might force' claims the consequence "
        "only possible (the mood rule, D008): 13292.174#1 is the modal variant (negative_consequences_modal), "
        "secondary, and the bridge 13292.174#2 (bridge_from_possible_harm), primary, carries its conclusion to the "
        "target. What 'dangerous' means (food safety, workers, animals, environment) is unstated."),
    '13292.104': (
        "Proposes another action, regulating the animal farming industry, as the better way to the goal the thesis's "
        "environmental argument pursues: the problem is the industry, not meat, so act on the industry 'rather than "
        "penalize those who wish to consume meat'. A means that reaches the goal more surely answers 13292.108's "
        "exception realize_more_effectively(governments,environmental_harm_from_meat_reduced,tax_meat) (OC6; D075, "
        "D089): its primary instance is better_means (comparison: more surely), a cq_answer to the game's "
        "CQ:other_means (22.1 CQ2), with role.move alternative_means, against; re-parented from 13292.13, a sibling of "
        "13292.108. The argument for the other course is a secondary instance (practical reasoning), and 'rather than' "
        "ties it to the thesis in a third (an argument from alternatives). Its own goal is the broader 'problems' of "
        "13292.13, not only environmental harm. It also supports 13292.13 by implying that taxing consumers misses the "
        "source of the problems; 13292.208 replies that meat production is itself the problem."),
    '13292.43': (
        "Ranks two consequences against 13292.9: the goods the tax serves (sustainability, animal welfare) press harder "
        "than the lost tax revenue 13292.9 says may follow. A ranking argues the exception of the argument it answers "
        "(OC3, OC4): 13292.9#1, the modal argument (13292.9 claims the loss only possible; OC15), at its exception "
        "good_outweighs(tax_meat,tax_revenue_reduced). Its instance is good_outweighs_bad, C003's answer scheme, with "
        "the goods named as one consequence (no reply splits them, OC11); a cq_answer to the game's CQ:good_outweighs "
        "(33.2 CQ3), against. The values it names are things the tax would promote, so they count as consequences (OC5). "
        "Its Kialo parent cites the outweighed consequence, so it is not re-parented."),
    '13292.192': (
        "Attacks the link 13292.6 states (less meat demand -> less deforestation) with an offsetting causal chain: less "
        "meat -> more crops -> deforestation for cropland. Each step's cause is only possible, so cause_to_effect_modal "
        "(§5 item 11): 13292.192#1 (reduced meat consumption -> more crop production), primary, and 13292.192#2 (more "
        "crop production -> deforestation for cropland). The target is 13292.6's stated "
        "would_be_realized(deforestation_slowed,reduced_meat_consumption); no critical question is about it, so "
        "premise_attack (§5 item 3), against. Its child 13292.200 replies that the claim was only that deforestation "
        "would slow or reverse."),
    '13292.147': (
        "Kialo con of 13292.2 (emissions), but it does not dispute the emissions: it puts forward another course, "
        "rationing meat through personal carbon budgets, for the environmental goal, without saying that it does better "
        "than the tax (its children compare it with a tax the other way, 13292.149, or by who bears the cost, "
        "13292.151). A course not said to do better answers nothing at the exception realize_more_effectively (D006): it "
        "is a proposal of another course, coded at the conclusion of the argument it replies to, 13292.108's "
        "should_be_performed(tax_meat) (an ancestor, hence re-parented), role action_amendment with move "
        "alternative_means, polarity against as Kialo places it (OC6; D066). Its own argument is positive consequences "
        "for the other course (rationing reduces carbon emissions), which the replay plays."),
    '13292.204': (
        "Denies the starvation link 13292.127 relies on: where people would starve from higher food costs, they eat "
        "little or no meat, so a meat tax cannot starve them by reducing their access to meat. A reply to a weighing "
        "claim attacks that claim's own instance (OC10): here the consequence it relies on, "
        "brings_about(tax_meat,starvation_from_reduced_meat_access), a premise of 13292.127#1 (bad_outweighs_good), "
        "about which no critical question is asked, so premise_attack (§5 item 3), against. The premise is 13292.127's "
        "own (inferred from 13292.7), so this is not re-parenting. Its argument is Other (absent exposure: the affected "
        "group is not exposed to the cause). Both annotators aimed it at the same link as a premise of the starvation "
        "argument, 13292.127's primary then and 13292.127#2 now, B as an answer to its likelihood question; D055's "
        "recoding moved it to the weighing's premise, which OC10 now supports: a question for the review (D086 round 1, "
        "with 13292.202's)."),
    '13292.233': (
        "Denies that the reduction a tax would cause is unhealthy: meta-analyses and reviews find that less meat, and "
        "vegetarian diets, bring better health, lower mortality and longer lives, especially when high intakes of red "
        "or processed meat are reduced. 13292.121 claims that harm only possible (OC15): the reply denies that the "
        "possibility is real, so it answers the modal premise may_bring_about(tax_meat,unhealthy_meat_underconsumption) "
        "of 13292.121#1, the game's CQ:modal_likelihood (33.2 CQ1), against. Its own argument is from expert opinion "
        "(the cited studies). Its qualifier ('especially when high intakes ... are reduced') leaves open people who "
        "already eat little meat, which is 13292.121's worry."),
    '13292.29': (
        "Anti-paternalist rebuttal: whatever the health facts (13292.8), the state should not act on them. It sits in the "
        "health branch and argues against 13292.115's conclusion should_be_performed(tax_meat) through a rule from "
        "personal liberty (the argument from rules, 55.2); the step 'the state stays out' -> 'no meat tax' is left "
        "implicit. A right that rules the action out whatever its consequences is not the outweighing exception (OC5): "
        "it is coded with the scheme that fits it, and names candidate C007. It does not address the externality "
        "arguments (13292.218, 13292.108), which are not about self-regarding choices, so it is coded as a reply to "
        "13292.115 rather than as a new con of the thesis."),
    '13292.5': (
        "A further bad consequence of the tax that ranks nothing, offered against the thesis's environmental argument "
        "13292.108 (an ancestor, hence re-parented): the tax would deter regenerative agriculture, a way of producing "
        "meat that benefits the environment. A side effect that ranks nothing against the goal is negative "
        "consequences, a rebuttal at the conclusion of the argument it answers (OC2; D070), with role.move side_effect, "
        "against; 'would be deterred' is categorical. Its first clause also contests the generality of 13292.4 and "
        "13292.108, as 13292.137 does; it is coded by its distinctive point, what the tax would do."),
    '13292.123': (
        "Accepts 13292.11's diagnosis (subsidies keep the price below the real costs) and proposes another means to "
        "13292.118's goal: abolish the subsidies (or exempt meat from them) instead of taxing. A means that reaches the "
        "goal more surely answers 13292.118's exception realize_more_effectively(governments,"
        "meat_price_reflects_real_costs,tax_meat) (OC6; D075, D089): its primary instance is better_means (comparison: "
        "more surely), a cq_answer to the game's CQ:other_means (22.1 CQ2), with role.move alternative_means, against; "
        "re-parented from 13292.11 (13292.118 is an ancestor). The argument for the other course is a secondary "
        "instance; the incoherence sentence is a third, from opposition (11.2), coded as Other because walton.yml's 11.2 "
        "encoding is self-contradictory."),
    '13292.202': (
        "'A only for X': answers 13292.127's starvation objection by restricting the tax's scope (exceptions, or the tax "
        "only in non-third-world countries): a mitigation, an amended action (C006; OC6), "
        "tax_meat_outside_poor_countries, with no purpose stated beyond the context's, so scheme none (a bare proposal, "
        "§5 item 4), role action_amendment, polarity against; it still supports a restricted meat tax. OC6 puts a "
        "mitigation at the conclusion of the argument it answers: 13292.127#2's ¬should_be_performed(tax_meat), the "
        "argument that cites the starvation, where both annotators put it (at 13292.127's conclusion, when that argument "
        "was its primary). A reply to a weighing claim that attacks it addresses the link the weighing relies on "
        "(OC10), as its sibling 13292.204 does; this reply amends the action instead (the reference had it at that "
        "link, since format 1 required a formula of the primary instance once the weighing became primary, D055: a "
        "question for the review, D086 round 1). Stated as a possibility ('there can be', 'could')."),
    '13292.194': (
        "An amended action that mitigates the harm 13292.12 predicts (C006; OC6): the meat tax plus import measures "
        "(eco-tariffs; imports that verifiably meet environmental and animal-welfare standards, or pass a high enough "
        "tax barrier), tax_meat_with_import_safeguards. 'Can solve this issue' is can of capability (the mood rule, "
        "D008): positive consequences of the amended action, role action_amendment, polarity against 13292.12, at "
        "13292.12's predicted harm, the conclusion may_occur(domestic_meat_industry_pushed_out) of its primary causal "
        "step 13292.12#1, where both annotators put it. Two other readings are put to the review (D086 round 1): the "
        "target at 13292.12#3, the negative-consequences argument that cites the harm against the tax; or a cq_answer to "
        "28 CQ3 at 13292.12#1's interference, the mechanism blocked, which the draft does not take, since the measures "
        "are the agent's own action (case.json), unlike the substitutes of 13292.182 and .198. Its sibling 13292.45 "
        "(round 2) has the same target."),
    '13292.9': (
        "A possible consequence of the tax: it could reduce tax revenue by harming the meat industry, which pays "
        "substantial taxes. 'Could actually' claims the consequence only possible (the mood rule, D008): 13292.9#1 is "
        "the modal variant of negative consequences, secondary, and the bridge 13292.9#4 (bridge_from_possible_harm), "
        "primary, carries its conclusion to the thesis, an independent con. The consequence falls on the agent's own "
        "goal, the government's revenue: 33.3, which has no modal variant, so the modal variant of 33.2 with its bridge "
        "(OC15). Its two causal links are secondary steps whose cause is only proposed or possible (cause_to_effect_modal, "
        "§5 item 11): the tax -> harm to the meat industry (13292.9#2), and that harm, itself predicted, -> lower "
        "revenue (13292.9#3). The net claim "
        "presupposes that lost industry taxes exceed the meat tax's own yield. 13292.12 backs the first link with a "
        "mechanism; 13292.43 weighs the tax's other goods against the lost revenue."),
    '13292.2': (
        "Backs the circumstances S1 of 13292.108's environmental argument, for production: meat production creates "
        "significant greenhouse gas emissions, which contribute to climate change. The primary instance is the "
        "classification step that turns the causal facts into the parent's evaluative claim (meat production is an "
        "activity that is bad for the environment); the causal chain (production -> emissions -> climate change) is two "
        "secondary steps whose causes have occurred, so cause_to_effect (§5 item 11). No critical question is about "
        "current_circumstances, so premise_backing (§5 item 3), for. 13292.141 backs the first causal step with a "
        "mechanism (28 CQ1); 13292.135 and 13292.190 bear on how significant the emissions are (14 CQ1)."),
    '13292.180': (
        "Two Kialo parents (R7). Primary function: a further bad consequence of making meat dearer (meat is the easiest "
        "source of iron, so financial restrictions will raise anemia), an independent argument for 13292.7's conclusion; "
        "its causal step, reduced meat consumption -> more anemia (13292.180#2, cause_to_effect_modal: the cause, less "
        "meat eaten, is only possible, a predicted effect of the tax), backs the primary's consequence premise, the step "
        "from the tax to less meat eaten left implicit. "
        "Toward 13292.121 it backs the modal premise may_bring_about(tax_meat,unhealthy_meat_underconsumption) of "
        "13292.121#1 (D008), recorded in other_kialo_parents as a cq_answer to the game's CQ:modal_likelihood (33.2 "
        "CQ1). 'Organic iron' presumably means heme iron. 13292.182 and 13292.198 answer the causal step with "
        "substitutes (28 CQ3, at its exception); 13292.184, a reply to 13292.182 outside the batch, says that supplements "
        "are not always a good replacement (bioavailability)."),
}
for cid, text in NOTES.items():
    if cid not in ROUND1 and cid not in JOINED:
        refuse(f'{cid} is not a claim of round 1')
    G[cid]['notes']['text'] = text

G['13292.147']['notes']['candidate_pattern'] = 'a proposal of another course, not said to do better (C017; guidelines §3.1, OC6)'

# --- Kept: the reference's coding, and why ----------------------------------------------------------

NOTES_REWRITTEN = 'notes rewritten as one current account'
CHALLENGES_REWRITTEN = 'challenge notes rewritten without their history'
KEPT = {
    '13292.1': ([], 'a control (A, B and the reference agree on every core field): the thesis, scheme none.'),
    '13292.30': ([], 'a control: the practical analogy from taxes on ecologically damaging businesses elsewhere, at the thesis (guidelines §3 names it).'),
    '13292.36': (['instances.note', 'instances.scheme.confidence', 'notes'],
                 'a control (A, B and the reference agree on every core field). The developer\'s first reading differed '
                 '(backing of 13292.7#1\'s consequence premise, the added production costs its mechanism), but the claim '
                 'adds a consequence of its own (food security), so an independent argument for 13292.7\'s conclusion, as '
                 'all three have it (its note names backing as the alternative). '
                 f'{NOTES_REWRITTEN[0].upper() + NOTES_REWRITTEN[1:]} (the history of the causal steps\' coding is in this '
                 f'record and gold-format1.json); {LIFTED}.'),
    '13292.108': (['challenges', 'instances.premises', 'notes'],
                  'every variable bound (see the note): the thesis has no argument, so this fragment carries the environmental argument; §5 item 6 would leave seven of the eleven targets into it nothing to point at (the four at the stated circumstances keep theirs). Whether a fragment whose parent has no argument carries that argument is put to the review and the owner. '
                  f'{NOTES_REWRITTEN[0].upper() + NOTES_REWRITTEN[1:]}; {CHALLENGES_REWRITTEN}; the note of its efficacy '
                  'premise says how 13292.14 and 13292.61 bear on it (they rebut 13292.13\'s denial).'),
    '13292.25': (['notes'], f'the exception too_slight (OC9), the game\'s CQ:too_slight; {NOTES_REWRITTEN}.'),
    '13292.33': (['notes'], 'the likelihood question at 13292.7#1\'s consequence premise (13292.7 has no causal step of its own); notes rewritten.'),
    '13292.127': (['challenges', 'instances.premises', 'notes'],
                  f'the weighing against the class at 13292.108\'s exception (OC3, OC4, OC7, OC8); {NOTES_REWRITTEN} (the earlier codings, against 13292.115 and with Other, are in this record\'s entries); {CHALLENGES_REWRITTEN}, and {WORKSTREAM_E}; the note of 13292.127#2\'s starvation premise says which reply engages what (13292.204 the weighing\'s premise, 13292.202 this argument\'s conclusion).'),
    '13292.174': (['challenges', 'instances.premises', 'notes'], f'a possible side effect that ranks nothing: the modal variant and its bridge, a rebuttal at 13292.108\'s conclusion (OC2, OC15; D070); notes rewritten; {CHALLENGES_REWRITTEN}, and the note of its modal premise too.'),
    '13292.104': (['challenges', 'notes'], 'a better means at 13292.108\'s exception (OC6; D075, D089): the developer\'s reading matched; notes rewritten; challenge notes rewritten as current (the one at 13292.104#2\'s circumstances says where format 1 moved 13292.208, which round 2 takes up).'),
    '13292.43': (['notes'], 'the weighing at 13292.9#1\'s exception (OC3, OC4, OC15): the developer\'s reading matched; notes rewritten.'),
    '13292.192': (['instances.note', 'instances.scheme.confidence', 'notes'],
                  f'the attack on 13292.6\'s stated link with a modal causal chain; notes rewritten without the history of the causal steps\' coding; {LIFTED_192}.'),
    '13292.147': (['challenges', 'notes'], f'a proposal of another course at 13292.108\'s conclusion (OC6; D006, D066); notes rewritten (the earlier text described the answer at the exception it no longer is), and candidate C017 named, as OC6 asks; {CHALLENGES_REWRITTEN}, and {WORKSTREAM_E}.'),
    '13292.204': (['notes'], 'a premise attack on the starvation link the weighing relies on (OC10); the annotators\' reading (the same link as a premise of the starvation argument, now 13292.127#2, B\'s cq_answer to CQ:likelihood before D055\'s recoding) is put to the review with 13292.202\'s; notes rewritten (the earlier text called the weighing Other, as it was before D064 and D075).'),
    '13292.233': (['notes'], 'the modal premise of 13292.121#1, CQ:modal_likelihood (OC15); notes rewritten (the earlier text named CQ:likelihood, the question before the modal variant).'),
    '13292.29': (['notes'], 'the rule from personal liberty, a rebuttal at 13292.115\'s conclusion (OC5, C007); fit explicit, since both premise-role formulas are stated (§5 item 6); notes revised: the remark on a game question removed, since a right that rules the action out answers no question (OC5).'),
    '13292.5': (['challenges', 'notes'], f'a side effect that ranks nothing, a rebuttal at 13292.108\'s conclusion (OC2; D070); notes rewritten; {CHALLENGES_REWRITTEN}.'),
    '13292.123': (['challenges', 'notes'], f'a better means at 13292.118\'s exception (OC6; D075, D089); notes rewritten; {CHALLENGES_REWRITTEN}.'),
    '13292.194': (['challenges', 'notes'], f'positive consequences of the amended action, at 13292.12\'s predicted harm, where both annotators put it (the developer\'s own reading was 13292.12#3, the argument that cites the harm); two other readings (13292.12#3, and 28 CQ3 at 13292.12#1\'s interference) are put to the review; notes rewritten; {CHALLENGES_REWRITTEN}, and {WORKSTREAM_E}.'),
    '13292.8': ([], 'the classification of North American consumption as unhealthy, backing 13292.115\'s premise brings_about(...) (CQ:likelihood): the developer\'s first reading (a fragment) took 13292.115 for practical reasoning; it is positive consequences, which has no circumstances premise.'),
    '13292.46': ([], 'the argument from rules (the rule, tax by impact; the case, beef\'s impact; the conclusion, beef taxed higher) fits what the claim states; no consequence of the graded tax is stated.'),
}
for cid, (fields, why) in KEPT.items():
    if cid not in ROUND1:
        refuse(f'{cid} is not a claim of round 1')
    record(cid, fields, "Drafted: the reference's coding kept: " + why)

missing = [c for c in list(ROUND1) + list(JOINED) if c not in done]
if missing:
    refuse(f'claims of round 1 not drafted: {missing}')

gold['status'] = ('Development material (D086): made by language models (the batch\'s adjudication, and the developer\'s '
                  'drafts of the first gold, round by round: docs/d086-first-gold.md); no item is verified by a person yet.')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Round 1 of the first gold, draft 1 (D086; docs/d086-first-gold.md, d086-rounds.json): 28 claims drafted, "
             "and three more that the draft changed and so joined the round (13292.9, .2, .180). Changed: 13292.12, .141, "
             ".182, .198 and .202 name what they answer, which format 1 had kept them from (13292.9#2 and 13292.2#2 at the "
             "book's 28 CQ1, 13292.180#2's exception at 28 CQ3, and 13292.127#2's conclusion for the mitigation 13292.202); "
             "the fragments 13292.6, .11 and .218 bind only the variables of the formulas they state (guidelines §5 item "
             "6). Kept, with the reason in each item's record: the other 20 of the round and the three joined items, whose "
             "records of the retargeted claims are put right; among them 13292.108, which carries the environmental "
             "argument (a question for the review and the owner). Nine causal steps capped by the N8 note have their "
             "confidence set on their merits (0.55; 0.5 for 13292.192#2). Notes and challenge notes that told a coding's history, or that the game asks a "
             "question at the conclusion only until workstream E, are rewritten as one current account; candidate C017 "
             "named for 13292.147. Corrected before any review, after an independent check of its first version and an "
             "independent check of those corrections.")})

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
for c in ROUND1:
    rounds['claims'][c].update(draft=1, review='not reviewed')
for c, why in JOINED.items():
    rounds['claims'][c].update(draft=1, review='not reviewed', joined=1, joined_why=why)
with open(ROUNDS, 'wb') as fh:
    fh.write(json.dumps(rounds, indent=1, ensure_ascii=False).encode('utf-8') + b'\n')
print(f'{len(done)} items drafted; gold.json and d086-rounds.json written; checking the gold (validate.py)')
run = subprocess.run([sys.executable, '-I', DOC + 'validate.py'], capture_output=True, text=True)
print(run.stdout.strip()[-3000:])
if run.returncode != 0:
    raise SystemExit('validate.py refuses the drafted gold: fix it, then copy it to ' + COPY)
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
    print('copied to', COPY)
