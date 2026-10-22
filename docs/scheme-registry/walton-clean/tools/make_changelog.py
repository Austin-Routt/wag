"""Write CHANGELOG.md from changes.json, the check reports and the notes below.

usage: python3 -I make_changelog.py OUTDIR
"""
import collections
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from proposals import PROPOSALS  # noqa: E402

ID_MAP = [
    ("popular_practise", "popular_practice", "misspelling ('practise' for the noun 'practice'); also the predicate"),
    ("epistemic_ignorence", "epistemic_ignorance", "misspelling"),
    ("treat", "threat", "misspelling; the title is Argument From Threat"),
    ("exeptional", "exceptional", "misspelling"),
    ("exeption_to_rule", "exception_to_rule", "misspelling"),
]
from edits import ID_MAP as _IDS  # noqa: E402

NEW_ID = dict(_IDS)
assert NEW_ID == {o: n for o, n, _ in ID_MAP}, "id map out of sync with edits.py"
CATS = ["syntax", "typo", "variable", "template", "repair — needs approval", "repair — owner decision",
        "addition — owner decision"]
OWNER = ("repair — owner decision", "addition — owner decision")


def term_preds(f):
    """name/arity of every compound or atom in a formula (outer ¬ and builtins skipped)."""
    out = []
    tokens = re.findall(r"[A-Za-z_][A-Za-z0-9_]*|\(|\)|,|¬|'[^']*'|\S", f)
    pos = 0  # recursive descent over the tokens

    def parse():
        nonlocal pos
        if pos < len(tokens) and tokens[pos] == "¬":
            pos += 1
            return parse()
        tok = tokens[pos]
        pos += 1
        if pos < len(tokens) and tokens[pos] == "(":
            pos += 1
            nargs = 0
            while True:
                parse()
                nargs += 1
                if tokens[pos] == ",":
                    pos += 1
                    continue
                pos += 1  # ')'
                break
            out.append(f"{tok}/{nargs}")
        elif re.match(r"[a-z]", tok):
            out.append(f"{tok}/0")
    try:
        parse()
    except (IndexError, KeyError):
        pass
    return out


