#!/usr/bin/env python3
"""Check D084's formulas, synchronized copies and optional rename-only delta.

--before is the cleaned collection immediately before D084. It verifies the
whole parsed YAML change, so use that mode before any later semantic migration.
Without --before, the integrity checks remain valid after later migrations.
This tool never rewrites historical graph/run files or their evidence text.
"""
import argparse
import copy
import json
from pathlib import Path
import re

import yaml


TARGETS = {
    "group_2": ("conclusions", "is(M,Q)", "is_a(M,Q)"),
    "rule_physical_world_1": ("premises", "is(W,D)", "is_a(W,D)"),
}
ROLES = ("premises", "assumptions", "exceptions", "conclusions", "deletions", "guards")


def schemes(document):
    records = document["argument_schemes"]
    result = {record["id"]: record for record in records}
    if len(records) != len(result):
        raise ValueError("duplicate scheme IDs")
    return result


def require(value, message):
    if not value:
        raise ValueError(message)


def exact_delta(before, after):
    expected = copy.deepcopy(before)
    language = expected["language"]
    require("is/2" in language and "is_a/2" not in language, "baseline is not pre-D084")
    language["is_a/2"] = language.pop("is/2")
    records = schemes(expected)
    for sid, (role, old, new) in TARGETS.items():
        values = records[sid][role]
        require(values.count(old) == 1, f"baseline {sid} does not contain exactly one {old}")
        records[sid][role] = [new if value == old else value for value in values]
    require(expected == after, "the parsed collection changes more than the two formulas and language-key rename")


def formula_values(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == "formula":
                yield child
            else:
                yield from formula_values(child)
    elif isinstance(value, list):
        for child in value:
            yield from formula_values(child)


def check(root, before=None):
    library = root / "docs/scheme-registry/walton-clean/walton.yml"
    after = yaml.safe_load(library.read_text())
    records = schemes(after)
    require("is/2" not in after["language"], "reserved language key is/2 remains")
    require(after["language"].get("is_a/2") == "%s is %s.", "D084 changed its English template")
    for sid, record in records.items():
        for role in ROLES:
            for formula in record.get(role) or []:
                require(not re.search(r"(?<![A-Za-z0-9_])is\s*\(", formula), f"reserved predicate remains in {sid}/{role}")
    links = json.loads((root / "docs/scheme-registry/links.json").read_text())
    for sid, (role, old, new) in TARGETS.items():
        require(new in records[sid][role], f"{sid} is missing {new}")
        matching = [link for entry in links["entries"] for link in entry["walton_links"] if link["id"] == sid]
        require(any(new in set(formula_values(link)) for link in matching), f"links for {sid} do not contain {new}")
        require(all(old not in set(formula_values(link)) for link in matching), f"links for {sid} retain {old}")
        catalog = yaml.safe_load((root / f"wag/data/catalogue/schemes/{sid}.yml").read_text())
        require(new in catalog["formulas"][role], f"catalogue formula for {sid} is stale")
        require(new in set(formula_values(catalog["premise_map"])), f"catalogue premise map for {sid} is stale")
    require(library.read_bytes() == (root / "wag/data/annotation/walton.yml").read_bytes(), "annotation copy is not byte-identical")
    if before:
        exact_delta(yaml.safe_load(before.read_text()), after)
    return {"migration": "D084", "integrity": "pass", "exact_rename_only_delta": "pass" if before else "not requested", "schemes": len(records)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--before", type=Path)
    args = parser.parse_args()
    try:
        result = check(args.root, args.before)
    except (KeyError, TypeError, ValueError, OSError) as error:
        parser.exit(1, f"D084 migration rejected: {error}\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
