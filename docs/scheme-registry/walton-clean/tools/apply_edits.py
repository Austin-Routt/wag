"""Apply the recorded edits (tools/edits.py) to walton.yml.

Operations, applied in list order to the current text of a line:
  sub    : replace the substring `old` (must occur exactly once) by `new`
  set    : replace the whole line; `before` must equal the current line;
           `after` may contain "\n" (insertion) or be None (deletion)
  subs   : replace every occurrence of `old` by `new` on each of `lines`
  insert : add the line `text` after original line `line`
  delete : remove the line; `before` must equal it
  rename : replace the identifier `old` by `new` wherever it occurs as a whole
           identifier ([A-Za-z0-9_] boundaries); the set of lines touched must
           equal `lines`
Line numbers always refer to the ORIGINAL file.  Each concrete change is
recorded as one row (line, scheme, category, before, after, reason).

usage: python3 -I apply_edits.py ORIGINAL OUT.yml [--syntax-only | --with-proposals] [--json changes.json]
"""
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from edits import OPS, ID_MAP  # noqa: E402


def scheme_of_lines(lines):
    """Map original line number -> owning section (scheme id, 'language', 'meta')."""
    owner = {}
    cur = "meta"
    for i, l in enumerate(lines, 1):
        if l.startswith("language:"):
            cur = "language"
        elif l.startswith("argument_schemes:"):
            cur = "argument_schemes"
        m = re.match(r"\s*- id:\s*(\S+)", l)
        if m:
            cur = m.group(1)
        owner[i] = cur
    return owner


def main(argv):
    src, dst = argv[1], argv[2]
    syntax_only = "--syntax-only" in argv
    ops = list(OPS)
    if "--with-proposals" in argv:  # check/walton.proposals.yml only
        from proposals import PROPOSALS
        ops += PROPOSALS
    json_out = argv[argv.index("--json") + 1] if "--json" in argv else None
    with open(src, encoding="utf-8") as f:
        orig = f.read().split("\n")
    owner = scheme_of_lines(orig)
    cur = list(orig)          # current text per original line (None = deleted)
    rows = []

    def record(n, before, after, op, change="edit"):
        rows.append({
            "line": n,
            "change": change,
            "scheme": op.get("scheme") or owner[n],
            "scheme_after": ID_MAP.get(op.get("scheme") or owner[n], op.get("scheme") or owner[n]),
            "category": op["category"],
            "needs_owner_approval": op["category"].startswith("repair"),
            "before": before,
            "after": after,
            "reason": op["reason"],
        })

    for op in ops:
        if syntax_only and not op.get("syntax_only"):
            continue
        kind = op["op"]
        if kind == "sub":
            n = op["line"]
            text = cur[n - 1]
            if text is None or text.count(op["old"]) != 1:
                raise SystemExit(f"line {n}: {op['old']!r} occurs {0 if text is None else text.count(op['old'])} times in {text!r}")
            new = text.replace(op["old"], op["new"])
            cur[n - 1] = new
            record(n, text, new, op)
        elif kind == "set":
            n = op["line"]
            text = cur[n - 1]
            if text != op["before"]:
                raise SystemExit(f"line {n}: expected {op['before']!r}, found {text!r}")
            cur[n - 1] = op["after"]
            record(n, text, op["after"], op)
        elif kind == "subs":
            for n in op["lines"]:
                text = cur[n - 1]
                if text is None or op["old"] not in text:
                    raise SystemExit(f"line {n}: {op['old']!r} not in {text!r}")
                new = text.replace(op["old"], op["new"])
                cur[n - 1] = new
                record(n, text, new, op)
        elif kind == "insert":
            n = op["line"]
            cur[n - 1] = cur[n - 1] + "\n" + op["text"]
            record(n, "", op["text"], op, change="insert after")
        elif kind == "delete":
            n = op["line"]
            text = cur[n - 1]
            if text != op["before"]:
                raise SystemExit(f"line {n}: expected {op['before']!r}, found {text!r}")
            cur[n - 1] = None
            record(n, text, "", op, change="delete")
        elif kind == "rename":
            pat = re.compile(r"(?<![A-Za-z0-9_])" + re.escape(op["old"]) + r"(?![A-Za-z0-9_])")
            touched = []
            for i, text in enumerate(cur):
                if text is None or not pat.search(text):
                    continue
                new = pat.sub(op["new"], text)
                cur[i] = new
                touched.append(i + 1)
                record(i + 1, text, new, op)
            if touched != sorted(op["lines"]):
                raise SystemExit(f"rename {op['old']}: touched {touched}, expected {op['lines']}")
        else:
            raise SystemExit(f"unknown op {kind}")

    out = [t for t in cur if t is not None]
    with open(dst, "w", encoding="utf-8") as f:
        f.write("\n".join(out))
    if json_out:
        rows.sort(key=lambda r: r["line"])  # stable: keeps op order within a line
        with open(json_out, "w", encoding="utf-8") as f:
            json.dump(rows, f, indent=1, ensure_ascii=False)
            f.write("\n")
    print(f"{len(rows)} recorded changes -> {dst}")


if __name__ == "__main__":
    main(sys.argv)
