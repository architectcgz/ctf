#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
MANIFEST_PATH = ROOT / "harness" / "policies" / "script-layer-manifest.json"
SCRIPTS_DIR = ROOT / "scripts"
TOOLS_DIR = ROOT / "tools"
ENTRYPOINT_RE = re.compile(r"^(check|start|run|install|uninstall|doctor)-[a-z0-9-]+\.(sh|py)$")
ENTRYPOINT_PREFIXES = ("check-", "start-", "run-", "install-", "uninstall-", "doctor-")


def load_manifest() -> dict[str, object]:
    try:
        return json.loads(MANIFEST_PATH.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise SystemExit(f"FAIL: missing manifest: {MANIFEST_PATH.relative_to(ROOT)}")
    except json.JSONDecodeError as exc:
        raise SystemExit(f"FAIL: invalid manifest JSON: {MANIFEST_PATH.relative_to(ROOT)}: {exc}")


def rel(path: Path) -> str:
    return path.relative_to(ROOT).as_posix()


def fail(message: str, failures: list[str]) -> None:
    failures.append(message)


def check_script_root(manifest: dict[str, object], failures: list[str]) -> None:
    registered = {Path(path) for path in manifest["stable_entrypoints"]}
    actual = {
        path.relative_to(ROOT)
        for path in SCRIPTS_DIR.iterdir()
        if path.is_file() and path.name != "AGENTS.md"
    }

    missing = sorted(rel(ROOT / path) for path in registered - actual)
    extra = sorted(rel(ROOT / path) for path in actual - registered)
    if missing:
        fail("stable entrypoints missing from scripts/: " + ", ".join(missing), failures)
    if extra:
        fail("unregistered top-level files under scripts/: " + ", ".join(extra), failures)

    for path in sorted(actual):
        basename = path.name
        if not ENTRYPOINT_RE.fullmatch(basename):
            fail(
                f"scripts/ top-level file must use stable entrypoint naming: {path.as_posix()}",
                failures,
            )


def check_tools_root(manifest: dict[str, object], failures: list[str]) -> None:
    registered = {Path(path) for path in manifest["project_tools"]}
    actual = {
        path.relative_to(ROOT)
        for path in TOOLS_DIR.iterdir()
        if path.is_file() and path.name != "AGENTS.md"
    }
    extra_dirs = sorted(rel(path) for path in TOOLS_DIR.iterdir() if path.is_dir())

    missing = sorted(rel(ROOT / path) for path in registered - actual)
    extra = sorted(rel(ROOT / path) for path in actual - registered)
    if missing:
        fail("project tools missing from tools/: " + ", ".join(missing), failures)
    if extra:
        fail("unregistered top-level files under tools/: " + ", ".join(extra), failures)
    if extra_dirs:
        fail("unregistered directories under tools/: " + ", ".join(extra_dirs), failures)

    for path in sorted(actual):
        basename = path.name
        if basename.startswith(ENTRYPOINT_PREFIXES):
            fail(
                f"tools/ file must not look like a stable entrypoint: {path.as_posix()}",
                failures,
            )


def check_script_namespaces(manifest: dict[str, object], failures: list[str]) -> None:
    namespaces = {Path(path) for path in manifest["script_namespaces"].keys()}
    actual = {
        path.relative_to(ROOT)
        for path in SCRIPTS_DIR.iterdir()
        if path.is_dir()
    }

    missing = sorted(rel(ROOT / path) for path in namespaces - actual)
    extra = sorted(rel(ROOT / path) for path in actual - namespaces)
    if missing:
        fail("registered script namespaces missing from scripts/: " + ", ".join(missing), failures)
    if extra:
        fail("unregistered directories under scripts/: " + ", ".join(extra), failures)


def check_manifest_paths_exist(manifest: dict[str, object], failures: list[str]) -> None:
    all_paths = list(manifest["stable_entrypoints"]) + list(manifest["project_tools"]) + list(manifest["script_namespaces"].keys())
    for raw in all_paths:
        path = ROOT / raw
        if not path.exists():
            fail(f"manifest entry does not exist: {raw}", failures)


def main() -> int:
    manifest = load_manifest()
    failures: list[str] = []

    check_manifest_paths_exist(manifest, failures)
    check_script_root(manifest, failures)
    check_tools_root(manifest, failures)
    check_script_namespaces(manifest, failures)

    if failures:
        print("FAIL: script layer conventions drift detected", file=sys.stderr)
        for item in failures:
            print(f"- {item}", file=sys.stderr)
        return 1

    print("PASS: script layer conventions are aligned")
    print("- scripts/ top-level files stay as stable entrypoints")
    print("- tools/ top-level files stay as project tools")
    print("- scripts/ subdirectories stay within registered namespaces")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
