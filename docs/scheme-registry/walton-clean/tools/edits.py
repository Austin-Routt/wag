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
_D081 = ("D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the "
         "book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed ")

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
    # 4. D081 repairs (proposal D081-D084), decided under the owner's standing rule.
    # ------------------------------------------------------------------
    insert(495, "      - morally_acceptable_action(X,A)", OWN,
           _D081 + "p. 312, entry 4.6: empty conclusions: Carneades built no argument; the 'Therefore, ...' formula listed as the "
           "first premise is the conclusion (as in the category file B/4)"),
    delete(497, "      -  morally_acceptable_action(X,A)", OWN,
           _D081 + "p. 312, entry 4.6: moved to conclusions (line 495); it was the scheme's conclusion listed as a premise"),
    sub(908, "great_sacrifice(S,A,X)", "great_sacrifice(S,A,V)", OWN,
        _D081 + "p. 322, entry 20: undeclared X made the premise non-ground, and the value V occurred only in the conclusion; "
        "the template's third argument is 'the value %s', i.e. V"),
    sub(768, "if_treatment_required(X,A,A2,X)", "if_treatment_required(X,A,A2,X2)", OWN,
        _D081 + "p. 318, entry 12.1: the 4th argument repeats X where the template says 'the contrary treatment %s', i.e. X2; "
        "otherwise X2 occurs only in the conclusion (validator error). Preserve WAG's treatment-first variable roles."),
    sub(1557, "- reapplication_sequence", "- reapplication_sequence(A2,A,P)", OWN,
        _D081 + "pp. 340-341, entry 51: the chain links clear case A2 to A for the same property P. "
        "The former binary repair bound A but could reuse chain evidence concerning a different property."),
    sub(306, "  reapplication_sequence:", "  reapplication_sequence/3:", OWN,
        _D081 + "pp. 340-341, entry 51: template for the property-scoped reapplication_sequence(A2,A,P) (line 1557)"),
    sub(306, '"A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."',
        '"A chain of modus ponens steps leads from %s, in the clear area, through the grey area to %s, each step concerning %s."', OWN,
        _D081 + "pp. 340-341, entry 51: the template names both ends of the chain and its property, in our own words"),

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
# 2026-10-22: compendium 50, the Precedent Slippery Slope Argument, had no formalization anywhere (upstream included).
# Written for WAG by the owner's decision, from the compendium (p. 340) and the book's ch. 3 version (pp. 116-117):
# - one formula per printed premise, the Related Cases Premise split in two, so that CQ1 (would C0 set a precedent?)
#   and CQ2 (why would each intervening precedent occur?) each have one: the first pair (C0 binds C1) and the chain
#   from C1 to the end case Cn;
# - the intolerable case is taken to be the end of the series, Cn, as a slippery slope needs: the book prints C0 there
#   and in CQ3 (the compendium; the ch. 3 scan is unclear but looks the same), which can also be read as "treating C0
#   as an exception would be intolerable because of where it leads", a reason the chain premises then carry;
# - the conclusion is the negation of case_exception, the premise of the argument for an exceptional case
#   (exceptional), so that the slope meets the argument it counters (p. 117).
_PS = ("compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the "
       "owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first "
       "pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the "
       "conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)")
_PST = "template for precedent_slippery_slope (50), in our own words"
ADDITIONS += [
    insert(88, '  claimed_exception/2: "%s is put forward as an exception to the rule %s."', ADD, _PST),
    insert(199, '  intolerable_exception/2: "Making an exception of %s to the rule %s would be intolerable."', ADD, _PST),
    insert(290, '  precedent_chain/3: "From %s to %s runs a chain of cases, each so like the next that an exception for one '
                'commits us to an exception for the next under the rule %s."', ADD, _PST),
    insert(332, '  sets_precedent/3: "Making an exception of %s would commit us to making one of %s as well, since like '
                'cases must be treated alike under the rule %s."', ADD, _PST),
] + [insert(1573, text, ADD, _PS, scheme="precedent_slippery_slope") for text in [
    "  - id: precedent_slippery_slope",
    "    meta:",
    "      title: Precedent Slippery Slope Argument",
    "      source: >",
    "             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,",
    "             1995, p. 159. Douglas Walton, Slippery Slope Arguments. Newport News, Va.: Vale Press, 1992, pp. 93, 95.",
    "    variables: [C0,C1,Cn,R]",
    "    conclusions:",
    "      - ¬case_exception(C0)",
    "    premises:",
    "      - claimed_exception(C0,R)",
    "      - sets_precedent(C0,C1,R)",
    "      - precedent_chain(C1,Cn,R)",
    "      - intolerable_exception(Cn,R)",
    "",
]]
# 2026-10-22: the modal cause to effect (audit N8). Compendium 28 has a strong and a weak reading (the cause occurs or
# might; the effect will or might follow), and cause_to_effect formalizes only the strong one: it asserts that the cause has
# occurred, so a causal step of a policy not yet adopted had no scheme. Written for WAG by the owner's decision: the
# same causal generalization and exception as cause_to_effect, with may_occur for both the cause and the effect, so
# that hypothetical steps chain (a tax may come in, prices may rise, consumption may fall). The causal premise is not
# weakened: the book's weak reading has a weak generalization too, which the catalogue records as open.
_MOD = ("audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which "
        "cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with "
        "cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that "
        "hypothetical steps chain")
ADDITIONS += [
    insert(207, '  may_occur/1: "An event %s has happened, or could happen."', ADD,
           "template for cause_to_effect_modal (audit N8), in our own words"),
] + [insert(1749, text, ADD, _MOD, scheme="cause_to_effect_modal") for text in [
    "  - id: cause_to_effect_modal",
    "    meta:",
    "      title: Argument from Cause to Effect - Modal Form",
    "      source: >",
    "             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,",
    "             1995, pp. 140-141.",
    "    variables: [E1,E2]",
    "    conclusions:",
    "      - may_occur(E2)",
    "    premises:",
    "      - may_occur(E1)",
    "      - causes(E1,E2)",
    "    exceptions:",
    "      - interference(E1)",
    "",
]]
# ----------------------------------------------------------------------
# 2026-10-22: the analogy family in two-case form (decisions D001, D028 and
# D029; WAG's docs/proposals/D001-D028-D029.md, version 3, approved by the
# owner after two independent reviews). The book's analogy (7) and its
# practical counterparts (8.1, 8.2) use one letter for what holds in both
# cases, made relative to each case. Here each case has its own claim or
# action, joined by a correspondence premise; the difference exception
# tells against what the argument concludes and names it, sided in the
# practical pair; the book's CQ3 is a counter-analogy that rebuts, so
# more_on_point goes. The practical analogies conclude that the target
# action should (not) be performed, so that they rebut each other and meet
# the practical schemes.
_AN = ("D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false "
       "whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between "
       "the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, "
       "a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner "
       "decision of 2026-10-22 (proposal D001-D028-D029, version 3)")
