"""Tests of agreement.py's exact target (docs/annotation-records.md §11), with its structural version (F6).

Run: python3 -I docs/kialo-batch2/test_agreement.py
"""
import importlib.util
import os
import sys
import unittest

sys.dont_write_bytecode = True  # no __pycache__ beside agreement.py: the tree stays clean
HERE = os.path.dirname(os.path.abspath(__file__))
_spec = importlib.util.spec_from_file_location("agreement", os.path.join(HERE, "agreement.py"))
ag = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(ag)


def target(claim, instance, part="conclusion", formula=None):
    return {"target": {"actual": {"claim": claim, "instance": instance, "part": part, "formula": formula}}}


def claim_with(cid, secondary_conclusion):
    """A claim with a primary instance and a secondary one."""
    return {"kialo_claim_id": cid, "instances": [
        {"id": cid + "#1", "primary": True, "scheme": {"walton_id": "rule"}, "conclusion": {"formula": "p(" + cid.replace(".", "_") + ")"}},
        {"id": cid + "#2", "primary": False, "scheme": {"walton_id": "cause_to_effect_modal"}, "conclusion": {"formula": secondary_conclusion}},
    ]}


class ExactTarget(unittest.TestCase):
    def setUp(self):
        # Two claims whose secondary instances name the same scheme and
        # conclusion: only the claim tells their targets apart.
        self.S = {"c.1": claim_with("c.1", "may_occur(x)"), "c.2": claim_with("c.2", "may_occur( x )")}

    def value(self, item):
        return ag.exact(item, "target_instance", self.S)

    def test_primary_targets_of_different_claims_disagree(self):
        a, b = self.value(target("c.1", "c.1#1")), self.value(target("c.2", "c.2#1"))
        self.assertEqual((a, b), ("c.1 primary", "c.2 primary"))

    def test_secondary_targets_of_different_claims_disagree(self):
        a, b = self.value(target("c.1", "c.1#2")), self.value(target("c.2", "c.2#2"))
        self.assertEqual(a, "c.1 secondary:cause_to_effect_modal:may_occur(x)")
        self.assertEqual(b, "c.2 secondary:cause_to_effect_modal:may_occur(x)")

    def test_a_format1_target_is_its_claims_primary(self):
        self.assertEqual(self.value(target("c.1", None)), "c.1 primary")
        self.assertEqual(self.value(target("c.2", "primary instance")), "c.2 primary")

    def test_a_target_the_file_does_not_hold_is_unresolved(self):
        self.assertEqual(self.value(target("c.9", "c.9#1")), ag.UNRESOLVED)
        self.assertEqual(self.value(target("c.1", "c.1#7")), ag.UNRESOLVED)
        self.assertEqual(self.value(target("c.1", {"unresolved": "which instance"})), ag.UNRESOLVED)

    def test_no_target_is_empty(self):
        self.assertEqual(ag.exact({"target": {}}, "target_instance", self.S), "")



def item_at(formula="brings_about(a,g)", part="premise", instance="c.0#1"):
    """A claim backing a premise of c.0's primary instance; its target's fields as given."""
    return {"kialo_claim_id": "c.1",
            "instances": [{"id": "c.1#1", "primary": True,
                           "scheme": {"walton_id": "verbal_classification", "fit": "reconstructed"},
                           "conclusion": {"formula": "instance(o,g)"}}],
            "role": {"kind": "premise_backing", "polarity": "for", "cq": None},
            "target": {"kialo_parent": "c.0", "actual": {"claim": "c.0", "instance": instance,
                                                         "part": part, "formula": formula}}}


PARENT = {"kialo_claim_id": "c.0", "instances": [
    {"id": "c.0#1", "primary": True, "scheme": {"walton_id": "positive_consequences", "fit": "reconstructed"},
     "premises": [{"role": "premise", "formula": "brings_about(a,g)"}], "conclusion": {"formula": "should_be_performed(a)"}}]}
MARK = {"unresolved": "a background condition of the premise"}
RELATION = ("role", "cq", "target_part", "target_instance", "target_formula")


