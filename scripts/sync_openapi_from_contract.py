#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

import yaml


@dataclass(frozen=True)
class Endpoint:
    method: str
    path: str  # OpenAPI-style path, e.g. /challenges/{id}
    source: str  # for diagnostics


_METHODS = ("GET", "POST", "PUT", "PATCH", "DELETE")


def _to_openapi_path(path: str) -> str:
    # Convert /api/v1/foo/:id/bar -> /foo/{id}/bar
    path = path.strip()
    if path.startswith("/api/v1/"):
        path = path[len("/api/v1") :]
    path = re.sub(r"/:([A-Za-z_][A-Za-z0-9_]*)", r"/{\1}", path)
    return path


def parse_contract_endpoints(markdown: str) -> list[Endpoint]:
    endpoints: list[Endpoint] = []

    # Example headings we support:
    # ### 2.1 POST `/api/v1/auth/login`
    # ### 8.3 POST `/api/v1/admin/users` / PUT `/api/v1/admin/users/:id`
    heading_re = re.compile(r"^###\s+(\d+\.\d+)\s+(.*)$", re.MULTILINE)
    pair_re = re.compile(rf"\b({'|'.join(_METHODS)})\b\s+`([^`]+)`")

    for match in heading_re.finditer(markdown):
        heading_no = match.group(1)
        rest = match.group(2)
        for method, raw_path in pair_re.findall(rest):
            endpoints.append(
                Endpoint(
                    method=method.upper(),
                    path=_to_openapi_path(raw_path),
                    source=f"heading {heading_no}: {method} {raw_path}",
                )
            )

    # De-dup while preserving order
    seen: set[tuple[str, str]] = set()
    unique: list[Endpoint] = []
    for ep in endpoints:
        key = (ep.method, ep.path)
        if key in seen:
            continue
        seen.add(key)
        unique.append(ep)
    return unique


def _guess_tag(path: str) -> str:
    if path.startswith("/auth/"):
        return "Auth"
    if path.startswith("/challenges"):
        return "Challenges"
    if path.startswith("/instances"):
        return "Instances"
    if path.startswith("/contests"):
        return "Contests"
    if path.startswith("/notifications"):
        return "Notifications"
    if path.startswith("/teacher"):
        return "Teacher"
    if path.startswith("/users/me/skill-profile") or path.startswith("/users/me/recommendations"):
        return "Assessment"
    if path.startswith("/users/me/") or path.startswith("/users/"):
        return "Assessment"
    if path.startswith("/reports/"):
        return "Reports"
    if path.startswith("/admin/"):
        return "Admin"
    return "Misc"


def _infer_security(method: str, path: str) -> list[dict] | None:
    if path in ("/auth/login", "/auth/register"):
        return []
    if path == "/auth/refresh":
        return [{"refreshTokenCookie": []}]
    # Everything else requires bearer by default.
    return [{"bearerAuth": []}]


def _path_params(openapi_path: str) -> list[dict]:
    params = []
    for name in re.findall(r"{([A-Za-z_][A-Za-z0-9_]*)}", openapi_path):
        params.append(
            {
                "in": "path",
                "name": name,
                "required": True,
                "schema": {"$ref": "#/components/schemas/ID"} if name.endswith("_id") or name == "id" else {"type": "string"},
            }
        )
    return params


def _skeleton_operation(method: str, openapi_path: str) -> dict:
    tag = _guess_tag(openapi_path)
    op: dict = {
        "tags": [tag],
        "summary": f"TODO: {method} {openapi_path}",
        "responses": {
            "200": {
                "description": "OK",
                "content": {
                    "application/json": {
                        "schema": {"$ref": "#/components/schemas/ApiEnvelopeNull"},
                    }
                },
            }
        },
    }

    security = _infer_security(method, openapi_path)
    if security is not None:
        op["security"] = security

    params = _path_params(openapi_path)
    if params:
        op["parameters"] = params

    # Minimal requestBody placeholders for common cases
    if method in ("POST", "PUT", "PATCH") and openapi_path.endswith("/read") is False:
        # Keep it intentionally loose; real schema belongs to hand-edited OpenAPI.
        op["requestBody"] = {
            "required": False,
            "content": {
                "application/json": {
                    "schema": {"type": "object", "description": "TODO: fill request schema"},
                }
            },
        }

    return op


def _load_openapi(openapi_path: Path) -> dict:
    return yaml.safe_load(openapi_path.read_text("utf-8"))


def _dump_yaml_fragment(doc: dict) -> str:
    return yaml.safe_dump(doc, sort_keys=False, allow_unicode=True).rstrip() + "\n"


