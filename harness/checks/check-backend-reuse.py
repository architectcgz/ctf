#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path

from common import get_changed_files, load_reuse_reference_text, parse_diff_args, repo_files


BACKEND_PATTERNS = {
    "handler": [
        "code/backend/internal/module/**/api/**/*.go",
        "code/backend/internal/handler/**/*.go",
    ],
    "repository": [
        "code/backend/internal/module/**/infrastructure/**/*repository*.go",
        "code/backend/internal/module/**/infrastructure/repository.go",
    ],
    "port": [
        "code/backend/internal/module/**/ports/**/*.go",
    ],
    "job": [
        "code/backend/internal/module/**/application/**/*job*.go",
        "code/backend/internal/module/**/application/jobs/**/*.go",
        "code/backend/internal/module/**/application/**/*worker*.go",
    ],
    "mapper": [
        "code/backend/internal/module/**/*mapper*.go",
        "code/backend/internal/shared/mapper*/**/*.go",
    ],
    "readmodel": [
        "code/backend/internal/module/*_readmodel/**/*.go",
    ],
    "composition": [
        "code/backend/internal/app/composition/**/*.go",
        "code/backend/internal/module/**/runtime/module.go",
    ],
    "migration": [
        "code/backend/migrations/**/*.sql",
        "code/backend/internal/module/**/migrations/**/*.sql",
    ],
}


def match_category(path: str) -> str | None:
    target = Path(path)
    for category, patterns in BACKEND_PATTERNS.items():
        for candidate in repo_files(patterns):
            if Path(candidate) == target:
                return category
    return None


def module_name(path: str) -> str | None:
    parts = Path(path).parts
    try:
        idx = parts.index("module")
    except ValueError:
        return None
    if idx + 1 < len(parts):
        return parts[idx + 1]
    return None


def score_candidate(new_path: str, candidate: str) -> tuple[int, str]:
    score = 0
    new_module = module_name(new_path)
    candidate_module = module_name(candidate)
    if new_module and new_module == candidate_module:
        score += 5
    new_parts = {part.lower() for part in Path(new_path).parts}
    candidate_parts = {part.lower() for part in Path(candidate).parts}
    score += len(new_parts & candidate_parts)
    new_stem = Path(new_path).stem.lower()
    candidate_stem = Path(candidate).stem.lower()
    for token in ("repository", "handler", "service", "job", "worker", "mapper", "module"):
        if token in new_stem and token in candidate_stem:
            score += 2
    return score, candidate


def main() -> int:
    args = parse_diff_args("detect backend additions that should cite existing backend patterns")
    changed_files = get_changed_files(args)
    new_backend = [
        item.path
        for item in changed_files
        if item.is_added and item.path.startswith("code/backend/") and (item.path.endswith(".go") or item.path.endswith(".sql"))
    ]
    if not new_backend:
        print("PASS: no new backend Go or migration files added")
        return 0

    decision_text = load_reuse_reference_text()
    failures = 0

    for path in new_backend:
        category = match_category(path)
        if category is None:
            continue

        candidates = [candidate for candidate in repo_files(BACKEND_PATTERNS[category], exclude={path})]
        ranked = [candidate for score, candidate in sorted((score_candidate(path, c) for c in candidates), reverse=True) if score > 0]
        top_candidates = ranked[:5]
        if not top_candidates:
            continue

        if any(candidate in decision_text for candidate in top_candidates):
            continue

        failures += 1
        print(f"FAIL: {path} is a new backend {category} and should cite existing backend patterns:", file=sys.stderr)
        for candidate in top_candidates:
            print(f"- {candidate}", file=sys.stderr)
        print(
            "Please cite the closest backend implementation(s) in a task-scoped reuse decision file under .harness/reuse-decisions/ and explain whether to reuse, extend, split, or create new with reason.",
            file=sys.stderr,
        )

    if failures:
        return 1

    print("PASS: new backend files either have no comparable candidates or already cite existing patterns")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
