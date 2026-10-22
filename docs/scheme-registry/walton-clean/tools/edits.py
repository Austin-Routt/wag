"""Every edit made to walton.yml, in application order.

Line numbers are those of the ORIGINAL file (2068 lines).  Categories:
syntax, typo, variable, template, "repair — needs approval", and the
owner's decisions at the end: "repair — owner decision" (OWNER_REPAIRS)
and "addition — owner decision" (ADDITIONS).
"""

SYN = "syntax"
TYPO = "typo"
VAR = "variable"
TPL = "template"
REP = "repair — needs approval"
ADD = "addition — owner decision"
OWN = "repair — owner decision"

# scheme id renames (typos), old -> new
ID_MAP = {
    "popular_practise": "popular_practice",
    "epistemic_ignorence": "epistemic_ignorance",
    "treat": "threat",
    "exeptional": "exceptional",
    "exeption_to_rule": "exception_to_rule",
}


def sub(line, old, new, category, reason, **kw):
    return dict(op="sub", line=line, old=old, new=new, category=category, reason=reason, **kw)


def subs(lines, old, new, category, reason, **kw):
    return dict(op="subs", lines=lines, old=old, new=new, category=category, reason=reason, **kw)


def rename(old, new, lines, category, reason, **kw):
    return dict(op="rename", old=old, new=new, lines=lines, category=category, reason=reason, **kw)


def insert(line, text, category, reason, **kw):
    return dict(op="insert", line=line, text=text, category=category, reason=reason, **kw)


def delete(line, before, category, reason, **kw):
    return dict(op="delete", line=line, before=before, category=category, reason=reason, **kw)


def arity(line, key, n, reason=None, **kw):
    r = reason or (f"language key without /arity is never used by Carneades (it looks up "
                   f"name/arity); every scheme uses {key} with {n} argument(s) and the template has {n} %s")
    return sub(line, f"  {key}:", f"  {key}/{n}:", TPL, r, **kw)


