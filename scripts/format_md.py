#!/usr/bin/env python3
"""
Format Markdown files to wrap lines at 120 columns.

Rules:
- Do not modify fenced code blocks (``` ... ```).
- Do not modify indented code blocks (>=4 leading spaces).
- Do not wrap ATX headings (# ...).
- For list items, wrap and indent continuation lines to align under the text.
- For blockquotes (>), keep the blockquote prefix on wrapped continuation lines.
- For Markdown tables, leave unchanged; optionally convert to HTML if requested separately.

This script performs non-destructive wrapping by splitting only lines that exceed 120 columns.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

MAX_COL = 120

FENCE_RE = re.compile(r"^\s*```")
ATX_HEADER_RE = re.compile(r"^\s*#{1,6}\s+")
INDENTED_CODE_RE = re.compile(r"^(?:\t| {4,})")
BLOCKQUOTE_RE = re.compile(r"^(\s*>+\s+)(.*)$")
BULLET_RE = re.compile(r"^(\s*[-*+]\s+)(.*)$")
NUMBERED_RE = re.compile(r"^(\s*\d+\.\s+)(.*)$")
TABLE_ROW_RE = re.compile(r"^\s*\|.*\|\s*$")
TABLE_SEP_RE = re.compile(r"^\s*\|?(?:\s*:?[-]{3,}:?\s*\|)+\s*:?[-]{3,}:?\s*\|?\s*$")


def _bytelen(s: str) -> int:
    return len(s.encode("utf-8"))


def wrap_with_prefix(prefix: str, text: str, keep_prefix_on_continuation: bool) -> list[str]:
    # Split on whitespace to wrap at word boundaries
    words = text.split()
    if not words:
        return [prefix.rstrip()] if prefix.strip() and not text else [prefix + text]

    first_prefix = prefix
    cont_prefix = prefix if keep_prefix_on_continuation else (" " * len(prefix))

    lines: list[str] = []
    line = first_prefix
    for w in words:
        # If adding the next word would exceed max, break
        projected = _bytelen(line) + (0 if line.endswith(" ") or line == first_prefix else 1) + _bytelen(w)
        if projected > MAX_COL:
            lines.append(line.rstrip())
            line = cont_prefix + w
        else:
            if line == first_prefix or line.endswith(" "):
                line += w
            else:
                line += " " + w
    # Ensure last line also respects the byte-based limit
    if _bytelen(line) > MAX_COL:
        # Hard wrap at the last space before limit
        cur = line
        while _bytelen(cur) > MAX_COL and " " in cur.strip():
            # Find the split point by iterating characters until byte limit
            bytes_used = 0
            split_idx = 0
            for i, ch in enumerate(cur):
                ch_bytes = len(ch.encode("utf-8"))
                extra = ch_bytes
                if bytes_used + extra > MAX_COL:
                    break
                bytes_used += extra
                split_idx = i + 1
            # Backtrack to last space within split_idx
            back = cur[:split_idx].rstrip()
            last_space = back.rfind(" ")
            if last_space <= 0:
                break
            lines.append(back[:last_space].rstrip())
            cur = cont_prefix + back[last_space + 1 :] + cur[split_idx:]
        lines.append(cur.rstrip())
    else:
        lines.append(line.rstrip())
    return lines


def process_file(path: Path) -> str:
    out_lines: list[str] = []
    in_fence = False
    in_table = False
    table_lines: list[str] = []

    def flush_table():
        nonlocal table_lines
        if not table_lines:
            return
        # Determine if table requires wrapping (any line > MAX_COL)
        requires_wrap = any(len(l.rstrip("\n")) > MAX_COL for l in table_lines)
        if requires_wrap:
            # Convert markdown table to HTML to allow wrapping
            # Simple pipe-split; does not support escaped pipes.
            rows = [r.strip().strip("|") for r in table_lines if not TABLE_SEP_RE.match(r)]
            if rows:
                # Header
                hdr_cells = [c.strip() for c in rows[0].split("|")]
                out_lines.append('<table style="width:100%; word-break: break-word; white-space: normal;">')
                out_lines.append("  <thead>")
                out_lines.append("    <tr>")
                for c in hdr_cells:
                    out_lines.append(f"      <th>{c}</th>")
                out_lines.append("    </tr>")
                out_lines.append("  </thead>")
                # Body
                data_rows = rows[1:]
                if data_rows:
                    out_lines.append("  <tbody>")
                    for r in data_rows:
                        cells = [c.strip() for c in r.split("|")]
                        out_lines.append("    <tr>")
                        for c in cells:
                            out_lines.append(f"      <td>{c}</td>")
                        out_lines.append("    </tr>")
                    out_lines.append("  </tbody>")
                out_lines.append("</table>")
            else:
                out_lines.extend(table_lines)
        else:
            out_lines.extend(table_lines)
        table_lines = []

    with path.open("r", encoding="utf-8") as f:
        for raw in f:
            line = raw.rstrip("\n")

            # Handle fenced code blocks
            if FENCE_RE.match(line):
                flush_table()
                in_fence = not in_fence
                out_lines.append(line)
                continue

            if in_fence:
                out_lines.append(line)
                continue

            # Table detection: accumulate contiguous table lines
            if TABLE_ROW_RE.match(line) or TABLE_SEP_RE.match(line):
                if not in_table:
                    in_table = True
                    table_lines = []
                table_lines.append(line)
                continue
            else:
                if in_table:
                    flush_table()
                    in_table = False

            # Skip wrapping ATX headings
            if ATX_HEADER_RE.match(line):
                out_lines.append(line)
                continue

            # Blockquote
            m = BLOCKQUOTE_RE.match(line)
            if m:
                prefix, text = m.groups()
                # Wrap keeping blockquote prefix on continuation lines
                if _bytelen(line) > MAX_COL:
                    out_lines.extend(wrap_with_prefix(prefix, text, keep_prefix_on_continuation=True))
                else:
                    out_lines.append(line)
                continue

            # Bullets (support nested via leading spaces)
            m = BULLET_RE.match(line) or NUMBERED_RE.match(line)
            if m:
                prefix, text = m.groups()
                if _bytelen(line) > MAX_COL:
                    out_lines.extend(wrap_with_prefix(prefix, text, keep_prefix_on_continuation=False))
                else:
                    out_lines.append(line)
                continue

            # Skip wrapping indented code lines (after list/blockquote handling)
            if INDENTED_CODE_RE.match(line):
                out_lines.append(line)
                continue

            # Short lines unchanged
            if _bytelen(line) <= MAX_COL:
                out_lines.append(line)
                continue

            # Default: preserve leading spaces as prefix
            leading_spaces = len(line) - len(line.lstrip(" "))
            prefix = " " * leading_spaces
            text = line[leading_spaces:]
            out_lines.extend(wrap_with_prefix(prefix, text, keep_prefix_on_continuation=False))

    # Flush at EOF
    if in_table:
        flush_table()

    return "\n".join(out_lines) + "\n"


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("usage: format_md.py <file> [<file> ...]", file=sys.stderr)
        return 2
    paths = [Path(a) for a in argv[1:]]
    for p in paths:
        if not p.exists() or not p.is_file():
            print(f"warning: skipping non-file {p}", file=sys.stderr)
            continue
        new = process_file(p)
        with p.open("w", encoding="utf-8", newline="\n") as f:
            f.write(new)
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
