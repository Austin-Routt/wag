"""PROPOSED repairs, NOT applied to walton.clean.yml.

Two schemes cannot pass the Carneades validator without changing what a
premise asserts: a variable occurs in the conclusion or an exception but in
no premise ("variable not used in premises or deletions";
"variable declared but not used").  These are the smallest changes I found
that make them pass.  They are applied only to check/walton.proposals.yml, to
show that these two then pass the validator.  Each needs the owner's
decision.  Line numbers refer to the ORIGINAL file.

Until 2026-11-24 this file also proposed repairs of practical_necessary_condition
and practical_sufficient_condition (their agent P and other goal A2 were in no
premise).  They were withdrawn then: proposal D004-D075 (decisions D030 and
D090) repairs both schemes otherwise, in edits.py's block PRACTICAL.
"""
from edits import sub, insert, REP

PROPOSALS = [
    # ad_hominem: the argument S occurs only in the conclusion and an exception
    insert(1446, "      - advocates(A,S)", REP,
           "PROPOSAL ad_hominem: S (the argument) occurs in no premise; new premise 'A advocates argument S', "
           "which the conclusion 'A's argument S' presupposes"),
    insert(20, '  advocates/2: "%s advocates argument %s."', REP,
           "PROPOSAL ad_hominem: template for the new premise advocates(A,S)"),
    # regulative_rule: the agent A occurs only in the conclusion
    sub(1641, "- do_obligator", "- do_obligator(A)", REP,
        "PROPOSAL regulative_rule: the agent A occurs only in the conclusion; the norm premise is applied to A"),
    sub(118, "  do_obligator/0:", "  do_obligator/1:", REP, "PROPOSAL: template for do_obligator(A)"),
    sub(118, '"One must not do', '"%s must not do', REP, "PROPOSAL: template for do_obligator(A)"),
]
