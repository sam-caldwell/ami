#!/usr/bin/env python3
import sys
import re
import textwrap
from pathlib import Path

MAX_WIDTH = 120


def is_table_line(line: str) -> bool:
    # Consider as a table row if it has two or more pipes and isn't a code fence
    return line.count("|") >= 2 and not line.strip().startswith("```")


def collect_table(lines, start_idx):
    tbl = []
    i = start_idx
    while i < len(lines) and is_table_line(lines[i]):
        tbl.append(lines[i].rstrip("\n"))
        i += 1
    return tbl, i


def parse_markdown_table(md_lines):
    # Very lightweight parser: first non-separator is header; second is alignment; rest are rows.
    # Split on '|' and trim leading/trailing empty cells caused by outer pipes.
    def split_row(row):
        parts = [c.strip() for c in row.split("|")]
        # Remove leading/trailing empties if the row starts/ends with '|'
        if parts and parts[0] == "":
            parts = parts[1:]
        if parts and parts[-1] == "":
            parts = parts[:-1]
        return parts

    rows = [r for r in md_lines if r.strip()]
    if not rows:
        return None

    header = split_row(rows[0])
    if len(rows) > 1 and re.match(r"^\s*:?\-+?:?(\s*\|\s*:?\-+?:?)*\s*$", rows[1]):
        body_rows = [split_row(r) for r in rows[2:]]
    else:
        body_rows = [split_row(r) for r in rows[1:]]
    return header, body_rows


def table_to_html(md_lines):
    parsed = parse_markdown_table(md_lines)
    if not parsed:
        return "\n".join(md_lines) + "\n"
    header, body = parsed

    html = []
    html.append("<table>")
    if header:
        html.append("  <thead>")
        html.append("    <tr>")
        for h in header:
            html.append(f"      <th>{h}</th>")
        html.append("    </tr>")
        html.append("  </thead>")
    html.append("  <tbody>")
    for row in body:
        html.append("    <tr>")
        for cell in row:
            html.append(f"      <td>{cell}</td>")
        html.append("    </tr>")
    html.append("  </tbody>")
    html.append("</table>")
    return "\n".join(html) + "\n"


def wrap_paragraph(text: str) -> str:
    if not text.strip():
        return text
    # Preserve single spaces between sentences and avoid breaking long words
    wrapper = textwrap.TextWrapper(width=MAX_WIDTH, break_long_words=False, break_on_hyphens=False)
    return "\n".join(wrapper.fill(line).strip() for line in [" ".join(text.split())]) + "\n"


def should_wrap_line(line: str) -> bool:
    s = line.lstrip()
    # Skip headings, list items, blockquotes, code fences, and indented blocks
    if s.startswith(("#", "- ", "* ", ">", "```")):
        return False
    if line.startswith("    ") or line.startswith("\t"):
        return False
    return True


def format_markdown(content: str) -> str:
    lines = content.splitlines()
    out = []
    i = 0
    in_code = False
    para_buf = []

    def flush_para():
        nonlocal para_buf
        if not para_buf:
            return
        text = " ".join(l.strip() for l in para_buf)
        out.append(wrap_paragraph(text).rstrip("\n"))
        para_buf = []

    while i < len(lines):
        line = lines[i]

        # Code fence toggle
        if line.strip().startswith("```"):
            flush_para()
            in_code = not in_code
            out.append(line)
            i += 1
            continue

        if in_code:
            out.append(line)
            i += 1
            continue

        # Tables
        if is_table_line(line):
            flush_para()
            table_lines, next_i = collect_table(lines, i)
            is_wide = any(len(tl) > MAX_WIDTH for tl in table_lines)
            if is_wide:
                out.append(table_to_html(table_lines).rstrip("\n"))
            else:
                out.extend(table_lines)
            i = next_i
            continue

        # Blank line => flush paragraph
        if not line.strip():
            flush_para()
            out.append("")
            i += 1
            continue

        # Accumulate paragraph or emit as-is depending on wrap eligibility
        if should_wrap_line(line):
            para_buf.append(line)
        else:
            flush_para()
            out.append(line)

        i += 1

    flush_para()
    return "\n".join(out) + "\n"


def main(argv):
    paths = [Path(p) for p in argv[1:]]
    if not paths:
        paths = sorted(Path("docs").glob("*.md"))
    for p in paths:
        if not p.exists() or p.is_dir():
            continue
        orig = p.read_text(encoding="utf-8")
        fmt = format_markdown(orig)
        if fmt != orig:
            p.write_text(fmt, encoding="utf-8")
            print(f"formatted: {p}")
        else:
            print(f"unchanged: {p}")


if __name__ == "__main__":
    main(sys.argv)

