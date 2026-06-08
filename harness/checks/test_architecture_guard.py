from __future__ import annotations

from dataclasses import dataclass
from fnmatch import fnmatch


BACKEND_TEST_PATTERNS = (
    "code/backend/**/*_test.go",
    "code/backend/tests/**",
    "code/backend/internal/testutil/**",
    "code/backend/internal/**/testsupport/**",
)

FRONTEND_TEST_PATTERNS = (
    "code/frontend/src/**/*.test.ts",
    "code/frontend/src/**/*.test.tsx",
    "code/frontend/src/**/*.test.js",
    "code/frontend/src/**/*.test.jsx",
    "code/frontend/src/**/*.spec.ts",
    "code/frontend/src/**/*.spec.tsx",
    "code/frontend/src/**/*.spec.js",
    "code/frontend/src/**/*.spec.jsx",
)

SELF_CHECK_PATTERNS = (
    "scripts/check-test-architecture.sh",
    "harness/checks/check_test_architecture_guard.py",
    "harness/checks/test_architecture_guard.py",
    "harness/checks/test_test_architecture_guard.py",
    "harness/workflow-plugins/code-workflow/pre-commit-quick.d/*test*architecture*.sh",
    "harness/workflow-plugins/code-workflow/completion-full.d/*test*architecture*.sh",
    "scripts/lib/check-consistency/architecture.sh",
    "scripts/lib/check-consistency/navigation.sh",
    "harness/checks/check_local_harness_setup.sh",
    "harness/policies/script-layer-manifest.json",
    "harness/policies/local-harness-executables.txt",
)


@dataclass(frozen=True)
class CheckPlan:
    backend: bool
    frontend: bool
    self_check: bool
    backend_files: tuple[str, ...] = ()
    frontend_files: tuple[str, ...] = ()
    self_check_files: tuple[str, ...] = ()

    @property
    def has_work(self) -> bool:
        return self.backend or self.frontend or self.self_check


def matches_any(path: str, patterns: tuple[str, ...]) -> bool:
    return any(fnmatch(path, pattern) for pattern in patterns)


def plan_checks(paths: list[str] | tuple[str, ...]) -> CheckPlan:
    backend_files: list[str] = []
    frontend_files: list[str] = []
    self_check_files: list[str] = []

    for path in paths:
        if matches_any(path, BACKEND_TEST_PATTERNS):
            backend_files.append(path)
        if matches_any(path, FRONTEND_TEST_PATTERNS):
            frontend_files.append(path)
        if matches_any(path, SELF_CHECK_PATTERNS):
            self_check_files.append(path)

    return CheckPlan(
        backend=bool(backend_files),
        frontend=bool(frontend_files),
        self_check=bool(self_check_files),
        backend_files=tuple(sorted(set(backend_files))),
        frontend_files=tuple(sorted(set(frontend_files))),
        self_check_files=tuple(sorted(set(self_check_files))),
    )
