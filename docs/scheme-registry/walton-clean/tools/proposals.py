"""PROPOSED repairs, NOT applied to walton.clean.yml.

Four schemes cannot pass the Carneades validator without changing what a
premise asserts: a variable occurs in the conclusion, an assumption or an
exception but in no premise ("variable not used in premises or deletions";
"variable declared but not used").  These are the smallest changes I found
that make them pass.  They are applied only to check/walton.proposals.yml, to
show that the gate then passes for all 111 schemes.  Each needs the owner's
decision.  Line numbers refer to the ORIGINAL file.
"""
from edits import sub, insert, REP

PROPOSALS = [
    # ad_hominem: the argument S occurs only in the conclusion and an exception
    insert(1446, "      - advocates(A,S)", REP,
           "PROPOSAL ad_hominem: S (the argument) occurs in no premise; new premise 'A advocates argument S', "
           "which the conclusion 'A's argument S' presupposes"),
    insert(20, '  advocates/2: "%s advocates argument %s."', REP,
           "PROPOSAL ad_hominem: template for the new premise advocates(A,S)"),
    # practical_necessary_condition / practical_sufficient_condition: the
    # agent P and the other goal A2 occur only in an assumption and exceptions
    sub(973, "[X,A,A2,P]", "[X,A,P]", REP, "PROPOSAL practical_necessary_condition: A2 dropped (see line 985)"),
    sub(977, "goal_is_action(A)", "goal_is_action(P,A)", REP,
        "PROPOSAL practical_necessary_condition: the agent P occurs in no premise; the goal premise names the agent"),
    sub(985, "conflict_goals(P,A2,A)", "conflict_goals(P,A)", REP,
        "PROPOSAL practical_necessary_condition: the other goal A2 occurs in no premise; the exception says 'some other goals'"),
    sub(1001, "[X,A,A2,P]", "[X,A,P]", REP, "PROPOSAL practical_sufficient_condition: A2 dropped (see line 1013)"),
    sub(1005, "goal_is_action(A)", "goal_is_action(P,A)", REP,
        "PROPOSAL practical_sufficient_condition: the agent P occurs in no premise; the goal premise names the agent"),
    sub(1013, "conflict_goals(P,A2,A)", "conflict_goals(P,A)", REP,
        "PROPOSAL practical_sufficient_condition: the other goal A2 occurs in no premise"),
    sub(146, '  goal_is_action/1: "My goal is to bring about %s."', '  goal_is_action/2: "The goal of %s is to bring about %s."', REP,
        "PROPOSAL: template for goal_is_action(P,A) (used only by the two practical condition schemes)"),
    sub(97, "  conflict_goals/3:", "  conflict_goals/2:", REP, "PROPOSAL: conflict_goals(P,A)"),
    sub(97, "some other goals %s that", "some other goals that", REP, "PROPOSAL: conflict_goals(P,A)"),
    # regulative_rule: the agent A occurs only in the conclusion
    sub(1641, "- do_obligator", "- do_obligator(A)", REP,
        "PROPOSAL regulative_rule: the agent A occurs only in the conclusion; the norm premise is applied to A"),
    sub(118, "  do_obligator/0:", "  do_obligator/1:", REP, "PROPOSAL: template for do_obligator(A)"),
    sub(118, '"One must not do', '"%s must not do', REP, "PROPOSAL: template for do_obligator(A)"),
]
