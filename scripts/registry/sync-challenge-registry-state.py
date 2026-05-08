#!/usr/bin/env python3
from __future__ import annotations

import argparse
import base64
import json
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path

try:
    import yaml
except Exception as exc:  # pragma: no cover - runtime dependency check
    raise SystemExit(f"缺少 PyYAML，无法解析 challenge.yml: {exc}")


REPO_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_REGISTRY_ENV = REPO_ROOT / "docker/ctf/infra/registry/ctf-platform-registry.env"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="同步题目镜像 registry 状态到 images / challenges / awd_challenges / contest_awd_services"
    )
    parser.add_argument("pack_dir", help="题目包目录")
    parser.add_argument("--image-ref", required=True, help="最终镜像引用，例如 127.0.0.1:5000/jeopardy/web-notes-download:20260409")
    parser.add_argument("--digest", default="", help="可选；镜像 digest，例如 sha256:...")
    parser.add_argument(
        "--registry-env",
        default=str(DEFAULT_REGISTRY_ENV),
        help="registry 环境变量文件，默认 docker/ctf/infra/registry/ctf-platform-registry.env",
    )
    parser.add_argument("--db-container", default="ctf-postgres", help="PostgreSQL 容器名")
    parser.add_argument("--db-name", default="ctf", help="数据库名")
    parser.add_argument("--db-user", default="postgres", help="数据库用户")
    return parser.parse_args()


def log(message: str) -> None:
    print(f"[registry-sync] {message}")


def die(message: str) -> None:
    raise SystemExit(f"错误: {message}")


def sql_literal(value: str | None) -> str:
    if value is None:
        return "NULL"
    return "'" + value.replace("'", "''") + "'"


def parse_manifest(pack_dir: Path) -> dict:
    manifest = pack_dir / "challenge.yml"
    if not manifest.is_file():
        die(f"题目包缺少 challenge.yml: {manifest}")
    data = yaml.safe_load(manifest.read_text(encoding="utf-8")) or {}
    if not isinstance(data, dict):
        die(f"challenge.yml 结构非法: {manifest}")
    return data


def normalize_mode(data: dict) -> str:
    meta = data.get("meta") or {}
    mode = str(meta.get("mode") or "jeopardy").strip().lower()
    if mode not in {"jeopardy", "awd"}:
        die(f"不支持的题目模式: {mode}")
    return mode


def load_registry_env(env_file: Path) -> dict[str, str]:
    if not env_file.is_file():
        return {}
    payload: dict[str, str] = {}
    for line in env_file.read_text(encoding="utf-8").splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#") or "=" not in stripped:
            continue
        key, value = stripped.split("=", 1)
        payload[key.strip()] = value.strip()
    return payload


def split_image_ref(image_ref: str) -> tuple[str, str]:
    image_ref = image_ref.strip()
    if not image_ref:
        die("image_ref 不能为空")
    last_slash = image_ref.rfind("/")
    last_colon = image_ref.rfind(":")
    if last_colon > last_slash:
        return image_ref[:last_colon], image_ref[last_colon + 1 :]
    return image_ref, "latest"


def extract_repository_from_ref(image_ref: str) -> str:
    ref = image_ref.removeprefix("http://").removeprefix("https://")
    name, _ = split_image_ref(ref)
    first, sep, rest = name.partition("/")
    if not sep:
        return ""
    if "." in first or ":" in first or first == "localhost":
        return rest
    return name


def fetch_manifest_digest(image_ref: str, registry_env: dict[str, str]) -> str:
    server = registry_env.get("CTF_CONTAINER_REGISTRY_SERVER", "").strip()
    scheme = registry_env.get("CTF_CONTAINER_REGISTRY_SCHEME", "http").strip() or "http"
    username = registry_env.get("CTF_CONTAINER_REGISTRY_USERNAME", "").strip()
    password = registry_env.get("CTF_CONTAINER_REGISTRY_PASSWORD", "").strip()
    if not server:
        return ""

    repository = extract_repository_from_ref(image_ref)
    if not repository:
        return ""
    _, tag = split_image_ref(image_ref)
    url = f"{scheme}://{server}/v2/{repository}/manifests/{tag}"
    request = urllib.request.Request(url, method="HEAD")
    request.add_header("Accept", "application/vnd.docker.distribution.manifest.v2+json")
    if username and password:
        token = base64.b64encode(f"{username}:{password}".encode("utf-8")).decode("ascii")
        request.add_header("Authorization", f"Basic {token}")
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            return str(response.headers.get("Docker-Content-Digest", "")).strip()
    except urllib.error.URLError:
        return ""