def main(outdir):
    rows = json.load(open(os.path.join(outdir, "changes.json"), encoding="utf-8"))
    dump = json.load(open(os.path.join(outdir, "tools", "dump-clean.json"), encoding="utf-8"))
    reports = {}
    for k in ["original", "syntax-only", "clean", "proposals"]:
        p = os.path.join(outdir, "check", k + ".json")
        if os.path.exists(p):
            reports[k] = json.load(open(p, encoding="utf-8"))

    # predicate -> schemes using it (clean file); used specs
    users = collections.defaultdict(set)
    used_specs = set()
    for s in dump["schemes"]:
        for fld in ["premises", "assumptions", "exceptions", "conclusions"]:
            for f in s[fld] or []:
                for spec in term_preds(f):
                    users[spec.split("/")[0]].add(s["id"])
                    used_specs.add(spec)
    orphans = sorted(k for k in dump["language"] if k not in used_specs)

    L = []
    w = L.append
    w("# walton.yml cleanup: changelog")
    w("")
    nschemes = len(dump["schemes"])
    added = sorted({r["scheme"] for r in rows if r["category"].startswith("addition") and r["change"] == "insert after"
                    and r["scheme"] not in ("language",) and r["after"].lstrip().startswith("- id:")})
    changed = sorted({r["scheme_after"] for r in rows if r["category"] in OWNER and r["scheme"] != "language"} - set(added))
    cleanp = os.path.join(outdir, "check", "walton.clean.yml")
    nlines = len(open(cleanp, encoding="utf-8").read().split("\n")) if os.path.exists(cleanp) else "?"
    w("Input: `walton.yml` from `walton-argumentation-schemes-for-carneades-4-master` (2068 lines, 111 schemes, "
      "SHA-256 `1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c`). Output: `walton.yml` "
      f"({nlines} lines, after a 5-line notice of the modification that the licence asks for, `tools/header.txt`): "
      f"the same 111 schemes in the same order, {len(changed)} of them changed by the owner's decisions "
      f"({', '.join('`' + c + '`' for c in changed)}), and {len(added)} added by them "
      f"({', '.join('`' + a + '`' for a in added)}); {nschemes} in all. "
      "See [Changes decided by the owner](#changes-decided-by-the-owner). "
      "Every edit below is also in `changes.json` and is produced by "
      "`tools/apply_edits.py` from the list in `tools/edits.py`; `check/run.sh` rebuilds everything. Line numbers are "
      "those of the original file.")
    w("")

    # ---- gate ---------------------------------------------------------------
    w("## Gate")
    w("")
    w("`check/main.go` (Go, Carneades-4 fork at `carneades-4`, the repository's root), one process per file. "
      "Check 1 also decodes with yaml.v2 `UnmarshalStrict`, which rejects unknown keys (`assumption:`) and duplicate keys.")
    w("")
    w("| | original | syntax fixes only | **cleaned** | cleaned + 4 proposals (not applied) |")
    w("|---|---|---|---|---|")

    def cell(r, f):
        return f(r) if r else "–"
    R = [reports.get(k) for k in ["original", "syntax-only", "clean", "proposals"]]
    lines = [
        ("1 YAML / import error", lambda r: (r.get("yaml_error") or "none")[:45]),
        ("1 strict-decode errors", lambda r: str(len(r["strict_errors"]))),
        ("1 stderr during import", lambda r: "yes" if r["import_stderr"] else "none"),
        ("2 validator problems (before / after instantiation)", lambda r: f"{len(r['validator_before_instantiation'])} / {len(r['validator_after_instantiation'])}" if r["schemes"] else "–"),
        ("3 schemes failing InstantiateScheme checks", lambda r: f"{sum(1 for s in r['scheme_results'] or [] if not s['pass'])} of {r['schemes']}" if r["schemes"] else "–"),
        ("4 language problems", lambda r: str(len(r["language_problems"])) if r["schemes"] else "–"),
        ("schemes passing checks 2–4", lambda r: f"{r['schemes'] - len(r.get('scheme_failures') or {})} of {r['schemes']}" if r["schemes"] else "–"),
        ("**GATE**", lambda r: "**PASS**" if r["pass"] else "**FAIL**"),
    ]
    for name, f in lines:
        w("| " + name + " | " + " | ".join(cell(r, f) for r in R) + " |")
    w("")
    w("(The validator count before instantiation for the syntax-only file includes one entry for the term-parser errors "
      "printed to stderr; the audit's 137 is the same list without it.)")
    w("")
    clean = reports.get("clean")
    if clean and not clean["pass"]:
        w(f"**The cleaned file does not pass the whole gate.** Checks 1, 3 and 4 pass for all {nschemes} schemes, and check 2 "
          f"passes for {nschemes - 4}. The other four fail one validator rule only: every variable must occur in a premise. In "
          "`ad_hominem`, `practical_necessary_condition`, `practical_sufficient_condition` and `regulative_rule`, a variable "
          "occurs only in the conclusion, an assumption or an exception. Fixing that means adding or changing a premise, "
          "which the rules for this cleanup forbid. See [Schemes that cannot pass without a change of meaning]"
          "(#schemes-that-cannot-pass-without-a-change-of-meaning). With the four proposals there, the gate passes for all "
          f"{nschemes} (`proposals-check.txt`). These schemes still instantiate correctly (check 3), and WAG's loader does not "
          "run the validator.")
        w("")
    w("WAG's own lint (`wag-cli --lint --data DIR`, with the cleaned file as `schemes/core.yml` and a copy of WAG's NASCQ data) "
      "reports **0 errors and 1 warning** for the cleaned file. The warning is `no template for ¬/1`: WAG's convention is a "
      "`¬p/n` key per negated predicate, and without one a negation reads \"it is not the case that …\". "
      "For the syntax-only file it reports 58 language findings (`check/wag-lint-*.txt`). Its NASCQ findings come from "
      "WAG's data, which targets `core.yml`, and are not about this file.")
    w("")
    w("`check/semantic-diff.txt` compares what Carneades loads from the syntax-only file and from the cleaned file, "
      "after the renames below. The only scheme-level differences are the syntax fixes that restore a dropped formula "
      "(`popular_opinion`, `verbal_classification_arbitrariness`), the three variable declarations, the four "
      "approval-needed repairs, and the owner's decisions (below; the added scheme is listed, not compared). The six "
      "primed-variable schemes are identical apart from the variable names.")
    w("")

    # ---- diff size ----------------------------------------------------------
    w("## Diff size")
    w("")
    by_cat_rows = collections.Counter(r["category"] for r in rows)
    by_cat_lines = {c: len({r["line"] for r in rows if r["category"] == c}) for c in CATS}
    all_lines = {r["line"] for r in rows}
    diffp = os.path.join(outdir, "walton.diff")
    minus = plus = 0
    if os.path.exists(diffp):
        for l in open(diffp, encoding="utf-8"):
            if l.startswith("-") and not l.startswith("---"):
                minus += 1
            elif l.startswith("+") and not l.startswith("+++"):
                plus += 1
    w("| category | recorded edits | original lines touched |")
    w("|---|---|---|")
    for c in CATS:
        w(f"| {c} | {by_cat_rows.get(c, 0)} | {by_cat_lines[c]} |")
    w(f"| **total** | **{len(rows)}** | **{len(all_lines)}** distinct lines (a line can carry edits of several categories) |")
    w("")
    ndel = sum(1 for r in rows if r["change"] == "delete")
    nins = sum(1 for r in rows if r["change"] == "insert after" and r["category"] not in OWNER)
    nadd = sum(1 for r in rows if r["category"] in OWNER)
    w(f"Unified diff against the original (`walton.diff`): {minus} lines removed, {plus} added. That is "
      f"{minus - ndel} lines changed in place, {ndel} deleted (the duplicate keys `biased/1` and `good/1`, `posible/1`, "
      f"and the conclusion that was listed as a premise), {nins} inserted (the templates `cannot_present_sametime/3` "
      f"and `data_set/2`, and that conclusion under `conclusions:`); {nadd} of the recorded edits are the owner's "
      "decisions (below).")
    w("")
    kinds = collections.Counter()
    for r in rows:
        if r["category"] != "typo":
            continue
        why = r["reason"]
        k = ("template words" if why.startswith("template word") else
             "predicate names (incl. 1 key deleted)" if why.startswith(("misspelled predicate", "'posible'")) else
             "scheme ids" if why.startswith("scheme id") else
             "titles" if why.startswith("title") else
             "sources" if why.startswith("source") else "other")
        kinds[k] += 1
    w("Typo edits by kind: " + ", ".join(f"{k} {v}" for k, v in kinds.most_common()) + ". "
      "The `popular_practise` rename also renames the scheme id (counted under predicate names).")
    w("")

    # ---- the owner's decisions --------------------------------------------------
    w("## Changes decided by the owner")
    w("")
    w("These are not cleaning edits: they change what schemes say, or add one. Each was decided by the owner, on the "
      "date given; `tools/edits.py` records them as \"repair — owner decision\" and \"addition — owner decision\" "
      "(`OWNER_REPAIRS`, `ADDITIONS`), applied after the cleaning.")
    w("")
    w("| scheme | change | decided | why |")
    w("|---|---|---|---|")
    w("| `negative_consequences_reasoning` (new) | added from `A_reasoning/3_practical-reasoning.yml`, lines 511–524, a "
      "second block with the id `negative_consequences`; variables `[A,B]`; conclusion `¬should_be_performed(A)` in place "
      "of the block's `not_bring_about(A)`; exception `other_consequences(A)` | 2026-10-22 | It formalizes compendium "
      "33.3, Reasoning from Negative Consequences, which had no scheme: the master keeps only the first block of a "
      "duplicated id. The new id avoids the clash. `not_bring_about` is threat's predicate and meets no other consequence "
      "scheme, so the arguments could not rebut one another. Its two templates are in our own words, because the "
      "category file's repeat the book's premises. |")
    w("| `whole_into_parts` (new) | added from `A_reasoning/2_inductive-reasoning.yml`, under its upstream id | 2026-10-22 | "
      "It formalizes compendium 10.2, Division of the Whole into its Parts, which `walton.yml` lacked. Its formulas follow "
      "the book as printed (three parts stand for the book's n; only the positive reading). Its four templates are in "
      "our own words, because the category file's repeat the book. |")
    w("| `full_slippery_slope` (new) | added from `A_reasoning/5_casual-reasoning.yml`, under its upstream id; source filled "
      "in; `accept_all_steps` and `led_to_step` given the arguments `(C0,C2)` | 2026-10-22 | It formalizes compendium 53, "
      "the Full Slippery Slope Argument, which `walton.yml` lacked. The category file's source is empty, so the "
      "compendium's reference is used. Its two 0-ary atoms are about the series of cases at hand, and a 0-ary atom is "
      "one statement for every instance (\"0-ary atoms\", below). Its five templates are in our own words. |")
    w("| `precedent_slippery_slope` (new) | written for WAG: premises `claimed_exception(C0,R)`, `sets_precedent(C0,C1,R)`, "
      "`precedent_chain(C1,Cn,R)`, `intolerable_exception(Cn,R)`; conclusion `¬case_exception(C0)` | 2026-10-22 | It "
      "formalizes compendium 50, the Precedent Slippery Slope Argument, which no upstream file formalizes. One formula per "
      "printed premise, with the Related Cases Premise split into the first pair and the chain so that the book's CQ1 and "
      "CQ2 each have one. The intolerable case is taken to be the end of the series, Cn, as a slippery slope needs; the "
      "book prints C0 there and in CQ3 (the compendium, p. 340; the ch. 3 scan, p. 116, is unclear but looks the same), "
      "which can also be read as \"treating C0 as an exception would be intolerable because of where it leads\", a "
      "reason the chain premises then carry. A verifier with the book should confirm the reading. The conclusion negates `case_exception`, the premise of the "
      "argument for an exceptional case (`exceptional`), so that the slope meets the argument it counters (p. 117). Its "
      "templates are our own. |")
    w("| `positive_consequences`, `negative_consequences` | exception `other_consequences(A)` added; the template "
      "`other_consequences/1` rewritten | 2026-10-22 | The book's question about other, opposite consequences (33/CQ3) is "
      "printed for all of 33.1–33.4; `walton.yml` had its formula only in `negative_consequences_prudential` (33.4). An "
      "exception that holds defeats the argument whatever the weights, so the template now says that the opposite "
      "consequences outweigh the ones cited (it had said only that some exist, ungrammatically). |")
    w("| `slippery_slope_base_case` | conclusion `negative_consequences(A)` → `¬should_be_performed(A)` | 2026-10-22 | "
      "No other scheme uses `negative_consequences(A)`, so a slippery slope never met the consequence and practical "
      "reasoning schemes; the book's conclusion (49) is that the first step should not be taken. |")
    w("| `slippery_slope_base_case`, `slippery_slope_inductive_step` | source year 1999 → 1992 | 2026-10-22 | The "
      "book's compendium cites the same pages (49, pp. 93, 95) as 1992, as do `walton.yml`'s three other citations of the "
      "same book, and booksellers' listings date the Vale Press edition 1992. The book's ch. 4 (p. 114) cites p. 93, for "
      "gradualism, as \"Walton, 1999\"; its reference list, not checked, may give that as another work. |")
    w("| `negative_value` | conclusion `reason_retaining_goal(V,G)` → `reason_retracting_goal(V,G)` (new template); the "
      "`negative_affects/3` template rewritten | 2026-10-22 | Both said the opposite of the book's 19.2: the conclusion "
      "was copied from `positive_value`, and the template said a bad value supports commitment to the goal. Both "
      "templates are in our own words. |")
    w("")

    # ---- repairs list -------------------------------------------------------
    w("## Repairs that need the owner's approval")
    w("")
    w("| scheme | line | before | after | why |")
    w("|---|---|---|---|---|")
    rep = [
        ("moral_justification_ad_populum_excuse", "495, 497", "`conclusions:` empty; premise `morally_acceptable_action(X,A)`",
         "conclusion `morally_acceptable_action(X,A)`",
         "no conclusion, so Carneades built no argument. The formula's template starts \"Therefore, …\", and the category file B/4 has it as the conclusion"),
        ("sacrifice", "908", "`great_sacrifice(S,A,X)`", "`great_sacrifice(S,A,V)`",
         "X was undeclared (non-ground statement) and V occurred only in the conclusion; the template's third argument is \"the value %s\""),
        ("rhetorical_oppositions_normative", "768", "`if_treatment_required(X,A,A',X)`", "`if_treatment_required(X,A,A2,X2)`",
         "the 4th argument is \"the contrary treatment %s\" (X'), not X. Otherwise X' occurs only in the conclusion (validator error). The A'→A2 rename on the same line is a plain variable fix"),
        ("sorites_slippery_slope", "1557, 306", "premise `reapplication_sequence` (0-ary)", "`reapplication_sequence(A2,A)`; template `…from the clear area -%s- through the grey area to %s.`",
         "the conclusion's individual A occurred in no premise (validator error). The chain premise is the one that links the clear case A2 to A"),
    ]
    for r in rep:
        w("| " + " | ".join(r) + " |")
    w("")
    w("Three more edits change a declaration without touching a formula. They are recorded as `variable`, but they depart from "
      "\"leave declared-but-unused variables\", so the owner should see them. In `part_in_whole` (`[Y,X,P]`→`[Y,X,A]`, as in "
      "the category file A/2) and `verbal_classification_vagueness` (`[A,P]`→`[A,F]`), the unused P was the misnamed "
      "variable that the formulas use. Leaving it would keep a validator error and an extra dummy parameter. "
      "`verbal_slippery_slope` gets F declared; see the semantic concerns.")
    w("")

    # ---- id map -------------------------------------------------------------
    w("## Id map (old → new)")
    w("")
    w("| old id | new id | reason |")
    w("|---|---|---|")
    for o, n, why in ID_MAP:
        w(f"| `{o}` | `{n}` | {why} |")
    w("")
    w("No other id changed. The ids WAG's `core.yml` uses (`practical_reasoning1`, `positive_value`, `sign`, `credible_source`, "
      "`position_to_know`, `analogy`) are untouched. Hyphenated ids (`opposition-descriptive_1/2/3`, `opposition-normative`) "
      "are valid and were left; the category files spell some of them with underscores.")
    w("")
    w("### Predicate renames (old → new, everywhere they occur)")
    w("")
    preds = []
    seen = set()
    from edits import OPS
    for op in OPS:
        if op["op"] == "rename" and op["old"] not in seen:
            seen.add(op["old"])
            preds.append((op["old"], op["new"], len(op["lines"])))
    w("| old | new | lines |")
    w("|---|---|---|")
    for o, n, k in preds:
        w(f"| `{o}` | `{n}` | {k} |")
    w("| `true` | `is_true` | 2 (syntax: `true` is the boolean constant to the term parser) |")
    w("")

    # ---- edits by scheme ----------------------------------------------------
    w("## Every edit, grouped by scheme")
    w("")
    w("Format: **line** · category · reason, then the line before and after. An inserted line has an empty \"before\", "
      "a deleted line an empty \"after\". When a line has several edits, each is listed in the order applied, so one "
      "row's \"after\" is the next row's \"before\".")
    w("")
    groups = collections.OrderedDict()
    for r in sorted(rows, key=lambda r: (0 if r["scheme"] == "language" else 1, r["line"])):
        groups.setdefault(r["scheme"], []).append(r)

    def code(s):
        if s == "":
            return "_(none)_"
        s = s.replace("`", "'")
        return "`" + s + "`"

    for g, rs in groups.items():
        title = g
        if g == "language":
            title = "language (templates; \"used by\" names the schemes that use the predicate)"
        elif g in NEW_ID:
            title = f"{g} → {NEW_ID[g]}"
        w(f"### {title}")
        w("")
        for r in rs:
            extra = ""
            if g == "language":
                m = re.match(r"\s*([^:\s/]+)(/\d+)?:", r["after"] or r["before"])
                if m:
                    from edits import OPS as _OPS
                    rn = {op["old"]: op["new"] for op in _OPS if op["op"] == "rename"}
                    rn["true"] = "is_true"
                    u = sorted(users.get(rn.get(m.group(1), m.group(1)), []))
                    extra = f" (used by: {', '.join(u) if u else 'no scheme'})"
            kind = "" if r["change"] == "edit" else f" [{r['change']}]"
            w(f"- **{r['line']}**{kind} · {r['category']} · {r['reason']}{extra}")
            w(f"  - before: {code(r['before'])}")
            w(f"  - after: {code(r['after'])}")
        w("")

    # ---- cannot pass ----------------------------------------------------------
    w("## Schemes that cannot pass without a change of meaning")
    w("")
    w("All four instantiate correctly (every premise, assumption and exception, a conclusion, all ground), and their "
      "templates pass. They fail only the validator rule that each variable must occur in a premise. The rule exists "
      "because Carneades applies schemes by matching premises. `InstantiateScheme` with explicit parameters, which WAG "
      "uses, does not need it.")
    w("")
    w("| scheme | variable(s) in no premise | where they occur | proposal (in `tools/proposals.py`, not applied) |")
    w("|---|---|---|---|")
    w("| `ad_hominem` | S (the argument) | conclusion `bad_not_accepted(A,S)`, exception `stronger_than_character(S)` | add premise `advocates(A,S)` (\"%s advocates argument %s.\"), which the conclusion \"A's argument S\" presupposes. Walton's scheme has the argument only in the conclusion, so this adds a premise |")
    w("| `practical_necessary_condition` | P (agent), A2 (other goal) | assumption `possible_agent(X,P)`, exceptions `side_effect_action(P,X)`, `conflict_goals(P,A2,A)` | the premises are first-person (\"My goal…\", \"I have selected…\"). Proposal: `goal_is_action(P,A)` (\"The goal of %s is to bring about %s.\") and `conflict_goals(P,A)` (\"%s has some other goals that conflict with %s.\"), dropping A2 |")
    w("| `practical_sufficient_condition` | P, A2 | same | same |")
    w("| `regulative_rule` | A (agent) | conclusion `must_do(A,X)` | Grennan's scheme names the agent only in the conclusion. Proposal: apply the norm premise to A, `do_obligator(A)` (\"%s must not do what is prohibited -obligatory-.\"). This turns the general norm into a claim about A |")
    w("")
    w(f"With these {len(PROPOSALS)} edits (`walton.proposals.yml`), the validator reports 0 problems and the gate "
      f"passes for all {len(dump['schemes'])} schemes (`proposals-check.txt`).")
    w("")

    # ---- semantic concerns ----------------------------------------------------
    w("## Semantic concerns (left unchanged: these load and instantiate)")
    w("")
    for c in CONCERNS:
        w("- " + c)
    w("")
    w("### 0-ary atoms")
    w("")
    w("These 19 atoms stay 0-ary; only their language keys got `/0`. A 0-ary atom is one statement for every instance of "
      "its scheme, so answering it once answers it for all. That is right for a general principle and wrong for anything "
      "about the instance. I gave arguments only to `reapplication_sequence`, because the gate needed it. The others "
      "change which statement the scheme asserts and are left to the owner.")
    w("")
    w("| atom | scheme, role | what it is about | suggestion |")
    w("|---|---|---|---|")
    for a in ZERO_ARY:
        w("| " + " | ".join(a) + " |")
    w("")

    # ---- variables ------------------------------------------------------------
    w("## Unused variables")
    w("")
    w("- **Original, declared but used in no formula:** `part_in_whole` P and `verbal_classification_vagueness` P. "
      "Each was the misnamed form of a variable the formulas use but nobody declared (A and F respectively), so the declaration was "
      "corrected (lines 684, 817), not left. "
      "The primed variables P', X', A', R' (6 schemes) were declared, but the parser cannot read them; they were renamed P2, X2, A2, R2.")
    unused = (clean or {}).get("declared_but_unused_variables") or []
    w(f"- **Cleaned file, declared but used in no formula:** {len(unused)}.")
    w("- **Cleaned file, used but in no premise** (the validator counts these as unused): `ad_hominem` S; "
      "`practical_necessary_condition` P, A2; `practical_sufficient_condition` P, A2; `regulative_rule` A. See above.")
    w("- **Undeclared variables in the original:** A in `part_in_whole`, F in `verbal_classification_vagueness`, X in "
      "`sacrifice` (repaired to V), F in `verbal_slippery_slope` (declared). The cleaned file has none.")
    w("")

    # ---- not changed ----------------------------------------------------------
    w("## Left alone on purpose")
    w("")
    for c in LEFT:
        w("- " + c)
    w(f"- Orphan templates (no scheme uses them; {len(orphans)}): " + ", ".join(f"`{o}`" for o in orphans) + ".")
    w("")
    with open(os.path.join(outdir, "CHANGELOG.md"), "w", encoding="utf-8") as f:
        f.write("\n".join(L) + "\n")
    print("CHANGELOG.md written:", len(L), "lines")