OPS = [
    # ------------------------------------------------------------------
    # 1. The six one-line syntax fixes that make the file load
    #    (REPORT.md section 2).  --syntax-only applies just these.
    # ------------------------------------------------------------------
    sub(691, "-smaller_part_property", "- smaller_part_property", SYN,
        "no space after '-': YAML reads the exceptions value as a string, not a list", syntax_only=True),
    sub(732, "id:opposition", "id: opposition", SYN,
        "no space after 'id:': YAML reads a plain string and the next line fails "
        "(the file's load error 'line 733: mapping values are not allowed')", syntax_only=True),
    dict(op="set", line=869, before="      title: Argument From Values - Variant 1: Positive Value ",
         after='      title: "Argument From Values - Variant 1: Positive Value"', category=SYN,
         reason="unquoted title containing ': ' is a YAML error (trailing space dropped)", syntax_only=True),
    dict(op="set", line=884, before="      title: Argument From Values - Variant 2: Negative Value",
         after='      title: "Argument From Values - Variant 2: Negative Value"', category=SYN,
         reason="unquoted title containing ': ' is a YAML error", syntax_only=True),
    sub(1587, "-not_necessarily_true", "- not_necessarily_true", SYN,
        "no space after '-': YAML reads a string where a list is expected", syntax_only=True),
    sub(1602, "-rule_changed", "- rule_changed", SYN,
        "no space after '-': YAML reads a string where a list is expected", syntax_only=True),

    # ------------------------------------------------------------------
    # 2. Other syntax: keys the importer ignores, unparsable formula,
    #    duplicate YAML keys, a reserved word used as a predicate.
    # ------------------------------------------------------------------
    sub(418, "assumption:", "assumptions:", SYN,
        "misspelled key: yaml.v2 silently ignores 'assumption:', so generally_accepted(A) was dropped"),
    sub(838, "needed_nonarbitrary((A,F)", "needed_nonarbitrary(A,F)", SYN,
        "unbalanced '((': the term does not parse and InstantiateScheme silently dropped this premise"),
    delete(55, '  biased/1: "%s is biased."', SYN,
           "duplicate key biased/1 (also line 56); YAML keeps only the later one, so this line was "
           "never used; deleting it leaves the loaded language unchanged"),
    delete(150, '  good/1: "%s is good."', SYN,
           "duplicate key good/1, identical to line 149"),
    sub(372, "  true/1:", "  is_true/1:", SYN,
        "'true' is the boolean constant to the Carneades term parser, so 'true/1' is not a predicate "
        "symbol (validator); predicate renamed is_true, same template"),
    sub(431, "- true(A)", "- is_true(A)", SYN,
        "predicate 'true' renamed is_true (see line 372); the statement is unchanged"),

    # ------------------------------------------------------------------
    # 3. Variables
    # ------------------------------------------------------------------
    # primed variables: the term parser reads ' as the start of a quoted atom,
    # so every formula containing P' etc. was dropped and no argument was built
    subs([712, 714, 716], "P'", "P2", VAR,
         "primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme"),
    subs([725, 727, 729], "P'", "P2", VAR,
         "primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme"),
    subs([738, 740, 742], "P'", "P2", VAR,
         "primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme"),
    subs([751, 753, 755, 756], "X'", "X2", VAR,
         "primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme"),
    subs([764, 766], "X'", "X2", VAR,
         "primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme"),
    subs([764, 766, 768], "A'", "A2", VAR,
         "primed variable A' does not parse (' starts a quoted atom); renamed A2 within the scheme"),
    subs([777, 779, 781], "R'", "R2", VAR,
         "primed variable R' does not parse (' starts a quoted atom); renamed R2 within the scheme"),
    sub(684, "[Y,X,P]", "[Y,X,A]", VAR,
        "A is used in 3 formulas but undeclared (non-ground statements); P is declared but used nowhere "
        "(validator: 'declared but not used'); the category file A_reasoning/2 declares [Y,X,A]"),
    sub(817, "[A,P]", "[A,F]", VAR,
        "F is used in all 4 property formulas but undeclared (non-ground statements); P is declared "
        "but used nowhere (validator: 'declared but not used'); P is the misnamed F"),
    sub(1565, "[A1,A2,A,P]", "[A1,A2,A,P,F]", VAR,
        "F is used in individuals_propoerty(A1,F,A2) but undeclared (non-ground statement); declared"),

    # ------------------------------------------------------------------
    # 4. Repairs of broken encodings (need the owner's approval)
    # ------------------------------------------------------------------
    insert(495, "      - morally_acceptable_action(X,A)", REP,
           "empty conclusions: Carneades built no argument; the 'Therefore, ...' formula listed as the "
           "first premise is the conclusion (as in the category file B/4)"),
    delete(497, "      -  morally_acceptable_action(X,A)", REP,
           "moved to conclusions (line 495); it was the scheme's conclusion listed as a premise"),
    sub(908, "great_sacrifice(S,A,X)", "great_sacrifice(S,A,V)", REP,
        "undeclared X made the premise non-ground, and the value V occurred only in the conclusion; "
        "the template's third argument is 'the value %s', i.e. V"),
    sub(768, "if_treatment_required(X,A,A2,X)", "if_treatment_required(X,A,A2,X2)", REP,
        "the 4th argument repeats X where the template says 'the contrary treatment %s', i.e. X2; "
        "otherwise X2 occurs only in the conclusion (validator error)"),
    sub(1557, "- reapplication_sequence", "- reapplication_sequence(A2,A)", REP,
        "0-ary chaining premise: the conclusion's individual A occurred in no premise (validator error); "
        "the chain links the clear case A2 to A"),
    sub(306, "  reapplication_sequence:", "  reapplication_sequence/2:", REP,
        "template for the repaired reapplication_sequence(A2,A) (line 1557)"),
    sub(306, "from the clear area through the grey area.", "from the clear area -%s- through the grey area to %s.", REP,
        "template for the repaired reapplication_sequence(A2,A) (line 1557), in the file's '-...-' style"),

    # ------------------------------------------------------------------
    # 5. Language keys without /arity (29)
    # ------------------------------------------------------------------
    arity(10, "accept_saying", 0),
    arity(42, "attack_well_supported", 0),
    arity(43, "background", 0),
    arity(51, "being_credible", 0),
    arity(87, "character_revelant", 0),
    arity(102, "correctly_report", 0),
    arity(118, "do_obligator", 0),
    arity(132, "excuse", 0, "language key without /arity; regulative_rule uses 'excuse' with 0 arguments "
          "(excuse/1 is a different predicate, used by plea_for_excuse)"),
    sub(132, '"%s has an adequate', '"The agent has an adequate', TPL,
        "excuse is used with 0 arguments, so the template may not contain %s"),
    arity(147, "goal_prestige", 0),
    arity(208, "means_end", 1),
    arity(219, "more_important_ends", 0),
    arity(221, "more_reliable", 0),
    arity(242, "not_necessarily_true", 0),
    arity(243, "not_normal", 0),
    arity(251, "oficcial_rulebook", 3),
    arity(255, "other_consequences", 1),
    arity(260, "ought_to_continue", 2),
    arity(297, "proper_sequence", 0),
    arity(302, "prudential", 1),
    arity(313, "reference_class", 0),
    arity(323, "rule_changed", 0),
    arity(337, "should_not", 1),
    arity(346, "speakers_correct", 0),
    arity(354, "strong_advocate", 3),
    arity(357, "strongly_occur", 1, "language key without /arity; negative_consequences_prudential uses "
          "strongly_occur(X) with 1 argument"),
    sub(357, " consequences %s may", " consequences may", TPL,
        "strongly_occur is used with 1 argument (X); the template had a second %s"),
    arity(364, "sufficient_evidence", 0),
    arity(366, "synonymous", 0),
    arity(390, "waste_bad_thing", 2),

    # ------------------------------------------------------------------
    # 6. Templates whose %s count does not fit the predicate (14)
    # ------------------------------------------------------------------
    sub(17, "  action_fit_quality/1:", "  action_fit_quality/0:", TPL,
        "template has no %s for arity 1; no scheme uses the predicate, so the arity is set to fit the "
        "template instead of inventing an argument"),
    sub(50, "%v", "%s", TPL, "'%v' is not a %s verb (WAG lint error)"),
    sub(60, "bring about % more", "bring about %s more", TPL,
        "malformed verb '% m' printed '%!m(string=...)ore'"),
    sub(88, ", some cited bad consequences, %s, will follow.", ", some cited bad consequences will follow.", TPL,
        "cited_bad_consequences is used with 1 argument (X); the second %s printed '%!s(MISSING)'"),
    sub(97, "  conflict_goals/2:", "  conflict_goals/3:", TPL,
        "the template has 3 %s and both schemes use conflict_goals(P,A2,A) with 3 arguments"),
    sub(139, "in favor of %s.", "in favor of it.", TPL,
        "favor is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)'"),
    sub(200, '"% is %s."', '"%s is %s."', TPL, "malformed verb '% i' printed '%!i(string=...)s'"),
    sub(245, '"%is not required."', '"%s is not required."', TPL,
        "malformed verb '%i' printed '%!i(string=...)s'"),
    sub(256, "and A %s true", "and it is true", TPL,
        "the predicate has 2 arguments (A,C); 'A %s' was the variable name plus a stray %s ('A is true')"),
    sub(277, "in regard to %s.", "in regard to it.", TPL,
        "popular_practise is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)'"),
    sub(281, "  position_to_know_group/1:", "  position_to_know_group/2:", TPL,
        "the template has 2 %s and the scheme uses position_to_know_group(G,A)"),
    sub(328, '"%s has a great value."', '"%s has a great value %s."', TPL,
        "sacrifice_value is used with 2 arguments (A,V); the value V was printed as '%!(EXTRA ...)'"),
    sub(329, "  same_meaning/2:", "  same_meaning/3:", TPL,
        "the template has 3 %s and the scheme uses same_meaning(L,W,D)"),
    sub(350, '"A is a special type', '"%s is a special type', TPL,
        "special_proposition is used with 2 arguments (A,K); the first was written as the letter A"),

    # ------------------------------------------------------------------
    # 7. Missing templates for used predicates
    # ------------------------------------------------------------------
    insert(68, '  cannot_present_sametime/3: "%s cannot present %s and %s at the same time and in the same respect."', TPL,
           "opposition-normative uses cannot_present_sametime(A,X,X2) with 3 arguments; only /2 had a template"),
    insert(110, '  data_set/2: "%s is a set of data or supposed facts in case %s."', TPL,
           "abductive_backward uses data_set(D,A) with 2 arguments; only data_set/1 had a template"),

    # ------------------------------------------------------------------
    # 8. Typos: template words
    # ------------------------------------------------------------------
    sub(27, "also commined to", "also committed to", TYPO, "template word 'commined'"),
    sub(28, "has commiued herself", "has committed herself", TYPO, "template word 'commiued' (OCR)"),
    sub(40, "I wil l see", "I will see", TYPO, "template word 'wil l'"),
    sub(42, "Allegration", "Allegation", TYPO, "template word 'Allegration'"),
    sub(80, "In arguement", "In argument", TYPO, "template word 'arguement'"),
    sub(85, "argument depent on", "argument depends on", TYPO, "template word 'depent'"),
    sub(87, "the arguement was", "the argument was", TYPO, "template word 'arguement'"),
    sub(87, "of character revelant.", "of character relevant.", TYPO, "template word 'revelant'"),
    sub(102, "Oflicial rulelebooks", "Official rulebooks", TYPO, "template words 'Oflicial rulelebooks' (OCR)"),
    sub(118, "prohibite -obligator-", "prohibited -obligatory-", TYPO,
        "template words 'prohibite -obligator-' (cf. obligatory/1: 'prohibited -obligatory-')"),
    sub(121, "to achive.", "to achieve.", TYPO, "template word 'achive'"),
    sub(139, "as truei that", "as true that", TYPO, "template word 'truei'"),
    sub(151, "represen ts", "represents", TYPO, "template word 'represen ts' (OCR)"),
    sub(195, "by vinue of", "by virtue of", TYPO, "template word 'vinue' (OCR)"),
    sub(213, "Appearence", "Appearance", TYPO, "template word 'Appearence'"),
    sub(218, "More effecient", "More efficient", TYPO, "template word 'effecient'"),
    sub(218, "also brin about", "also bring about", TYPO, "template word 'brin'"),
    sub(231, "If thte situation", "If the situation", TYPO, "template word 'thte'"),
    sub(231, "would he true", "would be true", TYPO, "template word 'he' for 'be' (OCR)"),
    sub(256, "presum ption", "presumption", TYPO, "template word 'presum ption' (OCR)"),
    sub(312, "%s recieves the", "%s receives the", TYPO, "template word 'recieves'"),
    sub(325, "type o f action", "type of action", TYPO, "template word 'o f' (OCR)"),
    sub(344, "caryying", "carrying", TYPO, "template word 'caryying'"),
    sub(354, "a stron way", "a strong way", TYPO, "template word 'stron'"),
    sub(361, "availablle", "available", TYPO, "template word 'availablle'"),
    sub(378, "generalzation", "generalization", TYPO, "template word 'generalzation'"),

    # ------------------------------------------------------------------
    # 9. Typos: scheme ids and titles
    # ------------------------------------------------------------------
    sub(1143, "epistemic_ignorence", "epistemic_ignorance", TYPO, "scheme id 'ignorence'"),
    sub(1310, "id: treat", "id: threat", TYPO, "scheme id 'treat' for 'threat' (title: Argument From Threat)"),
    sub(1645, "id: exeptional", "id: exceptional", TYPO, "scheme id 'exeptional'"),
    sub(1976, "exeption_to_rule", "exception_to_rule", TYPO, "scheme id 'exeption'"),
    sub(571, "Popular Practise", "Popular Practice", TYPO, "title: 'Practise' (the noun is 'practice')"),
    sub(734, "Descriptive Scheme 2", "Descriptive Scheme 3", TYPO,
        "title of opposition-descriptive_3 repeated 'Scheme 2' (same as opposition-descriptive_2)"),
    sub(1145, "Ignorence", "Ignorance", TYPO, "title 'Ignorence'"),
    sub(1312, "From Treat", "From Threat", TYPO, "title 'Treat' for 'Threat'"),
    sub(1326, "Ad Baculum Treat", "Ad Baculum Threat", TYPO, "title 'Treat' for 'Threat'"),
    sub(1576, "Physical Worl Premise", "Physical World Premise", TYPO, "title 'Worl'"),
    sub(1591, "Physical Worl Premise", "Physical World Premise", TYPO, "title 'Worl'"),

    # ------------------------------------------------------------------
    # 10. Typos: OCR errors in sources
    # ------------------------------------------------------------------
    subs([410, 645, 657, 874, 889, 1550], "lnformal Logic", "Informal Logic", TYPO, "source: 'lnformal' (OCR l for I)"),
    sub(669, "Informal Iogic", "Informal Logic", TYPO, "source: 'Iogic' (OCR I for l)"),
    subs([788, 802], "Struktur rmd Funktion", "Struktur und Funktion", TYPO, "source: 'rmd' (OCR for 'und')"),
    subs([815, 831], "Argumentation Schmes", "Argumentation Schemes", TYPO, "source: 'Schmes'"),
    sub(901, "A Trtalise on", "A Treatise on", TYPO, "source: 'Trtalise' (OCR)"),
    subs([1037, 1050, 1063], "A Trtatise on", "A Treatise on", TYPO, "source: 'Trtatise' (OCR)"),
    sub(1024, "Artificial lntelligence", "Artificial Intelligence", TYPO, "source: 'lntelligence' (OCR l for I)"),
    subs([1111, 1127], "Arguments from lgnorance", "Arguments from Ignorance", TYPO, "source: 'lgnorance' (OCR l for I)"),
    sub(1148, "Quaerterly", "Quarterly", TYPO, "source: 'Quaerterly'"),
    sub(1608, "Wayrw Grennan", "Wayne Grennan", TYPO, "source: 'Wayrw' (OCR; cf. line 1635 'Wayne Grennan')"),

    # ------------------------------------------------------------------
    # 11. Typos: misspelled predicate names, renamed everywhere they occur
    #     (language key and formulas).  Runs last, after the key and word
    #     edits above.
    # ------------------------------------------------------------------
    delete(278, '  posible/1: "%s is possible."', TYPO,
           "'posible' is a misspelling of 'possible'; renamed in its only use (line 1105), which then "
           "uses the existing template possible/1 (line 284); a second possible/1 key would be a duplicate"),
    rename("posible", "possible", [1105], TYPO, "misspelled predicate 'posible'"),
    rename("all_parts_propert", "all_parts_property", [26, 674, 702], TYPO, "misspelled predicate"),
    rename("arguement_not_accepted", "argument_not_accepted", [38, 1461], TYPO, "misspelled predicate"),
    rename("beyond_contetion", "beyond_contention", [54, 1555], TYPO, "misspelled predicate"),
    rename("bring_relive", "bring_relief", [61, 1388], TYPO, "misspelled predicate ('relive' for 'relief')"),
    rename("cannot_recieve_evaluation", "cannot_receive_evaluation", [69, 753], TYPO, "misspelled predicate"),
    rename("certainity_no_need", "certainty_no_need", [80, 825], TYPO, "misspelled predicate"),
    rename("character_revelant", "character_relevant", [87, 1451], TYPO, "misspelled predicate"),
    rename("commined_to", "committed_to", [1408], TYPO,
           "misspelled predicate in commitment1's conclusion; committed_to/2 (line 96) is its template, "
           "and the category file B/2 has committed_to(A,Y)"),
    rename("commited_opposite", "committed_opposite", [91, 1497], TYPO, "misspelled predicate"),
    rename("easy_to_achive", "easy_to_achieve", [121, 1070], TYPO, "misspelled predicate"),
    rename("evidence_not_commited", "evidence_not_committed", [130, 1482], TYPO, "misspelled predicate"),
    rename("individuals_propoerty", "individuals_property", [194, 1570], TYPO, "misspelled predicate"),
    rename("not_has_propery", "not_has_property", [239, 1567], TYPO, "misspelled predicate"),
    rename("oficcial_rulebook", "official_rulebook", [251, 1599], TYPO, "misspelled predicate"),
    rename("part_propert", "part_property", [262, 686], TYPO, "misspelled predicate"),
    rename("popular_practise", "popular_practice", [277, 569, 579], TYPO,
           "misspelled predicate and scheme id ('practise' for the noun 'practice')"),
    rename("property_precion_needed", "property_precision_needed", [299, 822], TYPO, "misspelled predicate"),
    rename("reasonable_belive", "reasonable_believe", [309, 1699], TYPO, "misspelled predicate"),
    rename("recieves_evaluation", "receives_evaluation", [312, 756], TYPO, "misspelled predicate"),
    rename("situation_necessiy", "situation_necessity", [344, 1083], TYPO, "misspelled predicate"),
    rename("succesful_account", "successful_account", [359], TYPO, "misspelled predicate"),
    rename("succesful_diagram", "successful_diagram", [360, 1203], TYPO, "misspelled predicate"),
    rename("vaguness_recejted", "vagueness_rejected", [385, 819], TYPO, "misspelled predicate"),
]

