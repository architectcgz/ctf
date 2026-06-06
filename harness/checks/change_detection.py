#!/usr/bin/env python3
from __future__ import annotations

import argparse
import subprocess
from dataclasses import dataclass
from fnmatch import fnmatch
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

PROTECTED_PATTERNS = {
    "page": [
        "code/frontend/src/pages/**/*.vue",
        "code/frontend/src/views/**/*.vue",
        "code/frontend/src/components/**/*Page.vue",
        "code/frontend/src/components/**/*View.vue",
    ],
    "component": [
        "code/frontend/src/components/**/*.vue",
        "code/frontend/src/shared/ui/**/*.vue",
        "code/frontend/src/entities/**/ui/**/*.vue",
        "code/frontend/src/features/**/ui/**/*.vue",
        "code/frontend/src/widgets/**/*.vue",
    ],
    "hook": [
        "code/frontend/src/composables/use*.ts",
        "code/frontend/src/shared/**/use*.ts",
        "code/frontend/src/entities/**/model/use*.ts",
        "code/frontend/src/features/**/model/use*.ts",
        "code/frontend/src/components/**/use*.ts",
        "code/frontend/src/widgets/**/model/use*.ts",
    ],
    "service": [
        "code/backend/internal/**/*service*.go",
        "code/frontend/src/features/**/model/**/*Service*.ts",
    ],
    "handler": [
        "code/backend/internal/module/**/api/**/*.go",
        "code/backend/internal/handler/**/*.go",
    ],
    "repository": [
        "code/backend/internal/module/**/infrastructure/**/*repository*.go",
        "code/backend/internal/module/**/infrastructure/repository.go",
    ],
    "port": [
        "code/backend/internal/module/**/ports/**/*.go",
    ],
    "job": [
        "code/backend/internal/module/**/application/**/*job*.go",
        "code/backend/internal/module/**/application/jobs/**/*.go",
        "code/backend/internal/module/**/application/**/*worker*.go",
    ],
    "mapper": [
        "code/backend/internal/module/**/*mapper*.go",
        "code/backend/internal/shared/mapper*/**/*.go",
    ],
    "readmodel": [
        "code/backend/internal/module/*_readmodel/**/*.go",
    ],
    "composition": [
        "code/backend/internal/app/composition/**/*.go",
        "code/backend/internal/module/**/runtime/module.go",
    ],
    "store": [
        "code/frontend/src/stores/**/*.ts",
    ],
    "api": [
        "code/frontend/src/api/**/*.ts",
    ],
    "form": [
        "code/frontend/src/**/*Form*.vue",
        "code/frontend/src/**/*Form*.ts",
    ],
    "table": [
        "code/frontend/src/**/*Table*.vue",
        "code/frontend/src/**/*Table*.ts",
    ],
    "modal": [
        "code/frontend/src/**/*Modal*.vue",
        "code/frontend/src/**/*Drawer*.vue",
        "code/frontend/src/**/*Overlay*.vue",
    ],
    "layout": [
        "code/frontend/src/components/layout/**/*.vue",
        "code/frontend/src/**/*Layout*.vue",
    ],
    "schema": [
        "code/frontend/src/**/*schema*.ts",
        "code/backend/**/*.sql",
        "code/backend/**/*schema*.go",
        "challenges/**/*.yml",
        "challenges/**/*.yaml",
    ],
    "migration": [
        "code/backend/migrations/**/*.sql",
        "code/backend/internal/module/**/migrations/**/*.sql",
    ],
}


@dataclass(frozen=True)
class ChangedFile:
    status: str
    path: str


def run_git(*args: str) -> str:
    result = subprocess.run(
        ["git", *args],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout


def parse_diff_args(description: str) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("--staged", action="store_true", help="inspect staged diff")
    parser.add_argument("--base", help="base revision for compare mode")
    parser.add_argument("--head", default="HEAD", help="head revision for compare mode")
    args = parser.parse_args()

    if args.staged and args.base:
        parser.error("--staged and --base cannot be used together")

    return args


def get_changed_files(args: argparse.Namespace) -> list[ChangedFile]:
    if args.base:
        output = run_git("diff", "--name-status", "--diff-filter=ACMR", f"{args.base}...{args.head}")
    else:
        output = run_git("diff", "--cached", "--name-status", "--diff-filter=ACMR")

    changed: list[ChangedFile] = []
    for raw_line in output.splitlines():
        if not raw_line.strip():
            continue
        parts = raw_line.split("\t")
        status = parts[0][0]
        path = parts[-1]
        changed.append(ChangedFile(status=status, path=path))
    return changed


def matches_any(path: str, patterns: list[str]) -> bool:
    return any(fnmatch(path, pattern) for pattern in patterns)


def classify_protected_changes(changed_files: list[ChangedFile]) -> dict[str, list[str]]:
    matches: dict[str, list[str]] = {}
    for changed in changed_files:
        for change_type, patterns in PROTECTED_PATTERNS.items():
            if matches_any(changed.path, patterns):
                matches.setdefault(change_type, []).append(changed.path)
    return matches
