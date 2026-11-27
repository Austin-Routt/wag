"""Round 1 of the first gold, draft 1 (D086; docs/d086-first-gold.md §3).

    python3 -I recode_d086_r1.py

A one-off, kept as the record. It writes the developer's draft of round 1's 28 claims
(d086-rounds.json) into the gold, each item with an adjudication.recoded entry (by "developer
(D086 draft 1)") that names what changed and why, or says the reference's coding is kept and why:

- Changed. The four targets format 1 forced onto a claim's primary instance (adjudication.md R5)
  name the causal step they answer, as both annotators aimed them (annotation-records.md, "Open
  for the first gold's draft"): 13292.12 at 13292.9#2's causes(tax_meat,meat_industry_harmed),
  the book's 28 CQ1 (a mechanism for the link); 13292.141 at 13292.2#2's
  causes(meat_production,significant_ghg_emissions), 28 CQ2 (data offered for the link); 13292.182
  and 13292.198 at 13292.180#2's exception interference(reduced_meat_consumption), 28 CQ3, their
  own arguments concluding it (R8). The true fragments 13292.6, 13292.11 and 13292.218 follow
  guidelines §5 item 6: they bind only the variables of the formulas they state, the others
  {"unstated": why}, list those formulas only, and their conclusions, which name the action, are
  null; the challenge records at formulas they no longer list are left to the arguments they are
  fragments of, which record the same challenges.
- Kept. 13292.108 keeps every variable bound: the thesis has no argument, so the fragment carries
  the environmental argument, and §5 item 6 would leave eight of the eleven targets into it
  nothing to point at (a question for the review and the owner). The other 20 claims keep the
  reference's coding, which the developer's reading of each matched or, where it did not, the
  guidelines decide for the reference (workspace verify/d086/round1-readings.md).
- Notes. Where a note told the coding's history in dated additions, or described a coding the item
  no longer has, it is rewritten as one current account; the history is in the item's recoded
  entries and in gold-format1.json. 13292.147 names candidate C017, as OC6 asks.

It refuses any gold but the one the rounds were measured on (d086-rounds.json's reference), then
writes gold.json, sets each round-1 claim's draft to 1 and its review to "not reviewed" in
d086-rounds.json, checks the gold with validate.py, and copies it to the built-in copy
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
if any(rounds['claims'][c]['review'] != 'not drafted' for c in ROUND1):
    refuse('a claim of round 1 is drafted already')

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
    if part != 'conclusion' and formula not in [p['formula'] for p in want['premises']]:
        refuse(f'{formula} is not a formula of {iid}')


# --- Changed: the targets format 1 forced onto a primary instance --------------------------------

# 13292.12: the mechanism of 13292.9's first step.
set_target('13292.12', '13292.9', '13292.9#2', 'premise', 'causes(tax_meat,meat_industry_harmed)')
G['13292.12']['role'].update({'kind': 'cq_answer', 'cq': '28 CQ1', 'polarity': 'for'})
record('13292.12', ['role', 'target', 'notes'],
       "Retargeted from the modal premise of 13292.9#1 (format 1 required a formula of the primary instance; "
       "adjudication.md R5) to the causal step it backs, 13292.9#2's causes(tax_meat,meat_industry_harmed), as both "
       "annotators aimed it; a cq_answer to the book's 28 CQ1 there (a mechanism offered for the link; the game asks no "
       "question at the formula; guidelines §5 item 3), polarity for. The replay plays each claim's primary instance and a "
       "secondary one that carries it, so this claim is a cascade there now. Notes rewritten as one current account.")

# 13292.141: data for 13292.2's causal step.
set_target('13292.141', '13292.2', '13292.2#2', 'premise', 'causes(meat_production,significant_ghg_emissions)')
G['13292.141']['role'].update({'kind': 'cq_answer', 'cq': '28 CQ2', 'polarity': 'for'})
record('13292.141', ['role', 'target', 'notes'],
       "Retargeted from 13292.2#1's classification premise instance(meat_production,significant_ghg_source) (14 CQ1; "
       "format 1 required a formula of the primary instance) to the causal step it backs, 13292.2#2's "
       "causes(meat_production,significant_ghg_emissions), as both annotators aimed it; a cq_answer to the book's 28 CQ2 "
       "there (data offered for the link; a mechanism would answer 28 CQ1), polarity for. A cascade in the replay now. "
       "Notes rewritten.")

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
    record(cid, ['instances.conclusion', 'instances.scheme.description', 'role', 'target', 'notes'],
           "Retargeted from 13292.180#1's premise brings_about(tax_meat,more_anemia_cases) (CQ:likelihood; format 1 "
           "required a formula of the primary instance) to the causal step it answers, the exception "
           "interference(reduced_meat_consumption) of 13292.180#2, the book's 28 CQ3 (other factors that counteract the "
           "production of the effect; the game asks no question there), as both annotators aimed it (A an undercutter, "
           "B a cq_answer to 28 CQ3); its Other instance concludes that exception now (R8). A cascade in the replay now. "
           "Notes rewritten.")


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


PARENT_108 = "13292.108's environmental argument has it"
fr = fragment('13292.6', '13292.6#1', ('G', 'S2'),
              {'A': f'the action: {PARENT_108} (tax_meat); this claim does not state it',
               'S1': f'the circumstances: {PARENT_108}; this claim presupposes them, it does not state them',
               'P': f'the agent: {PARENT_108} (governments); this claim does not state it'},
              ['would_be_realized(deforestation_slowed,reduced_meat_consumption)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.108's argument (should_be_performed(tax_meat) there)."}
G['13292.6']['challenges'] = [c for c in G['13292.6']['challenges'] if c['formula'] == 'would_be_realized(deforestation_slowed,reduced_meat_consumption)']
record('13292.6', ['instances.bindings', 'instances.conclusion', 'instances.premises', 'challenges', 'notes'],
       "Guidelines §5 item 6 applied (the identity contract left it open for this draft): the fragment binds only the "
       "variables of the formula it states, G and S2 of would_be_realized(deforestation_slowed,reduced_meat_consumption); "
       "A, S1 and P are unstated, its premises are that formula alone, and its conclusion, which names A, is null. The "
       "challenge at would_bring_about(...) is left to 13292.108, whose own record has it. 13292.192's target, the stated "
       "link, is unchanged. Notes rewritten.")

PARENT_118 = "13292.118's argument has it"
fr = fragment('13292.11', '13292.11#1', ('S1',),
              {'A': f'the action: {PARENT_118} (tax_meat); this claim does not state it',
               'S2': f'the resulting state: {PARENT_118} (higher_meat_price); this claim does not state it',
               'G': f'the goal: {PARENT_118} (meat_price_reflects_real_costs); this claim does not state it',
               'P': f'the agent: {PARENT_118} (governments); this claim does not state it'},
              ['current_circumstances(meat_price_below_real_costs)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.118's argument (should_be_performed(tax_meat) there)."}
G['13292.11']['challenges'] = [c for c in G['13292.11']['challenges']
                               if c['instance'] == '13292.11#2' or c['formula'] == 'current_circumstances(meat_price_below_real_costs)']
record('13292.11', ['instances.bindings', 'instances.conclusion', 'instances.premises', 'challenges', 'notes'],
       "Guidelines §5 item 6 applied: 13292.11#1 binds S1 alone, the variable of the formula it states "
       "(current_circumstances(meat_price_below_real_costs)); A, S2, G and P are unstated, its premises are that formula "
       "alone, its conclusion null. The challenge at realize_more_effectively(...) is left to 13292.118, whose own record "
       "has it. 13292.11#2, the causal explanation, unchanged. Notes rewritten.")

fr = fragment('13292.218', '13292.218#1', ('S1', 'S2', 'G'),
              {'A': f'the action: {PARENT_118} (tax_meat, the "it" of the claim); this claim does not name it',
               'P': f'the agent: {PARENT_118} (governments); this claim does not state it'},
              ['current_circumstances(meat_price_below_real_costs)',
               'would_be_realized(meat_price_reflects_real_costs,higher_meat_price)'])
fr['conclusion'] = {'formula': None, 'en': "Not stated: the conclusion names the action, which this fragment leaves to "
                                           "13292.118's argument (should_be_performed(tax_meat) there)."}
G['13292.218']['challenges'] = [c for c in G['13292.218']['challenges']
                                if c['formula'] in ('current_circumstances(meat_price_below_real_costs)',
                                                    'would_be_realized(meat_price_reflects_real_costs,higher_meat_price)')]
record('13292.218', ['instances.bindings', 'instances.conclusion', 'instances.premises', 'challenges', 'notes'],
       "Guidelines §5 item 6 applied: the fragment binds S1, S2 and G, the variables of the formulas it states "
       "(current_circumstances(meat_price_below_real_costs), would_be_realized(meat_price_reflects_real_costs,"
       "higher_meat_price)); A and P are unstated, its premises are those two formulas, its conclusion null. The "
       "challenge at realize_more_effectively(...) is left to 13292.118, whose own record has it. 13292.224's target, the "
       "stated circumstances, is unchanged. Notes rewritten.")
instance('13292.11', '13292.11#1')['note'] = "The claim states the circumstances S1 of 13292.118's goal-based argument, and nothing else of it."


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
        "bad_outweighs(A,G). Guidelines §5 item 6 would leave A, S2, G and P unstated and the conclusion null, and eight "
        "of the eleven targets into this argument would point at nothing: whether a fragment whose parent has no "
        "argument carries that argument is a question the guidelines leave open (D086 round 1, for the review and the "
        "owner). Goal-based because no value is stated."),
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
        "higher_meat_price); the action and the agent are 13292.118's and are left unstated, and so is the conclusion. "
        "It backs 13292.118's premise would_be_realized(...), about which no critical question is asked, so "
        "premise_backing (§5 item 3); 'clarification' is a possible alternative role. 13292.224 backs the circumstances "
        "it states."),
    '13292.12': (
        "Gives the mechanism of 13292.9's 'by harming the meat industry': the tax would make imported meat cheaper "
        "(dumping, predatory pricing), and cheaper imports would push domestic producers out of the market. Two causal "
        "steps whose cause is only proposed, so cause_to_effect_modal (§5 item 11): 13292.12#2 (the tax -> cheaper "
        "imports) and 13292.12#1 (cheaper imports -> domestic producers pushed out), primary as the step that ends at "
        "the harm to the industry. Together they back the causal premise of 13292.9's first step, "
        "causes(tax_meat,meat_industry_harmed) (13292.9#2), which the book's 28 CQ1 asks about (how strong the causal "
        "generalization is; the game asks no question there): cq_answer with polarity for (§5 item 3). A mechanism "
        "answers 28 CQ1; data offered for the link would answer 28 CQ2. 13292.12#3 codes the same content as a con of "
        "the tax in its own right, as its children treat it (13292.131 backs the badness; 13292.45 and 13292.194 propose "
        "remedies)."),
    '13292.141': (
        "Backs 13292.2's claim that meat production releases significant greenhouse gas emissions, with a comparison of "
        "energy use: producing meat takes far more energy than growing the same nutrients from plants. Its argument is "
        "a causal step whose cause is an actual state (cause_to_effect: the energy-intensive production causes "
        "significant emissions); the link from energy use to emissions is supplied. It backs the causal premise of "
        "13292.2's own step, causes(meat_production,significant_ghg_emissions) (13292.2#2), with data offered for that "
        "link, which the book's 28 CQ2 asks about (whether the evidence warrants the generalization; the game asks no "
        "question there): cq_answer with polarity for (§5 item 3). It could also be read as evidence for 13292.2's "
        "classification premise instance(meat_production,significant_ghg_source) (14 CQ1). The comparison with plant "
        "sources also implies that they are more efficient (not coded)."),
    '13292.182': (
        "A substitute (iron supplements) lets people get enough iron without meat, so eating less meat need not bring "
        "more anemia: it counteracts the causal step 13292.180#2, from reduced meat consumption to more anemia, which is "
        "what the book's 28 CQ3 asks (other factors that would interfere with or counteract the production of the "
        "effect; the game asks no question there). So it is a cq_answer at that step's exception "
        "interference(reduced_meat_consumption), against, and its own argument (Other: no walton.yml scheme concludes "
        "it from a substitute) concludes the exception (R8). It does not deny that the tax reduces meat eating. Its "
        "sibling 13292.198 makes the same move with plant foods; 13292.184 replies that supplements are not always a "
        "good replacement."),
    '13292.198': (
        "Plant foods rich in iron (beans, lentils) can replace meat as a source of iron, so eating less meat need not "
        "bring more anemia: like its sibling 13292.182, it counteracts the causal step 13292.180#2 (reduced meat "
        "consumption -> more anemia), the book's 28 CQ3 (the game asks no question there): a cq_answer at that step's "
        "exception interference(reduced_meat_consumption), against; its Other instance concludes the exception (R8). It "
        "also implicitly contests 13292.180's premise that meat is the easiest source of iron. 13292.210 backs the "
        "substitution."),
    '13292.25': (
        "Concedes the burden ('the additional burden') and says it is small: it argues that the consequence 13292.7 "
        "cites is too slight to count, the exception too_slight(tax_meat,food_less_affordable_for_poor) of 13292.7#1 "
        "(OC9). It grants that the burden is bad, and does not say it is avoided (its sibling 13292.33 does). No scheme "
        "concludes too_slight, so its own argument is Other, concluding the exception (R8); the game asks CQ:too_slight "
        "there, so the role is cq_answer (§5 item 3), against. Its own words hedge ('could be as small as'); the fit and "
        "the confidence say so."),
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
        "chain tax -> production costs -> food security is coded as two causal steps whose cause is only proposed "
        "(cause_to_effect_modal, §5 item 11)."),
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
        "make food more expensive' (the mood rule, D008). Its children engage the starvation link, a premise of the "
        "weighing (OC10): 13292.202 amends the action, 13292.204 denies the link."),
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
        "group is not exposed to the cause)."),
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
        "§5 item 4), role action_amendment, polarity against 13292.127; it still supports a restricted meat tax. It is "
        "coded at the starvation link brings_about(tax_meat,starvation_from_reduced_meat_access), a premise of "
        "13292.127's weighing, since a reply to a weighing claim addresses the consequence it relies on (OC10); OC6 would "
        "put a mitigation at the conclusion of the argument it answers, here 13292.127#2's ¬should_be_performed(tax_meat), "
        "the argument that cites the starvation (a question for the review, D086 round 1). Stated as a possibility "
        "('there can be', 'could'). Its sibling 13292.204 instead denies the starvation link."),
    '13292.194': (
        "An amended action that mitigates the harm 13292.12 predicts (C006; OC6): the meat tax plus import measures "
        "(eco-tariffs; imports that verifiably meet environmental and animal-welfare standards, or pass a high enough "
        "tax barrier), tax_meat_with_import_safeguards. 'Can solve this issue' is can of capability (the mood rule, "
        "D008): positive consequences of the amended action, role action_amendment, polarity against 13292.12, at "
        "13292.12's predicted harm, the conclusion may_occur(domestic_meat_industry_pushed_out) of its primary causal "
        "step 13292.12#1. In format 2 the target could name 13292.12#3 instead, 13292.12's negative-consequences "
        "argument that cites the harm against the tax (a question for the review, D086 round 1; its sibling 13292.45, in "
        "round 2, has the same target). Read as blocking 13292.12's mechanism, it would be a cq_answer to 28 CQ3 "
        "(interference)."),
}
for cid, text in NOTES.items():
    if cid not in ROUND1:
        refuse(f'{cid} is not a claim of round 1')
    G[cid]['notes']['text'] = text

G['13292.147']['notes']['candidate_pattern'] = 'a proposal of another course, not said to do better (C017; guidelines §3.1, OC6)'

# --- Kept: the reference's coding, and why ----------------------------------------------------------

KEPT = {
    '13292.1': 'a control (A, B and the reference agree on every core field): the thesis, scheme none.',
    '13292.30': 'a control: the practical analogy from taxes on ecologically damaging businesses elsewhere, at the thesis (guidelines §3 names it).',
    '13292.36': 'a control; notes rewritten without the history of the causal steps\' coding (in this record and gold-format1.json).',
    '13292.108': 'every variable bound (see the note): the thesis has no argument, so this fragment carries the environmental argument; §5 item 6 would leave eight of the eleven targets into it nothing to point at. Whether a fragment whose parent has no argument carries that argument is put to the review and the owner. Notes rewritten as one current account.',
    '13292.25': 'the exception too_slight (OC9), the game\'s CQ:too_slight; notes rewritten as one current account.',
    '13292.33': 'the likelihood question at 13292.7#1\'s consequence premise (13292.7 has no causal step of its own); notes rewritten.',
    '13292.127': 'the weighing against the class at 13292.108\'s exception (OC3, OC4, OC7, OC8); notes rewritten as one current account (the earlier codings, against 13292.115 and with Other, are in this record\'s entries).',
    '13292.174': 'a possible side effect that ranks nothing: the modal variant and its bridge, a rebuttal at 13292.108\'s conclusion (OC2, OC15; D070); notes rewritten.',
    '13292.104': 'a better means at 13292.108\'s exception (OC6; D075, D089): the developer\'s reading matched; notes rewritten.',
    '13292.43': 'the weighing at 13292.9#1\'s exception (OC3, OC4, OC15): the developer\'s reading matched; notes rewritten.',
    '13292.192': 'the attack on 13292.6\'s stated link with a modal causal chain; notes rewritten without the history of the causal steps\' coding.',
    '13292.147': 'a proposal of another course at 13292.108\'s conclusion (OC6; D006, D066); notes rewritten (the earlier text described the answer at the exception it no longer is), and candidate C017 named, as OC6 asks.',
    '13292.204': 'a premise attack on the starvation link the weighing relies on (OC10); notes rewritten (the earlier text called the weighing Other, as it was before D064 and D075).',
    '13292.233': 'the modal premise of 13292.121#1, CQ:modal_likelihood (OC15); notes rewritten (the earlier text named CQ:likelihood, the question before the modal variant).',
    '13292.29': 'the rule from personal liberty, a rebuttal at 13292.115\'s conclusion (OC5, C007); fit explicit, since both premise-role formulas are stated (§5 item 6); notes revised: the remark on a game question removed, since a right that rules the action out answers no question (OC5).',
    '13292.5': 'a side effect that ranks nothing, a rebuttal at 13292.108\'s conclusion (OC2; D070); notes rewritten.',
    '13292.123': 'a better means at 13292.118\'s exception (OC6; D075, D089); notes rewritten.',
    '13292.202': 'scheme none and role action_amendment for the restricted tax, at the starvation link of the weighing (OC10); the alternative, OC6\'s conclusion of the argument that cites the starvation (13292.127#2), is put to the review; notes rewritten.',
    '13292.194': 'positive consequences of the amended action, at 13292.12\'s predicted harm; the alternative target, 13292.12#3, is put to the review with its sibling 13292.45 (round 2); notes rewritten.',
    '13292.8': 'the classification of North American consumption as unhealthy, backing 13292.115\'s premise brings_about(...) (CQ:likelihood): the developer\'s first reading (a fragment) took 13292.115 for practical reasoning; it is positive consequences, which has no circumstances premise.',
    '13292.46': 'the argument from rules (the rule, tax by impact; the case, beef\'s impact; the conclusion, beef taxed higher) fits what the claim states; no consequence of the graded tax is stated.',
}
for cid, why in KEPT.items():
    if cid not in ROUND1:
        refuse(f'{cid} is not a claim of round 1')
    fields = ['notes'] if cid in NOTES or cid == '13292.147' else []
    record(cid, fields, "Drafted: the reference's coding kept: " + why)

missing = [c for c in ROUND1 if c not in done]
if missing:
    refuse(f'round 1 claims not drafted: {missing}')

gold['status'] = ('Development material (D086): no item is verified by a person yet. The first gold is drafted from '
                  'it, round by round (docs/d086-first-gold.md).')
gold['history'].append({
    'date': WHEN, 'by': BY,
    'what': ("Round 1 of the first gold, draft 1 (D086; docs/d086-first-gold.md, d086-rounds.json): 28 claims drafted. "
             "Changed: 13292.12, .141, .182 and .198 name the causal step they answer (28 CQ1, 28 CQ2, 28 CQ3), which "
             "format 1 had kept them from; the fragments 13292.6, .11 and .218 bind only the variables of the formulas "
             "they state (guidelines §5 item 6). Kept, with the reason in each item's record: the other 21, among them "
             "13292.108, which carries the environmental argument (a question for the review and the owner). Notes "
             "rewritten as one current account where they told a history; candidate C017 named for 13292.147.")})

with open(GOLD, 'w', encoding='utf-8') as fh:
    fh.write(json.dumps(gold, indent=1, ensure_ascii=False))
for c in ROUND1:
    rounds['claims'][c].update(draft=1, review='not reviewed')
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
