"""Inter-annotator agreement for the Kialo batch 2 double annotation.

Usage (standard library only):

    python3 -I agreement.py annotator-A.json annotator-B.json \
        [--pilot ../kialo-pilot/pilot.json] [--families ../../wag/data/schemes/families.json] \
        [--json agreement.json]

For each field the guidelines keep apart (guidelines.md, section 2), it reports raw
agreement and Cohen's kappa over the items both annotators coded: the
primary scheme, its fit, the role, the polarity, the actual target (claim and
part), whether the claim was re-parented (derived from the target, as the
guidelines say the tooling computes it), and the evidence status. Where
both chose the same scheme it also compares the premises (by predicate and
role) and the bindings (by the words of each value). The scheme is also
compared at the level of families (families.json): two schemes of one family
agree there. With --pilot, the first pilot's 20 items are compared with each
annotator on the scheme (mapped to walton.yml IDs) and the role.
"""

import argparse
import collections
import json
import re
import sys

WAG_TO_WALTON = {
    "AS:practical_reasoning": "practical_reasoning1",
    "AS:practical_reasoning_goal": "practical_reasoning2",
    "AS:consequences": "positive_consequences",
    "AS:negative_consequences": "negative_consequences",
    "AS:values": "positive_value",
    "AS:sign": "sign",
    "AS:expert_opinion": "credible_source",
    "AS:position_to_know": "position_to_know",
    "AS:analogy": "analogy",
    "AS:other": "other",
}


def kappa(pairs):
    """Cohen's kappa for a list of (a, b) labels; None if undefined."""
    n = len(pairs)
    if n == 0:
        return None
    po = sum(a == b for a, b in pairs) / n
    ca = collections.Counter(a for a, _ in pairs)
    cb = collections.Counter(b for _, b in pairs)
    pe = sum(ca[k] * cb.get(k, 0) for k in ca) / (n * n)
    if pe == 1:
        return None  # chance agreement is 1: kappa is undefined
    return (po - pe) / (1 - pe)


def primary(item):
    inst = item.get("instances") or []
    for i in inst:
        if i.get("primary"):
            return i
    return inst[0] if inst else {}


def scheme_id(item):
    s = (primary(item).get("scheme") or {})
    sid = (s.get("walton_id") or "none").strip()
    return sid if sid else "none"


def fit(item):
    return ((primary(item).get("scheme") or {}).get("fit") or "none")


def part_kind(part):
    p = (part or "").lower()
    for k in ("conclusion", "premise", "exception", "inference", "argument"):
        if p.startswith(k):
            return "inference" if k == "argument" else k
    return p or "none"


def target(item):
    a = (item.get("target") or {}).get("actual") or {}
    return str(a.get("claim") or "none"), part_kind(a.get("part"))


def text(v):
    """A claim ID as WAG's Go measure reads it: text, or else none."""
    return v if isinstance(v, str) else ""


def re_parented(item):
    """Whether the claim is re-parented: its actual target claim is not its
    Kialo parent (guidelines §5, item 1). Derived, as the guidelines say the
    tooling computes it: the item's own re_parented is not read. WAG's Go
    measure (annotate.FieldOf) derives it alike."""
    t = item.get("target")
    a = t.get("actual") if isinstance(t, dict) else None
    if not isinstance(a, dict) or not a:
        return False
    return text(a.get("claim")) != text(t.get("kialo_parent"))


def predicates(item):
    out = collections.Counter()
    for p in primary(item).get("premises") or []:
        f = p.get("formula") or ""
        m = re.match(r"\s*(¬?)\s*([a-z_][A-Za-z0-9_]*)\s*\((.*)\)\s*$", f)
        if m:
            arity = len(split_args(m.group(3)))
            out[(p.get("role") or "premise", m.group(1) + m.group(2) + "/" + str(arity))] += 1
        else:
            out[(p.get("role") or "premise", f.strip())] += 1
    return out


def split_args(s):
    args, depth, cur = [], 0, ""
    for ch in s:
        if ch == "," and depth == 0:
            args.append(cur)
            cur = ""
            continue
        depth += ch == "("
        depth -= ch == ")"
        cur += ch
    if cur.strip():
        args.append(cur)
    return args


def words(v):
    return set(re.findall(r"[a-z0-9]+", str(v).lower())) - {"the", "a", "an", "of", "to", "in", "and", "is", "be", "would"}


def jaccard(a, b):
    if not a and not b:
        return 1.0
    inter = sum((a & b).values()) if isinstance(a, collections.Counter) else len(a & b)
    union = sum((a | b).values()) if isinstance(a, collections.Counter) else len(a | b)
    return inter / union if union else 1.0