class UnsettledRelation(unittest.TestCase):
    """Cases A1-A3 (d086-first-gold.md §6; D102): an unsettled relation is not scored."""

    def values(self, item):
        S = {"c.0": PARENT, "c.1": item}
        return {name: ag.field_value(name, item, S) for name in ag.FIELDS}

    def test_a1_unresolved_formula_leaves_the_relation_out(self):
        v = self.values(item_at(formula=MARK))
        for name in RELATION:
            self.assertEqual(v[name], ag.UNRESOLVED, name)
        self.assertEqual((v["scheme"], v["fit"], v["polarity"], v["target_claim"]),
                         ("verbal_classification", "reconstructed", "for", "c.0"))

    def test_a2_unresolved_part_leaves_the_relation_out(self):
        v = self.values(item_at(part=MARK))
        for name in RELATION:
            self.assertEqual(v[name], ag.UNRESOLVED, name)
        self.assertEqual(v["target_claim"], "c.0")

    def test_a3_unresolved_instance_alone_leaves_only_the_instance_out(self):
        v = self.values(item_at(instance=MARK))
        self.assertEqual(v["target_instance"], ag.UNRESOLVED)
        self.assertEqual((v["role"], v["target_part"], v["target_formula"]),
                         ("premise_backing", "premise", "brings_about(a,g)"))
        self.assertNotEqual(v["cq"], ag.UNRESOLVED)

    def test_a1_resolved_control_scores_the_relation(self):
        v = self.values(item_at())
        for name in RELATION:
            self.assertNotEqual(v[name], ag.UNRESOLVED, name)

    def test_a1_report_counts_the_unsettled_item(self):
        """The report: the item left out of the relation's fields and counted as unresolved there."""
        import json, subprocess, tempfile
        with tempfile.TemporaryDirectory() as d:
            files = []
            for name, it in (("a.json", item_at(formula=MARK)), ("b.json", item_at())):
                path = os.path.join(d, name)
                with open(path, "w", encoding="utf-8") as fh:
                    json.dump({"items": [PARENT, it]}, fh)
                files.append(path)
            out = os.path.join(d, "report.json")
            r = subprocess.run([sys.executable, "-I", "-B", os.path.join(HERE, "agreement.py"), *files, "--json", out],
                               capture_output=True, text=True)
            self.assertEqual(r.returncode, 0, r.stderr)
            fields = json.load(open(out, encoding="utf-8"))["fields"]
        for name in RELATION:
            self.assertEqual(fields[name]["unresolved"], 1, name)
        self.assertEqual(fields["polarity"].get("unresolved"), 0)
        self.assertEqual(fields["scheme"].get("unresolved"), 0)

class AfterTheReview(unittest.TestCase):
    """Cases A5-A7 (d086-first-gold.md §6), after the independent review of step 1."""

    def values(self, item):
        S = {"c.0": PARENT, "c.1": item}
        return {name: ag.field_value(name, item, S) for name in ag.FIELDS}

    def test_a5_a_reason_go_does_not_trim_is_a_reason(self):
        # U+001C to U+001F: Python's str.strip removes them, Go's strings.TrimSpace does not.
        v = self.values(item_at(formula={"unresolved": "\x1c\x1d\x1e\x1f"}))
        for name in RELATION:
            self.assertEqual(v[name], ag.UNRESOLVED, name)
        self.assertTrue(ag.is_marker({"unresolved": "\x1f"}, "unresolved"))
        self.assertFalse(ag.is_marker({"unresolved": "\u3000\t\x85"}, "unresolved"))

    def test_a6_a_target_that_is_not_an_object(self):
        for target in ("c.0", ["c.0"], {"actual": "c.0"}, {"actual": ["c.0"]}):
            with self.subTest(target=target):
                it = dict(item_at(), target=target)
                self.assertFalse(ag.relation_unsettled(it))
                self.assertEqual(ag.field_value("role", it, {}), "premise_backing")

    def test_a7_role_counts_and_the_pilot_follow_d102(self):
        import json, subprocess, tempfile
        pilot = {"items": [
            {"kialo_claim_id": "c.0", "scheme": {"id": "walton.yml positive_consequences"}, "role": {"kind": "none"}},
            {"kialo_claim_id": "c.1", "scheme": {"id": "walton.yml verbal_classification"}, "role": {"kind": "premise_attack"}}]}
        with tempfile.TemporaryDirectory() as d:
            files = []
            for name, it in (("a.json", item_at(formula=MARK)), ("b.json", item_at()), ("pilot.json", None)):
                path = os.path.join(d, name)
                with open(path, "w", encoding="utf-8") as fh:
                    json.dump(pilot if it is None else {"items": [PARENT, it]}, fh)
                files.append(path)
            out = os.path.join(d, "report.json")
            r = subprocess.run([sys.executable, "-I", "-B", os.path.join(HERE, "agreement.py"), files[0], files[1],
                                "--pilot", files[2], "--json", out], capture_output=True, text=True)
            self.assertEqual(r.returncode, 0, r.stderr)
            report = json.load(open(out, encoding="utf-8"))
        self.assertEqual(report["role_counts"]["A"], {"none": 1, ag.UNRESOLVED: 1})
        self.assertEqual(report["role_counts"]["B"], {"none": 1, "premise_backing": 1})
        # A's c.1 is left out of the pilot's role comparison; B's disagrees with the pilot.
        self.assertEqual(report["pilot"]["A"]["role_agreement"], 1.0)
        self.assertEqual(report["pilot"]["A"]["role_unresolved"], 1)
        self.assertEqual(report["pilot"]["B"]["role_agreement"], 0.5)
        self.assertEqual(report["pilot"]["B"]["role_unresolved"], 0)


