"""Inter-annotator agreement for the Kialo batch 2 double annotation.

Usage (standard library only):

    python3 -I agreement.py annotator-A.json annotator-B.json \
        [--pilot ../kialo-pilot/pilot.json] [--families ../../wag/data/schemes/families.json] \
        [--json agreement.json]

For each field the guidelines keep apart (guidelines.md, section 2), it reports raw
agreement and Cohen's kappa over the items both annotators coded: the
primary scheme, its fit, the role, the polarity, the actual target (claim and
part), whether the claim was re-parented (derived from the target, as the
guidelines say the tooling computes it), and the evidence status; and, since
the identity contract (docs/annotation-records.md §11, D092), the question
answered, the exact target (the instance it names, read in the annotator's
own file and compared by its scheme and conclusion, and the formula), the
primary instance's bindings, and every instance. A value that is unresolved,
or a target the annotator's own file does not hold, is left out of its
field and counted ("unresolved"). With --case, aliases are read as their
actions first. Both formats are read; WAG's Go measure (annotate.Agree)
computes the same figures. Where both chose the same scheme it also compares
the premises (by predicate and role) and the bindings (by the words of each
value). The scheme is also compared at the level of families
(families.json): two schemes of one family agree there. With --pilot, the
first pilot's 20 items are compared with each annotator on the scheme
(mapped to walton.yml IDs) and the role.
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


UNRESOLVED = "?"


def is_marker(v, kind):
    """{"unresolved": why} or {"unstated": why} (docs/annotation-records.md §8)."""
    return isinstance(v, dict) and len(v) == 1 and isinstance(v.get(kind), str) and v[kind].strip() != ""


def compact(f):
    return re.sub(r"\s+", "", f)


def signature(ins):
    """What an instance names: its scheme and its conclusion (as Go's signature)."""
    s = ins.get("scheme") if isinstance(ins.get("scheme"), dict) else {}
    c = ins.get("conclusion") if isinstance(ins.get("conclusion"), dict) else {}
    wid = s.get("walton_id") if isinstance(s.get("walton_id"), str) else ""
    f = c.get("formula") if isinstance(c.get("formula"), str) else ""
    return wid + ":" + compact(f)


INSTANCE_ID = re.compile(r"^(.+)#([1-9][0-9]*)$")


def target_instance(S, claim, ref):
    """The instance a target names, in the annotator's own file S: the one
    with that id (format 2), or the claim's primary (format 1, whose targets
    name none, or name it in words beginning "primary"); None if S does not
    hold it, or the target names another instance in words."""
    t = S.get(claim) if isinstance(claim, str) else None
    if t is None:
        return None
    if ref is None or isinstance(ref, str) and (ref == "" or ref.startswith("primary")):
        return primary(t)
    if not isinstance(ref, str) or not INSTANCE_ID.match(ref):
        return None
    for ins in t.get("instances") or []:
        if isinstance(ins, dict) and ins.get("id") == ref:
            return ins
    return None


def exact(item, which, S):
    """target_instance or target_formula, read in S (Go's FieldIn). The
    instance is compared as the targeting decision: the target claim's
    primary instance, or which secondary one; only what a field itself leaves
    open makes it unresolved."""
    a = (item.get("target") or {}).get("actual") if isinstance(item.get("target"), dict) else None
    if not isinstance(a, dict) or not a:
        return ""
    ref, f, claim = a.get("instance"), a.get("formula"), a.get("claim")
    if which == "target_formula":
        if is_marker(f, "unresolved"):
            return UNRESOLVED
        if isinstance(f, str) and f:
            return compact(f)
        part = a.get("part") if isinstance(a.get("part"), str) else ""
        if part == "inference":
            return ""
        if part == "conclusion" and not is_marker(ref, "unresolved"):
            ins = target_instance(S, claim, ref)
            c = (ins or {}).get("conclusion") if isinstance((ins or {}).get("conclusion"), dict) else {}
            if isinstance(c.get("formula"), str):
                return compact(c["formula"])
        return UNRESOLVED
    if is_marker(ref, "unresolved"):
        return UNRESOLVED
    ins = target_instance(S, claim, ref)
    if ins is None:
        return UNRESOLVED
    if ref is None or isinstance(ref, str) and (ref == "" or ref.startswith("primary")):
        return "primary"
    if ins.get("primary") is True:
        return "primary"
    return "secondary:" + signature(ins)


def question(item):
    """The question answered (role.cq): text, or none."""
    cq = (item.get("role") or {}).get("cq") if isinstance(item.get("role"), dict) else None
    return cq if isinstance(cq, str) else ""


def bindings(item):
    b = primary(item).get("bindings")
    if not isinstance(b, dict):
        return ""
    out = []
    for k in sorted(b):
        v = b[k]
        if is_marker(v, "unstated"):
            out.append(k + "=(unstated)")
        elif is_marker(v, "unresolved"):
            return UNRESOLVED
        else:
            out.append(k + "=" + ("" if v is None else str(v)))
    return ";".join(out)


def instances(item):
    return "|".join(sorted(signature(i) for i in item.get("instances") or [] if isinstance(i, dict)))


def canonical(S, case):
    """The annotations with each alias of the case's actions put back to its
    action, in bindings and formulas (Go's Case.CanonicalSet)."""
    canon = {}
    for a in case.get("actions") or []:
        canon[a["id"]] = a["id"]
        for al in a.get("aliases") or []:
            canon[al] = a["id"]
    if not canon:
        return S
    atom = re.compile(r"(?<![A-Za-z0-9_])[a-z_][A-Za-z0-9_]*(?![A-Za-z0-9_])(?!\s*\()")
    def walk(v, key):
        if isinstance(v, dict):
            return {k: (canon.get(y, y) if key == "bindings" and isinstance(y, str) else walk(y, k)) for k, y in v.items()}
        if isinstance(v, list):
            return [walk(y, key) for y in v]
        if isinstance(v, str) and key == "formula":
            return atom.sub(lambda m: canon.get(m.group(0), m.group(0)), v)
        return v
    return {k: walk(v, "") for k, v in S.items()}


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
    part = a.get("part")  # {"unresolved": why} is no part: target_part says so
    return str(a.get("claim") or "none"), part_kind(part if isinstance(part, str) else None)


def round3(x):
    """x to three decimals as WAG's Go measure rounds it (round3): half away
    from zero, on the same floating-point steps, so that the two agree on a
    tie such as 27/48."""
    return int(x * 1000 + (-0.5 if x < 0 else 0.5)) / 1000


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
    ap.add_argument("--case", help="case.json: read the aliases of its actions as the actions")
    args = ap.parse_args()
    A = {i["kialo_claim_id"]: i for i in json.load(open(args.a))["items"]}
    B = {i["kialo_claim_id"]: i for i in json.load(open(args.b))["items"]}
    if args.case:
        case = json.load(open(args.case))
        A, B = canonical(A, case), canonical(B, case)
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
        "target_part": lambda i: UNRESOLVED if is_marker(((i.get("target") or {}).get("actual") or {}).get("part"), "unresolved") else target(i)[1],
        "re_parented": re_parented,
        "evidence": lambda i: (i.get("evidence") or {}).get("status") or "none",
        "cq": lambda i: question(i),
        "target_instance": lambda i, S: exact(i, "target_instance", S),
        "target_formula": lambda i, S: exact(i, "target_formula", S),
        "bindings": bindings,
        "instances": instances,
    }

    def value(name, item, S):
        fn = fields[name]
        return fn(item, S) if name in ("target_instance", "target_formula") else fn(item)
    report = {"items": len(ids), "fields": {}, "disagreements": collections.defaultdict(list)}
    report["coverage"] = {"assigned":len(assigned), "basis":"provided_assignment" if args.assigned else "observed_union", "annotated_A":sum(k in A for k in assigned), "annotated_B":sum(k in B for k in assigned), "both":len(ids), "neither":sum(k not in A and k not in B for k in assigned)}
    for name in fields:
        every = [(value(name, A[k], A), value(name, B[k], B)) for k in ids]
        # An unresolved value is neither agreement nor disagreement: left out,
        # and counted (docs/annotation-records.md §11).
        keep = [(k, a, b) for k, (a, b) in zip(ids, every) if UNRESOLVED not in (a, b)]
        pairs = [(a, b) for _, a, b in keep]
        report["fields"][name] = {"agreement": round3(sum(a == b for a, b in pairs) / len(pairs)) if pairs else None,
                                  "kappa": None if kappa(pairs) is None else round3(kappa(pairs)),
                                  "unresolved": len(every) - len(keep)}
        report["fields"][name]["all_assigned_agreement"] = sum(a == b for a,b in pairs)/len(assigned) if assigned else None
        for k, a, b in keep:
            if a != b:
                report["disagreements"][name].append({"claim": k, "A": a, "B": b})

    # The scheme at the level of families.
    same_fam = 0
    for k in ids:
        a, b = scheme_id(A[k]), scheme_id(B[k])
        if a == b or (fam.get(a, set()) & fam.get(b, set())):
            same_fam += 1
    report["fields"]["scheme_family"] = {"agreement": round3(same_fam / len(ids)) if ids else None}

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
        "premise_jaccard_mean": round3(sum(prem) / len(prem)) if prem else None,
        "binding_word_jaccard_mean": round3(sum(bind) / len(bind)) if bind else None,
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
            out[who] = {"scheme_agreement": round3(sum(a == b for a, b in sp) / len(sp)) if sp else None,
                        "scheme_kappa": None if kappa(sp) is None else round3(kappa(sp)),
                        "role_agreement": round3(sum(a == b for a, b in rp) / len(rp)) if rp else None}
        out["items"] = len(pid)
        report["pilot"] = out

    text = json.dumps(report, indent=1, ensure_ascii=False, default=dict)
    if args.json:
        with open(args.json, "w") as f:
            f.write(text)
    print(text)


if __name__ == "__main__":
    sys.exit(main())