_PA = ("D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in "
       "each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be "
       "performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met "
       "nothing). The book prints no questions for them; they get a difference exception, sided: against the action in "
       "the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)")
_ANT = "template for the analogy family's two-case form (D001, D028, D029), in our own words"
_RW = ("the template nearly repeated the book's premise and had a stray comma; rewritten in our own words for the "
       "two-case form (D028, D029)")
ANALOGY = [
    insert(103, '  corresponds/2: "In the other case, the counterpart of %s is %s."', OWN, _ANT),
    insert(115, '  difference_against/3: "%s and %s differ in a way that tells against %s."', OWN, _ANT),
    insert(115, '  difference_for/3: "%s and %s differ in a way that tells in favour of %s."', OWN, _ANT),
    dict(op="set", line=321, before='  right_to_do/2: "The right thing to do in %s, was to carry out action %s."',
         after='  right_to_do/2: "In %s, %s was the right course."', category=OWN, reason=_RW),
    dict(op="set", line=402, before='  wrong_to_do/2: "The wrong thing to do in %s, was to carry out action %s."',
         after='  wrong_to_do/2: "In %s, %s was the wrong course."', category=OWN, reason=_RW),
    sub(646, "[A1,A2,X]", "[A1,A2,X1,X2]", OWN, _PA),
    sub(648, "carry_out_right(A2,X)", "should_be_performed(X2)", OWN, _PA),
    sub(650, "right_to_do(A1,X)", "right_to_do(A1,X1)", OWN, _PA),
    insert(651, "      - corresponds(X1,X2)", OWN, _PA),
    insert(651, "    exceptions:", OWN, _PA),
    insert(651, "      - difference_against(A1,A2,X2)", OWN, _PA),
    sub(658, "[A1,A2,X]", "[A1,A2,X1,X2]", OWN, _PA),
    sub(660, "carry_out_wrong(A2,X)", "¬should_be_performed(X2)", OWN, _PA),
    sub(662, "wrong_to_do(A1,X)", "wrong_to_do(A1,X1)", OWN, _PA),
    insert(663, "      - corresponds(X1,X2)", OWN, _PA),
    insert(663, "    exceptions:", OWN, _PA),
    insert(663, "      - difference_for(A1,A2,X2)", OWN, _PA),
    sub(1719, "[A,C]", "[P1,P2,C1,C2]", OWN, _AN),
    sub(1720, "[A]", "[P2]", OWN, _AN),
    sub(1722, "similar_case(C)", "similar_to(C1,C2)", OWN, _AN),
    sub(1723, "in_case(A,C)", "in_case(P1,C1)", OWN, _AN),
    insert(1723, "      - corresponds(P1,P2)", OWN, _AN),
    sub(1725, "relevant_differences(C)", "difference_against(C1,C2,P2)", OWN, _AN),
    delete(1726, "      - more_on_point(A,C)", OWN, _AN),
]
# ----------------------------------------------------------------------
# 2026-10-23: sided exceptions for the consequence schemes (decisions D055
# and D059; WAG's docs/proposals/D055-D059.md, version 5, approved by the
# owner after four independent reviews and a final check). The book's
# 33/CQ3 had one formula, other_consequences(A), shared by the arguments
# for and against an action, so that one argument that it held defeated
# every consequence argument about the action on both sides. Each scheme
# now has an exception on its own side that names the consequence it
# cites: an argument for A is defeated by bad consequences that outweigh
# its G, one against A by good ones that outweigh its G (or B). The
# prudential form's likelihood assumption names the consequence too, and
# it concludes that X should be done, so that it meets the practical
# schemes about X. Templates are our own; other_consequences/1,
# prudential/1 and strongly_occur/1 are left as orphan templates.
_SE = ("D055: the exception other_consequences(A), shared by the arguments for and against A, let one argument "
       "that opposite consequences outweigh defeat every consequence argument about A on both sides; now an exception "
       "on the scheme's own side that names the consequence it cites, so that it defeats only the argument it is "
       "about. Owner decision of 2026-10-23 (proposal D055-D059, version 5)")
_PR = ("D055, D059: the prudential form's likelihood assumption named only the action, so two warnings about one "
       "action shared it; its conclusion prudential(X) met no other scheme. Now strongly_occur(X,C), the sided "
       "exception bad_outweighs(X,C), and the conclusion should_be_performed(X), which meets the practical schemes "
       "about X. Owner decision of 2026-10-23 (proposal D055-D059, version 5)")
_SET = "template for the consequence schemes' sided exceptions (D055, D059), in our own words"
CONSEQUENCES = [
    insert(49, '  bad_outweighs/2: "The action %s has bad consequences that outweigh %s."', OWN, _SET),
    insert(152, '  good_outweighs/2: "The action %s has good consequences that outweigh %s."', OWN, _SET),
    insert(357, '  strongly_occur/2: "If %s is not done, %s is likely to follow."', OWN, _SET),
    sub(1289, "prudential(X)", "should_be_performed(X)", OWN, _PR),
    sub(1294, "other_consequences(X)", "bad_outweighs(X,C)", OWN, _PR),
    sub(1296, "strongly_occur(X)", "strongly_occur(X,C)", OWN, _PR),
    sub(1896, "other_consequences(A)", "good_outweighs(A,G)", OWN, _SE, scheme="negative_consequences"),
    sub(1897, "other_consequences(A)", "good_outweighs(A,B)", OWN, _SE, scheme="negative_consequences_reasoning"),
    sub(1925, "other_consequences(A)", "bad_outweighs(A,G)", OWN, _SE, scheme="positive_consequences"),
]
# D084 is an executable-vocabulary migration only. The English word "is"
# must not be globally renamed; only these two formulas and their key change.
_IS = ("D084: is/2 is reserved for GoCHR assignment; use is_a/2 for the ordinary "
       "two-place predicate, preserving its argument order and English wording. "
       "No coextensivity premise or other source reinterpretation is added.")
