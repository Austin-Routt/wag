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

if __name__ == "__main__":
    unittest.main()
