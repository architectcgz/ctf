#!/usr/bin/env python3
from __future__ import annotations

import sys

from common import (
    API_PATTERNS,
    extract_request_signatures,
    get_changed_files,
    load_reuse_reference_text,
    parse_diff_args,
    read_text,
    repo_files,
    url_segments,
)


def is_api_wrapper(path: str) -> bool:
    return (
        path.endswith(".ts")
        and path.startswith("code/frontend/src/api/")
        and "/__tests__/" not in path
        and not path.endswith("/index.ts")
        and not path.endswith("/request.ts")
        and not path.endswith("/contracts.ts")
    )


def api_score(new_api: str, candidate_api: str) -> tuple[int, set[tuple[str, str]]]:
    new_signatures = extract_request_signatures(read_text(new_api))
    candidate_signatures = extract_request_signatures(read_text(candidate_api))
    shared_matches: set[tuple[str, str]] = set()

    for new_method, new_url in new_signatures:
        new_segments = url_segments(new_url)
        for old_method, old_url in candidate_signatures:
            old_segments = url_segments(old_url)
            shared_prefix = tuple(a for a, b in zip(new_segments, old_segments) if a == b)
            if new_method == old_method and len(shared_prefix) >= 2:
                shared_matches.add((new_method, new_url))
            elif len(shared_prefix) >= 3:
                shared_matches.add((new_method, new_url))

    return len(shared_matches), shared_matches


def main() -> int:
    args = parse_diff_args("detect duplicate api wrappers")
    changed_files = get_changed_files(args)
    new_api_files = [item.path for item in changed_files if item.is_added and is_api_wrapper(item.path)]
    if not new_api_files:
        print("PASS: no new API wrapper files added")
        return 0

    decision_text = load_reuse_reference_text()
    existing_api_files = repo_files(API_PATTERNS, exclude=set(new_api_files))
    failures = 0

    for new_api in new_api_files:
        candidates = []
        for candidate in existing_api_files:
            if not is_api_wrapper(candidate):
                continue
            score, shared_matches = api_score(new_api, candidate)
            if score > 0 and shared_matches:
                candidates.append((score, candidate))

        candidates.sort(reverse=True)
        top_candidates = [candidate for _, candidate in candidates[:3]]
        if not top_candidates:
            continue

        if any(candidate in decision_text for candidate in top_candidates):
            continue

        failures += 1
        print(f"FAIL: {new_api} overlaps with existing API wrappers:", file=sys.stderr)
        for candidate in top_candidates:
            print(f"- {candidate}", file=sys.stderr)
        print(
            "Please cite the existing API wrappers in a task-scoped reuse decision file under .harness/reuse-decisions/ and explain why they cannot be extended. "
            "If this is a reusable pattern, also update harness/reuse/index.yaml after the task.",
            file=sys.stderr,
        )

    if failures:
        return 1

    print("PASS: new API wrappers either have no overlap or already cite existing wrappers")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