# A soft hyphen (U+00AD) stands for the dash in the first Bench-Capon citation of the two value schemes, so most
# renderers print "pp.429448". Found while drafting the catalogue, 2026-10-22.
OPS += [sub(n, "pp.429\u00ad448", "pp. 429-448", TYPO, "source: a soft hyphen (U+00AD) in place of the page range's dash")
        for n in (872, 887)]

# ----------------------------------------------------------------------
# Repairs the owner decided, 2026-10-22. Applied after every edit above,
# so the renames do not touch them.
_NEG = ("the conclusion is used by no other scheme, so a slippery slope never meets the consequence and practical "
        "reasoning schemes' should_be_performed(A); the book's conclusion (49) is that A should not be brought about")
_YEAR = ("the book's compendium cites these pages (49, pp. 93, 95) as 1992, as do walton.yml's three other citations "
         "of the same book, and booksellers' listings date the Vale Press edition 1992; the book's ch. 4 (p. 114) cites "
         "p. 93 for gradualism as 'Walton, 1999', which the book's reference list (not checked) may give as another work")
_NV = ("the conclusion was copied from positive_value and says the opposite of the book's 19.2 (the negative value is "
       "a reason to retract commitment to the goal)")
OWNER_REPAIRS = [
    dict(op="set", line=255,
         before='  other_consequences/1: "There are some good or bad consequences of doing %s should be taken into account."',
         after='  other_consequences/1: "Doing %s also has consequences of the opposite kind that outweigh the ones cited."',
         category=OWN,
         reason="the exception now carries the book's 33/CQ3 in the three consequence schemes (below), and an exception that "
                "holds defeats the argument whatever the weights; so it must say that the opposite consequences outweigh "
                "the cited ones, not only that some exist. The old template was also ungrammatical. In our own words"),
    sub(2000, "negative_consequences(A)", "¬should_be_performed(A)", OWN, _NEG),
    sub(1997, "1999", "1992", OWN, _YEAR),
    sub(2009, "1999", "1992", OWN, _YEAR),
    sub(892, "reason_retaining_goal(V,G)", "reason_retracting_goal(V,G)", OWN, _NV),
    insert(308, '  reason_retracting_goal/2: "%s is a reason to give up commitment to goal %s."', OWN,
           "template for negative_value's new conclusion, in our own words"),
    dict(op="set", line=229,
         before='  negative_affects/3: "The fact that value %s is negative affects the interpretation and therefore the '
                'evaluation of goal %s of agent %s, if value is bad, it supports commitment to goal."',
         after='  negative_affects/3: "That value %s is negative bears on how goal %s of agent %s is judged: a bad value '
               'counts against commitment to the goal."',
         category=OWN, reason="the template said that a bad value supports commitment to the goal, the opposite of the "
                              "book's 19.2; rewritten in our own words"),
]