def _indent(text: str, spaces: int) -> str:
    pad = " " * spaces
    return "".join(pad + line if line.strip() else line for line in text.splitlines(keepends=True))


def _find_path_block_range(lines: list[str], openapi_path: str) -> tuple[int, int] | None:
    """
    Returns (start_index, end_index) for the YAML block of a path item:
    - start_index points to the line: "  /foo/bar:"
    - end_index is the first index AFTER the block.
    """
    needle = f"  {openapi_path}:"
    start = None
    for i, line in enumerate(lines):
        if line.rstrip("\n") == needle:
            start = i
            break
    if start is None:
        return None

    # Block ends at next path key (2-space indent + "/") or EOF.
    for j in range(start + 1, len(lines)):
        if re.match(r"^  /[^ ].*:\s*$", lines[j]):
            return (start, j)
    return (start, len(lines))


def sync_openapi(contract_path: Path, openapi_path: Path) -> tuple[bool, list[str]]:
    contract_md = contract_path.read_text("utf-8")
    endpoints = parse_contract_endpoints(contract_md)

    openapi_text = openapi_path.read_text("utf-8")
    doc = yaml.safe_load(openapi_text)
    if not isinstance(doc, dict):
        raise ValueError("OpenAPI root must be a mapping/object.")

    paths = doc.setdefault("paths", {})
    if not isinstance(paths, dict):
        raise ValueError("OpenAPI 'paths' must be a mapping/object.")

    changed = False
    notes: list[str] = []
    lines = openapi_text.splitlines(keepends=True)

    for ep in endpoints:
        path_item = paths.get(ep.path)
        if path_item is None:
            op = _skeleton_operation(ep.method, ep.path)
            snippet = f"  {ep.path}:\n" + _indent(_dump_yaml_fragment({ep.method.lower(): op}), 4)
            if lines and not lines[-1].endswith("\n"):
                lines[-1] = lines[-1] + "\n"
            lines.append(snippet)
            # Update parsed doc for subsequent lookups in this run.
            paths[ep.path] = {ep.method.lower(): op}
            changed = True
            notes.append(f"added path {ep.method} {ep.path} ({ep.source})")
            continue

        if not isinstance(path_item, dict):
            notes.append(f"skip non-mapping path item: {ep.path}")
            continue

        m = ep.method.lower()
        if m in path_item:
            continue

        op = _skeleton_operation(ep.method, ep.path)
        method_block = _indent(_dump_yaml_fragment({m: op}), 4)

        block_range = _find_path_block_range(lines, ep.path)
        if block_range is None:
            # Fallback: append as a new path block (better than losing the endpoint).
            snippet = f"  {ep.path}:\n" + method_block
            lines.append(snippet)
            paths[ep.path] = {m: op}
        else:
            _, end = block_range
            lines.insert(end, method_block)
            # Update end offset for potential subsequent inserts is not needed here (we don't reuse `end`).
            path_item[m] = op

        changed = True
        notes.append(f"added method {ep.method} under {ep.path} ({ep.source})")

    if changed:
        new_text = "".join(lines)
        # Validate the resulting YAML is still parseable.
        yaml.safe_load(new_text)
        openapi_path.write_text(new_text, "utf-8")

    return changed, notes


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description="Sync OpenAPI paths from api-contract headings.")
    parser.add_argument(
        "--contract",
        default="docs/contracts/api-contract-v1.md",
        help="Path to api-contract markdown (relative to repo root).",
    )
    parser.add_argument(
        "--openapi",
        default="docs/contracts/openapi-v1.yaml",
        help="Path to OpenAPI yaml (relative to repo root).",
    )
    parser.add_argument("--check", action="store_true", help="Do not write; exit non-zero if changes needed.")
    args = parser.parse_args(list(argv))

    contract_path = Path(args.contract)
    openapi_path = Path(args.openapi)

    if not contract_path.exists():
        print(f"[sync_openapi] missing contract file: {contract_path}", file=sys.stderr)
        return 2
    if not openapi_path.exists():
        print(f"[sync_openapi] missing openapi file: {openapi_path}", file=sys.stderr)
        return 2

    before = openapi_path.read_text("utf-8")
    changed, notes = sync_openapi(contract_path, openapi_path)
    after = openapi_path.read_text("utf-8") if changed else before

    if notes:
        for line in notes:
            print(f"[sync_openapi] {line}")

    if args.check:
        if before != after:
            print("[sync_openapi] OpenAPI is out of sync; run sync (or let pre-commit hook do it).", file=sys.stderr)
            return 1
        return 0

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
