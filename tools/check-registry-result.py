#!/usr/bin/env python3
"""Fail closed on a registry tool failure; tolerate only the named D042 debt.

The program is built first so its own exit status can be distinguished from a
Go build/runtime failure. Delete the allowance when D042 is implemented.
"""
import argparse
from pathlib import Path
import re


def check(status, stdout, stderr):
    if stderr.strip():
        return "registry emitted stderr"
    lines = stdout.splitlines()
    if len(lines) < 2 or not re.fullmatch(r"[1-9][0-9]* schemes, [1-9][0-9]* distinct patterns; related pairs: map\[.*\]", lines[0]):
        return "registry did not emit its complete report header"
    summary = re.fullmatch(r"([0-9]+) error\(s\), ([0-9]+) warning\(s\)", lines[-1])
    if not summary:
        return "registry did not emit its completion summary"
    errors = [line for line in lines[1:-1] if line.startswith("ERROR ")]
    warnings = [line for line in lines[1:-1] if line.startswith("warning ")]
    if len(errors) + len(warnings) != len(lines) - 2:
        return "registry emitted an unrecognized report line"
    if (len(errors), len(warnings)) != tuple(map(int, summary.groups())):
        return "registry summary does not match its findings"
    if not errors:
        return None if status == 0 else "registry failed without a reported finding"
    known = re.compile(r"ERROR composition_generic = division_generic: the same pattern under two names \(.+\): merge them, or declare them aliases")
    if len(errors) != 1 or not known.fullmatch(errors[0]):
        return "registry found an error other than the one recorded in D042"
    return None if status == 1 else "registry exit status disagrees with its known finding"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--status", required=True, type=int)
    parser.add_argument("stdout", type=Path)
    parser.add_argument("stderr", type=Path)
    args = parser.parse_args()
    error = check(args.status, args.stdout.read_text(), args.stderr.read_text())
    if error:
        parser.exit(1, f"Registry gate failed: {error}\n")
    print("Registry gate passed (only the explicitly recorded D042 debt is allowed).")


if __name__ == "__main__":
    main()