def family_of(families):
    fam = collections.defaultdict(set)
    for f in families.get("families", []):
        for m in f.get("members", []):
            if m.get("scheme"):
                fam[m["scheme"]].add(f["id"])
    return fam


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("a")
    ap.add_argument("b")
    ap.add_argument("--pilot")
    ap.add_argument("--families")
    ap.add_argument("--json")
    ap.add_argument("--assigned", help="items.json: all assigned claims, including missing replies")
    args = ap.parse_args()
    A = {i["kialo_claim_id"]: i for i in json.load(open(args.a))["items"]}
    B = {i["kialo_claim_id"]: i for i in json.load(open(args.b))["items"]}
    assigned = [i["kialo_claim_id"] for i in json.load(open(args.assigned))["items"]] if args.assigned else list(dict.fromkeys([*A,*B]))
    if len(set(assigned)) != len(assigned):
        ap.error("assignment list contains duplicate claim IDs")
    ids = [k for k in assigned if k in A and k in B]
    fam = family_of(json.load(open(args.families))) if args.families else {}

    fields = {
        "scheme": scheme_id,
        "fit": fit,
        "role": lambda i: (i.get("role") or {}).get("kind") or "none",
        "polarity": lambda i: (i.get("role") or {}).get("polarity") or "none",
        "target_claim": lambda i: target(i)[0],
        "target_part": lambda i: target(i)[1],
        "re_parented": re_parented,
        "evidence": lambda i: (i.get("evidence") or {}).get("status") or "none",
    }
    report = {"items": len(ids), "fields": {}, "disagreements": collections.defaultdict(list)}
    report["coverage"] = {"assigned":len(assigned), "basis":"provided_assignment" if args.assigned else "observed_union", "annotated_A":sum(k in A for k in assigned), "annotated_B":sum(k in B for k in assigned), "both":len(ids), "neither":sum(k not in A and k not in B for k in assigned)}
    for name, fn in fields.items():
        pairs = [(fn(A[k]), fn(B[k])) for k in ids]
        report["fields"][name] = {"agreement": round(sum(a == b for a, b in pairs) / len(pairs), 3) if pairs else None,
                                  "kappa": None if kappa(pairs) is None else round(kappa(pairs), 3)}
        report["fields"][name]["all_assigned_agreement"] = sum(a == b for a,b in pairs)/len(assigned) if assigned else None
        for k, (a, b) in zip(ids, pairs):
            if a != b:
                report["disagreements"][name].append({"claim": k, "A": a, "B": b})

    # The scheme at the level of families.
    same_fam = 0
    for k in ids:
        a, b = scheme_id(A[k]), scheme_id(B[k])
        if a == b or (fam.get(a, set()) & fam.get(b, set())):
            same_fam += 1
    report["fields"]["scheme_family"] = {"agreement": round(same_fam / len(ids), 3) if ids else None}

    # Premises and bindings where the scheme agrees.
    prem, bind, sup = [], [], []
    for k in ids:
        if scheme_id(A[k]) != scheme_id(B[k]) or scheme_id(A[k]) in ("none", "other"):
            continue
        prem.append(jaccard(predicates(A[k]), predicates(B[k])))
        ba, bb = primary(A[k]).get("bindings") or {}, primary(B[k]).get("bindings") or {}
        for v in set(ba) & set(bb):
            bind.append(jaccard(words(ba[v]), words(bb[v])))
        sa = collections.Counter(p.get("status") for p in primary(A[k]).get("premises") or [])
        sb = collections.Counter(p.get("status") for p in primary(B[k]).get("premises") or [])
        sup.append((sa.get("supplied", 0), sb.get("supplied", 0)))
    report["same_scheme"] = {
        "items": len(prem),
        "premise_jaccard_mean": round(sum(prem) / len(prem), 3) if prem else None,
        "binding_word_jaccard_mean": round(sum(bind) / len(bind), 3) if bind else None,
        "supplied_premises_A_B": [sum(a for a, _ in sup), sum(b for _, b in sup)],
    }
    report["scheme_counts"] = {"A": collections.Counter(scheme_id(A[k]) for k in ids),
                               "B": collections.Counter(scheme_id(B[k]) for k in ids)}
    report["role_counts"] = {"A": collections.Counter(fields["role"](A[k]) for k in ids),
                             "B": collections.Counter(fields["role"](B[k]) for k in ids)}

    if args.pilot:
        P = {i["kialo_claim_id"]: i for i in json.load(open(args.pilot))["items"]}
        pid = [k for k in ids if k in P]

        def pilot_scheme(i):
            s = i["scheme"]["id"]
            for wag, w in WAG_TO_WALTON.items():
                if s.startswith(wag):
                    return w
            m = re.match(r"walton\.yml (\w+)", s)
            if m:
                return m.group(1)
            return "new" if s.startswith("NEW") else "none"

        role_map = {"new_pro": "independent_pro"}
        out = {}
        for who, X in (("A", A), ("B", B)):
            sp = [(pilot_scheme(P[k]), scheme_id(X[k])) for k in pid]
            rp = [(role_map.get(P[k]["role"]["kind"], P[k]["role"]["kind"]), fields["role"](X[k])) for k in pid]
            out[who] = {"scheme_agreement": round(sum(a == b for a, b in sp) / len(sp), 3) if sp else None,
                        "scheme_kappa": None if kappa(sp) is None else round(kappa(sp), 3),
                        "role_agreement": round(sum(a == b for a, b in rp) / len(rp), 3) if rp else None}
        out["items"] = len(pid)
        report["pilot"] = out

    text = json.dumps(report, indent=1, ensure_ascii=False, default=dict)
    if args.json:
        with open(args.json, "w") as f:
            f.write(text)
    print(text)


if __name__ == "__main__":
    sys.exit(main())
