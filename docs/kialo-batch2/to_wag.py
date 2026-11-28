"""Play annotated Kialo claims as a WAG game, through a running wag-server.

Usage (standard library only):

    python3 -I to_wag.py ANNOTATIONS.json items.json http://127.0.0.1:8080 \
        [--report report.json] [--export game.json]

The thesis is the claim of a debate (as the formula the annotations give its
conclusion, e.g. should_be_performed(tax_meat)). Then each annotated claim,
top-down, is played as the move its annotation describes, by Pro or Con:

  - an argument for or against a statement (independent pro/con, rebuttal,
    premise backing or attack, side effect): a move about that statement, with
    the annotated scheme if the game has it, else with Other;
  - an answer to one of the game's critical questions about an argument that
    was played: that answer, with Other when the annotation codes it so;
  - a claim whose target is the inference of an argument that was played
    (OC18): an answer to that argument's question there, with Other when
    the annotation codes it so or the game lacks its scheme; or, for an
    undercutter, an undercut of that argument (with Other when the game
    lacks its scheme). No other role is played there;
  - a claim whose primary instance is carried by a secondary one (a bridge,
    whose entry premise its modal argument concludes): the primary, then the
    secondary for that premise, with its scheme; claims at the secondary
    instance then play against it;
  - a fragment, a clarification, or a claim with no scheme: not played.

A move the game rejects is retried with Other (the claim's literal text as
its premise), and the report says why. When the same side moves twice, the
other passes. The game's export (--export) can be loaded in the browser
("Load an exported game") to look at the annotated discussion.

Each claim's entry in the report keeps what the annotation intended
(`intended`: scheme, role, target, question) apart from what was played
(`executed`), and says how they relate (`integration`):

  - as_intended: played with its scheme, in the role the annotation gives,
    its argument concluding the annotation's own conclusion (or, for a bare
    denial, as the challenge the guidelines prescribe; for an inference
    target, as an undercutter of its argument);
  - conclusion_changed: played as intended in scheme and role, but its
    argument concludes another formula than the annotation's conclusion: the
    game took the move with another proposition (cases T6, T10, T13, T16). The
    engine binds an answer variable to the statement answered (expert
    opinion's S, carneades/answers.go), so an expert answering through a
    question, or arguing for the statement, is credited with that statement,
    whatever the claim itself concludes (13292.224 and .233). The row gives
    both formulas (`conclusion_change`);
  - coverage_gap: its scheme is not one the game has, so it was played with
    Other: the game lacks the scheme, the annotation is not at fault;
  - fragment: the claim states a premise of its target's argument (fit
    "fragment": the scheme is the target's, not the claim's), so it is
    played as the claim's own support of that premise, with Other;
  - failed: the game has its scheme, but rejected the intended move, and the
    claim was played with Other or not at all: a failed test of the scheme,
    of its questions or of the annotation, even when the Other move is legal;
    or a claim played in its role whose annotated conclusion, or its
    argument's, is missing: there is nothing to compare (cases T7, T8, T12);
    or a claim coded Other that lost its named question's route;
  - duplicate: a challenge of a premise another claim already challenged;
  - cascade: its target was itself not played as intended, so the premise
    or statement it is about is not on the board: played with Other, or, for
    a claim at its conclusion, labelled conclusion_changed, the move then
    made about the target's annotated conclusion and refused (case T22); or
    its target is a secondary instance of its claim (format 2 names it) that
    was not played: only each claim's primary instance is played, and a
    secondary one that carries it. A row labelled conclusion_changed is never
    a cascade: its move is the first the replay tries, so nothing was
    refused before it (case T24);
  - not_played: an unresolved target, or a malformed one (a target instance,
    part or formula that is neither text, null nor {"unresolved": why}, which
    validate.py refuses); a clarification, a claim with no scheme and no
    target, or no side; or a claim whose scheme the game lacks, whose move with
    Other the game refused.

Each row whose move made an argument also says whether that argument
concludes the annotation's own conclusion (`conclusion_preserved`: true or
false, null when either formula is missing), whatever its label: a coverage
gap or a fragment played with Other at its target may conclude the target, not
the claim (case T17). The formulas are compared without Go's whitespace and
with not(...) as ¬ (T9, T21). Not compared: a challenge, which makes no
argument; a move at an inference, an undercutter or an answer through a
question whose locus is the argument, which concludes undercut(<its target>)
(T19); and a new Other argument, which concludes the statement the game makes
from the claim's English (null, T20).
"""

