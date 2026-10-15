"""Compare what Carneades loads from the syntax-only file and from the cleaned
file, scheme by scheme, after applying the declared renames (scheme ids,
predicates, primed variables).  Whatever differs is a change of what a scheme
asserts and must be one of the recorded syntax fixes or repairs.

usage: python3 -I semantic_diff.py dump-syntax-only.json dump-clean.json
"""
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from edits import OPS, ID_MAP  # noqa: E402

PRED_MAP = {op["old"]: op["new"] for op in OPS if op["op"] == "rename"}
PRED_MAP["true"] = "is_true"
VAR_MAP = {
    "opposition-descriptive_1": {"P'": "P2"},
    "opposition-descriptive_2": {"P'": "P2"},
    "opposition-descriptive_3": {"P'": "P2"},
    "opposition-normative": {"X'": "X2"},
    "rhetorical_oppositions_normative": {"X'": "X2", "A'": "A2"},
    "rhetorical_oppositions_descriptive": {"R'": "R2"},
}


def norm(f, vmap):
    for a, b in vmap.items():
        f = f.replace(a, b)
    def pred(m):
        return PRED_MAP.get(m.group(0), m.group(0))
    return re.sub(r"[A-Za-z_][A-Za-z0-9_]*", pred, f)


def main(a, b):
    A = json.load(open(a, encoding="utf-8"))
    B = json.load(open(b, encoding="utf-8"))
    sa, sb = A["schemes"], B["schemes"]
    assert len(sa) == len(sb), (len(sa), len(sb))
    n = 0
    for x, y in zip(sa, sb):
        oid = x["id"]
        nid = ID_MAP.get(oid, oid)
        if nid != y["id"]:
            print(f"ORDER/ID MISMATCH {oid} -> {y['id']}")
            n += 1
            continue
        vm = VAR_MAP.get(oid, {})
        for field in ["variables", "premises", "assumptions", "exceptions", "conclusions"]:
            before = [norm(f, vm) for f in (x[field] or [])]
            after = list(y[field] or [])
            if before != after:
                n += 1
                print(f"{nid:40s} {field:12s} {before} -> {after}")
    print(f"{n} scheme-level difference(s) after renames")
    la, lb = A["language"], B["language"]
    ka = {norm(k.split("/")[0], {}) + ("/" + k.split("/")[1] if "/" in k else ""): k for k in la}
    added = sorted(set(lb) - set(ka))
    removed = sorted(set(ka) - set(lb))
    changed = sorted(k for k in set(ka) & set(lb) if la[ka[k]] != lb[k])
    print(f"language: {len(la)} -> {len(lb)} keys; keys only in clean: {len(added)}, only in original: {len(removed)}, same key with a changed template: {len(changed)}")
    for k in added:
        print("  + ", k, repr(lb[k]))
    for k in removed:
        print("  - ", ka[k], repr(la[ka[k]]))
    for k in changed:
        print("  ~ ", k, repr(la[ka[k]]), "->", repr(lb[k]))


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2])
