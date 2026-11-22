"""Adverse controls for the CI wrapper, including failures the old grep lost."""
import importlib.util
from pathlib import Path
import unittest
import sys

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("registry_result", Path(__file__).with_name("check-registry-result.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
HEADER = "116 schemes, 115 distinct patterns; related pairs: map[identical:1]\n"
KNOWN = "ERROR composition_generic = division_generic: the same pattern under two names (all formulas): merge them, or declare them aliases\n"


class RegistryGate(unittest.TestCase):
    def test_expected_debt_and_clean_result(self):
        self.assertIsNone(module.check(1, HEADER + KNOWN + "1 error(s), 0 warning(s)\n", ""))
        self.assertIsNone(module.check(0, HEADER + "0 error(s), 0 warning(s)\n", ""))

    def test_failures_cannot_hide_without_error_lines(self):
        for status, out, err in [(1, "", "build failed"), (2, "", ""),
                                 (1, HEADER + "0 error(s), 0 warning(s)\n", ""),
                                 (0, "", ""), (1, HEADER + KNOWN, "")]:
            with self.subTest(status=status, out=out, err=err):
                self.assertIsNotNone(module.check(status, out, err))

    def test_extra_errors_and_inconsistent_summary(self):
        cases = [HEADER + KNOWN + "ERROR other: broken\n2 error(s), 0 warning(s)\n",
                 HEADER + KNOWN + "0 error(s), 0 warning(s)\n",
                 HEADER + "panic: failure\n0 error(s), 0 warning(s)\n"]
        for out in cases:
            with self.subTest(out=out):
                self.assertIsNotNone(module.check(1, out, ""))


if __name__ == "__main__":
    unittest.main()