class DiscussionInstances(unittest.TestCase):
    """Case RC10 (d086-first-gold.md §6; D096, annotation-records.md §11 and §13)."""

    def test_rc10_instances_leave_the_discussions_out(self):
        disc = {"id": "c.0#2", "primary": False, "by": "the discussion",
                "scheme": {"walton_id": "practical_reasoning2", "fit": "reconstructed"},
                "conclusion": {"formula": "should_be_performed(a)"}}
        with_it = {**PARENT, "instances": PARENT["instances"] + [disc]}
        self.assertEqual(ag.instances(with_it), ag.instances(PARENT))
        # A target at it is compared by its scheme and conclusion, as at a secondary instance.
        S = {"c.0": with_it}
        it = {"target": {"actual": {"claim": "c.0", "instance": "c.0#2", "part": "conclusion",
                                    "formula": "should_be_performed(a)"}}}
        self.assertEqual(ag.exact(it, "target_instance", S), "c.0 secondary:practical_reasoning2:should_be_performed(a)")



# Go's unicode.IsSpace, written out here so that the test does not read the constant it checks
# (F4 of the external review of exchange 3; WAG's check of it, the planted defect 02-c).
GO_SPACES = [0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x20, 0x85, 0xA0, 0x1680, 0x2000, 0x2001, 0x2002, 0x2003,
             0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000]


def _go_space(name):
    """The GO_SPACE a script assigns, read from its source (the scripts run when imported)."""
    import ast
    with open(os.path.join(HERE, name + ".py"), encoding="utf-8") as fh:
        tree = ast.parse(fh.read())
    found = [ast.literal_eval(n.value) for n in tree.body if isinstance(n, ast.Assign)
             and any(isinstance(t, ast.Name) and t.id == "GO_SPACE" for t in n.targets)]
    if len(found) != 1:
        raise AssertionError(f"{name}.py assigns GO_SPACE {len(found)} times")
    return found[0]


def goal_instance(iid, goal, primary=False, by=None, order=1, spaced=False, unresolved=False):
    """An instance of t.1 for its goal: practical reasoning concluding should_be_performed(tax), its
    premises in the order given, a formula with spaces if asked, a premise unresolved if asked."""
    ps = [{"role": "premise", "formula": "goal_of_agent( g , " + goal + " )" if spaced else "goal_of_agent(g," + goal + ")"},
          {"role": "premise", "formula": "would_bring_about(tax," + goal + ")"},
          {"role": "exception", "formula": "side_effect(tax)"}]
    if unresolved:
        ps[1]["formula"] = {"unresolved": "which effect"}
    ins = {"id": iid, "primary": primary, "scheme": {"walton_id": "practical_reasoning2"},
           "premises": ps[::order], "conclusion": {"formula": "should_be_performed(tax)"}}
    if by:
        ins["by"] = by
    return ins


def target_claim(*instances):
    return {"kialo_claim_id": "t.1", "instances": list(instances)}