# ----------------------------------------------------------------------
# Additions the owner decided. Applied after the repairs above.
#
# 2026-10-22: compendium 33.3, Reasoning from Negative Consequences. The
# category file A_reasoning/3_practical-reasoning.yml formalizes it in a
# second block with the id negative_consequences (lines 511-524 there);
# the master has only the first block of that id, so 33.3 had no scheme.
# It comes in under a new id, with its empty variable list filled in, and
# with templates in our own words: the category file's own templates
# repeat the book's premises (and one has a typo). Its conclusion is
# ¬should_be_performed(A), like negative_consequences', instead of the
# category file's not_bring_about(A), threat's predicate, so that the
# consequence arguments can rebut one another; and it has the exception
# other_consequences(A), like the other consequence schemes (below).
_R33 = ("compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file "
        "A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the "
        "master; added under a new id by the owner's decision of 2026-10-22")
_T33 = "template for the 33.3 scheme added below, in our own words (the category file's repeats the book)"
ADDITIONS = [
    insert(21, '  agent_bring_about/2: "Bringing about %s would lead to %s."', ADD, _T33),
    insert(46, '  bad_for_goals/1: "%s would harm the agent\'s goals."', ADD, _T33),
] + [insert(1897, text, ADD, _R33, scheme="negative_consequences_reasoning") for text in [
    "  - id: negative_consequences_reasoning",
    "    meta:",
    "      title: Argument From Consequences - Reasoning From Negative Consequences",
    "      source: >",
    "             Douglas Walton, A Pragmatic Theory of Fallacy.",
    "             Tuscaloosa and London: University of Alabama Press, 1995. pp. 155-156.",
    "             Douglas Walton, Scare Tactics. Dordrecht: Kluwer, 2000, p. 123.",
    "    variables: [A,B]",
    "    conclusions:",
    "      - ¬should_be_performed(A)",
    "    premises:",
    "      - agent_bring_about(A,B)",
    "      - bad_for_goals(B)",
    "    exceptions:",
    "      - other_consequences(A)",
    "      ",
]]
# 2026-10-22: the book's question about other, opposite consequences (33/CQ3, printed for 33.1-33.4) had a formula
# only in negative_consequences_prudential (33.4); the other three consequence schemes get the same exception.
_OC = ("the book's 33/CQ3 (other, opposite consequences) applies to 33.1-33.4; walton.yml had its formula, the "
       "exception other_consequences, only in negative_consequences_prudential")
