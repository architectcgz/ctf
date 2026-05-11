from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
CHALLENGES_ROOT = REPO_ROOT / "challenges" / "jeopardy"
PACKS_DIR = CHALLENGES_ROOT / "packs"
DIST_DIR = CHALLENGES_ROOT / "dist"
TARGETS_FILE = REPO_ROOT / "scripts" / "challenges" / "data" / "jeopardy_real_training_targets.json"
MATRIX_DOC = REPO_ROOT / "docs" / "design" / "jeopardy-80-真实训练题去重矩阵.md"
REPORT_DOC = REPO_ROOT / "docs" / "reports" / "2026-05-09-jeopardy-80-pack-verification.md"

IMAGE_TAG = "20260509"
WEB_PORT = 8080
