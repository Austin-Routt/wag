"""Inter-annotator agreement for the Kialo batch 2 double annotation.

Usage (the standard library; --case also reads walton.yml, with PyYAML, as
validate.py does, its scheme ids as walton.yml has them, as the Go measure
reads them):

    python3 -I agreement.py annotator-A.json annotator-B.json \
        [--pilot ../kialo-pilot/pilot.json] [--families ../../wag/data/schemes/families.json] \
        [--json agreement.json]

For each field the guidelines keep apart (guidelines.md, section 2), it reports raw
agreement and Cohen's kappa over the items both annotators coded: the
primary scheme, its fit, the role, the polarity, the actual target (claim and
part), whether the claim was re-parented (derived from the target, as the
guidelines say the tooling computes it), and the evidence status; and, since
the identity contract (docs/annotation-records.md §11, D092), the question
answered, the exact target (its claim with the instance it names, read in
the annotator's own file and compared by its scheme and conclusion, and the
formula), the
primary instance's bindings, and every instance. A value that is unresolved,
or a target the annotator's own file does not hold, is left out of its
field and counted ("unresolved"). With --case, an alias of an action is read
as the action first, where an action stands: in the binding of a variable in
an action's place of the instance's scheme (walton.yml,
wag/data/annotation/positions.json) and in an action's place of a formula. A
goal or a consequence spelled like an alias keeps its name. Both formats are
read; WAG's Go measure (annotate.Agree) computes the same figures. Where
both chose the same scheme it also compares the premises (by predicate and
role) and the bindings (by the words of each value). The scheme is also
compared at the level of families
(families.json): two schemes of one family agree there. With --pilot, the
first pilot's 20 items are compared with each annotator on the scheme
(mapped to walton.yml IDs) and the role.
"""

import argparse
import collections
import json
import os
import re
import sys

