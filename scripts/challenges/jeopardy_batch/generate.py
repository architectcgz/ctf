from __future__ import annotations

import argparse

from .models import BuildResult, Target
from .pack_io import write_dist_zip, write_pack
from .registry import BUILDERS
from .targets import filter_targets, load_targets, print_summary, sync_matrix_doc, validate_targets
from .paths import MATRIX_DOC, REPO_ROOT

def main() -> None:
    parser = argparse.ArgumentParser(description="生成 Jeopardy 80 扩展题包")
    parser.add_argument("--write", action="store_true", help="实际写入题包与 dist")
    parser.add_argument("--slug", action="append", default=[], help="只生成指定 slug，可重复传入")
    parser.add_argument(
        "--sync-docs",
        action="store_true",
        help="只同步去重矩阵文档，不生成题包",
    )
    args = parser.parse_args()

    targets = load_targets()
    validate_targets(targets)
    sync_matrix_doc(targets)

    if args.sync_docs:
        print(f"已同步: {MATRIX_DOC.relative_to(REPO_ROOT)}")
        return

    chosen = filter_targets(targets, args.slug)
    print_summary(chosen)
    if not args.write:
        return

    for target in chosen:
        build = build_target(target)
        pack_written = write_pack(target, build)
        zip_written = write_dist_zip(target)
        state = "write" if pack_written or zip_written else "skip"
        print(f"[{state}] {target.slug}")

def build_target(target: Target) -> BuildResult:
    builder = BUILDERS.get(target.kind)
    if builder is None:
        raise SystemExit(f"缺少 builder: {target.kind}")
    result = builder(target)
    if result.flag_type == "static" and not result.flag_value:
        raise SystemExit(f"静态题缺少 flag: {target.slug}")
    return result
