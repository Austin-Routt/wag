#!/usr/bin/env python3
"""Validate D081's scoped formulas, source links and optional exact YAML delta."""
import argparse
import copy
import json
from pathlib import Path

import yaml

from check_d084_migration import formula_values, require, schemes


CHAIN = "reapplication_sequence(A2,A,P)"
TEXT = "A chain of modus ponens steps leads from %s, in the clear area, through the grey area to %s, each step concerning %s."
EXPECTED = {
    "moral_justification_ad_populum_excuse": ("conclusions", "morally_acceptable_action(X,A)"),
    "sacrifice": ("premises", "great_sacrifice(S,A,V)"),
    "rhetorical_oppositions_normative": ("premises", "if_treatment_required(X,A,A2,X2)"),
    "sorites_slippery_slope": ("premises", CHAIN),
}
# The forms the repairs replaced (the upstream file, primes renamed by the
# cleaning). Without --before, as CI runs it, the check must still fail if
# one of them comes back beside its repair.
REPLACED = {
    "moral_justification_ad_populum_excuse": ("premises", "morally_acceptable_action(X,A)"),
    "sacrifice": ("premises", "great_sacrifice(S,A,X)"),
    "rhetorical_oppositions_normative": ("premises", "if_treatment_required(X,A,A2,X)"),
    "sorites_slippery_slope": ("premises", "reapplication_sequence(A2,A)"),
}


def exact_delta(before, after):
    expected = copy.deepcopy(before)
    language = expected["language"]
    require("reapplication_sequence/2" in language and "reapplication_sequence/3" not in language, "baseline is not pre-D081")
    del language["reapplication_sequence/2"]
    language["reapplication_sequence/3"] = TEXT
    record = schemes(expected)["sorites_slippery_slope"]
    old = "reapplication_sequence(A2,A)"
    require(record["premises"].count(old) == 1, "baseline does not contain the binary chain")
    record["premises"] = [CHAIN if value == old else value for value in record["premises"]]
    require(expected == after, "D081 changed executable library content beyond the scoped sorites premise and template")


def check(root, before=None):
    library = root / "docs/scheme-registry/walton-clean/walton.yml"
    after = yaml.safe_load(library.read_text())
    records = schemes(after)
    require("reapplication_sequence/2" not in after["language"], "binary chain key remains")
    require(after["language"].get("reapplication_sequence/3") == TEXT, "chain template does not name all three arguments")
    require(library.read_bytes() == (root / "wag/data/annotation/walton.yml").read_bytes(), "annotation copy is stale")
    links = json.loads((root / "docs/scheme-registry/links.json").read_text())
    for sid, (role, formula) in EXPECTED.items():
        require(formula in records[sid][role], f"{sid} is missing {formula}")
        matching = [link for entry in links["entries"] for link in entry["walton_links"] if link["id"] == sid]
        require(any(formula in set(formula_values(link)) for link in matching), f"{sid} has stale links")
        catalog = yaml.safe_load((root / f"wag/data/catalogue/schemes/{sid}.yml").read_text())
        require(formula in catalog["formulas"][role], f"{sid} has stale catalogue formulas")
        require(formula in set(formula_values(catalog["premise_map"])), f"{sid} has stale catalogue source maps")
    for sid, (role, formula) in REPLACED.items():
        require(formula not in records[sid][role], f"{sid} still has the replaced {formula}")
        catalog = yaml.safe_load((root / f"wag/data/catalogue/schemes/{sid}.yml").read_text())
        require(formula not in catalog["formulas"][role], f"{sid}'s catalogue still has the replaced {formula}")
    if before:
        exact_delta(yaml.safe_load(before.read_text()), after)
    return {"migration": "D081", "integrity": "pass", "exact_scoped_delta": "pass" if before else "not requested"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--before", type=Path)
    args = parser.parse_args()
    try:
        result = check(args.root, args.before)
    except (KeyError, TypeError, ValueError, OSError) as error:
        parser.exit(1, f"D081 migration rejected: {error}\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
