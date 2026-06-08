from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import test_architecture_guard


class TestArchitectureGuardClassificationTest(unittest.TestCase):
    def test_backend_test_files_trigger_backend_guard(self) -> None:
        plan = test_architecture_guard.plan_checks(
            ["code/backend/internal/module/runtime/service_test.go"]
        )

        self.assertTrue(plan.backend)
        self.assertFalse(plan.frontend)

    def test_backend_testsupport_files_trigger_backend_guard(self) -> None:
        plan = test_architecture_guard.plan_checks(
            ["code/backend/internal/module/contest/testsupport/db.go"]
        )

        self.assertTrue(plan.backend)

    def test_frontend_test_files_trigger_frontend_guard(self) -> None:
        plan = test_architecture_guard.plan_checks(
            ["code/frontend/src/features/auth/model/useAuth.test.ts"]
        )

        self.assertTrue(plan.frontend)
        self.assertFalse(plan.backend)

    def test_guard_maintenance_files_trigger_self_check(self) -> None:
        plan = test_architecture_guard.plan_checks(
            [
                "harness/checks/check_test_architecture_guard.py",
                "harness/checks/check_local_harness_setup.sh",
                "scripts/lib/check-consistency/navigation.sh",
                "harness/policies/script-layer-manifest.json",
            ]
        )

        self.assertTrue(plan.self_check)
        self.assertFalse(plan.backend)
        self.assertFalse(plan.frontend)

    def test_non_test_implementation_files_do_not_trigger_test_architecture(self) -> None:
        plan = test_architecture_guard.plan_checks(
            [
                "code/backend/internal/module/runtime/service.go",
                "code/frontend/src/features/auth/model/useAuth.ts",
            ]
        )

        self.assertFalse(plan.backend)
        self.assertFalse(plan.frontend)
        self.assertFalse(plan.self_check)

    def test_workflow_stages_run_test_architecture_guard(self) -> None:
        root = Path(__file__).resolve().parents[2]
        expected_calls = {
            root / "harness/workflow-plugins/code-workflow/pre-commit-quick.d/40-test-architecture.sh": "scripts/check-test-architecture.sh --changed-from-env",
            root / "harness/workflow-plugins/code-workflow/completion-full.d/40-test-architecture.sh": "scripts/check-test-architecture.sh --working-tree",
        }

        for path, expected in expected_calls.items():
            with self.subTest(path=path):
                self.assertTrue(path.exists(), f"missing workflow plugin: {path}")
                self.assertIn(expected, path.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