def run_psql(
    *,
    db_container: str,
    db_user: str,
    db_name: str,
    sql: str,
    capture: bool,
) -> str:
    cmd = [
        "docker",
        "exec",
        "-i",
        db_container,
        "psql",
        "-v",
        "ON_ERROR_STOP=1",
        "-U",
        db_user,
        "-d",
        db_name,
    ]
    if capture:
        cmd.extend(["-At", "-F", "\t"])
    cmd.extend(["-c", sql])
    completed = subprocess.run(cmd, check=True, text=True, capture_output=capture)
    return completed.stdout.strip() if capture else ""


def query_rows(args: argparse.Namespace, sql: str) -> list[list[str]]:
    output = run_psql(
        db_container=args.db_container,
        db_user=args.db_user,
        db_name=args.db_name,
        sql=sql,
        capture=True,
    )
    if not output:
        return []
    return [line.split("\t") for line in output.splitlines() if line.strip()]


def exec_sql(args: argparse.Namespace, sql: str) -> None:
    run_psql(
        db_container=args.db_container,
        db_user=args.db_user,
        db_name=args.db_name,
        sql=sql,
        capture=False,
    )


def query_single_row(args: argparse.Namespace, sql: str) -> list[str] | None:
    rows = query_rows(args, sql)
    return rows[0] if rows else None


def lookup_jeopardy_challenge(args: argparse.Namespace, slug: str) -> tuple[int, int]:
    row = query_single_row(
        args,
        f"""
SELECT id, COALESCE(image_id, 0)
FROM public.challenges
WHERE deleted_at IS NULL
  AND package_slug = {sql_literal(slug)}
ORDER BY id
LIMIT 1;
""".strip(),
    )
    if row is None:
        die(f"未找到 Jeopardy 题目 package_slug={slug}")
    return int(row[0]), int(row[1])


def lookup_awd_challenge(args: argparse.Namespace, slug: str) -> tuple[int, int]:
    row = query_single_row(
        args,
        f"""
SELECT
  id,
  COALESCE((COALESCE(NULLIF(runtime_config, '')::jsonb, '{{}}'::jsonb)->>'image_id'), '0')
FROM public.awd_challenges
WHERE deleted_at IS NULL
  AND slug = {sql_literal(slug)}
ORDER BY id
LIMIT 1;
""".strip(),
    )
    if row is None:
        die(f"未找到 AWD 题目 slug={slug}")
    return int(row[0]), int(row[1] or "0")


def lookup_image_by_id(args: argparse.Namespace, image_id: int) -> dict | None:
    if image_id <= 0:
        return None
    row = query_single_row(
        args,
        f"""
SELECT id, name, tag, COALESCE(source_type, '')
FROM public.images
WHERE id = {image_id}
LIMIT 1;
""".strip(),
    )
    if row is None:
        return None
    return {"id": int(row[0]), "name": row[1], "tag": row[2], "source_type": row[3]}


def lookup_image_by_ref(args: argparse.Namespace, name: str, tag: str) -> dict | None:
    row = query_single_row(
        args,
        f"""
SELECT id, name, tag, COALESCE(source_type, '')
FROM public.images
WHERE name = {sql_literal(name)}
  AND tag = {sql_literal(tag)}
LIMIT 1;
""".strip(),
    )
    if row is None:
        return None
    return {"id": int(row[0]), "name": row[1], "tag": row[2], "source_type": row[3]}


def update_image_row(args: argparse.Namespace, image_id: int, name: str, tag: str, digest: str) -> None:
    digest_sql = sql_literal(digest) if digest else "NULL"
    exec_sql(
        args,
        f"""
UPDATE public.images
SET name = {sql_literal(name)},
    tag = {sql_literal(tag)},
    status = 'available',
    digest = {digest_sql},
    verified_at = CURRENT_TIMESTAMP,
    last_error = '',
    deleted_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = {image_id};
""".strip(),
    )


def insert_image_row(args: argparse.Namespace, name: str, tag: str, digest: str) -> int:
    digest_sql = sql_literal(digest) if digest else "NULL"
    row = query_single_row(
        args,
        f"""
INSERT INTO public.images(
  name,
  tag,
  status,
  digest,
  source_type,
  last_error,
  verified_at,
  created_at,
  updated_at
)
VALUES (
  {sql_literal(name)},
  {sql_literal(tag)},
  'available',
  {digest_sql},
  'manual',
  '',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
RETURNING id;
""".strip(),
    )
    if row is None:
        die(f"创建镜像记录失败: {name}:{tag}")
    return int(row[0])