import argparse
import json
import os
import re
import sys
import urllib.error
import urllib.request

# The game's scheme for each walton.yml scheme it plays (its Carneades id), read
# from the server's data (/api/meta) when the game is made, so that a replay on
# other data (wag-server --data on a frozen folder) plays that data's schemes.
# credible_source, which the game does not play, is played as expert opinion.
WALTON_TO_WAG = {}


def game_schemes(meta):
    out = {"credible_source": "AS:expert_opinion"}
    for s in (meta or {}).get("schemes") or []:
        if s.get("carneades_id") and not s.get("optional"):
            out[s["carneades_id"]] = s["id"]
    return out


# How a better means does better, as the gold records it and as the game
# takes it (D089).
COMPARISON = {"more surely": "surely", "sooner": "sooner", "at lower cost": "cheaper"}
FOR_ROLES = {"independent_pro", "premise_backing"}
AGAINST_ROLES = {"independent_con", "rebuttal", "premise_attack", "side_effect", "undercutter", "alternative_means"}


def call(base, path, body=None):
    req = urllib.request.Request(base + path, method="POST" if body is not None else "GET",
                                 data=None if body is None else json.dumps(body).encode(),
                                 headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req) as r:
            data = r.read()
    except urllib.error.HTTPError as e:
        return {"result": {"error": f"HTTP {e.code}: {e.read().decode(errors='replace').strip()}"}}
    try:
        return json.loads(data)
    except ValueError:
        return data.decode()


def primary(item):
    for i in item.get("instances") or []:
        if i.get("primary"):
            return i
    inst = item.get("instances") or []
    return inst[0] if inst else {}


def atom(text):
    t = re.sub(r"[^0-9a-z]+", "_", str(text).lower()).strip("_")
    return ("x_" + t) if t[:1].isdigit() else (t or "x_empty")


def words(value):
    """A binding as the player would type it: atoms back to words."""
    return str(value).replace("_", " ").strip()


def side_of(item, kialo_relation):
    """The side that makes the move, and whether it argues for or against
    its target: from the annotated polarity or role, else the Kialo
    relation (pro, con)."""
    role = (item.get("role") or {})
    pol = role.get("polarity")
    kind = role.get("kind")
    if pol not in ("for", "against"):
        if kind in FOR_ROLES:
            pol = "for"
        elif kind in AGAINST_ROLES:
            pol = "against"
        elif kialo_relation in ("pro", "con"):
            pol = "for" if kialo_relation == "pro" else "against"
        else:
            return None, None
    side = "pro" if pol == "for" else "con"
    if kialo_relation in ("pro", "con") and kind in ("action_amendment", "alternative_means", "none", "clarification"):
        side = kialo_relation  # a proposal of its own, by the side that filed it
    return side, pol


# The characters Go's strings.TrimSpace removes (Python's str.strip also removes U+001C-U+001F):
# a marker's reason is blank exactly when Go finds it blank (case A5), as in agreement.py.
GO_SPACE = "\t\n\v\f\r \x85\xa0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000"


