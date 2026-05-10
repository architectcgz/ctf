#!/usr/bin/env python3
from __future__ import annotations

import sys

from common import classify_protected_changes, get_changed_files, load_reuse_decision_text, parse_diff_args, validate_reuse_decision


def main() -> int:
    args = parse_diff_args("verify that protected changes include a reuse decision")
    changed_files = get_changed_files(args)
    protected = classify_protected_changes(changed_files)

    if not protected:
        print("PASS: no protected page/component/hook/api/store/schema changes in diff")
        return 0

    protected_paths = sorted({path for paths in protected.values() for path in paths})
    reuse_decision = load_reuse_decision_text()
    if not reuse_decision:
        print("FAIL: .harness/reuse-decision.md is missing or empty", file=sys.stderr)
        print("Protected changes:", file=sys.stderr)
        for path in protected_paths:
            print(f"- {path}", file=sys.stderr)
        return 1

    errors = validate_reuse_decision(reuse_decision, protected_paths)
    if errors:
        print("FAIL: reuse decision is incomplete for protected changes", file=sys.stderr)
        for error in errors:
            print(f"- {error}", file=sys.stderr)
        print("Protected changes:", file=sys.stderr)
        for path in protected_paths:
            print(f"- {path}", file=sys.stderr)
        return 1

    print("PASS: reuse decision present for protected changes")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