def ensure_image_row(args: argparse.Namespace, current_image_id: int, name: str, tag: str, digest: str) -> int:
    current = lookup_image_by_id(args, current_image_id)
    canonical = lookup_image_by_ref(args, name, tag)

    if canonical is not None and current is not None and canonical["id"] != current["id"]:
        update_image_row(args, canonical["id"], name, tag, digest)
        return int(canonical["id"])

    if current is not None:
        update_image_row(args, int(current["id"]), name, tag, digest)
        return int(current["id"])

    if canonical is not None:
        update_image_row(args, int(canonical["id"]), name, tag, digest)
        return int(canonical["id"])

    return insert_image_row(args, name, tag, digest)


def update_jeopardy_challenge(args: argparse.Namespace, challenge_id: int, image_id: int) -> None:
    exec_sql(
        args,
        f"""
UPDATE public.challenges
SET image_id = {image_id},
    updated_at = CURRENT_TIMESTAMP
WHERE id = {challenge_id};
""".strip(),
    )


def update_awd_challenge(args: argparse.Namespace, challenge_id: int, image_id: int, image_ref: str) -> None:
    exec_sql(
        args,
        f"""
UPDATE public.awd_challenges
SET runtime_config = jsonb_set(
      jsonb_set(
        COALESCE(NULLIF(runtime_config, '')::jsonb, '{{}}'::jsonb),
        '{{image_ref}}',
        to_jsonb({sql_literal(image_ref)}::text),
        true
      ),
      '{{image_id}}',
      to_jsonb({image_id}::bigint),
      true
    )::text,
    updated_at = CURRENT_TIMESTAMP
WHERE id = {challenge_id};

UPDATE public.contest_awd_services
SET runtime_config = jsonb_set(
      COALESCE(NULLIF(runtime_config, '')::jsonb, '{{}}'::jsonb),
      '{{challenge_runtime}}',
      COALESCE(COALESCE(NULLIF(runtime_config, '')::jsonb, '{{}}'::jsonb)->'challenge_runtime', '{{}}'::jsonb)
      || jsonb_build_object('image_ref', {sql_literal(image_ref)}::text, 'image_id', {image_id}),
      true
    )::text,
    service_snapshot = jsonb_set(
      COALESCE(NULLIF(service_snapshot, '')::jsonb, '{{}}'::jsonb),
      '{{runtime_config}}',
      COALESCE(COALESCE(NULLIF(service_snapshot, '')::jsonb, '{{}}'::jsonb)->'runtime_config', '{{}}'::jsonb)
      || jsonb_build_object('image_ref', {sql_literal(image_ref)}::text, 'image_id', {image_id}),
      true
    )::text,
    updated_at = CURRENT_TIMESTAMP
WHERE awd_challenge_id = {challenge_id};
""".strip(),
    )


def main() -> int:
    args = parse_args()
    pack_dir = Path(args.pack_dir).resolve()
    manifest = parse_manifest(pack_dir)
    mode = normalize_mode(manifest)
    meta = manifest.get("meta") or {}
    runtime = manifest.get("runtime") or {}
    slug = str(meta.get("slug") or "").strip()
    runtime_type = str(runtime.get("type") or "").strip()
    if not slug:
      die("challenge.yml meta.slug 不能为空")
    if runtime_type != "container":
      die(f"题目 {slug} 的 runtime.type={runtime_type!r}，不是容器题")

    image_name, image_tag = split_image_ref(args.image_ref)
    registry_env = load_registry_env(Path(args.registry_env).resolve())
    digest = args.digest.strip() or fetch_manifest_digest(args.image_ref, registry_env)

    log(f"题目目录: {pack_dir}")
    log(f"mode={mode} slug={slug}")
    log(f"image_ref={args.image_ref}")
    if digest:
        log(f"digest={digest}")

    if mode == "jeopardy":
        challenge_id, current_image_id = lookup_jeopardy_challenge(args, slug)
        image_id = ensure_image_row(args, current_image_id, image_name, image_tag, digest)
        update_jeopardy_challenge(args, challenge_id, image_id)
        log(f"已同步 Jeopardy 题目 challenge_id={challenge_id} image_id={image_id}")
        return 0

    challenge_id, current_image_id = lookup_awd_challenge(args, slug)
    image_id = ensure_image_row(args, current_image_id, image_name, image_tag, digest)
    update_awd_challenge(args, challenge_id, image_id, args.image_ref)
    log(f"已同步 AWD 题目 awd_challenge_id={challenge_id} image_id={image_id}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
