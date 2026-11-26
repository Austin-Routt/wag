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


if __name__ == "__main__":
    unittest.main()
