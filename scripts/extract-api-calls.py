#!/usr/bin/env python3
"""Extract Wanderlog API call sites from the decompiled JavaScript bundle.

The decompiled Hermes output is not valid, idiomatic source code, so this keeps
the parser deliberately small and tolerant. It records two views:

* wrappedAxios(...) call sites, with method/url when they are literal strings
* every quoted /api/... endpoint string, including strings outside wrappedAxios
"""

from __future__ import annotations

import argparse
import json
import re
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Iterable


STRING_RE = re.compile(r"""(["'])(?P<value>(?:\\.|(?!\1).)*?)\1""")
API_STRING_RE = re.compile(r"""(["'])(?P<value>/api/[^"'\\\s{}),;]+)(\1)""")
URL_FIELD_RE = re.compile(r"""(?:^|[,{]\s*)url\s*:\s*(?P<expr>[^,\n}]+)""")
METHOD_FIELD_RE = re.compile(r"""(?:^|[,{]\s*)method\s*:\s*(?P<expr>[^,\n}]+)""")


@dataclass(frozen=True)
class ExtractedCall:
    kind: str
    line: int
    method: str | None
    url: str | None
    url_expr: str | None
    method_expr: str | None
    source: str


def line_number(text: str, offset: int) -> int:
    return text.count("\n", 0, offset) + 1


def decode_js_string(expr: str) -> str | None:
    expr = expr.strip()
    match = STRING_RE.fullmatch(expr)
    if not match:
        return None
    value = match.group("value")
    try:
        return bytes(value, "utf-8").decode("unicode_escape")
    except UnicodeDecodeError:
        return value


def find_matching_paren(text: str, open_index: int) -> int | None:
    depth = 0
    quote: str | None = None
    escaped = False
    for i in range(open_index, len(text)):
        ch = text[i]
        if quote is not None:
            if escaped:
                escaped = False
            elif ch == "\\":
                escaped = True
            elif ch == quote:
                quote = None
            continue
        if ch in ("'", '"', "`"):
            quote = ch
            continue
        if ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
            if depth == 0:
                return i
    return None


def excerpt(source: str, limit: int = 220) -> str:
    compact = " ".join(source.strip().split())
    if len(compact) <= limit:
        return compact
    return compact[: limit - 3] + "..."


def extract_wrapped_axios_calls(text: str) -> Iterable[ExtractedCall]:
    needle = "wrappedAxios"
    start = 0
    while True:
        idx = text.find(needle, start)
        if idx < 0:
            return
        paren = text.find("(", idx + len(needle))
        if paren < 0:
            return
        if text[idx + len(needle) : paren].strip():
            start = idx + len(needle)
            continue
        end = find_matching_paren(text, paren)
        if end is None:
            start = paren + 1
            continue

        call_source = text[idx : end + 1]
        url_expr = None
        method_expr = None
        url = None
        method = "GET"

        url_match = URL_FIELD_RE.search(call_source)
        if url_match:
            url_expr = url_match.group("expr").strip()
            url = decode_js_string(url_expr)

        method_match = METHOD_FIELD_RE.search(call_source)
        if method_match:
            method_expr = method_match.group("expr").strip()
            decoded = decode_js_string(method_expr)
            method = decoded.upper() if decoded else None

        yield ExtractedCall(
            kind="wrappedAxios",
            line=line_number(text, idx),
            method=method,
            url=url,
            url_expr=url_expr,
            method_expr=method_expr,
            source=excerpt(call_source),
        )
        start = end + 1


def extract_endpoint_strings(text: str) -> Iterable[ExtractedCall]:
    seen: set[tuple[int, str]] = set()
    for match in API_STRING_RE.finditer(text):
        url = match.group("value")
        key = (match.start(), url)
        if key in seen:
            continue
        seen.add(key)
        line_start = text.rfind("\n", 0, match.start()) + 1
        line_end = text.find("\n", match.end())
        if line_end < 0:
            line_end = len(text)
        yield ExtractedCall(
            kind="endpointString",
            line=line_number(text, match.start()),
            method=None,
            url=url,
            url_expr=match.group(0),
            method_expr=None,
            source=excerpt(text[line_start:line_end]),
        )


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--input",
        default="artifacts/decompiled/wanderlog_decompiled.js",
        help="Path to the decompiled JS file",
    )
    parser.add_argument(
        "--output",
        default="artifacts/api-contracts/decompiled_calls.json",
        help="Path to write extracted calls as JSON",
    )
    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.output)
    text = input_path.read_text(errors="replace")

    wrapped = list(extract_wrapped_axios_calls(text))
    endpoints = list(extract_endpoint_strings(text))
    data = {
        "source": str(input_path),
        "wrappedAxiosCount": len(wrapped),
        "endpointStringCount": len(endpoints),
        "wrappedAxios": [asdict(call) for call in wrapped],
        "endpointStrings": [asdict(call) for call in endpoints],
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")
    print(
        f"wrote {output_path} "
        f"({len(wrapped)} wrappedAxios calls, {len(endpoints)} endpoint strings)"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