D084 = [
    sub(200, "  is/2:", "  is_a/2:", OWN, _IS),
    sub(933, "is(M,Q)", "is_a(M,Q)", OWN, _IS, scheme="group_2"),
    sub(1584, "is(W,D)", "is_a(W,D)", OWN, _IS, scheme="rule_physical_world_1"),
]
# ----------------------------------------------------------------------
# 2026-11-24: practical reasoning (decisions D004, D005, D030, D063, D064,
# D070 and D075, taken under the owner's standing rule of 2026-11-22, and
# the owner's D088-D091 of 2026-11-24; WAG's docs/proposals/D004-D075.md,
# version 3, after two independent reviews). Practical reasoning 1 and 2
# state the agent's goal, goal_of_agent(P,G), and the conflicting goal
# scoped to the agent, incompatible_goal(P,G); the goal's possibility and
# the agent's are apart (possible_goal(G), possible_bring_about(P,A)); the
# three "another action would do better" exceptions name the agent; the
# side-effect exception becomes the claim that A's bad consequences
# outweigh G, D055's bad_outweighs(A,G), a plain side effect being a
# rebuttal (WAG's reading of the source's side-effect exception). The
# necessary- and sufficient-condition schemes take the schemes library's
# repair (the agent P in each premise, the set of means S, G for the goal),
# with the best-choice reading (a strictly better means defeats the
# recommendation; the sufficient scheme has no mere-alternative exception;
# the necessary scheme's objection is that the goal can be reached without
# any means of the set), bad_outweighs(X,G) and incompatible_goal(P,G), and
# conclude should_be_performed(X); practical inference takes the same
# formulas, keeps its own incompatible_consequences, scoped, and concludes
# should_be_performed(A). Three answer schemes conclude exceptions: C003's
# pair, a consequence of the other sign that matters more (one of them is
# also the side effect that outweighs), and another means that does better
# for the same agent. Templates are our own, and read right with plural
# agents. Templates left unused are listed as orphans by the gate.
_PRV = "proposal D004-D075, version 3"
_STD = "Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (" + _PRV + ")"
_PG = ("D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the "
       "conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the "
       "goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the "
       "exceptions that another action would do better name the agent; D070, D091: the side-effect exception "
       "side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad "
       "consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. " + _STD)
_NC = ("D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was "
       "about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart "
       "from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice "
       "reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the "
       "objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); "
       "D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they "
       "conclude should_be_performed(X), which meets the practical schemes. " + _STD)
_SC = ("D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the "
       "necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no "
       "mere-alternative exception, which its own premise granted. " + _STD)
_PI = ("D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now "
       "should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and "
       "keeps its own objection that the action works against its goal, scoped to the agent. " + _STD)
_AW = ("D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) "
       "from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that "
       "outweighs. " + _STD)
_BM = ("D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), "
       "the agent bound by the answered argument's goal premise and able to take the other means. " + _STD)
_PT = ("template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, "
       "D075, D088-D091), in our own words")
