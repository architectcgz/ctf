#!/usr/bin/env python3
from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import common


def build_reuse_decision(existing_code_searched: str) -> str:
    return f"""# Reuse Decision

## Change type
layout

## Existing code searched
{existing_code_searched}

## Similar implementations found
- `code/frontend/src/shared/ui/layout/TopNav.vue`

## Decision
extend_existing

## Reason
reuse the existing shared workspace shell.

## Files to modify
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
"""


class ReuseDecisionValidationTest(unittest.TestCase):
    def test_accepts_real_shared_file_path(self) -> None:
        document = build_reuse_decision("- `code/frontend/src/shared/ui/layout/TopNav.vue`")

        self.assertEqual(common.validate_reuse_decision(document), [])

    def test_accepts_policy_root_shorthand(self) -> None:
        document = build_reuse_decision("- `code/frontend/src/pages/...`")

        self.assertEqual(common.validate_reuse_decision(document), [])

    def test_rejects_entries_without_configured_root_or_existing_repo_path(self) -> None:
        document = build_reuse_decision("- `docs/non-existent-guideline.md`")

        errors = common.validate_reuse_decision(document)

        self.assertIn(
            "reuse decision Existing code searched must mention a configured search root or an existing repo file/directory",
            errors,
        )

    def test_classifies_pages_route_files_as_page_surface(self) -> None:
        protected = common.classify_protected_changes(
            [common.ChangedFile(status="A", path="code/frontend/src/pages/dashboard/DashboardRoutePage.vue")]
        )

        self.assertEqual(protected["page"], ["code/frontend/src/pages/dashboard/DashboardRoutePage.vue"])

    def test_classifies_shared_layout_files_as_component_surface(self) -> None:
        protected = common.classify_protected_changes(
            [common.ChangedFile(status="A", path="code/frontend/src/shared/ui/layout/TopNav.vue")]
        )

        self.assertEqual(protected["component"], ["code/frontend/src/shared/ui/layout/TopNav.vue"])


if __name__ == "__main__":
    unittest.main()
