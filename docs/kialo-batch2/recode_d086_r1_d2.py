"""Round 1 of the first gold, draft 2 (D086; docs/d086-first-gold.md §11).

    python3 -I recode_d086_r1_d2.py

A one-off, kept as the record. It takes the first exchange of the review of round 1 into the gold:
the other reviewer's item review of draft 1 (2026-11-27) and the developer's answer, which the
reviewer's addendum of the same day reconciled (in the workspace, verify/chatgpt-d086-2026-11-27/
REVIEW.md §3 and verify/ethix-2026-11-27/REVIEW.md §2). Each changed item gets an
adjudication.recoded entry (by "developer (D086 draft 2)") that names what changed and why.

- Accepted objections. 13292.25: its ease premise supplied, the claim's hedge in its note.
  13292.104: #2's fit reconstructed (its goal premise is supplied; guidelines §5 item 6, R18); #1's
  comparison supplied, explained; #1's link explicit, the claim's 'the appropriate solution'
  projected to 13292.108's goal. 13292.43: the tax's effect supplied, its evaluation explicit.
  13292.127: the starvation link supplied in both instances (13292.7 grounds only the dearer food),
  #1's evaluation explicit as #2's. 13292.192: #2 primary, its conclusion reaching the target (§5
  item 5); the note says what the chain does not show. 13292.194: at 13292.12#3's conclusion, as
  13292.202 (OC6); the reviewer's optional second relation noted, not coded.
- Partly accepted. 13292.8: the note says the backing is partial; the target kept (the reviewer's
  unresolved target, a question for the guidelines). With the reconciliation's conditions:
  13292.123: better_means kept, its comparison and its link to the goal (in both instances) supplied
  and scoped (subsidy removal corrects 13292.11's distortion, not 13292.218's externalities), the
  goal 13292.118's (D075), the other reading (a proposal of another course) recorded as an
  alternative. 13292.218: the action named but not represented; the circumstances inferred from
  13292.11 (R19); the stated link explicit, its projection to 13292.118's higher_meat_price and its
  scope (the price rise that internalizes the named externalities) explained.
- Refinements of agreed items. 13292.11: the subsidies as the claim presents them. 13292.6: the
  modal causal reading recorded as an alternative.
- Corrections of agreed items whose records draft 2's changes made false, from the independent
  check of draft 2 (wag-workspace verify/d086/rv/): 13292.204's account of the premise it attacks;
  13292.12's of where 13292.194 is, and a quotation made exact. Quotations in the notes are the
  claims' words, checked against items.json.
- Unchanged: 13292.108 (question 1 goes to the owner) and the 17 other agreed items.

It refuses any gold but draft 1's, then writes gold.json; in d086-rounds.json it records each claim's
first exchange (exchanges), sets the changed claims to draft 2, not reviewed, the unchanged agreed
ones to agreed at their revision, 13292.108 to the owner, and the owner's decisions each claim's
coding waits for (pending: Q1, the two layers of 13292.108; G2, the causal exception's scope); it
checks the gold with validate.py and copies it to the built-in copy
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
DRAFT1 = 'a9ddd68d02ad05496292484f64e538d1c7955245f607d929d112b27a9d34fdb5'  # gold.json at 0daebfe
WHEN = '2026-11-27'
BY = 'developer (D086 draft 2)'
WHERE = ('wag-workspace verify/chatgpt-d086-2026-11-27/REVIEW.md §3; verify/ethix-2026-11-27/REVIEW.md §2 '
         '(the reconciliation)')


def refuse(why):
    raise SystemExit('recode_d086_r1_d2: ' + why + ': nothing written')


raw = open(GOLD, 'rb').read()
if hashlib.sha256(raw).hexdigest() != DRAFT1:
    refuse('gold.json is not draft 1 of round 1 (run it once, on that gold)')
rounds = json.load(open(ROUNDS, encoding='utf-8'))
gold = json.loads(raw)
G = {it['kialo_claim_id']: it for it in gold['items']}
ROUND1 = [c for c, e in rounds['claims'].items() if e['round'] == 1 or e.get('joined') == 1]
if len(ROUND1) != 31 or any(rounds['claims'][c]['draft'] != 1 or rounds['claims'][c]['review'] != 'not reviewed'
                            for c in ROUND1):
    refuse('round 1 is not the 31 claims at draft 1, not reviewed')

done = {}


def record(cid, fields, what):
    if cid in done:
        refuse(f'{cid} is recorded twice')
    adj = G[cid].setdefault('adjudication', {'changed_fields': [], 'reason': 'agreed'})
    adj.setdefault('recoded', []).append({'date': WHEN, 'by': BY, 'fields': sorted(fields), 'what': what})
    done[cid] = True


def instance(cid, iid):
    for ins in G[cid]['instances']:
        if ins['id'] == iid:
            return ins
    refuse(f'{cid} has no instance {iid}')


def premise(cid, iid, formula, status):
    """The one premise of the instance with this formula, which has the status given."""
    hits = [p for p in instance(cid, iid)['premises'] if p['formula'] == formula]
    if len(hits) != 1 or hits[0]['status'] != status:
        refuse(f'{iid} has no single {status} premise {formula}')
    return hits[0]


def restate(p, status, note, cite=None):
    p['status'] = status
    p['note'] = note
    if cite:
        p['cite'] = cite
    else:
        p.pop('cite', None)


def notes(cid, text, expect):
    if expect not in G[cid]['notes']['text']:
        refuse(f'{cid}\'s notes are not draft 1\'s')
    G[cid]['notes']['text'] = text


# --- Accepted objections ---------------------------------------------------------------------------

# 13292.25: the claim says the burden could be that small, not that the change is easy.
restate(premise('13292.25', '13292.25#1', 'easy_to_achieve(change_food_consumption_pattern)', 'explicit'), 'supplied',
        "supplied: the claim says the burden 'could be as small as' a change of diet, a possibility; that the change "
        "is easy for those the tax burdens, which the argument needs to conclude that the consequence is too slight, "
        "is the reconstruction's")
notes('13292.25',
      "Concedes the burden ('the additional burden') and says it could be small: as small as a change in one's eating "
      "pattern. Its premises: changing one's eating pattern avoids the added cost (stated), and is easy to do "
      "(supplied: the claim says only that the burden could be that small, not that the change is easy for everyone "
      "the tax burdens). It argues that the consequence 13292.7 cites is too slight to count, the exception "
      "too_slight(tax_meat,food_less_affordable_for_poor) of 13292.7#1 (OC9), not that the tax does not bring it "
      "about: its sibling 13292.33 says that, pointing to cheaper protein. No scheme concludes too_slight, so its own "
      "argument is Other, concluding the exception (R8); the game asks CQ:too_slight there, so the role is cq_answer "
      "(§5 item 3), against. The hedge ('could be as small as') is the claim's, recorded in the ease premise's note; "
      "the confidence is the estimate that a careful annotator would choose the same (guidelines §3), not the "
      "strength of the claim.",
      'Its own words hedge')
record('13292.25', ['instances.premises', 'notes'],
       "D086 exchange 1, the reviewer's objection accepted: the ease premise easy_to_achieve(change_food_consumption_"
       "pattern) is supplied, not explicit (the claim says the burden 'could be as small as' a change of diet, not that "
       "the change is easy for those it burdens), with the hedge in its note; the note no longer says that the "
       "confidence carries the hedge (confidence is the expected agreement, guidelines §3). Scheme, role, target "
       "unchanged.")

# 13292.104: #2's fit; #1's comparison and its inferred link.
x2 = instance('13292.104', '13292.104#2')
if x2['scheme']['fit'] != 'explicit':
    refuse('13292.104#2 is not explicit')
x2['scheme']['fit'] = 'reconstructed'
p = premise('13292.104', '13292.104#2',
            'would_bring_about(regulate_animal_farming_industry,farming_industry_not_meat_is_problem,animal_farming_industry_regulated)',
            'explicit')
p['note'] = ("analytic: regulating the industry yields a regulated industry. This S2 is close to the action restated, "
             "which §5 item 4 asks practical reasoning to avoid (S2 a state distinct from the action): a question for "
             "the guidelines (D086 exchange 1)")
restate(premise('13292.104', '13292.104#1',
                'more_effective_means(governments,regulate_animal_farming_industry,environmental_harm_from_meat_reduced,tax_meat)',
                'explicit'), 'supplied',
        "supplied: the claim says the industry, not meat, is the problem, and that regulating it is 'the appropriate "
        "solution'; that regulation reaches the goal more surely than the tax, the comparison better_means needs, is "
        "reconstructed from that diagnosis, not stated. 'Rather than penalize those who wish to consume meat' adds a "
        "reason of fairness to consumers, which is not this comparison (D089)")
restate(premise('13292.104', '13292.104#1', 'brings_about(regulate_animal_farming_industry,environmental_harm_from_meat_reduced)',
                'inferred'), 'explicit',
        "the claim calls regulating the industry 'the appropriate solution' to the problems its Kialo parent 13292.13 "
        "names ('the problems associated with its production and consumption', meat's), the environmental harm among "
        "them: the formula's S2 is that harm, 13292.108's goal, to which this claim's solution is projected (13292.104#2 "
        "states it of the problems in general)")
notes('13292.104',
      "Proposes another action, regulating the animal farming industry, for the goal the thesis's environmental "
      "argument pursues: the problem is the industry, not meat, so act on the industry 'rather than penalize those who "
      "wish to consume meat'. Read as a better means, it answers 13292.108's exception realize_more_effectively("
      "governments,environmental_harm_from_meat_reduced,tax_meat) (OC6; D075, D089): its primary instance is "
      "better_means, a cq_answer to the game's CQ:other_means (22.1 CQ2), with role.move alternative_means, against; "
      "re-parented from 13292.13, a sibling of 13292.108. The comparison better_means needs, that regulation reaches "
      "the goal more surely than the tax, is supplied: the claim diagnoses the source of the problem and calls "
      "regulating it the appropriate solution, from which the comparison is reconstructed, not stated; 'rather than "
      "penalize those who wish to consume meat' adds a reason of fairness to consumers, which is not that comparison "
      "(D089). The argument for the other course is a secondary instance (practical reasoning, reconstructed: its "
      "goal premise is supplied); 'rather than' ties it to the thesis in a third (an argument from alternatives). Its "
      "own goal is the broader 'problems' of 13292.13, not only environmental harm. It also supports 13292.13 by "
      "implying that taxing consumers misses the source of the problems; 13292.208 replies that meat production is "
      "itself the problem.",
      'Proposes another action')
record('13292.104', ['instances.premises', 'instances.scheme.fit', 'notes'],
       "D086 exchange 1, the reviewer's objections accepted: 13292.104#2's fit is reconstructed, since its goal premise "
       "is supplied (guidelines §5 item 6; R18 checks it since this draft); #1's comparison more_effective_means(...) "
       "is supplied, not explicit, its note saying how it is reconstructed from the claim's diagnosis and that the "
       "fairness reason is not it; #1's link brings_about(regulate_animal_farming_industry,...) is explicit, the "
       "claim's own 'the appropriate solution' projected to 13292.108's goal (13292.13, which it first cited, names "
       "the problems but does not say that regulating solves them: the independent check of draft 2, W7). #2's S2, "
       "close to the action restated (§5 item 4), is noted as a question for the guidelines. Better_means kept (the "
       "reviewer: a defensible choice, which the record now exposes). Notes rewritten.")

# 13292.43: the tax's effect presupposed; the evaluation the claim's own.
restate(premise('13292.43', '13292.43#1', 'brings_about(tax_meat,sustainability_and_animal_welfare)', 'explicit'), 'supplied',
        "supplied: presupposed by 'the primary reason to tax meat': that the tax would serve these goods is not "
        "argued; the claim ranks them")
restate(premise('13292.43', '13292.43#1', 'good(sustainability_and_animal_welfare)', 'inferred'), 'explicit',
        "'other, more pressing issues': the claim's own evaluation")
notes('13292.43',
      "Ranks two consequences against 13292.9: the goods the tax would serve (sustainability, animal welfare) press "
      "harder than the lost tax revenue 13292.9 says may follow. A ranking argues the exception of the argument it "
      "answers (OC3, OC4): 13292.9#1, the modal argument (13292.9 claims the loss only possible; OC15), at its "
      "exception good_outweighs(tax_meat,tax_revenue_reduced). Its instance is good_outweighs_bad, C003's answer "
      "scheme, with the goods named as one consequence (no reply splits them, OC11); a cq_answer to the game's "
      "CQ:good_outweighs (33.2 CQ3), against. That the tax would bring these goods about is presupposed, not argued: "
      "the premise is supplied, and the ranking does not show it. That they are good is the claim's own evaluation "
      "('other, more pressing issues'). The values it names are things the tax would promote, so they count as "
      "consequences (OC5). Its Kialo parent cites the outweighed consequence, so it is not re-parented.",
      'Ranks two consequences against 13292.9')
record('13292.43', ['instances.premises', 'notes'],
       "D086 exchange 1, the reviewer's objection accepted: brings_about(tax_meat,sustainability_and_animal_welfare) "
       "is supplied (presupposed: the claim ranks the goods, it does not argue that the tax brings them), and "
       "good(sustainability_and_animal_welfare) explicit (the claim's own evaluation, 'more pressing issues'), not "
       "inferred without a cited claim (R19). Notes rewritten.")

# 13292.127: the starvation link presupposed, 13292.7 grounding only the dearer food.
LINK_127 = 'brings_about(tax_meat,starvation_from_reduced_meat_access)'
restate(premise('13292.127', '13292.127#1', LINK_127, 'inferred'), 'supplied',
        "supplied: presupposed, as in 13292.127#2; 13292.7 grounds only the first step, the tax making food dearer, "
        "not the step from there to starvation")
restate(premise('13292.127', '13292.127#2', LINK_127, 'inferred'), 'supplied',
        "supplied: presupposed by 'starvation from decreased access to meat'; 13292.7 grounds only the first step, the "
        "tax making food dearer, not the step from there to starvation. 13292.204 denies this link at the weighing's "
        "own premise (13292.127#1, OC10), and 13292.202 amends the action at this argument's conclusion")
restate(premise('13292.127', '13292.127#1', 'bad(starvation_from_reduced_meat_access)', 'inferred'), 'explicit',
        '"a more severe threat"')
notes('13292.127',
      "The claim ranks two consequences: starvation from reduced access to meat is a more severe threat than any "
      "consequence of high meat consumption, the harms the pros cite. A ranking argues the exception of the argument it "
      "weighs against (OC3, OC4). It names a class, so it is coded against one argument of the class (OC7): its Kialo "
      "parent, 13292.7, argues against the tax and is not in it, so the earliest argument of the class on the board, "
      "13292.108's practical reasoning, at its exception bad_outweighs(tax_meat,environmental_harm_from_meat_reduced) "
      "(D070, D091); re-parented from 13292.7. That one argument stands for the class: the claim ranks starvation above "
      "any consequence of high meat consumption, more than this coding states. The weighing is primary because the "
      "words state the comparison (OC8): an instance of bad_outweighs_good, C003's answer scheme, a cq_answer to the "
      "game's CQ:bad_outweighs (22.1 CQ5), against. The starvation consequence is a secondary negative-consequences "
      "instance (13292.127#2). The link from the tax to starvation is presupposed, not stated, and supplied in both "
      "instances: 13292.7 grounds only its first step, the tax making food dearer ('would make food more expensive'), "
      "not the step from dearer food to starvation; the claim presupposes the link without a 'could', so the "
      "categorical scheme (the mood rule, D008). That starvation is bad is the claim's own evaluation ('a more severe "
      "threat'), explicit in both. Its children engage the starvation consequence: 13292.202 amends the action, at the "
      "conclusion of the starvation argument 13292.127#2 (OC6), and 13292.204 denies the link the weighing relies on "
      "(OC10).",
      'The claim ranks two consequences')
record('13292.127', ['instances.premises', 'notes'],
       "D086 exchange 1, the reviewer's objection accepted: the link brings_about(tax_meat,starvation_from_reduced_"
       "meat_access) is supplied in both instances, not inferred from 13292.7, which grounds only the dearer food, not "
       "the step to starvation; #1's bad(starvation_from_reduced_meat_access) is explicit, as #2's, from the same words "
       "('a more severe threat'). The note says that one argument stands for the class the claim ranks against (OC7), "
       "and no longer derives the starvation link from 13292.7's prediction. Notes rewritten.")

# 13292.192: the instance that reaches the target is primary.
a, b = instance('13292.192', '13292.192#1'), instance('13292.192', '13292.192#2')
if not (a['primary'] and not b['primary']):
    refuse('13292.192#1 is not the primary')
a['primary'], b['primary'] = False, True
notes('13292.192',
      "Attacks the link 13292.6 states (less meat demand -> less deforestation) with an offsetting causal chain: less "
      "meat -> more crops -> deforestation for cropland. Each step's cause is only possible, so cause_to_effect_modal "
      "(§5 item 11): 13292.192#1 (reduced meat consumption -> more crop production) and 13292.192#2 (more crop "
      "production -> deforestation for cropland), primary, since its conclusion reaches the target (§5 item 5). The "
      "target is 13292.6's stated would_be_realized(deforestation_slowed,reduced_meat_consumption); no critical "
      "question is about it, so premise_attack (§5 item 3), against. The chain shows only that some deforestation may "
      "still occur, which does not by itself contradict its slowing: the attack needs a further premise the claim does "
      "not state, that the land cleared for crops would match or exceed the pasture spared. Its child 13292.200 "
      "replies so: the claim was only that deforestation would slow or reverse.",
      'Attacks the link 13292.6 states')
record('13292.192', ['instances.primary', 'notes'],
       "D086 exchange 1, the reviewer's objection accepted: 13292.192#2, whose conclusion may_occur(deforestation_for_"
       "cropland) reaches the target, is primary (§5 item 5), keeping its id; #1 secondary. The note says that the "
       "chain shows only that some deforestation may occur, not that it does not slow, and names the net-effect "
       "premise the claim leaves unstated (13292.200's reply). Notes rewritten.")

# 13292.8: a background condition, backed in part.
notes('13292.8',
      "Backs a background condition of 13292.115's premise brings_about(tax_meat,reduced_unhealthy_meat_consumption): "
      "that there is unhealthy (excess) meat consumption for the tax to reduce. That condition is not the premise, "
      "which says that the tax would reduce such consumption: the claim backs it only in part, and the step from excess "
      "consumption to the tax's reducing it is not argued. The target is the nearest formula the format offers (a "
      "target names no component of a premise), and the relation is partial. positive_consequences has CQ:likelihood "
      "(33.1 CQ1) on that premise, so the role is cq_answer with polarity for (R2). The step from 'more meat than "
      "dietary guidelines recommend' to unhealthy consumption rests on the guidelines' authority, coded as a "
      "secondary expert-opinion instance. "
      "Its children treat it as stating the problem: 13292.3 (curbing it cuts health costs) and 13292.29 (it is a "
      "personal choice).",
      'Backs the presupposition of 13292.115')
record('13292.8', ['notes'],
       "D086 exchange 1, the reviewer's objection partly accepted: the note no longer equates the background condition "
       "(there is excess consumption) with the premise it backs (the tax would reduce it), as its 'i.e.' did; it says "
       "the backing is partial and the target the nearest formula the format offers. Not taken: the reviewer's "
       "unresolved or approximate target, since the format has no approximate target and an unresolved one could not "
       "be verified (R17); how to target a background condition that backs a premise in part is a question for the "
       "guidelines, to the owner. Scheme, role and target unchanged.")

# 13292.194: at the conclusion of the argument that cites the harm, as 13292.202.
tgt = G['13292.194']['target']['actual']
if (tgt['instance'], tgt['part']) != ('13292.12#1', 'conclusion'):
    refuse('13292.194 is not at 13292.12#1\'s conclusion')
c3 = instance('13292.12', '13292.12#3')
if (c3.get('conclusion') or {}).get('formula') != '¬should_be_performed(tax_meat)':
    refuse('13292.12#3 does not conclude ¬should_be_performed(tax_meat)')
tgt.update({'claim': '13292.12', 'instance': '13292.12#3', 'part': 'conclusion', 'formula': '¬should_be_performed(tax_meat)'})
notes('13292.194',
      "An amended action that mitigates the harm 13292.12 predicts (C006; OC6): the meat tax plus import measures "
      "(eco-tariffs; imports that verifiably meet environmental and animal-welfare standards, or pass a high enough tax "
      "barrier), tax_meat_with_import_safeguards. 'Can solve this issue' is can of capability (the mood rule, D008): "
      "positive consequences of the amended action, role action_amendment, against. As for 13292.202 (OC6), the "
      "amendment is at the conclusion of the argument that cites the harm against the tax, 13292.12#3's "
      "¬should_be_performed(tax_meat): an amended policy answers the argument against the tax, and its conclusion is "
      "not the negation of may_occur(domestic_meat_industry_pushed_out), the predicted harm, where the reference and "
      "both annotators put it. Nor is it a cq_answer to 28 CQ3 at 13292.12#1's interference: that would say that "
      "something under the unamended tax already blocks the mechanism, while the claim proposes a changed policy. Its "
      "sibling 13292.45 (round 2) is coded the same way when its round comes.",
      'An amended action that mitigates')
record('13292.194', ['notes', 'target'],
       "D086 exchange 1, the reviewer's objection accepted (the developer's own first reading): retargeted from "
       "13292.12#1's conclusion may_occur(domestic_meat_industry_pushed_out) to 13292.12#3's ¬should_be_performed("
       "tax_meat), the conclusion of the argument that cites the harm against the tax, as 13292.202 is (OC6). The note "
       "gives the reason as an amended policy against a blocking circumstance that already holds, not as whose action "
       "the measures are. The reviewer's optional second relation, the measures as interference with 13292.12#1's "
       "mechanism, is not coded (a claim has one target, and the claim proposes a changed policy, not a blocker under "
       "the tax): 13292.12's challenge at that interference names the measures. Scheme, role and polarity unchanged. "
       "Notes rewritten.")


# --- Partly accepted, with the reconciliation's conditions -----------------------------------------

# 13292.123: better_means kept; the comparison and the link supplied and scoped; the other reading recorded.
x1 = instance('13292.123', '13292.123#1')
restate(premise('13292.123', '13292.123#1',
                'more_effective_means(governments,abolish_meat_subsidies,meat_price_reflects_real_costs,tax_meat)',
                'explicit'), 'supplied',
        "supplied: 'instead' offers the other course; that it reaches the goal more surely than the tax, the comparison "
        "better_means needs, is reconstructed, not stated. Its scope: abolishing the subsidies removes the distortion "
        "13292.11 names (a price held below cost by subsidies); it does not put into the price the external costs "
        "13292.218 names, which the tax would")
restate(premise('13292.123', '13292.123#1', 'brings_about(abolish_meat_subsidies,meat_price_reflects_real_costs)', 'inferred'),
        'supplied',
        "supplied: 'distortions caused by subsidies should be removed by abolishing those subsidies' says "
        "that abolishing them removes the distortion; that the price then reflects the real costs, external costs "
        "included, is the reconstruction's, and holds of the subsidies' part only")
restate(premise('13292.123', '13292.123#2',
                'would_be_realized(meat_price_reflects_real_costs,subsidy_distortions_removed)', 'inferred'),
        'supplied',
        "supplied: that the price reflects the real costs once the subsidies' distortion is removed is the "
        "reconstruction's, as in 13292.123#1, and holds of the subsidies' part only; 13292.118 states the goal, not "
        "this link")
alts = x1['scheme']['alternatives']
if [x['walton_id'] for x in alts] != ['practical_reasoning2']:
    refuse('13292.123#1\'s alternatives are not draft 1\'s')
alts[0]['why_not'] = (
    "the other reading (the reviewer's, D086 exchange 1): a proposal of another course, not said to do better (C017; "
    "OC6), its primary the argument for abolishing the subsidies (13292.123#2) at 13292.118's conclusion. Kept as an "
    "alternative: the draft reads 'instead' as offering the other course for 13292.118's goal, as 13292.104 does for "
    "13292.108's, its comparison supplied and scoped")
notes('13292.123',
      "Accepts 13292.11's diagnosis (subsidies keep the price below the real costs) and proposes another means: abolish "
      "the subsidies (or exempt meat from them) instead of taxing. Read as a better means, it answers 13292.118's "
      "exception realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat) (OC6; D075, D089), as "
      "13292.104 does 13292.108's: its primary instance is better_means, a cq_answer to the game's CQ:other_means "
      "(22.1 CQ2), with role.move alternative_means, against; re-parented from 13292.11 (13292.118 is an ancestor). The "
      "goal is 13292.118's, as D075 asks of a better means. The comparison and the link from abolishing the subsidies "
      "to that goal (in both instances) are supplied, not stated: 'instead' offers the other course and does not say "
      "that it reaches the goal more surely. Their scope is narrower than the goal: abolishing the subsidies removes the distortion "
      "13292.11 names (a price held below cost by subsidies); it does not put into the price the external costs "
      "13292.218 names, which the tax would. The other reading, a proposal of another course not said to do better "
      "(C017; OC6) at 13292.118's conclusion, is recorded as an alternative of the primary instance. The argument for "
      "the other course is a secondary instance; the incoherence sentence is a third, from opposition (11.2), coded as "
      "Other because walton.yml's 11.2 encoding is self-contradictory.",
      "Accepts 13292.11's diagnosis")
record('13292.123', ['instances.premises', 'instances.scheme.alternatives', 'notes'],
       "D086 exchange 1, the reviewer's objection partly accepted, on the conditions of its reconciliation: better_means "
       "kept, with the same goal as 13292.118's argument (D075); its comparison more_effective_means(...) and its link "
       "brings_about(abolish_meat_subsidies,meat_price_reflects_real_costs) supplied, not explicit or inferred, their "
       "notes and the item's scoping them (subsidy removal corrects the distortion 13292.11 names, not the "
       "externalities 13292.218 names), and the same link in 13292.123#2, would_be_realized(meat_price_reflects_"
       "real_costs,subsidy_distortions_removed), supplied too, not inferred from 13292.118, which states the goal, not "
       "the link (the independent check of draft 2, W7); the reviewer's reading, a proposal of another course not "
       "said to do better (C017; OC6), recorded as the primary's alternative. Role and target unchanged. Notes "
       "rewritten.")

# 13292.218: the action named but not represented; the circumstances inferred; the link's scope.
x = instance('13292.218', '13292.218#1')
if not isinstance(x['bindings'].get('A'), dict):
    refuse('13292.218#1 binds A')
x['bindings']['A'] = {'unstated': (
    "the action: named in the claim, the meat tax ('it' in the Kialo text, resolved in the literal), but not "
    "represented in this fragment, which states no formula with it: the transition from the tax to the higher price, "
    "would_bring_about, is 13292.118's")}
restate(premise('13292.218', '13292.218#1', 'current_circumstances(meat_price_below_real_costs)', 'explicit'), 'inferred',
        "inferred from the sibling 13292.11, which states it: this claim elaborates the circumstances (it names the costs "
        "the price leaves out: externalities, from climate change, water pollution and disease risks) without stating "
        "them as a proposition of its own ('it would factor in externalities', into the price 13292.118 speaks of, "
        "presupposes that the price leaves them out). 13292.224 backs them here, where their externalities are named", cite='13292.11')
p = premise('13292.218', '13292.218#1', 'would_be_realized(meat_price_reflects_real_costs,higher_meat_price)', 'explicit')
p['note'] = ("'It would factor in externalities': the price the tax raises includes the external costs. The "
             "formula's S2 is 13292.118's higher_meat_price, to which this claim's price with the externalities in it is "
             "projected, so that it backs that premise. Its scope: the price rise that internalizes the named externalities "
             "under this tax, not any higher price; the restriction is this claim's words, the projection 13292.118's "
             "reconstruction")
notes('13292.218',
      "'It' is the meat tax of 13292.118 (the literal resolves it). The claim restates 13292.118's link from the higher "
      "price to its goal with content: the price would then include the external costs (climate change, water "
      "pollution, disease risks). A fragment of 13292.118's goal-based argument (guidelines §3; §5 item 6): its stated "
      "link is would_be_realized(meat_price_reflects_real_costs,higher_meat_price), explicit, and the circumstances it "
      "elaborates, current_circumstances(meat_price_below_real_costs), are inferred from its sibling 13292.11, which "
      "states them (the claim presupposes them, naming the costs left out). The action is named ('it') but not "
      "represented here, since the fragment states no formula with it: the transition from the tax, would_bring_about, "
      "is 13292.118's; so is the agent, and the conclusion is left unstated. The S2 of the stated link is 13292.118's "
      "higher_meat_price, projected from the claim's price with the externalities in it: the link holds of the price "
      "rise that internalizes those externalities, not of any higher price. It backs 13292.118's premise "
      "would_be_realized(...), about which no critical question is asked, so premise_backing (§5 item 3); "
      "'clarification' is a possible alternative role. 13292.224 backs the circumstances it elaborates.",
      "'It' is the meat tax of 13292.118")
record('13292.218', ['instances.bindings', 'instances.premises', 'notes'],
       "D086 exchange 1, the reviewer's objection partly accepted, with its reconciliation's scope: A's reason says that "
       "the action is named ('it') but not represented in the fragment, which is not the same as unstated; the "
       "circumstances current_circumstances(meat_price_below_real_costs) are inferred from 13292.11, which states them "
       "(R19), this claim elaborating them; the stated link would_be_realized(meat_price_reflects_real_costs,higher_"
       "meat_price) stays explicit, its note saying how this claim's price with the externalities in it is projected "
       "to 13292.118's higher_meat_price and that the link holds of the price rise that internalizes those "
       "externalities, not of any higher price. Role and target unchanged. Notes rewritten.")


# --- Refinements of agreed items ------------------------------------------------------------------

p = premise('13292.11', '13292.11#2', 'has_occurred(meat_subsidies)', 'explicit')
p['note'] = ("explicit: 'due to subsidies' presents the subsidies as existing, so the cause has occurred as the claim "
             "has it (not a hypothetical cause)")
OLD_11 = 'the subsidies exist, so the cause has occurred and cause_to_effect applies'
if OLD_11 not in G['13292.11']['notes']['text']:
    refuse('13292.11\'s notes are not draft 1\'s')
G['13292.11']['notes']['text'] = G['13292.11']['notes']['text'].replace(
    OLD_11, 'the claim presents the subsidies as existing, so the cause has occurred as the claim has it, and '
            'cause_to_effect applies')
record('13292.11', ['instances.premises', 'notes'],
       "D086 exchange 1, the reviewer's refinement taken: the note and the occurrence premise's note say that the claim "
       "presents the subsidies as existing, rather than assert that they exist. Coding unchanged.")

alts = instance('13292.6', '13292.6#1')['scheme']['alternatives']
if [x['walton_id'] for x in alts] != ['cause_to_effect', 'positive_consequences']:
    refuse('13292.6#1\'s alternatives are not draft 1\'s')
alts.append({'walton_id': 'cause_to_effect_modal', 'why_not': (
    "a causal reading of the link fits the modal form, its cause (less meat eaten) only possible (§5 item 11), which "
    "the reason against cause_to_effect does not rule out; the claim offers the link as the goal its parent's "
    "resulting state realizes (S2 -> G), so the fragment of 13292.108's practical reasoning")})
record('13292.6', ['instances.scheme.alternatives'],
       "D086 exchange 1, the reviewer's refinement taken: cause_to_effect_modal recorded as an alternative of "
       "13292.6#1, since the reason against cause_to_effect (the cause has not occurred) does not rule out the modal "
       "form. Coding unchanged.")


# --- Corrections of agreed items that draft 2's changes made false (the independent check of draft 2) --

# 13292.204: draft 2 made the premise it attacks supplied in 13292.127 (E2).
OLD_204 = "The premise is 13292.127's own (inferred from 13292.7), so this is not re-parenting."
if OLD_204 not in G['13292.204']['notes']['text']:
    refuse('13292.204\'s notes are not draft 1\'s')
G['13292.204']['notes']['text'] = G['13292.204']['notes']['text'].replace(
    OLD_204, "The premise is 13292.127's own (supplied: the claim presupposes the link, and 13292.7 grounds only its "
             "first step, the tax making food dearer), so this is not re-parenting.")
record('13292.204', ['notes'],
       "D086 draft 2, after the independent check of draft 2 (E2): the note's account of the premise it attacks "
       "corrected, which draft 2 made supplied in 13292.127 (the note said inferred from 13292.7). Coding unchanged.")

# 13292.12: draft 2 moved 13292.194 to 13292.12#3's conclusion (W1); a quotation of 13292.9 made exact.
ch = [c for c in G['13292.12']['challenges'] if c.get('instance') == '13292.12#1' and c.get('cq') == '28 CQ3']
OLD_12C = ("the import measures 13292.45 and 13292.194 propose would block it; they are coded as amended actions at "
           "13292.12's predicted harm (OC6), not as answers here")
if len(ch) != 1 or ch[0].get('note') != OLD_12C:
    refuse('13292.12\'s challenge at the interference is not draft 1\'s')
ch[0]['note'] = ("the import measures 13292.45 and 13292.194 propose would block it; they are coded as amended actions "
                 "(OC6), not as answers here: 13292.194 at 13292.12#3's conclusion, the argument that cites the harm "
                 "against the tax, and 13292.45, until its round, at this argument's conclusion, the predicted harm")
t = G['13292.12']['notes']['text']
OLD_12Q = "Gives the mechanism of 13292.9's 'by harming the meat industry':"
OLD_12N = ("(13292.131 backs the badness; 13292.45 and 13292.194 propose remedies, amended actions at 13292.12's "
           "predicted harm)")
if t.count(OLD_12Q) != 1 or t.count(OLD_12N) != 1:
    refuse('13292.12\'s notes are not draft 1\'s')
G['13292.12']['notes']['text'] = t.replace(
    OLD_12Q, "Gives the mechanism of 13292.9's 'by harming the trillion-dollar meat industry':").replace(
    OLD_12N, "(13292.131 backs the badness; 13292.45 and 13292.194 propose remedies, amended actions: 13292.194 at "
             "13292.12#3's conclusion, 13292.45, until its round, at 13292.12#1's, the predicted harm)")
record('13292.12', ['challenges', 'notes'],
       "D086 draft 2, after the independent check of draft 2 (W1): the challenge at 13292.12#1's interference and the "
       "notes say where 13292.194 now is, at 13292.12#3's conclusion, not at the predicted harm; the quotation of "
       "13292.9 made exact ('the trillion-dollar meat industry'). Coding unchanged.")


# --- d086-rounds.json: the first exchange, the states of draft 2, the decisions pending ------------

def rev(it):
    x = {k: v for k, v in it.items() if k not in ('adjudication', 'verified')}
    return hashlib.sha256(json.dumps(x, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8')).hexdigest()


ACCEPTED = ('13292.25', '13292.104', '13292.43', '13292.127', '13292.192')
PARTLY = ('13292.123', '13292.218')
REFINED = ('13292.11', '13292.6')
# The developer's answer where it is none of those (the independent check of draft 2, N3).
OTHER = {'13292.8': ('objection', 'partly accepted: the note rewritten; the unresolved or approximate target not '
                                  'taken, a question for the guidelines (to the owner)'),
         '13292.194': ('objection', 'accepted; its optional second relation, the interference at 13292.12#1, noted, '
                                    'not coded')}
# Agreed in exchange 1, corrected in draft 2 where draft 2's changes made their records false (the
# independent check of draft 2, E2 and W1): back to not reviewed, at draft 2.
CORRECTED = ('13292.204', '13292.12')
# Q1: the targets on 13292.108's conclusion and on its supplied premises, which move to a separate
# reconstruction under the two-layer rule (fragments carry no conclusion), and 13292.108 itself;
# G2: the items with a causal instance, whose exception changes with the scope, and the answers at
# 13292.180#2's exception (wag-workspace verify/ethix-2026-11-27/scripts/dependencies.py).
PENDING = {'13292.108': ['Q1'], '13292.5': ['Q1'], '13292.6': ['Q1'], '13292.147': ['Q1'], '13292.174': ['Q1'],
           '13292.104': ['Q1'], '13292.127': ['Q1'],
           '13292.2': ['G2'], '13292.9': ['G2'], '13292.11': ['G2'], '13292.12': ['G2'], '13292.36': ['G2'],
           '13292.141': ['G2'], '13292.180': ['G2'], '13292.182': ['G2'], '13292.198': ['G2'], '13292.192': ['G2']}
if set(done) != set(ACCEPTED) | set(PARTLY) | set(REFINED) | set(OTHER) | set(CORRECTED):
    refuse('the items recorded are not the ones changed')
for c in ROUND1:
    e = rounds['claims'][c]
    if c in ACCEPTED:
        reviewer, developer = 'objection', 'accepted'
    elif c in PARTLY:
        reviewer, developer = 'objection', 'partly accepted, on the conditions of the reconciliation'
    elif c in REFINED:
        reviewer, developer = 'agreed, with a refinement', 'refinement taken'
    elif c in OTHER:
        reviewer, developer = OTHER[c]
    elif c == '13292.108':
        reviewer, developer = 'objection', 'agreed in direction; a rule for the owner first (question 1)'
    else:
        reviewer, developer = 'agreed', 'agreed'
    e['exchanges'] = [{'exchange': 1, 'date': WHEN, 'draft': 1, 'reviewer': reviewer, 'developer': developer,
                       'record': WHERE}]
    if c in done:
        e.update(draft=2, review='not reviewed', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])
    elif c == '13292.108':
        e.update(review='to the owner', agreed_rev=None, agreed_by=None, agreed_date=None, objections=[])
    else:
        e.update(review='agreed', agreed_rev=rev(G[c]), agreed_by='the reviewer and the developer', agreed_date=WHEN,
                 objections=[])
    if c in PENDING:
        e['pending'] = PENDING[c]
if set(PENDING) - set(ROUND1):
    refuse('a pending claim is not of round 1')
rounds['note'] = (
    'The rounds of the first gold (D086; docs/d086-first-gold.md): which claims each round takes and why '
    '(d086_rounds.py, from the reference named by its hash), and where each claim stands: its latest draft, the '
    "review's state (not drafted, not reviewed, objections, agreed, to the owner), the revision of the item agreed "
    'on (verify_gold.py writes a verification only at that revision), its review exchanges (exchanges), and the '
    "owner's decisions its coding waits for (pending; verify_gold.py writes no verification while there are any).")

AGREED = sum(1 for c in ROUND1 if rounds['claims'][c]['review'] == 'agreed')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Round 1 of the first gold, draft 2 (D086; docs/d086-first-gold.md §11, d086-rounds.json): the first "
             "exchange of the review taken in. Accepted objections: 13292.25, .104, .43, .127, .192 and .194 (its "
             "optional second relation noted, not coded); partly accepted: 13292.8 (the note rewritten, the target "
             "kept), and, on the conditions of the reconciliation, 13292.123 and .218; refinements of agreed items: "
             "13292.11 and .6. Corrected after the independent check of draft 2: 13292.204 and .12, agreed items whose "
             "records draft 2's changes made false. Premise provenance follows guidelines §2 item 3 and §5 item 6, "
             f"now checked (R18, R19). 13292.108 goes to the owner (question 1); the {AGREED} other agreed items are "
             f"agreed at their revision. {len(PENDING)} of the round's claims wait for the owner's decisions "
             f"(pending): {sum(1 for v in PENDING.values() if 'Q1' in v)} for Q1, the two layers of 13292.108, and "
             f"{sum(1 for v in PENDING.values() if 'G2' in v)} for G2, the causal exception's scope.")})

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
with open(ROUNDS, 'wb') as fh:
    fh.write(json.dumps(rounds, indent=1, ensure_ascii=False).encode('utf-8') + b'\n')
print(f'{len(done)} items recoded; gold.json and d086-rounds.json written; checking the gold (validate.py)')
run = subprocess.run([sys.executable, '-I', DOC + 'validate.py'], capture_output=True, text=True)
print(run.stdout.strip()[-3000:])
if run.returncode != 0:
    raise SystemExit('validate.py refuses the drafted gold: fix it, then copy it to ' + COPY)
if os.path.isdir(COPY) and os.environ.get('NO_COPY') != '1':
    shutil.copyfile(GOLD, COPY + 'gold.json')
    print('copied to', COPY)