_AT = "template for the answer schemes (D064, D075, D089), in our own words"
PRACTICAL = [
    insert(11, '  acceptable_action/2: "For %s, %s is an acceptable thing to do."', OWN, _PT),
    insert(29, '  alternative_action/3: "%s can reasonably hold that at least one of the means %s is needed to bring '
               'about %s."', OWN, _PT),
    insert(29, '  alternative_action_sufficient/3: "%s can reasonably hold that each of the means %s would be enough '
               'to bring about %s."', OWN, _PT),
    insert(60, '  bring_about_more_effectively/3: "There is a way for %s to reach the state %s more effectively than '
               '%s."', OWN, _PT),
    insert(144, '  goal_achievable_without_any/3: "%s could reach %s without using any of the means %s."', OWN, _PT),
    insert(146, '  goal_of_agent/2: "One of the goals of %s is %s."', OWN, _PT),
    insert(185, '  incompatible_consequences/3: "For %s, %s would have consequences that work against %s."', OWN, _PT),
    insert(186, '  incompatible_goal/2: "Another goal of %s is incompatible with %s."', OWN, _PT),
    insert(206, '  matters_more/2: "%s matters more than %s."', ADD, _AT),
    insert(216, '  more_acceptable_bringing/3: "%s would rather bring about %s than leave %s undone."', OWN, _PT),
    insert(217, '  more_effective_means/4: "For %s, %s would realize %s more effectively than %s: more surely, sooner '
                'or at lower cost."', ADD, _AT),
    insert(227, '  necessary_action/4: "%s chose %s, among the means %s, as an acceptable or the best means needed '
                'for %s."', OWN, _PT),
    insert(290, '  practicality_action/2: "As far as %s can tell, nothing that cannot be changed stands in the way of '
                'bringing about %s."', OWN, _PT),
    insert(296, '  promote_more_effectively/3: "There is a way for %s to promote the value %s more effectively than '
                '%s."', OWN, _PT),
    insert(304, '  realize_more_effectively/3: "There is a way for %s to realize the goal %s more effectively than '
                '%s."', OWN, _PT),
    insert(362, '  sufficient_action/4: "%s chose %s, among the means %s, as an acceptable or the best means that '
                'would be enough for %s."', OWN, _PT),
    # practical_inference (22.1)
    sub(949, "[A,G]", "[A,G,P]", OWN, _PI),
    sub(951, "carry_out_action(A)", "should_be_performed(A)", OWN, _PI),
    sub(953, "have_goal(G)", "goal_of_agent(P,G)", OWN, _PI),
    sub(956, "other_goals_conflict(G)", "incompatible_goal(P,G)", OWN, _PI),
    sub(957, "more_efficient_alternative_actions(A,G)", "realize_more_effectively(P,G,A)", OWN, _PI),
    sub(958, "incompatible_consequences(A,G)", "bad_outweighs(A,G)", OWN, _PI),
    insert(958, "      - incompatible_consequences(P,A,G)", OWN, _PI),
    sub(960, "possible_action(A)", "possible_bring_about(P,A)", OWN, _PI),
    # practical_necessary_condition (22.2)
    sub(973, "[X,A,A2,P]", "[X,G,P,S]", OWN, _NC),
    sub(975, "bringing_about_action(X)", "should_be_performed(X)", OWN, _NC),
    sub(977, "goal_is_action(A)", "goal_of_agent(P,G)", OWN, _NC),
    sub(978, "alternative_action(X,A)", "alternative_action(P,S,G)", OWN, _NC),
    sub(979, "necessary_action(X,A)", "necessary_action(P,X,S,G)", OWN, _NC),
    sub(980, "practicality_action(X)", "practicality_action(P,X)", OWN, _NC),
    sub(981, "more_acceptable_bringing(A,X)", "more_acceptable_bringing(P,G,X)", OWN, _NC),
    sub(983, "alternative_means(A,X)", "realize_more_effectively(P,G,X)", OWN, _NC),
    insert(983, "      - goal_achievable_without_any(P,G,S)", OWN, _NC),
    sub(984, "side_effect_action(P,X)", "bad_outweighs(X,G)", OWN, _NC),
    sub(985, "conflict_goals(P,A2,A)", "incompatible_goal(P,G)", OWN, _NC),
    sub(987, "possible_action(A)", "possible_goal(G)", OWN, _NC),
    sub(988, "acceptable_action(X)", "acceptable_action(P,X)", OWN, _NC),
    sub(989, "possible_agent(X,P)", "possible_bring_about(P,X)", OWN, _NC),
    # practical_sufficient_condition (22.3)
    sub(1001, "[X,A,A2,P]", "[X,G,P,S]", OWN, _SC),
    sub(1003, "bringing_about_action(X)", "should_be_performed(X)", OWN, _SC),
    sub(1005, "goal_is_action(A)", "goal_of_agent(P,G)", OWN, _SC),
    sub(1006, "alternative_action(X,A)", "alternative_action_sufficient(P,S,G)", OWN, _SC),
    sub(1007, "sufficient_action(X,A)", "sufficient_action(P,X,S,G)", OWN, _SC),
    sub(1008, "practicality_action(X)", "practicality_action(P,X)", OWN, _SC),
    sub(1009, "more_acceptable_bringing(A,X)", "more_acceptable_bringing(P,G,X)", OWN, _SC),
    sub(1011, "alternative_means(A,X)", "realize_more_effectively(P,G,X)", OWN, _SC),
    sub(1012, "side_effect_action(P,X)", "bad_outweighs(X,G)", OWN, _SC),
    sub(1013, "conflict_goals(P,A2,A)", "incompatible_goal(P,G)", OWN, _SC),
    sub(1015, "possible_action(A)", "possible_goal(G)", OWN, _SC),
    sub(1016, "acceptable_action(X)", "acceptable_action(P,X)", OWN, _SC),
    sub(1017, "possible_agent(X,P)", "possible_bring_about(P,X)", OWN, _SC),
    # practical_reasoning1 (22.4)
    sub(1935, "[A,S1,S2,G,V]", "[A,S1,S2,G,V,P]", OWN, _PG),
    insert(1938, "      - goal_of_agent(P,G)", OWN, _PG),
    sub(1946, "possible(A)", "possible_bring_about(P,A)", OWN, _PG),
    sub(1948, "bring_about_more_effectively(S2,A)", "bring_about_more_effectively(P,S2,A)", OWN, _PG),
    sub(1949, "realize_more_effectively(G,A)", "realize_more_effectively(P,G,A)", OWN, _PG),
    sub(1950, "promote_more_effectively(V,A)", "promote_more_effectively(P,V,A)", OWN, _PG),
    sub(1951, "side_effects(A,S1)", "bad_outweighs(A,G)", OWN, _PG),
    insert(1951, "      - incompatible_goal(P,G)", OWN, _PG),
    # practical_reasoning2 (22.1)
    sub(1959, "[A,S1,S2,G]", "[A,S1,S2,G,P]", OWN, _PG),
    insert(1962, "      - goal_of_agent(P,G)", OWN, _PG),
    sub(1967, "possible(A)", "possible_bring_about(P,A)", OWN, _PG),
    sub(1968, "possible(G)", "possible_goal(G)", OWN, _PG),
    sub(1970, "bring_about_more_effectively(S2,A)", "bring_about_more_effectively(P,S2,A)", OWN, _PG),
    sub(1971, "realize_more_effectively(G,A)", "realize_more_effectively(P,G,A)", OWN, _PG),
    sub(1973, "side_effects(A,S1)", "bad_outweighs(A,G)", OWN, _PG),
    sub(1974, "incompatible_goal(G)", "incompatible_goal(P,G)", OWN, _PG),
] + [insert(1926, text, ADD, _AW, scheme=sid) for sid, text in [
    ("bad_outweighs_good", "  - id: bad_outweighs_good"),
    ("bad_outweighs_good", "    meta:"),
    ("bad_outweighs_good", "      title: Weighing a Bad Consequence against a Good One"),
    ("bad_outweighs_good", "      source: >"),
    ("bad_outweighs_good", "        WAG's answer scheme (decisions D064 and D075; candidate C003, docs/candidates.md),"),
    ("bad_outweighs_good", "        after Walton's comparative weighing of good and bad consequences: Douglas Walton,"),
    ("bad_outweighs_good", "        The New Dialectic: Conversational Contexts of Argument. University of Toronto Press,"),
    ("bad_outweighs_good", "        1998, pp. 159, 170. Not in the compendium."),
    ("bad_outweighs_good", "    variables: [A,H,G]"),
    ("bad_outweighs_good", "    conclusions:"),
    ("bad_outweighs_good", "      - bad_outweighs(A,G)"),
    ("bad_outweighs_good", "    premises:"),
    ("bad_outweighs_good", "      - brings_about(A,H)"),
    ("bad_outweighs_good", "      - bad(H)"),
    ("bad_outweighs_good", "      - matters_more(H,G)"),
    ("bad_outweighs_good", ""),
    ("good_outweighs_bad", "  - id: good_outweighs_bad"),
    ("good_outweighs_bad", "    meta:"),
    ("good_outweighs_bad", "      title: Weighing a Good Consequence against a Bad One"),
    ("good_outweighs_bad", "      source: >"),
    ("good_outweighs_bad", "        WAG's answer scheme (decisions D064 and D075; candidate C003, docs/candidates.md),"),
    ("good_outweighs_bad", "        after Walton's comparative weighing of good and bad consequences: Douglas Walton,"),
    ("good_outweighs_bad", "        The New Dialectic: Conversational Contexts of Argument. University of Toronto Press,"),
    ("good_outweighs_bad", "        1998, pp. 159, 170. Not in the compendium."),
    ("good_outweighs_bad", "    variables: [A,H,G]"),
    ("good_outweighs_bad", "    conclusions:"),
    ("good_outweighs_bad", "      - good_outweighs(A,G)"),
    ("good_outweighs_bad", "    premises:"),
    ("good_outweighs_bad", "      - brings_about(A,H)"),
    ("good_outweighs_bad", "      - good(H)"),
    ("good_outweighs_bad", "      - matters_more(H,G)"),
    ("good_outweighs_bad", ""),
]] + [insert(1975, text, ADD, _BM, scheme="better_means") for text in [
    "  - id: better_means",
    "    meta:",
    "      title: Argument from a Better Means",
    "      source: >",
    "        WAG's answer scheme (decisions D075 and D089), the sketch better_means of WAG's review of",
    "        2026-10-15 (part 6, §4), repaired; after the book's questions on other means and efficiency",
    "        (Walton, Reed and Macagno, Argumentation Schemes, 2008: 22.1 and 22.4, and ch. 3, p. 96).",
    "        Not in the compendium.",
    "    variables: [A,A2,G,P]",
    "    conclusions:",
    "      - realize_more_effectively(P,G,A)",
    "    premises:",
    "      - goal_of_agent(P,G)",
    "      - possible_bring_about(P,A2)",
    "      - brings_about(A2,G)",
    "      - more_effective_means(P,A2,G,A)",
    "",
]]