class StructuralTarget(unittest.TestCase):
    """Cases TS1-TS6 (d086-first-gold.md §6; F6 of the external review of exchange 3):
    target_instance_v2 names a secondary target by its structure; target_instance keeps its
    meaning. TS7, the Go measure on the same fixtures, is TestStructuralTargetMatchesPython."""

    def values(self, S, instance, formula="should_be_performed(tax)", part="conclusion"):
        item = target("t.1", instance, part, formula)
        return ag.field_value("target_instance", item, S), ag.field_value("target_instance_v2", item, S)

    def test_ts1_two_reconstructions_with_one_signature(self):
        S = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env"),
                                 goal_instance("t.1#3", "health"))}
        v1a, v2a = self.values(S, "t.1#2")
        v1b, v2b = self.values(S, "t.1#3")
        self.assertEqual(v1a, v1b)  # the historical field: one signature
        self.assertEqual(v1a, "t.1 secondary:practical_reasoning2:should_be_performed(tax)")
        self.assertNotEqual(v2a, v2b)
        self.assertEqual(v2a, "t.1 secondary:practical_reasoning2:should_be_performed(tax)"
                              "[exception=side_effect(tax);premise=goal_of_agent(g,env);premise=would_bring_about(tax,env)]by=")

    def test_ts2_the_same_reconstruction_however_written(self):
        A = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env"))}
        B = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env", order=-1, spaced=True))}
        self.assertEqual(self.values(A, "t.1#2"), self.values(B, "t.1#2"))

    def test_ts3_a_primary_target_is_its_claims_primary(self):
        A = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True))}
        B = {"t.1": target_claim(goal_instance("t.1#1", "health", primary=True))}
        self.assertEqual(self.values(A, "t.1#1"), ("t.1 primary", "t.1 primary"))
        self.assertEqual(self.values(B, "t.1#1"), ("t.1 primary", "t.1 primary"))

    def test_ts4_a_premise_unresolved(self):
        S = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env", unresolved=True))}
        v1, v2 = self.values(S, "t.1#2")
        self.assertEqual(v1, "t.1 secondary:practical_reasoning2:should_be_performed(tax)")
        self.assertEqual(v2, ag.UNRESOLVED)

    def test_ts5_by_the_discussion_or_the_claims_own(self):
        S = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env"),
                                 goal_instance("t.1#3", "env", by="the discussion"))}
        self.assertEqual(self.values(S, "t.1#2")[0], self.values(S, "t.1#3")[0])
        self.assertNotEqual(self.values(S, "t.1#2")[1], self.values(S, "t.1#3")[1])
        self.assertTrue(self.values(S, "t.1#3")[1].endswith("]by=the discussion"))

    def test_ts6_an_unsettled_relation_leaves_both_out(self):
        S = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), goal_instance("t.1#2", "env"))}
        self.assertEqual(self.values(S, "t.1#2", formula={"unresolved": "which formula"}), (ag.UNRESOLVED, ag.UNRESOLVED))
        self.assertIn("target_instance_v2", ag.RELATION_FIELDS)


class MalformedShapes(unittest.TestCase):
    """Cases TS8 and TS10 (d086-first-gold.md §6; the independent check of F3, F5 and F6), read as Go
    reads them; TestTS9ParityWithPythonOnEveryShape compares every value with the Go measure."""

    def test_ts8_premises_that_are_not_a_list_are_none(self):
        for premises in ("ab", {"x": 1, "y": 2}, 5, True):
            with self.subTest(premises=premises):
                ins = goal_instance("t.1#2", "env")
                ins["premises"] = premises
                S = {"t.1": target_claim(goal_instance("t.1#1", "env", primary=True), ins)}
                item = target("t.1", "t.1#2", "conclusion", "should_be_performed(tax)")
                self.assertEqual(ag.field_value("target_instance_v2", item, S),
                                 "t.1 secondary:practical_reasoning2:should_be_performed(tax)[]by=")

    def test_ts10_the_primary_as_go_reads_it(self):
        first = {"id": "t.1#1", "primary": False, "conclusion": {"formula": "first(x)"}}
        for value in ("true", 1):
            with self.subTest(primary=value):
                later = {"id": "t.1#2", "primary": value, "conclusion": {"formula": "second(x)"}}
                claim = target_claim(first, later)
                self.assertIs(ag.primary(claim), first)
                item = target("t.1", None, "conclusion", None)
                self.assertEqual(ag.field_value("target_formula", item, {"t.1": claim}), "first(x)")
        # The boolean true marks the primary wherever it is; the first object stands in only when none is marked.
        marked = dict(first, id="t.1#2", primary=True)
        self.assertIs(ag.primary(target_claim(first, marked)), marked)
        self.assertEqual(ag.primary(target_claim(7, first))["id"], "t.1#1")

    def test_ts10_instances_that_are_not_a_list_or_not_objects(self):
        item = target("t.1", None, "conclusion", "should_be_performed(tax)")
        for insts in (5, "ab", None, []):
            with self.subTest(instances=insts):
                claim = {"kialo_claim_id": "t.1", "instances": insts}
                self.assertEqual(ag.instances(claim), "")
                self.assertEqual(ag.primary(claim), {})
                self.assertEqual(ag.field_value("target_instance", item, {"t.1": claim}), ag.UNRESOLVED)
        # An entry that is not an object is an instance of no scheme and no conclusion, as in Go.
        claim = target_claim(7, goal_instance("t.1#1", "env", primary=True))
        self.assertEqual(ag.instances(claim), ":|practical_reasoning2:should_be_performed(tax)")
        # An empty instance is an instance: the claim's primary.
        self.assertEqual(ag.field_value("target_instance", item, {"t.1": target_claim({})}), "t.1 primary")