CONCERNS = [
    "`pop_scheme`: two conclusions, `is_true(A)` and `should_accept(A)`. Carneades builds two arguments that share the "
    "premises and the undercutter. The book's conclusion is a disjunction (\"A is true, or you should accept A\"). The owner should decide "
    "whether to split the scheme or keep one conclusion.",
    "`common_folk_ad_populum`: premise and conclusion look swapped. The premise `being_credible` reads \"Therefore, you ought "
    "to take what I say as … more credible\" and the conclusion `background` reads \"The speaker is an ordinary person…\". "
    "The scheme has no variables, so every instance is the same argument.",
    "`common_folks_group_ad_populum`: the conclusion `accept_saying` is 0-ary, so every instance concludes the same statement.",
    "`from_example` (`both_properties`), `fear` (`prevent(Y)`) and `bias_ad_hominem` (`morally_bad(A)`): a premise has a "
    "\"Therefore, …\" template, so premise and conclusion may be swapped.",
    "`division_generic` is a verbatim copy of `composition_generic` (parts → whole). Division should go from the whole to its "
    "parts; the category file A/2 has a `whole_into_parts` scheme.",
    "`opposition-normative`: the premise `receives_evaluation(A,X2)` and the conclusion `cannot_receive_evaluation(A,X2)` are "
    "about the same evaluation X2; the premise is presumably about X. The 3-place premise uses `cannot_present_sametime`, while the "
    "unused template `cannot_opposite_evaluation/3` (\"If %s receives the evaluation %s. It cannot receive the opposite "
    "evaluation %s…\") looks like the intended predicate. I added a `cannot_present_sametime/3` template worded like the /2 one, "
    "and did not change the formula.",
    "The other opposition schemes (`opposition-descriptive_1/2/3`, `rhetorical_oppositions_descriptive`; book p. 317–318) "
    "have the content errors the audit lists (REPORT.md 3.5). I renamed their primed variables and changed nothing else; "
    "they were not checked against the book.",
    "`evidence_hypothesis_falsification` concludes `¬A`, which restates its premise `event_false(A)`; modus tollens gives `¬H`. "
    "Its assumption is `sufficient_correlation(A,H)`, while verification has `sufficient_correlation(H,A)` (different statements).",
    "`witness_testimony`: `in_domain(D,S)` reverses the order every other scheme uses (`in_domain(S,D)`). It is inherited from "
    "upstream Carneades.",
    "`bias`: the conditional premise `not_taken_both(X,C)` is about X, not the biased arguer A, so X is otherwise unconstrained.",
    "`two_person_practical`: the exceptions `alternative_means(A,X)` and `side_effect_other_act(B,A)` do not fit the order "
    "of their templates' arguments (\"alternative means of realizing %s other than %s\"; \"Carrying out %s … %s other goals\").",
    "`exceptional` (was `exeptional`): `case_exception(A)` is both a premise and an assumption, and `rule_can_waived(X)` uses X, "
    "which no other formula shares.",
    "`verbal_slippery_slope`: I declared F, the property in `individuals_property(A1,F,A2)`. Every other formula uses P for "
    "the property, so F is probably P. As written, the conditional is not linked to P.",
    "`threat` (was `treat`): the premise `position_to_know(X)` (\"I am in position to bring about %s\") shares its name with the "
    "position-to-know predicate `position_to_know/2`. Other predicates used with two arities: `data_set/1`–`/2`, `excuse/0`–`/1`, "
    "`group_member/1`–`/2`, `cannot_present_sametime/2`–`/3`, `relevant_differences/1`–`/2`. They are different statements and easy to confuse.",
    "`abductive_backward`: `data_set(D,A)` has 2 arguments, but `abductive_forward` has `data_set(D)`. I added a `data_set/2` "
    "template (\"%s is a set of data or supposed facts in case %s.\"); the role of A there is a guess.",
    "`exception_to_rule` (was `exeption_to_rule`): the content is upstream Carneades `precedent` with a new id and title. "
    "The README says Walton renamed the scheme, but the formulas were not changed. Check against the book (AS57).",
    "Templates I fitted to the formula as used, where the original template suggests the formula may have lost an argument: "
    "`strongly_occur(X)` (template had the consequences C too), `cited_bad_consequences(X)` (the consequences; C occurs in "
    "`assert_bad(X,C)`), `excuse` in `regulative_rule` (template had the agent; `excuse/1` is taken by `plea_for_excuse`).",
    "`biased/1` had two templates. The loaded one, \"Arguer %s is biased.\", is kept so loading is unchanged. It is also used "
    "for witnesses and sources (`credible_source`, `witness_testimony`), where upstream's deleted \"%s is biased.\" fits better.",
    "`waste` now uses `possible(X)`, whose existing template reads \"Action %s is possible.\", though X is a goal. "
    "`practical_reasoning2`'s `possible(G)` has the same mismatch; the deleted `posible/1` read \"%s is possible.\".",
    "Titles: `verbal_classification_vagueness` has the title of `definition_to_verbal_classification`; `rule_mental_world` has "
    "an empty title; `slippery_slope_base_case` and `slippery_slope_inductive_step` share a title; `abduction` has no `meta`.",
]

