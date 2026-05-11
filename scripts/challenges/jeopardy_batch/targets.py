from __future__ import annotations

import json
import re

from .models import Target
from .paths import MATRIX_DOC, PACKS_DIR, REPO_ROOT, TARGETS_FILE


def load_targets() -> list[Target]:
    data = json.loads(TARGETS_FILE.read_text(encoding="utf-8"))
    return [Target(**item) for item in data]


def validate_targets(targets: list[Target]) -> None:
    counts: dict[str, int] = {}
    for target in targets:
        counts[target.category] = counts.get(target.category, 0) + 1

    expected = {
        "crypto": 11,
        "forensics": 11,
        "misc": 11,
        "pwn": 10,
        "reverse": 11,
        "web": 11,
    }
    if counts != expected:
        raise SystemExit(f"分类配额错误: {counts}")

    for field_name in ("slug", "kind", "primary_skill", "primary_action", "training_goal"):
        values = [getattr(target, field_name) for target in targets]
        if len(values) != len(set(values)):
            raise SystemExit(f"{field_name} 存在重复")

    overlap = []
    for target in targets:
        pack_dir = PACKS_DIR / target.slug
        if pack_dir.exists() and not existing_pack_matches_target(pack_dir, target):
            overlap.append(target.slug)
    if overlap:
        raise SystemExit(f"新增 slug 与现有 pack 冲突: {sorted(overlap)}")


def existing_pack_matches_target(pack_dir, target: Target) -> bool:
    challenge_path = pack_dir / "challenge.yml"
    if not challenge_path.exists():
        return False
    text = challenge_path.read_text(encoding="utf-8")
    patterns = [
        rf"^\s*slug:\s*{re.escape(target.slug)}\s*$",
        rf'^\s*title:\s*"{re.escape(target.title)}"\s*$',
        rf"^\s*category:\s*{re.escape(target.category)}\s*$",
        rf"^\s*-\s*topic:{re.escape(target.kind)}\s*$",
    ]
    return all(re.search(pattern, text, re.M) for pattern in patterns)


def filter_targets(targets: list[Target], requested: list[str]) -> list[Target]:
    if not requested:
        return targets
    want = set(requested)
    chosen = [target for target in targets if target.slug in want]
    missing = sorted(want - {target.slug for target in chosen})
    if missing:
        raise SystemExit(f"未找到目标 slug: {missing}")
    return chosen


def print_summary(targets: list[Target]) -> None:
    counts: dict[str, int] = {}
    for target in targets:
        counts[target.category] = counts.get(target.category, 0) + 1
    print(f"targets={len(targets)} categories={counts}")


def sync_matrix_doc(targets: list[Target]) -> None:
    grouped: dict[str, list[Target]] = {}
    for target in targets:
        grouped.setdefault(target.category, []).append(target)

    order = ["crypto", "forensics", "misc", "reverse", "pwn", "web"]
    lines = [
        "# Jeopardy 80 真实训练题去重矩阵",
        "",
        "这份矩阵服务本轮 Jeopardy 题目扩容，目标是把新增 65 道题严格收口到“真实训练题、可验证、不可换皮”。",
        "",
        "去重规则：",
        "",
        "1. `主知识点` 不重复",
        "2. `主动作` 不重复",
        "3. `主训练目标` 不重复",
        "",
        "只要核心训练点发生重合，就视为换皮，需要换题型，不接受仅改题名或附件名。",
        "",
    ]
    for category in order:
        items = grouped.get(category, [])
        lines.append(f"## {category.title()}（{len(items)}）")
        lines.append("")
        for target in items:
            lines.append(f"- `{target.slug}`")
            lines.append(f"  - 主知识点：{target.primary_skill}")
            lines.append(f"  - 主动作：{target.primary_action}")
            lines.append(f"  - 主训练目标：{target.training_goal}")
        lines.append("")
    MATRIX_DOC.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