ADDITIONS += [insert(n, text, ADD, _OC) for n in (1896, 1925) for text in ("    exceptions:", "      - other_consequences(A)")]
# 2026-10-22: compendium 10.2 and 53 had no scheme in walton.yml; the upstream category files formalize them
# (whole_into_parts, full_slippery_slope). Both come in under their upstream ids, with templates in our own words
# (the category files' repeat the book's premises). full_slippery_slope also gets the compendium's reference as its
# source (the category file's is empty), and arguments for its two 0-ary atoms, which are about the series of cases at
# hand: a 0-ary atom is one statement for every instance (CHANGELOG, "0-ary atoms"; walton-audit REPORT §6).
_WP = ("compendium 10.2, Division of the Whole into its Parts, formalized only in the category file "
       "A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22")
_FS = ("compendium 53, the Full Slippery Slope Argument, formalized only in the category file "
       "A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's "
       "reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand")
_WPT = "template for whole_into_parts (10.2), in our own words (the category file's repeats the book)"
_FST = "template for full_slippery_slope (53), in our own words (the category file's repeats the book)"
ADDITIONS += [
    insert(13, '  accept_all_steps/2: "Opinion runs so that whoever comes to accept one case in the series from %s to %s '
               'will be carried on to accept the next."', ADD, _FST),
    insert(195, '  initial_presumption/1: "At the outset, %s is presumed acceptable."', ADD, _FST),
    insert(204, '  leads_to_next/3: "In the chain of cases from %s through %s to %s, every case brings on the following '
                'one, through causal, precedent or analogical links."', ADD, _FST),
    insert(204, '  led_to_step/2: "Accepting a case in the series from %s to %s will in fact carry people on to the next '
                'one."', ADD, _FST),
    insert(231, '  not_acceptable_step/1: "%s cannot be accepted after all, whatever the initial presumption."', ADD, _FST),
    insert(261, '  part_is/2: "The part %s has %s."', ADD, _WPT),
    insert(262, '  part_whole/5: "%s can have %s only if at least one of %s, %s and %s has it."', ADD, _WPT),
    insert(390, '  whole_is/2: "The whole %s has %s."', ADD, _WPT),
    insert(390, '  whole_of/4: "%s is made up of the parts %s, %s and %s."', ADD, _WPT),
] + [insert(705, text, ADD, _WP, scheme="whole_into_parts") for text in [
    "  - id: whole_into_parts",
    "    meta:",
    "      title: Argument From Division - Division of the Whole into its Parts",
    "      source: >",
    "              Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation.",
    "              Notre Dame, Ind.: University of Notre Dame Press, 1969, pp. 231-233,234-241.",
    "    variables: [X,X1,X2,X3,A]",
    "    conclusions:",
    "      - whole_is(X,A)",
    "    premises:",
    "      - whole_of(X,X1,X2,X3)",
    "      - part_whole(X,A,X1,X2,X3)",
    "      - part_is(X1,A)",
    "     ",
]] + [insert(1573, text, ADD, _FS, scheme="full_slippery_slope") for text in [
    "  - id: full_slippery_slope",
    "    meta:",
    "      title: Full Slippery Slope Argument",
    "      source: >",
    "             Douglas Walton, Slippery Slope Arguments. Newport News, Va.: Vale Press, 1992, pp. 200, 203.",
    "    variables: [C0,C1,C2]",
    "    conclusions:",
    "      - not_acceptable_step(C0)",
    "    premises:",
    "      - initial_presumption(C0)",
    "      - leads_to_next(C0,C1,C2)",
    "      - accept_all_steps(C0,C2)",
    "    assumptions:",
    "      - led_to_step(C0,C2)",
    "",
]]
OPS += OWNER_REPAIRS + ADDITIONS