ZERO_ARY = [
    ("`goal_prestige`", "snob_appeal_ad_populum, premise", "the respondent (\"you\"), who has no variable", "leave, or add a respondent variable"),
    ("`background`", "common_folk_ad_populum, conclusion", "the speaker; the scheme has no variables", "see concerns (swapped)"),
    ("`being_credible`", "common_folk_ad_populum, premise", "the speaker; no variables", "see concerns (swapped)"),
    ("`accept_saying`", "common_folks_group_ad_populum, conclusion", "\"what I say\"; no variable", "a variable for what is said"),
    ("`more_important_ends`", "end_and_means_2, premise", "general principle", "fine as 0-ary"),
    ("`not_normal`", "negative_normal_expectations, conclusion", "\"the situation\"; no variable", "e.g. `not_normal(A)`"),
    ("`sufficient_evidence`", "negative_practical, assumption", "the search for evidence about A", "e.g. `sufficient_evidence(A)`"),
    ("`more_reliable`", "sign, exception", "\"the sign\" = A", "`more_reliable(A)` (as in cq_drafts.json)"),
    ("`reference_class`", "retroductive, assumption", "\"this event\"; no event variable", "leave"),
    ("`attack_well_supported`", "ad_hominem, assumption", "the character attack on A", "`attack_well_supported(A)`"),
    ("`character_relevant`", "ad_hominem, assumption", "\"the argument\" = S", "`character_relevant(S)`"),
    ("`proper_sequence`", "gradualism, assumption", "the sequence P1…P4", "e.g. `proper_sequence(P1,P4)`"),
    ("`synonymous`", "rule_physical_world_1, premise", "general principle", "fine as 0-ary"),
    ("`not_necessarily_true`", "rule_physical_world_1, exception", "\"First premise\" = `is(W,D)`", "`not_necessarily_true(W,D)`"),
    ("`correctly_report`", "rule_physical_world_2, premise", "general principle", "fine as 0-ary"),
    ("`rule_changed`", "rule_physical_world_2, exception", "the rule of `official_rulebook(I,D,W)`", "e.g. `rule_changed(I,D,W)`"),
    ("`speakers_correct`", "rule_mental_world, premise", "general principle", "fine as 0-ary"),
    ("`do_obligator`", "regulative_rule, premise", "general norm (\"One must…\")", "see the regulative_rule proposal"),
    ("`excuse`", "regulative_rule, exception", "the agent's excuse (template had %s)", "needs a name other than `excuse/1`, e.g. `has_excuse(A)`"),
]

LEFT = [
    "Grammar and wording in templates (e.g. \"you will a member\", \"There are evidence\", \"If %s than %s\", \"an typical\"). "
    "Only misspelled words and OCR splits were fixed.",
    "`common_folk_ad_populum` id, `do_obligator` name, `rule_can_waived` name: not misspellings.",
    "Titles beginning \"Argumentation From …\" (`sacrifice`, `regulative_rule`), empty sources (`retroductive`, `cause_to_effect`, "
    "`defeasible_modus_ponens`, `verbal_classification`), and stray spaces in sources.",
    "A space before `(` in `carry_out_act (A,X)` and `data_account (A1,D)`, and the double space in `-  vagueness_rejected(A)`: "
    "the parser accepts them.",
    "The stray \" s\" in the `data_account/2` template, and the soft hyphen in \"pp.429­448\".",
    "`langugage:` occurs only in the category file `B_source-based-arguments/1_…yml`, not in `walton.yml`.",
    "Negated formulas have no `¬p/n` WAG templates (WAG lint warning, see Gate).",
]

if __name__ == "__main__":
    main(sys.argv[1])