# ----------------------------------------------------------------------
# 2026-11-25: the consequences (decisions D008, D065, D066, D069, D072 and
# D074, taken under the owner's standing rule of 2026-11-22 and refined by
# the owner on 2026-11-25; WAG's docs/proposals/D008-D074.md, version 5 as
# amended, after independent reviews of versions 1 and 2, a check of version
# 3, a review of version 4's application, the owner's answer on D053, D074
# and D008, and a check of version 5 and its application). The modal variants of 33.1
# and 33.2 conclude may_be_worth_performing(A,G) and may_be_unwise_to_perform(A,G)
# from may_bring_about(A,G), and meet should_be_performed(A) only through a
# bridge, one per side, whose premise matters_to_decision(A,G) its arguer
# supports and whose exception worth_finding_out_first(A,G) is the critic's
# (D008). 33.1 and 33.2 and their modal variants gain too_slight(A,G),
# neutral as to polarity (D053, the owner's declaration), and the sided
# bearer pair, gain_discounted_by_bearer(A,G) for A and
# harm_discounted_by_bearer(A,G) against it (D065); 33.4 gains
# fails_to_avert(X,C) (D069). D074's objection is an undercutter of the
# specific argument, not a formula, so 33.3 does not change. harm_reduction is
# WAG's argument from what happens anyway (D072). Templates are our own.
_C3V = "proposal D008-D074, version 5"
_C3 = ("Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (" + _C3V + ")")
_MOD = ("D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps "
        "the modality in its premise and its conclusion, which names the consequence, and meets the thesis only "
        "through its bridge. " + _C3)
_BR = ("D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its "
       "arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth "
       "finding out first), both naming A and G; an entry condition, not a weight. " + _C3)
_HR = ("D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each "
       "occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm "
       "overall would not fall. " + _C3)
_EX = ("D065: the consequence schemes gain too_slight(A,G) (how serious), neutral as to polarity (D053, declared "
       "by the owner), and the bearer exception of their side (on whom), a sided flip. " + _C3)
