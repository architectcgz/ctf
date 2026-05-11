#!/usr/bin/env python3
from __future__ import annotations

import sys

from common import (
    HOOK_FEATURE_KEYWORDS,
    HOOK_PATTERNS,
    extract_api_imports,
    extract_keyword_hits,
    extract_name_tokens,
    get_changed_files,
    load_reuse_reference_text,
    parse_diff_args,
    read_text,
    repo_files,
)


def hook_score(new_hook: str, candidate_hook: str) -> tuple[int, set[str], set[str], set[str]]:
    new_text = read_text(new_hook)
    candidate_text = read_text(candidate_hook)

    shared_name_tokens = extract_name_tokens(new_hook) & extract_name_tokens(candidate_hook)
    shared_keywords = extract_keyword_hits(new_text, HOOK_FEATURE_KEYWORDS) & extract_keyword_hits(
        candidate_text, HOOK_FEATURE_KEYWORDS
    )
    shared_api_imports = extract_api_imports(new_text) & extract_api_imports(candidate_text)

    score = len(shared_name_tokens) * 3 + len(shared_keywords) * 2 + len(shared_api_imports) * 3
    return score, shared_name_tokens, shared_keywords, shared_api_imports


def main() -> int:
    args = parse_diff_args("detect new hooks that duplicate existing hook patterns")
    changed_files = get_changed_files(args)
    new_hooks = [item.path for item in changed_files if item.is_added and item.path.endswith(".ts") and "/use" in item.path]
    if not new_hooks:
        print("PASS: no new hook files added")
        return 0

    decision_text = load_reuse_reference_text()
    existing_hooks = repo_files(HOOK_PATTERNS, exclude=set(new_hooks))
    failures = 0

    for new_hook in new_hooks:
        candidates = []
        for candidate in existing_hooks:
            score, shared_tokens, shared_keywords, shared_api_imports = hook_score(new_hook, candidate)
            if score >= 7 and (shared_tokens or shared_keywords or shared_api_imports):
                candidates.append((score, candidate))

        candidates.sort(reverse=True)
        top_candidates = [candidate for _, candidate in candidates[:3]]
        if not top_candidates:
            continue

        if any(candidate in decision_text for candidate in top_candidates):
            continue

        failures += 1
        print(f"FAIL: {new_hook} looks close to existing hooks:", file=sys.stderr)
        for candidate in top_candidates:
            print(f"- {candidate}", file=sys.stderr)
        print(
            "Please reference the similar hooks in .harness/reuse-decision.md before creating another one-off hook. "
            "If this is a reusable pattern, also update .harness/reuse-index.yaml after the task.",
            file=sys.stderr,
        )

    if failures:
        return 1

    print("PASS: new hooks either have no close match or already cite similar hooks")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
