from __future__ import annotations

import tempfile
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

from .models import BuildResult, Target
from .paths import DIST_DIR, IMAGE_TAG, PACKS_DIR


def render_challenge_yml(target: Target, build: BuildResult) -> str:
    lines = [
        "api_version: v1",
        "kind: challenge",
        "",
        "meta:",
        f"  slug: {target.slug}",
        f"  title: \"{target.title}\"",
        f"  category: {target.category}",
        f"  difficulty: {target.difficulty}",
        f"  points: {target.points}",
        "  tags:",
        f"    - topic:{target.kind}",
        f"    - kp:{target.primary_skill}",
        "",
        "content:",
        "  statement: statement.md",
        "  attachments:",
    ]
    if build.attachments:
        for rel_path, name in build.attachments:
            lines.extend(
                [
                    f"    - path: {rel_path}",
                    f"      name: {name}",
                ]
            )
    else:
        lines.append("    []")

    lines.extend(["", "flag:"])
    lines.append(f"  type: {build.flag_type}")
    if build.flag_type == "static":
        lines.append("  prefix: flag")
        lines.append(f"  value: {build.flag_value}")
    else:
        lines.append("  prefix: flag")

    lines.extend(["", "hints:"])
    for idx, hint in enumerate(build.hints, start=1):
        lines.extend(
            [
                f"  - level: {idx}",
                f"    title: Hint {idx}",
                f"    content: {yaml_quote(hint)}",
            ]
        )
    if not build.hints:
        lines.append("  []")

    lines.extend(["", "runtime:"])
    lines.append(f"  type: {build.runtime_type}")
    if build.runtime_type == "container":
        lines.extend(
            [
                "  image:",
                f"    ref: 127.0.0.1:5000/jeopardy/{target.slug}:{IMAGE_TAG}",
            ]
        )
        if build.runtime_port:
            lines.extend(
                [
                    "  service:",
                    "    protocol: tcp",
                    f"    port: {build.runtime_port}",
                ]
            )
    return "\n".join(lines).rstrip() + "\n"


def yaml_quote(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace("\"", "\\\"")
    return f"\"{escaped}\""


def write_text(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content.rstrip() + "\n", encoding="utf-8")


def write_pack(target: Target, build: BuildResult) -> bool:
    pack_dir = PACKS_DIR / target.slug
    if pack_dir.exists():
        return False
    PACKS_DIR.mkdir(parents=True, exist_ok=True)
    tmp_dir = Path(tempfile.mkdtemp(prefix=f".{target.slug}.", dir=PACKS_DIR))

    write_text(tmp_dir / "challenge.yml", render_challenge_yml(target, build))
    write_text(tmp_dir / "statement.md", build.statement)
    write_text(tmp_dir / "writeup" / "solution.md", build.solution)
    write_text(tmp_dir / "writeup" / "solve.py", build.solve_py)

    for rel_path, data in build.files.items():
        dest = tmp_dir / rel_path
        dest.parent.mkdir(parents=True, exist_ok=True)
        if isinstance(data, bytes):
            dest.write_bytes(data)
            if dest.name == "challenge.bin":
                dest.chmod(0o755)
        else:
            raise TypeError(f"unexpected file payload type for {rel_path}")

    solve_path = tmp_dir / "writeup" / "solve.py"
    solve_path.chmod(0o755)
    tmp_dir.rename(pack_dir)
    return True


def write_dist_zip(target: Target) -> bool:
    DIST_DIR.mkdir(parents=True, exist_ok=True)
    zip_path = DIST_DIR / f"{target.slug}.zip"
    if zip_path.exists():
        return False
    pack_dir = PACKS_DIR / target.slug
    tmp_zip = DIST_DIR / f".{target.slug}.zip.tmp"
    with ZipFile(tmp_zip, "w", compression=ZIP_DEFLATED) as zf:
        for path in sorted(pack_dir.rglob("*")):
            if path.is_file():
                arcname = Path(target.slug) / path.relative_to(pack_dir)
                zf.write(path, arcname.as_posix())
    tmp_zip.rename(zip_path)
    return True
