#!/usr/bin/env python3
from __future__ import annotations

import sys

from common import (
    PAGE_FEATURE_KEYWORDS,
    PAGE_PATTERNS,
    extract_keyword_hits,
    extract_name_tokens,
    get_changed_files,
    load_reuse_decision_text,
    parse_diff_args,
    read_text,
    repo_files,
)


def similarity_score(new_path: str, candidate_path: str) -> tuple[int, set[str], set[str]]:
    new_text = read_text(new_path)
    candidate_text = read_text(candidate_path)
    new_keywords = extract_keyword_hits(new_text, PAGE_FEATURE_KEYWORDS)
    candidate_keywords = extract_keyword_hits(candidate_text, PAGE_FEATURE_KEYWORDS)
    shared_keywords = new_keywords & candidate_keywords

    new_tokens = extract_name_tokens(new_path)
    candidate_tokens = extract_name_tokens(candidate_path)
    shared_tokens = new_tokens & candidate_tokens

    score = len(shared_keywords) * 3 + len(shared_tokens) * 2
    if "WorkspaceDirectoryToolbar" in new_text and "WorkspaceDirectoryToolbar" in candidate_text:
        score += 3
    if "WorkspaceDataTable" in new_text and "WorkspaceDataTable" in candidate_text:
        score += 3
    if "PagePaginationControls" in new_text and "PagePaginationControls" in candidate_text:
        score += 2

    return score, shared_keywords, shared_tokens


def is_page_like(path: str) -> bool:
    return path.endswith(".vue") and ("Page.vue" in path or "/views/" in path)


def main() -> int:
    args = parse_diff_args("detect new pages that duplicate existing page patterns")
    changed_files = get_changed_files(args)
    new_pages = [item.path for item in changed_files if item.is_added and is_page_like(item.path)]
    if not new_pages:
        print("PASS: no new page files added")
        return 0

    decision_text = load_reuse_decision_text()
    existing_pages = repo_files(PAGE_PATTERNS, exclude=set(new_pages))
    failures = 0

    for new_page in new_pages:
        scored_candidates = []
        for candidate in existing_pages:
            score, shared_keywords, shared_tokens = similarity_score(new_page, candidate)
            if len(shared_keywords) >= 4 or score >= 10:
                scored_candidates.append((score, candidate, shared_keywords, shared_tokens))

        scored_candidates.sort(reverse=True)
        top_candidates = [candidate for _, candidate, _, _ in scored_candidates[:3]]
        if not top_candidates:
            continue

        if any(candidate in decision_text for candidate in top_candidates):
            continue

        failures += 1
        print(f"FAIL: {new_page} looks structurally similar to existing pages:", file=sys.stderr)
        for candidate in top_candidates:
            print(f"- {candidate}", file=sys.stderr)
        print(
            "Please update .harness/reuse-decision.md to reference these files and explain "
            "why you are reusing, extending, refactoring, or creating a new page.",
            file=sys.stderr,
        )

    if failures:
        return 1

    print("PASS: new pages either have no strong match or already cite similar pages")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