REPO = os.path.normpath(os.path.dirname(os.path.abspath(__file__)) + "/../..") + "/"
WALTON = REPO + "docs/scheme-registry/walton-clean/walton.yml"
POSITIONS = REPO + "wag/data/annotation/positions.json"

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
    instance is compared as the targeting decision: the target claim with its
    primary instance, or with which secondary one, so that two targets of
    different claims disagree; only what a field itself leaves open makes it
    unresolved."""
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
        return claim + " primary"
    if ins.get("primary") is True:
        return claim + " primary"
    return claim + " secondary:" + signature(ins)


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


# Go's strings.TrimSpace trims what unicode.IsSpace calls space; Python's
# str.strip() also trims U+001C to U+001F, so the agreement's own reading of
# formulas trims Go's set, and reads a formula as the Go measure does.
GO_SPACE = "\t\n\v\f\r \x85\xa0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000"


def split_formula(f):
    """"¬p(a,q(b),c)" as p and its top-level arguments, the name and each
    argument trimmed (Go's splitFormula)."""
    f = f.strip(GO_SPACE)
    if f.startswith("¬"):
        f = f[1:]
    i = f.find("(")
    if i < 0 or not f.endswith(")"):
        return f, []
    args, depth, cur = [], 0, ""
    for ch in f[i + 1:-1]:
        if ch == "," and depth == 0:
            args.append(cur.strip(GO_SPACE))
            cur = ""
            continue
        if ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
        cur += ch
    if cur.strip(GO_SPACE):
        args.append(cur.strip(GO_SPACE))
    return f[:i].strip(GO_SPACE), args


def typed_vars(scheme, places):
    """A scheme's variables in an agent's or an action's place (Go's typedVars;
    validate.py's typed)."""
    out = {}
    for key in ("premises", "assumptions", "exceptions", "conclusions"):
        for f in scheme.get(key) or []:
            name, args = split_formula(f)
            for kind in ("agent", "action"):
                for i in places[kind].get(f"{name}/{len(args)}", []):
                    if i - 1 < len(args):
                        out.setdefault(args[i - 1], kind)
    return out


def case_canon(case):
    """Each action of the case record and each alias, to its action. An
    action or an alias that is an actor's id, empty, or given twice is refused,
    as Go's parseCase refuses it, so that the two measures never read one
    case two ways (validate.py checks the rest of the record)."""
    actors = {a.get("id") for a in case.get("actors") or [] if isinstance(a, dict)}
    canon = {}
    for a in case.get("actions") or []:
        x = a.get("id")
        if x in actors:
            raise SystemExit(f"case: {x!r} is an actor, so it cannot be an action")
        if not x or x in canon:
            raise SystemExit(f"case: the action {x!r} is empty or given twice")
        canon[x] = x
    for a in case.get("actions") or []:
        for al in a.get("aliases") or []:
            if al in actors:
                raise SystemExit(f"case: {al!r} is an actor, so it cannot be an alias of {a['id']}")
            if not al or al in canon:
                raise SystemExit(f"case: {al!r} is an action or an alias already, so it cannot be an alias of {a['id']}")
            canon[al] = a["id"]
    return canon


def canonical(S, case, schemes, places):
    """The annotations with each alias of the case's actions put back to its
    action where an action stands (Go's Case.CanonicalSet): in the binding of
    a variable in an action's place of the instance's scheme, and in an
    action's place of a formula: one positions.json types, or, in an
    instance's formulas, one where its scheme has an action variable, so that
    a binding and the formulas it fills are read alike; nested terms
    included. A goal or a consequence spelled like an alias keeps its name:
    read as the action, it would make two different annotations agree. A
    formula with nothing to put back is left as it is."""
    canon = case_canon(case)
    if not canon:
        return S

    def term(f, extra):
        g = f.strip(GO_SPACE)
        if g.startswith("¬"):
            y, changed = term(g[1:], extra)
            return ("¬" + y, True) if changed else (f, False)
        name, args = split_formula(g)
        if not args:
            return f, False
        key = f"{name}/{len(args)}"
        at = {i - 1 for i in places["action"].get(key, [])} | extra.get(key, set())
        changed = False
        for i, x in enumerate(args):
            if "(" in x:
                y, ch = term(x, extra)
                if ch:
                    args[i], changed = y, True
            elif i in at and canon.get(x, x) != x:
                args[i], changed = canon[x], True
        return (name + "(" + ",".join(args) + ")", True) if changed else (f, False)

    def scheme_of(ins):
        sc = ins.get("scheme")
        wid = sc.get("walton_id") if isinstance(sc, dict) else None
        return schemes.get(wid) if isinstance(wid, str) else None

    def action_places(s):
        """Go's schemeActionPlaces: the places of the scheme's formulas that
        hold one of its action variables, by predicate and arity."""
        acts = {v for v, kind in typed_vars(s, places).items() if kind == "action"}
        out = {}
        for key in ("premises", "assumptions", "exceptions", "conclusions"):
            for f in s.get(key) or []:
                name, args = split_formula(f)
                for i, a in enumerate(args):
                    if a in acts:
                        out.setdefault(f"{name}/{len(args)}", set()).add(i)
        return out

    def walk(v, key, extra):
        if isinstance(v, dict):
            acts = set()
            if "bindings" in v:
                s = scheme_of(v)
                acts = {x for x, kind in typed_vars(s, places).items() if kind == "action"} if s is not None else set()
                extra = action_places(s) if s is not None else {}
            out = {}
            for k, y in v.items():
                if k == "bindings" and isinstance(y, dict):
                    out[k] = {x: (canon.get(b, b) if isinstance(b, str) and x in acts else b) for x, b in y.items()}
                else:
                    out[k] = walk(y, k, extra)
            return out
        if isinstance(v, list):
            return [walk(y, key, extra) for y in v]
        if isinstance(v, str) and key == "formula":
            return term(v, extra)[0]
        return v
    return {k: walk(v, "", {}) for k, v in S.items()}


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


# The fields compared, each a function of an item (target_instance and target_formula also of the
# annotator's own items, S, in which a target's instance is read).
FIELDS = {
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

# D102: an item whose target formula or part is unresolved has no settled relation to its target, so
# its role, its question and its target's part, instance and formula are not compared: each is
# unresolved, left out and counted. Its scheme, fit, polarity and target claim are compared as before.
# An unresolved instance alone (a target claim not yet annotated) leaves only target_instance out.
RELATION_FIELDS = ("role", "cq", "target_part", "target_instance", "target_formula")


def relation_unsettled(item):
    """Whether the item's target leaves its relation unresolved (D102): its formula or its part is a
    marker {"unresolved": why}."""
    actual = (item.get("target") or {}).get("actual") or {}
    return isinstance(actual, dict) and (is_marker(actual.get("formula"), "unresolved")
                                         or is_marker(actual.get("part"), "unresolved"))


def field_value(name, item, S):
    """The item's value for a field, UNRESOLVED where it is not compared."""
    if name in RELATION_FIELDS and relation_unsettled(item):
        return UNRESOLVED
    fn = FIELDS[name]
    return fn(item, S) if name in ("target_instance", "target_formula") else fn(item)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("a")
    ap.add_argument("b")
    ap.add_argument("--pilot")
    ap.add_argument("--families")
    ap.add_argument("--json")
    ap.add_argument("--assigned", help="items.json: all assigned claims, including missing replies")
    ap.add_argument("--case", help="case.json: read the aliases of its actions as the actions, where an action stands")
    args = ap.parse_args()
    A = {i["kialo_claim_id"]: i for i in json.load(open(args.a))["items"]}
    B = {i["kialo_claim_id"]: i for i in json.load(open(args.b))["items"]}
    if args.case:
        import yaml
        case = json.load(open(args.case))
        schemes = {s["id"]: s for s in yaml.safe_load(open(WALTON, encoding="utf-8"))["argument_schemes"]}
        places = json.load(open(POSITIONS, encoding="utf-8"))
        A, B = canonical(A, case, schemes, places), canonical(B, case, schemes, places)
    assigned = [i["kialo_claim_id"] for i in json.load(open(args.assigned))["items"]] if args.assigned else list(dict.fromkeys([*A,*B]))
    if len(set(assigned)) != len(assigned):
        ap.error("assignment list contains duplicate claim IDs")
    ids = [k for k in assigned if k in A and k in B]
    fam = family_of(json.load(open(args.families))) if args.families else {}

    fields = FIELDS
    value = field_value
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