def norm(formula):
    """A formula as the replay compares it: without the whitespace Go's strings.Fields removes
    (GO_SPACE; case T21: a tab, U+00A0 or U+3000 is no difference, as validate.py and agreement.py
    compare formulas), and not(...) written as ¬."""
    f = re.sub("[" + re.escape(GO_SPACE) + "]+", "", formula or "")
    if f.startswith("not(") and f.endswith(")"):
        f = "¬" + f[4:-1]
    return f


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("annotations")
    ap.add_argument("items")
    ap.add_argument("base")
    ap.add_argument("--report")
    ap.add_argument("--export")
    ap.add_argument("--case", help="the case record (wag-case/1) whose profile the game has: its agent, actors and "
                    "actions (D094 (b)); default: case.json beside the items file")
    args = ap.parse_args()
    ann = {i["kialo_claim_id"]: i for i in json.load(open(args.annotations))["items"]}
    items = json.load(open(args.items))
    base = args.base.rstrip("/")
    thesis_id = items["discussion"]["thesis_claim_id"]
    th = ann.get(thesis_id, {})
    claim = (primary(th).get("conclusion") or {}).get("formula") or "should_be_performed(tax_meat)"
    claim = claim.replace(" ", "")
    # The game's profile is the case's (D094 (b)): its agent, and the actions
    # the discussion deliberates, each with its actor and aliases.
    case_path = args.case or os.path.join(os.path.dirname(os.path.abspath(args.items)), "case.json")
    case = json.load(open(case_path, encoding="utf-8"))
    st = call(base, "/api/games", {"title": "Kialo 13292, annotated (batch 2)", "mode": "debate", "allow_other": True,
                                   "seats": [{"name": "Pro", "kind": "human"}, {"name": "Con", "kind": "human"}],
                                   "claim": claim, "concede_on_pass": False, "max_rounds": 0, "case": case})
    gid = st["id"]
    WALTON_TO_WAG.update(game_schemes(call(base, "/api/meta")))
    played = {}  # claim id -> {"arg": aN, "conclusion": stmt id}
    played_inst = {}  # instance id -> the same: each claim's primary instance, and a secondary one that carries it
    report = []
    discussion = []  # D096: the instances by the discussion, played apart from the claims' moves
    backs = {}  # a claim played as a fragment -> {"formula": the premise it backs, "trec": the argument that has it}

    def state():
        return call(base, f"/api/games/{gid}")

    def move(side, action, why):
        s = state()
        me = "Pro" if side == "pro" else "Con"
        if s["active"] != me:
            call(base, f"/api/games/{gid}/move", {"action": {"type": "abstain"}, "why": "keeps the Kialo order"})
        return call(base, f"/api/games/{gid}/move", {"action": action, "why": why})

    def play_carried(cid, a, inst, side):
        """A secondary instance whose conclusion is a premise of the claim's
        primary instance carries it (a bridge's modal argument, whose
        conclusion is the bridge's entry premise): it is played for that
        premise, with its scheme, when the game has it. What each did."""
        out = []
        prem = {norm(p.get("formula")) for p in inst.get("premises") or [] if p.get("role") in ("premise", "assumption")}
        for sec in a.get("instances") or []:
            if sec is inst or sec.get("primary") or sec.get("by") == "the discussion":
                continue
            sw = (sec.get("scheme") or {}).get("walton_id")
            sconc = norm((sec.get("conclusion") or {}).get("formula"))
            wag2 = WALTON_TO_WAG.get(sw)
            if not wag2 or not sconc or sconc not in prem:
                continue
            act = {"type": "support", "target_stmt_id": sconc, "scheme_id": wag2,
                   "variable_bindings": {k: words(v) for k, v in (sec.get("bindings") or {}).items()},
                   "text": (sec.get("conclusion") or {}).get("en") or sconc}
            why = f"Kialo {cid}: its instance {sec.get('id')}, {sw}, for the premise it carries"
            r = move(side, act, why)
            res = r.get("result") or {}
            fixed = re.findall(r"leave (\w+) out", res.get("error") or "")
            if not res.get("accepted") and fixed:
                act = dict(act, variable_bindings={k: v for k, v in act["variable_bindings"].items() if k not in fixed})
                r = move(side, act, why)
                res = r.get("result") or {}
            rec = {"instance": sec.get("id"), "scheme": sw, "wag_id": wag2, "premise": sconc}
            if res.get("accepted"):
                arg = (res.get("arg_ids") or [None])[0]
                s = state()
                played_inst[sec.get("id")] = {"arg": arg, "conclusion": s["arguments"][arg]["conclusion"] if arg in s["arguments"] else None}
                rec.update(played=True, arg=arg)
            else:
                rec.update(played=False, error=res.get("error"))
            out.append(rec)
        return out

    def play_discussion(owner, a, side):
        """D096 (annotation-records.md §13): an instance by the discussion is the argument the
        discussion needs for the claim it is recorded at. It is played as an argument for its
        conclusion, that claim's statement (its primary instance's conclusion) or its negation, with
        its scheme, by the claim's side, or the other side when it argues against the statement. The
        game must have its scheme: it is not played with Other, which would leave its premises off the
        board. It is recorded as a played instance, so that claims at it are played there."""
        stmt = norm((primary(a).get("conclusion") or {}).get("formula"))
        for ins in a.get("instances") or []:
            if ins.get("by") != "the discussion":
                continue
            sw = (ins.get("scheme") or {}).get("walton_id")
            wag2 = WALTON_TO_WAG.get(sw)
            conc = norm((ins.get("conclusion") or {}).get("formula"))
            rec = {"claim": owner, "instance": ins.get("id"), "scheme": sw, "wag_id": wag2, "conclusion": conc}
            # The move argues the instance's own conclusion (RC16): the claim's statement, supported by the
            # claim's side; its negation, by the other side, as an attack on the statement, or, when the
            # statement is itself a negation, as support for what it negates.
            if conc == stmt:
                kind, target, against = "support", stmt, False
            elif conc == "¬" + stmt:
                kind, target, against = "attack", stmt, True
            elif stmt.startswith("¬") and conc == stmt[1:]:
                kind, target, against = "support", conc, True
            else:
                kind, target, against = None, None, None
            if not wag2:
                rec.update(played=False, error=f"the game has no scheme for {sw}")
            elif not stmt:
                rec.update(played=False, error=f"{owner} has no statement for it to argue")
            elif kind is None:
                rec.update(played=False, error=f"it concludes {conc or 'nothing'}, neither {owner}'s statement ({stmt}) "
                                               "nor its negation (R20)")
            else:
                who = side if not against else ("con" if side == "pro" else "pro")
                act = {"type": kind, "target_stmt_id": target, "scheme_id": wag2,
                       "variable_bindings": {k: words(v) for k, v in (ins.get("bindings") or {}).items()},
                       "text": (ins.get("conclusion") or {}).get("en") or conc}
                why = f"Kialo {owner}: the discussion's instance {ins.get('id')}, {sw} (D096)"
                r = move(who, act, why)
                res = r.get("result") or {}
                fixed = re.findall(r"leave (\w+) out", res.get("error") or "")
                if not res.get("accepted") and fixed:
                    act = dict(act, variable_bindings={k: v for k, v in act["variable_bindings"].items() if k not in fixed})
                    r = move(who, act, why)
                    res = r.get("result") or {}
                if res.get("accepted"):
                    arg = (res.get("arg_ids") or [None])[0]
                    st = state()
                    played_inst[ins.get("id")] = {"arg": arg,
                                                  "conclusion": st["arguments"][arg]["conclusion"] if arg in st["arguments"] else None}
                    rec.update(played=True, arg=arg, side=who, target_stmt=target)
                else:
                    rec.update(played=False, error=res.get("error"))
            discussion.append(rec)

    def played_with_other(claim_id):
        """Whether the claim was played with Other where its annotation
        meant a scheme: its premises are not on the board."""
        for e in report:
            if e["claim"] == claim_id:
                return e.get("played") in ("other", "other-new", "inference-other") and e.get("integration") != "fragment"
        return False

    def conclusion_changed(claim_id):
        """The claim's row if its argument concludes another formula than its
        annotated one (case T22): its annotated conclusion is not that
        argument's, so a claim at its conclusion is about a statement the
        argument does not put on the board."""
        for e in report:
            if e["claim"] == claim_id:
                return e if e.get("integration") == "conclusion_changed" else None
        return None

    # The thesis's instances by the discussion go first: claims answer them (D096).
    play_discussion(thesis_id, th, "pro")
    # A claim that answers a sibling played later waits for it (one more pass).
    order = [it for it in items["items"] if it["kialo_claim_id"] != thesis_id and it["kialo_claim_id"] in ann]
    later = {}
    for n, it in enumerate(order):
        later[it["kialo_claim_id"]] = n
    queue, deferred = [], []
    for n, it in enumerate(order):
        t = str((((ann[it["kialo_claim_id"]].get("target") or {}).get("actual")) or {}).get("claim") or "")
        (deferred if later.get(t, -1) > n else queue).append(it)
    for it in queue + deferred:
        cid = it["kialo_claim_id"]
        a = ann.get(cid)
        inst = primary(a)
        scheme = (inst.get("scheme") or {})
        wid = scheme.get("walton_id") or "none"
        rel = (it.get("kialo_parents") or [{}])[0].get("relation")
        side, pol = side_of(a, rel)
        entry = {"claim": cid, "scheme": wid, "role": (a.get("role") or {}).get("kind")}
        tgt0 = ((a.get("target") or {}).get("actual") or {})
        # Format 2 (docs/annotation-records.md) names the target's instance;
        # format 1's target is its claim's primary instance.
        tinst = tgt0.get("instance")
        entry["intended"] = {"walton_id": wid, "wag_id": WALTON_TO_WAG.get(wid), "fit": scheme.get("fit"),
                             "role": (a.get("role") or {}).get("kind"), "cq": (a.get("role") or {}).get("cq"),
                             "polarity": pol, "target_claim": tgt0.get("claim"),
                             "target_instance": tinst if isinstance(tinst, str) else ("unresolved" if tinst else None),
                             "target_part": tgt0.get("part"), "target_formula": tgt0.get("formula"),
                             "conclusion": (inst.get("conclusion") or {}).get("formula")}
        # Format 2 keeps unresolved targets as development material. They
        # authorize no executable target: do not normalize a marker as a
        # string, fall back to the parent's primary instance, or create a
        # different new argument. Preserve the annotation in the report.
        unresolved, malformed = {}, []
        for field in ("instance", "part", "formula"):
            value = tgt0.get(field)
            if value is None or isinstance(value, str):
                continue
            if (isinstance(value, dict) and set(value) == {"unresolved"}
                    and isinstance(value["unresolved"], str) and value["unresolved"].strip(GO_SPACE)):
                unresolved[field] = value["unresolved"]
            else:
                malformed.append(field)
        if malformed:
            # A value the format does not allow (validate.py refuses it, R16):
            # no move is made from it, and the replay goes on (case U7).
            entry["played"] = "no: malformed target: " + ", ".join(malformed)
            entry["executed"], entry["integration"] = None, "not_played"
            entry["malformed_target"] = malformed
            report.append(entry)
            continue
        if unresolved:
            entry["played"] = "no: unresolved target: " + ", ".join(unresolved)
            entry["executed"], entry["integration"] = None, "not_played"
            entry["unresolved_target"] = unresolved
            report.append(entry)
            continue
        tprimary = primary(ann.get(str(tgt0.get("claim")), {})).get("id")
        secondary_target = isinstance(tinst, str) and tprimary and tinst != tprimary
        if secondary_target and tinst not in played_inst:
            # A secondary instance is played only when it carries its claim's
            # primary one (play_carried): another one's premises are not on
            # the board.
            entry["played"] = f"no: its target, the instance {tinst}, is a secondary instance of {tgt0.get('claim')}, which is not played"
            entry["executed"], entry["integration"] = None, "cascade"
            report.append(entry)
            continue
        # The target instance on the board: its argument and its conclusion.
        trec = played_inst.get(tinst) if secondary_target else played.get(str(tgt0.get("claim") or ""))
        # A claim played as a fragment supports the premise it backs: on the
        # board that formula is a premise of the argument it backs, and the
        # conclusion of the fragment's own Other argument. A move at that
        # premise of the fragment is made at the argument it backs (RC14); the
        # fragment's other premises are not on the board (RC15).
        back = backs.get(str(tgt0.get("claim") or ""))
        if back and str(tgt0.get("part", "")).startswith("premise") and norm(tgt0.get("formula")) == back["formula"]:
            trec = back["trec"]
        if wid == "none" and side == "con" and str(tgt0.get("part", "")).startswith("premise") and trec:
            # A bare denial of a premise is a challenge (audit P5): the
            # premise's author must argue for it.
            arg, f = trec["arg"], (tgt0.get("formula") or "").replace(" ", "")
            cq = "CQ:premise_acceptable"
            for o in call(base, f"/api/games/{gid}/moves?target={arg}") or []:
                if isinstance(o, dict) and o.get("locus") == "premise" and o.get("stmt") == f and o.get("question"):
                    cq = o["cq"]
                    break
            r = move(side, {"type": "question", "target_arg_id": arg, "answers_cq_id": cq, "target_premise_id": f},
                     f"Kialo {cid}: a bare denial, played as a challenge")
            res = r.get("result") or {}
            entry["played"] = "challenge" if res.get("accepted") else "no: challenge: " + str(res.get("error"))
            entry["executed"] = {"how": "challenge", "cq": cq, "target_arg": arg, "premise": f, "accepted": bool(res.get("accepted")),
                                 "error": res.get("error")}
            err = str(res.get("error") or "")
            entry["integration"] = ("as_intended" if res.get("accepted") else "duplicate" if err.startswith("Q2")
                                    else "cascade" if played_with_other(str(tgt0["claim"])) else "failed")
            report.append(entry)
            continue
        if side is None or wid == "none":
            entry["played"] = "no: " + ("no side" if side is None else "no scheme: " + str(scheme.get("fit")))
            entry["executed"], entry["integration"] = None, "not_played"
            report.append(entry)
            continue
        tgt = ((a.get("target") or {}).get("actual") or {})
        tclaim, tpart = str(tgt.get("claim") or ""), str(tgt.get("part") or "conclusion")
        # The statement the move is about.
        stmt, changed_target = None, False
        if tclaim == thesis_id and tpart.startswith("conclusion"):
            stmt = claim
        elif trec:
            if tpart.startswith("conclusion"):
                stmt = trec["conclusion"]
                changed = None if secondary_target else conclusion_changed(tclaim)
                if changed:
                    # T22: the move is about the target's annotated conclusion, not the formula its
                    # argument concludes; where the game refuses it (S7), a cascade (below).
                    stmt, changed_target = changed["conclusion_change"]["annotated"], True
            elif tpart.startswith("premise") or tpart.startswith("exception"):
                f = tgt.get("formula") or (tpart.split(":", 1)[1].strip() if ":" in tpart else "")
                stmt = f.replace(" ", "") or None
        literal = " ".join(a.get("literal") or [it["text"]])[:300]
        wag = WALTON_TO_WAG.get(wid)
        bindings = {k: words(v) for k, v in (inst.get("bindings") or {}).items()}
        cq = (a.get("role") or {}).get("cq")
        conc = norm((inst.get("conclusion") or {}).get("formula"))
        conc_en = (inst.get("conclusion") or {}).get("en") or literal
        # An instance whose conclusion is neither the target nor its opposite
        # (an amended action, an alternative) is an argument of its own.
        if stmt and conc and conc not in (stmt, "¬" + stmt) and stmt != "¬" + conc and (a.get("role") or {}).get("kind") in (
                "action_amendment", "alternative_means", "independent_pro", "independent_con", "none", "clarification"):
            stmt = None
        actions = []
        # Polarity is the stance on the target's conclusion. At an exception
        # it flips: a move against the argument argues that the exception
        # holds, and a move for it argues that it does not.
        at_exception = tpart.startswith("exception")
        kind = ("attack" if pol == "for" else "support") if at_exception else ("support" if pol == "for" else "attack")
        # A fragment that states a premise of its target's argument carries the
        # target's scheme, not one of its own: it supports the premise.
        fragment = scheme.get("fit") == "fragment" and stmt and tpart.startswith("premise")
        if fragment:
            wag = None
            # The premise it backs is on the board in the argument it backs, whether or not the
            # fragment's own move is accepted (RC20): a move at it is made there (RC14).
            backs[cid] = {"formula": norm(stmt), "trec": trec}
        if (a.get("role") or {}).get("kind") == "cq_answer" and wag and cq and str(cq).startswith("CQ:") and trec:
            act = {"type": "attack" if pol == "against" else "support", "target_arg_id": trec["arg"],
                   "answers_cq_id": cq, "scheme_id": wag, "variable_bindings": bindings, "text": literal}
            if tpart.startswith("premise"):
                f = tgt.get("formula") or (tpart.split(":", 1)[1] if ":" in tpart else "")
                if f:
                    act["target_premise_id"] = f.strip().replace(" ", "")
            actions.append(("cq", act))
        elif (a.get("role") or {}).get("kind") == "cq_answer" and wid == "other" and cq and str(cq).startswith("CQ:") and trec:
            # An answer coded with Other is played through its question too,
            # with Other: the game refuses it where the question refuses Other.
            act = {"type": "attack" if pol == "against" else "support", "target_arg_id": trec["arg"],
                   "answers_cq_id": cq, "scheme_id": "AS:other", "variable_bindings": {"P1": literal}, "text": literal}
            if tpart.startswith("premise"):
                f = tgt.get("formula") or (tpart.split(":", 1)[1] if ":" in tpart else "")
                if f:
                    act["target_premise_id"] = f.strip().replace(" ", "")
            actions.append(("cq-other", act))
        if tpart == "inference":
            # An inference is an argument, not an independent statement.
            # Format 2 deliberately leaves its formula null (OC18): the
            # engine constructs undercut(target_arg) for this exact instance.
            if not trec or not trec.get("arg"):
                entry["played"] = f"no: its inference target {tinst or tclaim} is not on the board"
                entry["executed"], entry["integration"] = None, "cascade"
                report.append(entry)
                continue
            role = (a.get("role") or {}).get("kind")
            if role == "cq_answer":
                # An argument-locus CQ is still an inference target. Keep the
                # CQ route assembled above, including answers to sign's
                # CQ:other_event; its backing/Other policy belongs to the game.
                # If unavailable or refused, do not silently replace the CQ
                # answer by an unrelated new statement or generic undercut.
                game_cq = bool(cq) and str(cq).startswith("CQ:")
                if not actions and not wag and game_cq:
                    # The game lacks the claim's scheme: answer the same question
                    # with Other, as a rejected move is retried with Other.
                    actions.append(("cq-other", {"type": "attack" if pol == "against" else "support",
                                                 "target_arg_id": trec["arg"], "answers_cq_id": cq,
                                                 "scheme_id": "AS:other", "variable_bindings": {"P1": literal},
                                                 "text": literal}))
                if not actions:
                    # No question of the game to answer: the game lacks it, so the
                    # annotation is not at fault (a coverage gap); with the scheme
                    # and the question there, a failed route.
                    entry["played"] = "no: its inference CQ answer has no playable question route"
                    entry["executed"] = None
                    entry["integration"] = "coverage_gap" if not (wag and game_cq) else "failed"
                    report.append(entry)
                    continue
            elif pol == "against" and role == "undercutter":
                actions = []
                if wag:
                    actions.append(("inference", {"type": "attack", "attack_type": "undercutter",
                        "target_arg_id": trec["arg"], "scheme_id": wag,
                        "variable_bindings": bindings, "text": literal}))
                actions.append(("inference-other", {"type": "attack", "attack_type": "undercutter",
                    "target_arg_id": trec["arg"], "scheme_id": "AS:other",
                    "variable_bindings": {"P1": literal}, "text": literal}))
            else:
                entry["played"] = "no: an inference target requires a CQ answer or an undercutter against that argument"
                entry["executed"], entry["integration"] = None, "failed"
                report.append(entry)
                continue
        if tpart != "inference" and stmt and wag:
            actions.append(("statement", {"type": kind, "target_stmt_id": stmt, "scheme_id": wag,
                                          "variable_bindings": bindings, "text": literal}))
        if tpart != "inference" and stmt:
            actions.append(("other", {"type": kind, "target_stmt_id": stmt, "scheme_id": "AS:other",
                                      "variable_bindings": {"P1": literal}, "text": literal}))
        if tpart != "inference" and not stmt:
            if wag:
                actions.append(("new", {"type": "support", "scheme_id": wag, "variable_bindings": bindings, "text": literal}))
            actions.append(("other-new", {"type": "support", "scheme_id": "AS:other", "text": literal,
                                          "variable_bindings": {"P1": literal, "C": conc_en}}))
        comparison = COMPARISON.get(inst.get("comparison"))
        for how, act in actions:
            if comparison and wag and act.get("scheme_id") == wag:
                act["comparison"] = comparison
        errors = []
        for how, act in actions:
            r = move(side, act, f"Kialo {cid}: {entry['role']}, {wid}")
            res = r.get("result") or {}
            fixed = re.findall(r"leave (\w+) out", res.get("error") or "")
            if not res.get("accepted") and fixed:  # the statement fixes them: as the browser, leave them out
                act = dict(act, variable_bindings={k: v for k, v in act["variable_bindings"].items() if k not in fixed})
                r = move(side, act, f"Kialo {cid}: {entry['role']}, {wid}")
                res = r.get("result") or {}
            if res.get("accepted"):
                arg = (res.get("arg_ids") or [None])[0]
                s = state()
                concl = s["arguments"][arg]["conclusion"] if arg in s["arguments"] else None
                played[cid] = {"arg": arg, "conclusion": concl}
                played_inst[inst.get("id") or cid] = played[cid]
                entry["played"] = how
                entry["arg"] = arg
                entry["executed"] = {"how": how, "scheme_id": act.get("scheme_id"), "cq": act.get("answers_cq_id"),
                                     "target_arg": act.get("target_arg_id"), "target_stmt": act.get("target_stmt_id"),
                                     "type": act.get("type"), "arg": arg, "conclusion": concl}
                if act.get("attack_type"):
                    entry["executed"]["attack_type"] = act["attack_type"]
                break
            errors.append(f"{how}: {res.get('error')}")
        else:
            entry["played"] = "no: " + ("; ".join(errors) if errors else f"its target ({tclaim}, {tpart}) is not on the board")
            entry["executed"] = None
        if errors and entry.get("played") in ("statement", "other", "new", "cq", "other-new", "cq-other", "inference", "inference-other"):
            entry["fallback_after"] = errors
        how = entry.get("played")
        if how in ("cq", "statement", "new", "inference"):
            carried = play_carried(cid, a, inst, side)
            if carried:
                entry["secondary"] = carried
        if cid in played:
            play_discussion(cid, a, side)
        if fragment and how == "other":
            entry["integration"] = "fragment"
        elif how in ("cq", "statement", "new", "inference"):
            # Played with its scheme; as intended if in the intended role:
            # a CQ answer as a CQ answer, anything else as a move about a
            # statement or a new argument.
            intended_cq = entry["intended"]["role"] == "cq_answer" and str(cq or "").startswith("CQ:")
            entry["integration"] = "as_intended" if intended_cq == (how == "cq") else "failed"
        elif how == "inference-other":
            entry["integration"] = "as_intended" if wid == "other" else "coverage_gap" if not wag else "failed"
        elif how == "cq-other":
            # Coded with Other, answering its question: its scheme and role. An
            # answer the game could not take with the claim's own scheme, played
            # with Other through the same question, is a coverage gap.
            entry["integration"] = "as_intended" if wid == "other" else "coverage_gap" if not wag else "failed"
        elif how == "other" and wid == "other":
            # Other was the annotation's own scheme, not a substitute for a
            # missing named scheme. A direct statement route still fails to
            # preserve an explicitly named game-CQ answer if that route failed.
            # Its conclusion is compared below, as every route's is.
            intended_cq = entry["intended"]["role"] == "cq_answer" and str(cq or "").startswith("CQ:")
            entry["integration"] = "failed" if intended_cq else "as_intended"
        elif how in ("other", "other-new"):
            entry["integration"] = "coverage_gap" if not wag else "failed"
        else:
            entry["integration"] = "failed" if wag else "not_played"
        # The annotation's own conclusion, kept or changed (F3 of the external
        # review of exchange 3; cases T6, T7, T8, T10-T17). A move the game took
        # in the intended scheme and role can still conclude another formula:
        # the engine binds an answer variable to the statement answered
        # (carneades/answers.go), and a statement can fix a conclusion
        # variable (the retry above leaves it out). An undercutter of an
        # inference (OC18) concludes undercut(its argument) by design, and a
        # challenge makes no argument: neither is compared.
        if how == "other-new":
            # A new Other argument concludes the statement the game makes from
            # the claim's English (its C), not a formula: not compared (T20).
            entry["conclusion_preserved"] = None
        elif how in ("cq", "statement", "new", "other", "cq-other") and tpart != "inference":
            expected = norm(entry["intended"].get("conclusion"))
            actual = norm((entry.get("executed") or {}).get("conclusion"))
            entry["conclusion_preserved"] = (expected == actual) if expected and actual else None
            if entry["integration"] == "as_intended" and not (expected and actual):
                entry["integration"] = "failed"  # nothing to compare (T7, T8, T12)
            elif entry["integration"] == "as_intended" and expected != actual:
                entry["integration"] = "conclusion_changed"
                entry["conclusion_change"] = {"annotated": expected, "executed": actual}
        if entry["integration"] in ("failed", "not_played") and \
                (played_with_other(tclaim) or changed_target) and \
                any("S8" in x or "S7" in x for x in errors):
            entry["integration"] = "cascade"
        report.append(entry)

    # An instance by the discussion at a claim that was not played is not on the board either.
    done = {d["instance"] for d in discussion}
    for cid, a in ann.items():
        for ins in a.get("instances") or []:
            if ins.get("by") == "the discussion" and ins.get("id") not in done:
                discussion.append({"claim": cid, "instance": ins.get("id"),
                                   "scheme": (ins.get("scheme") or {}).get("walton_id"), "played": False,
                                   "error": f"its claim {cid} is not on the board"})
    s = state()
    summary = {"game": gid, "claim": claim, "claim_label": s["statements"].get(claim, {}).get("label"),
               "played": sum(1 for e in report if not str(e.get("played", "")).startswith("no")),
               "with_their_scheme": sum(1 for e in report if e.get("played") in ("cq", "statement", "new", "inference")),
               "as_challenges": sum(1 for e in report if e.get("played") == "challenge"),
               "with_other": sum(1 for e in report if e.get("played") in ("other", "other-new", "cq-other", "inference-other")),
               "secondary_played": sum(1 for e in report for x in e.get("secondary") or [] if x.get("played")),
               "not_played": sum(1 for e in report if str(e.get("played", "")).startswith("no")),
               "integration": {k: sum(1 for e in report if e.get("integration") == k)
                               for k in ("as_intended", "conclusion_changed", "fragment", "coverage_gap", "failed", "duplicate", "cascade", "not_played")},
               "data_sha256": call(base, f"/api/games/{gid}/export").get("data_sha256"),
               "discussion": discussion,
               "moves": report}
    if args.report:
        json.dump(summary, open(args.report, "w"), indent=1, ensure_ascii=False)
    if args.export:
        exp = call(base, f"/api/games/{gid}/export")
        json.dump(exp, open(args.export, "w"), ensure_ascii=False)
    print(json.dumps({k: v for k, v in summary.items() if k != "moves"}, indent=1))


if __name__ == "__main__":
    sys.exit(main())