class CompactParity(unittest.TestCase):
    """compact() removes what Go's strings.Fields removes, and keeps U+001C-U+001F, as Go does."""

    def test_go_spaces_are_removed(self):
        for cp in GO_SPACES:
            with self.subTest(codepoint=hex(cp)):
                self.assertEqual(ag.compact("p(a" + chr(cp) + "b)"), "p(ab)")

    def test_what_go_keeps_is_kept(self):
        for cp in (0x1C, 0x1D, 0x1E, 0x1F, 0x200B, 0xFEFF):
            with self.subTest(codepoint=hex(cp)):
                f = "p(a" + chr(cp) + "b)"
                self.assertEqual(ag.compact(f), f)
                self.assertNotEqual(ag.compact(f), ag.compact("p(ab)"))

    def test_the_three_copies_are_gos_set(self):
        want = "".join(chr(cp) for cp in GO_SPACES)
        for name in ("agreement", "validate", "to_wag"):
            with self.subTest(module=name):
                self.assertEqual(sorted(_go_space(name)), sorted(want))



class GoReading(unittest.TestCase):
    """AG1 and AG2 (docs/annotation-records.md §11; N6 of the independent check of F1, F2 and F4): a
    value is compared as the Go measure reads it."""

    def test_the_scheme_as_written(self):
        # AG1: Go's FieldIn reads walton_id as written; agreement.py took Python's whitespace off it.
        item = lambda wid: {"instances": [{"primary": True, "scheme": {"walton_id": wid}}]}
        for wid in (" analogy", "analogy ", "analogy\u001c", "\u001canalogy"):
            with self.subTest(walton_id=repr(wid)):
                self.assertEqual(ag.field_value("scheme", item(wid), {}), wid)
                self.assertNotEqual(ag.field_value("scheme", item(wid), {}), ag.field_value("scheme", item("analogy"), {}))
        self.assertEqual(ag.field_value("scheme", item("analogy"), {}), "analogy")

    def test_split_formula_trims_gos_spaces_only(self):
        # AG2: Go's splitFormula trims with strings.TrimSpace; U+001C to U+001F are kept.
        self.assertEqual(ag.split_formula("p( a\u001c , b\u00a0)"), ("p", ["a\u001c", "b"]))
        for cp in (0x1C, 0x1D, 0x1E, 0x1F):
            with self.subTest(codepoint=hex(cp)):
                self.assertEqual(ag.split_formula("p(a" + chr(cp) + ",b)"), ("p", ["a" + chr(cp), "b"]))
                self.assertEqual(ag.split_formula("q" + chr(cp) + "(a)"), ("q" + chr(cp), ["a"]))
                # At the formula's ends too (the check's T4: Python's strip there took them off).
                self.assertEqual(ag.split_formula(chr(cp) + "p(a)"), (chr(cp) + "p", ["a"]))
                self.assertEqual(ag.split_formula("p(a)" + chr(cp)), ("p(a)" + chr(cp), []))
        for cp in GO_SPACES:
            with self.subTest(codepoint=hex(cp)):
                self.assertEqual(ag.split_formula(chr(cp) + "p(" + chr(cp) + "a" + chr(cp) + ")" + chr(cp)), ("p", ["a"]))

if __name__ == "__main__":
    unittest.main()
