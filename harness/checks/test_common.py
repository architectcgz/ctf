#!/usr/bin/env python3
from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import change_detection

class ProtectedSurfaceClassificationTest(unittest.TestCase):
    def test_classifies_pages_route_files_as_page_surface(self) -> None:
        protected = change_detection.classify_protected_changes(
            [change_detection.ChangedFile(status="A", path="code/frontend/src/pages/dashboard/DashboardRoutePage.vue")]
        )

        self.assertEqual(protected["page"], ["code/frontend/src/pages/dashboard/DashboardRoutePage.vue"])

    def test_classifies_shared_layout_files_as_component_surface(self) -> None:
        protected = change_detection.classify_protected_changes(
            [change_detection.ChangedFile(status="A", path="code/frontend/src/shared/ui/layout/TopNav.vue")]
        )

        self.assertEqual(protected["component"], ["code/frontend/src/shared/ui/layout/TopNav.vue"])


if __name__ == "__main__":
    unittest.main()
