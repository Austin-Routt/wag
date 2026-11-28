"""Tests of agreement.py's exact target (docs/annotation-records.md §11).

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


if __name__ == "__main__":
    unittest.main()