_P33 = "D069: the prudential form (33.4) gains fails_to_avert(X,C), the critic's to support. " + _C3
_TA = "template for the consequences' new formulas (D008, D065, D069, D072), in our own words"
CONSEQUENCES_C3 = [
    insert(138, '  fails_to_avert/2: "Doing %s would fail to avert %s."', OWN, _TA),
    insert(142, '  gain_discounted_by_bearer/2: "In deciding on action %s, the gain %s counts for less because of whom it would go to."', OWN, _TA),
    insert(164, '  happens_anyway/2: "Whether or not %s is done, %s will happen."', ADD, _TA),
    insert(164, '  harm_discounted_by_bearer/2: "In deciding on action %s, the harm %s counts for less because of whom it would fall on."', OWN, _TA),
    insert(205, '  less_harmful_per_occurrence/2: "Doing %s would make each occurrence of %s less harmful."', ADD, _TA),
    insert(206, '  matters_to_decision/2: "The chance that action %s brings about %s is large enough, given what that consequence would be, to bear on whether the action should be performed."', ADD, _TA),
    insert(206, '  may_be_unwise_to_perform/2: "Action %s may be unwise to perform because of %s."', ADD, _TA),
    insert(206, '  may_be_worth_performing/2: "Action %s may be worth performing for the sake of %s."', ADD, _TA),
    insert(206, '  may_bring_about/2: "The action %s may bring about %s."', ADD, _TA),
    insert(230, '  no_net_reduction/2: "Doing %s would not lower the harm that %s does overall."', ADD, _TA),
    insert(295, '  preventable_otherwise/2: "A course open to whoever would do %s could prevent %s."', ADD, _TA),
    insert(370, '  too_slight/2: "In deciding on action %s, %s is too slight a consequence to count."', OWN, _TA),
    insert(392, '  worth_finding_out_first/2: "It is worth finding out first whether action %s would bring about %s, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing."', ADD, _TA),
    # negative_consequences_prudential (33.4)
    insert(1294, "      - fails_to_avert(X,C)", OWN, _P33, scheme="negative_consequences_prudential"),
    # negative_consequences (33.2)
    insert(1896, "      - too_slight(A,G)", OWN, _EX, scheme="negative_consequences"),
    insert(1896, "      - harm_discounted_by_bearer(A,G)", OWN, _EX, scheme="negative_consequences"),
    # positive_consequences (33.1)
    insert(1925, "      - too_slight(A,G)", OWN, _EX, scheme="positive_consequences"),
    insert(1925, "      - gain_discounted_by_bearer(A,G)", OWN, _EX, scheme="positive_consequences"),
]
CONSEQUENCES_C3 += [insert(1926, text, ADD, _MOD, scheme='positive_consequences_modal') for text in [
    '  - id: positive_consequences_modal',
    '    meta:',
    '      title: Argument from Possible Positive Consequences',
    '      source: >',
    "        WAG's modal variant of the argument from positive consequences (decision D008), after Douglas",
    '        Walton, A Pragmatic Theory of Fallacy, The University of Alabama Press, Tuscaloosa and London, 1995,',
    '        pp. 155-156, and the questions on consequences in Walton, Reed and Macagno, Argumentation Schemes,',
    '        2008, pp. 332-333, which allow that a consequence will, may or must occur.',
    '    variables: [A,G]',
    '    conclusions:',
    '      - may_be_worth_performing(A,G)',
    '    premises:',
    '      - may_bring_about(A,G)',
    '      - good(G)',
    '    exceptions:',
    '      - bad_outweighs(A,G)',
    '      - too_slight(A,G)',
    '      - gain_discounted_by_bearer(A,G)',
    '',
]]
CONSEQUENCES_C3 += [insert(1926, text, ADD, _MOD, scheme='negative_consequences_modal') for text in [
    '  - id: negative_consequences_modal',
    '    meta:',
    '      title: Argument from Possible Negative Consequences',
    '      source: >',
    "        WAG's modal variant of the argument from negative consequences (decision D008), after Douglas",
    '        Walton, A Pragmatic Theory of Fallacy, The University of Alabama Press, Tuscaloosa and London, 1995,',
    '        pp. 155-156, and the questions on consequences in Walton, Reed and Macagno, Argumentation Schemes,',
    '        2008, pp. 332-333, which allow that a consequence will, may or must occur.',
    '    variables: [A,G]',
    '    conclusions:',
    '      - may_be_unwise_to_perform(A,G)',
    '    premises:',
    '      - may_bring_about(A,G)',
    '      - bad(G)',
    '    exceptions:',
    '      - good_outweighs(A,G)',
    '      - too_slight(A,G)',
    '      - harm_discounted_by_bearer(A,G)',
    '',
]]
CONSEQUENCES_C3 += [insert(1926, text, ADD, _BR, scheme='bridge_from_possible_good') for text in [
    '  - id: bridge_from_possible_good',
    '    meta:',
    '      title: From a Possible Good Consequence to the Action (bridge)',
    '      source: >',
    "        WAG's bridge (decision D008, refined 2026-11-25), on the pattern of decisions D011 and D063: an",
    '        explicit warrant, with a premise its arguer supports and an exception of its own, from the modal',
    '        conclusion of positive_consequences_modal to the recommendation. Not in the compendium.',
    '    variables: [A,G]',
    '    conclusions:',
    '      - should_be_performed(A)',
    '    premises:',
    '      - may_be_worth_performing(A,G)',
    '      - matters_to_decision(A,G)',
    '    exceptions:',
    '      - worth_finding_out_first(A,G)',
    '',
]]
CONSEQUENCES_C3 += [insert(1926, text, ADD, _BR, scheme='bridge_from_possible_harm') for text in [
    '  - id: bridge_from_possible_harm',
    '    meta:',
    '      title: From a Possible Bad Consequence to Not Acting (bridge)',
    '      source: >',
    "        WAG's bridge (decision D008, refined 2026-11-25), on the pattern of decisions D011 and D063: an",
    '        explicit warrant, with a premise its arguer supports and an exception of its own, from the modal',
    '        conclusion of negative_consequences_modal to the rejection. Not in the compendium.',
    '    variables: [A,G]',
    '    conclusions:',
    '      - ¬should_be_performed(A)',
    '    premises:',
    '      - may_be_unwise_to_perform(A,G)',
    '      - matters_to_decision(A,G)',
    '    exceptions:',
    '      - worth_finding_out_first(A,G)',
    '',
]]
CONSEQUENCES_C3 += [insert(1926, text, ADD, _HR, scheme='harm_reduction') for text in [
    '  - id: harm_reduction',
    '    meta:',
    '      title: Argument from What Happens Anyway (harm reduction)',
    '      source: >',
    "        WAG's scheme (decision D072, refined 2026-11-25; candidate C022), not in the compendium: an outcome",
    '        that happens whether or not the action is done, made less harmful each time by the action.',
    '    variables: [A,E]',
    '    conclusions:',
    '      - should_be_performed(A)',
    '    premises:',
    '      - happens_anyway(A,E)',
    '      - less_harmful_per_occurrence(A,E)',
    '    exceptions:',
    '      - preventable_otherwise(A,E)',
    '      - no_net_reduction(A,E)',
    '',
]]

# ----------------------------------------------------------------------
# 2026-11-26: the rest of the first family (decisions D022, D023, D031,
# D033, D060, D062 and D063, taken under the owner's standing rule of
# 2026-11-22, and D093, found in drafting and in the review of the
# proposal's first version, decided under the owner's instruction of
# 2026-11-25 to go with the recommended choices; WAG's
# docs/proposals/D022-D093.md, version 2, after independent reviews). The
# recommendations to the one addressed conclude should_be_performed(A) or
# its negation (D063); threat names the threatened consequence (D060);
# danger concludes about the action (D062); conclusions that name an agent
# keep it and meet the recommendation through a bridge whose premise its
# arguer supports (D063); the commitment form of sunk costs (D031); waste's
# reassessment (D033); end and means, negative and graded (D022), with
# borrowed questions (D023); and the repairs of D093. Templates are our own.
_C4V = "proposal D022-D093"
_C4 = "Decided under the owner's standing rule of 2026-11-22 (" + _C4V + ")"
_C4N = "Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (" + _C4V + ")"
_D063 = ("D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim "
         "the consequence and practical-reasoning schemes conclude, so that it meets them. " + _C4)
_D060 = "D060: the threat names the threatened consequence C where the book does (35). " + _C4
_D062 = "D062: danger concludes about the action, not the danger (37). " + _C4
_D022 = "D022: a variant of end and means 1, one reading of 22.6.1. " + _C4
_D022G = ("D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the "
          "question of the opposite value, as end and means 1 does (D093(g)). " + _C4 + "; D093: " + _C4N)
_D023 = "D023: a question borrowed from a related scheme, its source recorded in the proposal. " + _C4
_D031 = "D031: the commitment form of sunk costs (25), a scheme of its own. " + _C4
_D033 = "D033: the reassessment from now on (24, CQ2), an exception, the critic's. " + _C4
_D093 = ("D093: found in drafting step C4 and in the review of its first version: practical reasoning's "
         "formulas for the goal and the questions, the goal's possibility, and the templates of the changed "
         "schemes. " + _C4N)
_BR = ("D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its "
       "arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the "
       "continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). " + _C4 + "; "
       "D093: " + _C4N)
