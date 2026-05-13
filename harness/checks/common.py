#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from dataclasses import dataclass
from fnmatch import fnmatch
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REUSE_DECISIONS_DIR = ROOT / ".harness" / "reuse-decisions"
REUSE_HISTORY_PATH = ROOT / "harness" / "reuse" / "history.md"
REUSE_INDEX_PATH = ROOT / "harness" / "reuse" / "index.yaml"
POLICY_DIR = ROOT / "harness" / "policies"
TASK_SCOPED_REUSE_DECISION_HINT = ".harness/reuse-decisions/<task-slug>.md"

SEARCH_ROOTS = [
    "code/frontend/src/views",
    "code/frontend/src/components",
    "code/frontend/src/features",
    "code/frontend/src/widgets",
    "code/frontend/src/composables",
    "code/frontend/src/api",
    "code/frontend/src/stores",
    "code/backend/internal/module",
    "code/backend/internal/app/composition",
    "code/backend/internal/model",
    "code/backend/migrations",
]

PROTECTED_PATTERNS = {
    "page": [
        "code/frontend/src/views/**/*.vue",
        "code/frontend/src/components/**/*Page.vue",
        "code/frontend/src/components/**/*View.vue",
    ],
    "component": [
        "code/frontend/src/components/**/*.vue",
        "code/frontend/src/widgets/**/*.vue",
    ],
    "hook": [
        "code/frontend/src/composables/use*.ts",
        "code/frontend/src/features/**/model/use*.ts",
        "code/frontend/src/components/**/use*.ts",
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

PAGE_PATTERNS = PROTECTED_PATTERNS["page"]
HOOK_PATTERNS = PROTECTED_PATTERNS["hook"]
API_PATTERNS = PROTECTED_PATTERNS["api"]

PAGE_FEATURE_KEYWORDS = [
    "filter",
    "search",
    "table",
    "pagination",
    "columns",
    "modal",
    "drawer",
    "form",
    "submit",
    "query",
    "mutation",
    "detail",
    "list",
    "toolbar",
    "workspace-directory",
]

HOOK_FEATURE_KEYWORDS = [
    "query",
    "list",
    "detail",
    "filter",
    "pagination",
    "submit",
    "mutation",
    "drawer",
    "modal",
    "workspace",
    "route",
]

STOP_TOKENS = {
    "page",
    "view",
    "modal",
    "drawer",
    "form",
    "table",
    "use",
    "data",
    "item",
}

REQUEST_SIGNATURE_RE = re.compile(
    r"method:\s*['\"](?P<method>GET|POST|PUT|PATCH|DELETE)['\"].*?"
    r"url:\s*(?P<quote>['\"`])(?P<url>.*?)(?P=quote)",
    re.DOTALL,
)
API_IMPORT_RE = re.compile(r"from\s+['\"](?P<path>@/api[^'\"]+)['\"]")
TOKEN_RE = re.compile(r"[A-Z]?[a-z]+|[0-9]+")


@dataclass(frozen=True)
class ChangedFile:
    status: str
    path: str

    @property
    def is_added(self) -> bool:
        return self.status == "A"


@dataclass(frozen=True)
class ReuseDecisionDocument:
    path: str
    text: str


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


def load_reuse_decision_documents() -> list[ReuseDecisionDocument]:
    documents: list[ReuseDecisionDocument] = []
    if REUSE_DECISIONS_DIR.is_dir():
        for path in sorted(REUSE_DECISIONS_DIR.glob("*.md")):
            if not path.is_file():
                continue
            text = path.read_text(encoding="utf-8").strip()
            if text:
                documents.append(ReuseDecisionDocument(path=path.relative_to(ROOT).as_posix(), text=text))

    return documents


def load_reuse_decision_text() -> str:
    return "\n\n".join(document.text for document in load_reuse_decision_documents())


def reuse_decision_destination_hint() -> str:
    return TASK_SCOPED_REUSE_DECISION_HINT


def load_reuse_reference_text() -> str:
    parts = [load_reuse_decision_text()]
    for path in (REUSE_INDEX_PATH, REUSE_HISTORY_PATH):
        if path.is_file():
            parts.append(path.read_text(encoding="utf-8"))
    return "\n".join(part for part in parts if part)


def validate_reuse_decision(text: str, protected_paths: list[str] | None = None) -> list[str]:
    errors: list[str] = []
    required_sections = [
        "## Change type",
        "## Existing code searched",
        "## Similar implementations found",
        "## Decision",
        "## Reason",
        "## Files to modify",
    ]
    for section in required_sections:
        if section not in text:
            errors.append(f"missing section: {section}")

    if "待填写" in text or "TBD" in text or "TODO" in text:
        errors.append("reuse decision still contains placeholders")

    if not any(root in text for root in SEARCH_ROOTS):
        errors.append("reuse decision does not mention any configured search roots")

    if not any(
        decision in text
        for decision in (
            "reuse_existing",
            "extend_existing",
            "refactor_existing",
            "create_new_with_reason",
        )
    ):
        errors.append("reuse decision does not contain a valid decision value")

    if protected_paths is not None:
        for path in protected_paths:
            if path not in text:
                errors.append(f"reuse decision does not mention changed file: {path}")

    return errors


def mentioned_protected_paths(text: str, protected_paths: list[str]) -> list[str]:
    return sorted(path for path in protected_paths if path in text)


def repo_files(patterns: list[str], exclude: set[str] | None = None) -> list[str]:
    exclude = exclude or set()
    results: set[str] = set()
    for pattern in patterns:
        for file_path in ROOT.glob(pattern):
            if file_path.is_file():
                rel = file_path.relative_to(ROOT).as_posix()
                if rel not in exclude:
                    results.add(rel)
    return sorted(results)


def read_text(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def extract_name_tokens(path: str) -> set[str]:
    stem = Path(path).stem
    pieces = re.split(r"[^A-Za-z0-9]+", stem)
    tokens: set[str] = set()
    for piece in pieces:
        for token in TOKEN_RE.findall(piece):
            lower = token.lower()
            if len(lower) > 1 and lower not in STOP_TOKENS:
                tokens.add(lower)
    return tokens


def extract_keyword_hits(text: str, keywords: list[str]) -> set[str]:
    lower = text.lower()
    return {keyword for keyword in keywords if keyword in lower}


def extract_api_imports(text: str) -> set[str]:
    return {match.group("path") for match in API_IMPORT_RE.finditer(text)}


def normalize_url(url: str) -> str:
    normalized = re.sub(r"\$\{[^}]+\}", ":param", url)
    normalized = normalized.replace("\\/", "/")
    return normalized


def extract_request_signatures(text: str) -> set[tuple[str, str]]:
    signatures = set()
    for match in REQUEST_SIGNATURE_RE.finditer(text):
        signatures.add((match.group("method"), normalize_url(match.group("url"))))
    return signatures


def url_segments(url: str) -> tuple[str, ...]:
    return tuple(segment for segment in url.strip("/").split("/") if segment and segment != ":param")


def load_json_yaml(relative_path: str) -> object:
    return json.loads((ROOT / relative_path).read_text(encoding="utf-8"))


def fail(message: str) -> int:
    print(message, file=sys.stderr)
    return 1
