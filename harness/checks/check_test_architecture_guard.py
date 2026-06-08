#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).resolve().parent))

import test_architecture_guard


def run_git(*args: str) -> str:
    result = subprocess.run(
        ["git", *args],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout


def changed_working_tree_files() -> list[str]:
    tracked = run_git("diff", "--name-only", "--diff-filter=ACMR", "HEAD")
    untracked = run_git("ls-files", "--others", "--exclude-standard")
    return sorted(set(filter(None, [*tracked.splitlines(), *untracked.splitlines()])))


def changed_staged_files() -> list[str]:
    output = run_git("diff", "--cached", "--name-only", "--diff-filter=ACMR")
    return sorted(set(filter(None, output.splitlines())))


def changed_from_env() -> list[str]:
    raw = os.environ.get("WORKFLOW_CHANGED_FILES", "")
    return sorted(set(filter(None, raw.splitlines())))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run test architecture checks for changed test code.")
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--working-tree", action="store_true", help="inspect working tree changes")
    group.add_argument("--staged", action="store_true", help="inspect staged changes")
    group.add_argument("--changed-from-env", action="store_true", help="read WORKFLOW_CHANGED_FILES")
    group.add_argument("--files", nargs="+", help="inspect explicit files")
    return parser.parse_args()


def run_check(label: str, command: list[str]) -> int:
    print(f"[test-architecture] {label}", flush=True)
    result = subprocess.run(command, cwd=ROOT)
    if result.returncode == 0:
        print(f"  PASS - {label}", flush=True)
    else:
        print(f"  FAIL - {label}", flush=True)
    return result.returncode


def main() -> int:
    args = parse_args()

    if args.files:
        plan = test_architecture_guard.plan_checks(args.files)
    elif args.staged:
        plan = test_architecture_guard.plan_checks(changed_staged_files())
    elif args.changed_from_env:
        plan = test_architecture_guard.plan_checks(changed_from_env())
    else:
        plan = test_architecture_guard.plan_checks(changed_working_tree_files())

    if not plan.has_work:
        print("[test-architecture] no test architecture-sensitive changes detected")
        return 0

    fail = 0

    if plan.self_check:
        fail |= run_check(
            "guard self tests",
            ["python3", "-m", "unittest", "harness.checks.test_test_architecture_guard"],
        )

    if plan.frontend:
        command = ["bash", "scripts/check-frontend-test-guard.sh", "--files", *plan.frontend_files]
        fail |= run_check("frontend test guard", command)

    if plan.backend:
        fail |= run_check(
            "backend test architecture guard",
            ["bash", "-lc", "cd code/backend && go test ./tests/architecture -count=1"],
        )

    return 1 if fail else 0


if __name__ == "__main__":
    raise SystemExit(main())