_TA = "template for the formulas of proposal D022-D093, in our own words"
PRACTICAL_C4 = [
    sub(40, 'I hereby assert that in fact I will see to it that %s occurs if you bring about %s.', 'I will bring about %[2]s if you go ahead with %[1]s.', OWN, _D060 + " The template, in our own words."),
    sub(250, 'If the respondent bring about %s then %s will occur.', 'If the respondent goes ahead with %s, then %s will occur.', OWN, _D062 + " The template, in our own words."),
    sub(252, 'But the only way for you to prevent %s is bring about %s.', 'The only way for you to prevent %s is %s.', OWN, _D093 + " The template, in our own words."),
    sub(260, 'Therefore, %s ought to continue trying to realize %s.', '%s ought to continue trying to realize %s.', OWN, _D093 + " The template, in our own words."),
    sub(295, 'Therefore, you ought to prevent %s if possible.', 'You ought to prevent %s if you can.', OWN, _D093 + " The template, in our own words."),
    sub(344, 'As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this.', 'In the view of %s, %s would be a necessary or sufficient condition for realizing %s, a view reported to %s.', OWN, _D093 + " The template, in our own words."),
    sub(353, 'If %s stops trying to realize %s now, all previous efforts will be wasted.', 'If %s stopped trying to realize %s now, all the previous efforts would be wasted.', OWN, _D093 + " The template, in our own words."),
    sub(367, '%s intends to realize %s, and tells %s this.', '%s, intending to realize %s, told %s so.', OWN, _D093 + " The template, in our own words."),
    sub(390, "If all %s 's previous attempts to realize %s are wasted, that would be a bad thing.", 'If all the previous efforts of %s to realize %s were wasted, that would be a bad thing.', OWN, _D093 + " The template, in our own words."),
    insert(87, '  choice_at/2: "%s, there is a choice between %s and not doing so."', ADD, _TA),
    insert(88, '  cited_bad_consequences/2: "If you go ahead with %s, the bad consequence cited, %s, will follow."', OWN, _TA),
    insert(98, '  continues_trying/2: "%s would continue the effort to realize %s."', ADD, _TA),
    insert(112, '  deliberating_agent/2: "It falls to %s to decide whether to go ahead with %s."', ADD, _TA),
    insert(205, '  less_bad_as_means/2: "Being a means to the good end %[2]s would make %[1]s less bad."', ADD, _TA),
    insert(207, '  means_bad_outweighs/2: "As a means to the good end %[2]s, %[1]s would have bad consequences that matter more than that end."', OWN, _TA),
    insert(208, '  means_good_outweighs/2: "As a means to the bad end %[2]s, %[1]s would have good consequences that matter more than that end."', ADD, _TA),
    insert(277, '  position_to_bring_about/1: "I am in a position to bring about %s."', OWN, _TA),
    insert(290, '  precommitted/4: "What %[1]s did or undertook %[3]s is a commitment, binding %[4]s, to %[2]s."', ADD, _TA),
    insert(310, '  reassessment_against_continuing/2: "Leaving aside past losses that cannot be recovered, a reassessment of the costs and benefits from now on goes against %s continuing to try to realize %s."', OWN, _TA),
    sub(1343, 'ought_to_bring_about(X)', 'should_be_performed(X)', OWN, _D063, scheme='fear'),
    sub(1357, 'should_not(Y)', '¬should_be_performed(X)', OWN, _D062, scheme='danger'),
    sub(1318, 'not_bring_about(X)', '¬should_be_performed(X)', OWN, _D063, scheme='threat'),
    sub(1320, 'cited_bad_consequences(X)', 'cited_bad_consequences(X,C)', OWN, _D060, scheme='threat'),
    sub(1321, 'position_to_know(X)', 'position_to_bring_about(C)', OWN, _D060, scheme='threat'),
    sub(1331, 'do_other(X)', 'should_be_performed(X)', OWN, _D063, scheme='ad_baculum'),
    sub(1305, 'do_other(A)', 'should_be_performed(A)', OWN, _D063, scheme='pragmatic_alternative'),
    sub(525, 'carry_out_action(X)', 'should_be_performed(X)', OWN, _D063, scheme='appeal_to_vanity_ad_populum'),
    sub(1031, 'person_goal(P,G)', 'goal_of_agent(P,G)', OWN, _D093, scheme='practical_goal'),
    sub(1080, 'should_carry_out(A,Y)', 'should_do(A,Y)', OWN, _D063, scheme='two_person_practical'),
    sub(1085, 'alternative_means(A,X)', 'realize_more_effectively(A,X,Y)', OWN, _D093, scheme='two_person_practical'),
    sub(1086, 'goals_higher_priority(A,X)', 'incompatible_goal(A,X)', OWN, _D093, scheme='two_person_practical'),
    sub(1087, 'side_effect_other_act(B,A)', 'bad_outweighs(Y,X)', OWN, _D093, scheme='two_person_practical'),
    sub(1089, 'possible_action(Y)', 'possible_bring_about(A,Y)', OWN, _D093, scheme='two_person_practical'),
    sub(1105, 'possible(X)', 'possible_goal(X)', OWN, _D093, scheme='waste'),
    insert(1031, '    exceptions:', OWN, _D023, scheme='practical_goal'),
    insert(1031, '      - realize_more_effectively(P,G,X)', OWN, _D023, scheme='practical_goal'),
    insert(1031, '      - incompatible_goal(P,G)', OWN, _D023, scheme='practical_goal'),
    insert(1031, '      - bad_outweighs(X,G)', OWN, _D023, scheme='practical_goal'),
    insert(1105, '    exceptions:', OWN, _D033, scheme='waste'),
    insert(1105, '      - reassessment_against_continuing(A,X)', OWN, _D033, scheme='waste'),
    insert(1044, '    exceptions:', OWN, _D023, scheme='end_and_means_1'),
    insert(1044, '      - means_bad_outweighs(X,Y)', OWN, _D023, scheme='end_and_means_1'),
]
PRACTICAL_C4 += [insert(1045, text, ADD, _D022G, scheme='end_and_means_bad') for text in [
    '  - id: end_and_means_bad',
    '    meta:',
    '      title: Ends and Means - a Bad End Condemns Its Means',
    '      source: >',
    '        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge',
    '        University Press, ch. 9, entry 22.6.1 (its negative reading), p. 325. Chaim Perelman and Lucie',
    '        Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation, Notre Dame, Ind.: University of',
    '        Notre Dame Press, 1969, pp. 273-278. As the schemes library has it (decision D022), with the',
    '        question of the opposite value that end and means 1 borrows (D093).',
    '    variables: [X,Y]',
    '    conclusions:',
    '      - bad(X)',
    '    premises:',
    '      - means_to_end(X,Y)',
    '      - bad(Y)',
    '    exceptions:',
    '      - means_good_outweighs(X,Y)',
    '',
]]
PRACTICAL_C4 += [insert(1045, text, ADD, _D022, scheme='end_and_means_less_bad') for text in [
    '  - id: end_and_means_less_bad',
    '    meta:',
    '      title: Ends and Means - a Good End Makes a Bad Means Less Bad',
    '      source: >',
    '        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge',
    '        University Press, ch. 9, entry 22.6.1 (its graded reading), p. 325. Chaim Perelman and Lucie',
    '        Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation, Notre Dame, Ind.: University of',
    '        Notre Dame Press, 1969, pp. 273-278. As the schemes library has it (decision D022).',
    '    variables: [X,Y]',
    '    conclusions:',
    '      - less_bad_as_means(X,Y)',
    '    premises:',
    '      - means_to_end(X,Y)',
    '      - good(Y)',
    '      - bad(X)',
    '',
]]
PRACTICAL_C4 += [insert(1106, text, ADD, _BR, scheme='bridge_from_agent') for text in [
    '  - id: bridge_from_agent',
    '    meta:',
    '      title: From What an Agent Should Do to the Action (bridge)',
    '      source: >',
    "        WAG's bridge (decision D063, refined 2026-11-22; D093): an explicit warrant, with a premise its",
    '        arguer supports, from a conclusion that names an agent, should_do(P,A), to the recommendation',
    '        should_be_performed(A), when it falls to P to decide whether to go ahead with A. Not in the',
    '        compendium.',
    '    variables: [P,A]',
    '    conclusions:',
    '      - should_be_performed(A)',
    '    premises:',
    '      - should_do(P,A)',
    '      - deliberating_agent(P,A)',
    '',
]]
PRACTICAL_C4 += [insert(1106, text, ADD, _BR, scheme='bridge_from_continuing') for text in [
    '  - id: bridge_from_continuing',
    '    meta:',
    '      title: From Continuing to Try to the Action (bridge)',
    '      source: >',
    "        WAG's bridge (decision D063, refined 2026-11-22: continuing an activity binds to the continuing",
    "        action, not only to its goal; D093): from waste's conclusion, that an agent ought to continue trying",
    '        to realize a goal, to the recommendation of an action that continues the effort, when the decision',
    "        is the agent's, with practical reasoning's questions for the means it chooses. Not in the",
    '        compendium.',
    '    variables: [P,G,A]',
    '    conclusions:',
    '      - should_be_performed(A)',
    '    premises:',
    '      - ought_to_continue(P,G)',
    '      - continues_trying(A,G)',
    '      - deliberating_agent(P,A)',
    '    assumptions:',
    '      - possible_bring_about(P,A)',
    '    exceptions:',
    '      - realize_more_effectively(P,G,A)',
    '      - bad_outweighs(A,G)',
    '',
]]
PRACTICAL_C4 += [insert(2033, text, ADD, _D031, scheme='sunk_costs_precommitment') for text in [
    '  - id: sunk_costs_precommitment',
    '    meta:',
    '      title: Argument from Sunk Costs (the commitment form)',
    '      source: >',
    '        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge',
    "        University Press, pp. 135-136, and ch. 9, entry 25, pp. 326-327. Douglas Walton, 'The Sunk Costs",
    "        Fallacy or Argument from Waste', Argumentation 16 (2002), pp. 488-489. As the schemes library has it",
    '        (decision D031).',
    '    variables: [P,A,T1,T2]',
    '    conclusions:',
    '      - should_do(P,A)',
    '    premises:',
    '      - choice_at(T2,A)',
    '      - precommitted(P,A,T1,T2)',
    '',
]]

