from __future__ import annotations

import argparse
import re
import subprocess
from collections import Counter
from dataclasses import dataclass
from pathlib import Path

from .paths import PACKS_DIR, REPORT_DOC, REPO_ROOT
from .targets import load_targets as load_new_targets


FLAG_RE = re.compile(r"flag\{[a-zA-Z0-9_\-]+\}")


@dataclass(frozen=True)
class VerifyTarget:
    slug: str
    title: str
    category: str


def load_targets(include_new_only: bool) -> list[VerifyTarget]:
    if include_new_only:
        return [VerifyTarget(slug=item.slug, title=item.title, category=item.category) for item in load_new_targets()]
    targets = []
    for pack_dir in sorted(item for item in PACKS_DIR.iterdir() if item.is_dir()):
        text = (pack_dir / "challenge.yml").read_text(encoding="utf-8")
        slug = re.search(r"^\s*slug:\s*(\S+)\s*$", text, re.M)
        title = re.search(r'^\s*title:\s*"(.*?)"\s*$', text, re.M)
        category = re.search(r"^\s*category:\s*(\S+)\s*$", text, re.M)
        if not slug or not category:
            raise SystemExit(f"题包缺少元数据: {pack_dir}")
        targets.append(
            VerifyTarget(
                slug=slug.group(1),
                title=title.group(1) if title else pack_dir.name,
                category=category.group(1),
            )
        )
    return targets


def expected_flag(pack_dir: Path) -> str | None:
    text = (pack_dir / "challenge.yml").read_text(encoding="utf-8")
    match = re.search(r"^\s*value:\s*(flag\{[^\n]+\})\s*$", text, re.M)
    return match.group(1) if match else None


def verify_target(target: VerifyTarget) -> tuple[str, str, str]:
    pack_dir = PACKS_DIR / target.slug
    solve_path = pack_dir / "writeup" / "solve.py"
    if not solve_path.exists():
        return target.slug, "missing-solve", ""
    try:
        proc = subprocess.run(
            ["python3", str(solve_path)],
            cwd=pack_dir,
            capture_output=True,
            text=True,
            timeout=30,
        )
    except subprocess.TimeoutExpired as exc:
        output = ((exc.stdout or "") + "\n" + (exc.stderr or "")).strip()
        return target.slug, "solver-timeout", output
    output = (proc.stdout + "\n" + proc.stderr).strip()
    if proc.returncode != 0:
        return target.slug, f"solver-exit-{proc.returncode}", output
    match = FLAG_RE.search(output)
    if not match:
        return target.slug, "no-flag", output
    actual = match.group(0)
    expected = expected_flag(pack_dir)
    if expected and actual != expected:
        return target.slug, "flag-mismatch", output
    return target.slug, "ok", actual


def write_report(targets: list[VerifyTarget], results: list[tuple[str, str, str]]) -> None:
    status_map = {slug: (status, detail) for slug, status, detail in results}
    counts = Counter(target.category for target in targets)
    lines = [
        "# Jeopardy 80 扩容题包验证报告",
        "",
        f"- 题包总数：`{len(targets)}`",
        f"- 分类分布：`{dict(sorted(counts.items()))}`",
        "",
        "## 结果",
        "",
    ]
    for target in targets:
        status, detail = status_map[target.slug]
        lines.append(f"- `{target.slug}` `{status}`")
        if detail:
            lines.append(f"  - 输出：`{detail}`")
    lines.append("")
    REPORT_DOC.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="验证新增 Jeopardy 题包")
    parser.add_argument("--slug", action="append", default=[], help="只验证指定 slug")
    parser.add_argument("--new-only", action="store_true", help="只验证新增 65 题")
    parser.add_argument("--write-report", action="store_true", help="写入验证报告")
    args = parser.parse_args()

    targets = load_targets(include_new_only=args.new_only)
    if args.slug:
        wanted = set(args.slug)
        targets = [target for target in targets if target.slug in wanted]
        missing = wanted - {target.slug for target in targets}
        if missing:
            raise SystemExit(f"未找到目标 slug: {sorted(missing)}")

    results = []
    for target in targets:
        results.append(verify_target(target))
        print(f"{target.slug}: {results[-1][1]}")

    failed = [item for item in results if item[1] != "ok"]
    if args.write_report:
        write_report(targets, results)
        print(f"report: {REPORT_DOC.relative_to(REPO_ROOT)}")
    if failed:
        raise SystemExit(1)
