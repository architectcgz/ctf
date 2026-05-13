#!/usr/bin/env python3
from __future__ import annotations

import sys

from common import (
    classify_protected_changes,
    get_changed_files,
    load_reuse_decision_documents,
    mentioned_protected_paths,
    parse_diff_args,
    reuse_decision_destination_hint,
    validate_reuse_decision,
)


def main() -> int:
    args = parse_diff_args("verify that protected changes include a reuse decision")
    changed_files = get_changed_files(args)
    protected = classify_protected_changes(changed_files)

    if not protected:
        print("PASS: no protected reuse-first changes in diff")
        return 0

    protected_paths = sorted({path for paths in protected.values() for path in paths})
    decision_documents = load_reuse_decision_documents()
    if not decision_documents:
        print(
            f"FAIL: no reuse decision documents found; create or update {reuse_decision_destination_hint()}",
            file=sys.stderr,
        )
        print("Protected changes:", file=sys.stderr)
        for path in protected_paths:
            print(f"- {path}", file=sys.stderr)
        return 1

    coverage: dict[str, list[tuple[str, list[str]]]] = {path: [] for path in protected_paths}
    for document in decision_documents:
        covered_paths = mentioned_protected_paths(document.text, protected_paths)
        if not covered_paths:
            continue
        errors = validate_reuse_decision(document.text)
        for path in covered_paths:
            coverage[path].append((document.path, errors))

    uncovered: list[str] = []
    for path in protected_paths:
        matches = coverage[path]
        if not matches:
            uncovered.append(path)
            continue
        if not any(not errors for _, errors in matches):
            uncovered.append(path)

    if uncovered:
        print("FAIL: protected changes are not fully covered by valid reuse decision documents", file=sys.stderr)
        for path in uncovered:
            matches = coverage[path]
            if not matches:
                print(
                    f"- {path}: not referenced by any reuse decision document; use {reuse_decision_destination_hint()}",
                    file=sys.stderr,
                )
                continue

            print(f"- {path}: referenced only by incomplete reuse decision documents", file=sys.stderr)
            seen_docs: set[str] = set()
            for doc_path, errors in matches:
                if doc_path in seen_docs:
                    continue
                seen_docs.add(doc_path)
                print(f"  - {doc_path}", file=sys.stderr)
                for error in errors:
                    print(f"    - {error}", file=sys.stderr)
        return 1

    used_documents = sorted(
        {
            doc_path
            for path in protected_paths
            for doc_path, errors in coverage[path]
            if not errors
        }
    )
    print("PASS: protected changes are covered by reuse decision documents")
    for doc_path in used_documents:
        print(f"- {doc_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