# ----------------------------------------------------------------------
# 2026-11-26: the answer scheme for the bridges' exception (decision D095 (3),
# taken by the owner on 2026-11-26; WAG's docs/proposals/D095.md, version 2,
# after an independent review). It concludes worth_finding_out_first(A,G)
# from the conditions of the owner's declaration under D053 (2026-11-25): an
# investigation I that would settle whether A brings about G, done before the
# decision on A must be taken, whose costs and risks are worth bearing.
# Templates are our own.
_D095 = ("D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the "
         "conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the "
         "question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)")
_TD = "template for the answer scheme's premises (D095), in our own words"
FINDING_OUT_D095 = [
    insert(119, '  done_in_time/2: "%s can be done before the decision on action %s must be taken."', ADD, _TD),
    insert(332, '  settles_whether/3: "%s would settle, reliably enough for the decision, whether action %s brings about %s."', ADD, _TD),
    insert(393, '  worth_its_costs/3: "The costs and risks of %s are worth bearing to know whether action %s brings about %s before that possibility counts in the decision."', ADD, _TD),
]
FINDING_OUT_D095 += [insert(1926, text, ADD, _D095, scheme='finding_out_first') for text in [
    '  - id: finding_out_first',
    '    meta:',
    '      title: Argument for Finding Out First',
    '      source: >',
    "        WAG's answer scheme (decision D095 (3), 2026-11-26), on the owner's declaration under D053",
    "        (2026-11-25) of the bridges' exception worth_finding_out_first(A,G): it concludes the exception",
    '        from the conditions the declaration names. Not in the compendium. After Douglas Walton, The New',
    '        Dialectic: Conversational Contexts of Argument, University of Toronto Press, Toronto, 1998, p. 203',
    '        (a shift to information-seeking within a critical discussion), and Douglas Walton, Fundamentals of',
    '        Critical Argumentation, Cambridge University Press, Cambridge, 2006, p. 85 (delaying a decision to',
    '        collect information).',
    '    variables: [A,G,I]',
    '    conclusions:',
    '      - worth_finding_out_first(A,G)',
    '    premises:',
    '      - settles_whether(I,A,G)',
    '      - done_in_time(I,A)',
    '      - worth_its_costs(I,A,G)',
    '',
]]

OPS += OWNER_REPAIRS + ADDITIONS + ANALOGY + CONSEQUENCES + D084 + PRACTICAL + CONSEQUENCES_C3 + PRACTICAL_C4 + FINDING_OUT_D095
