#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${BACKEND_DIR}/../.." && pwd)"

APP_ENV="${APP_ENV:-dev}"
CTF_CONTAINER_FLAG_GLOBAL_SECRET="${CTF_CONTAINER_FLAG_GLOBAL_SECRET:-dev-integration-secret-123456789}"
CTF_FLAG_SECRET="${CTF_FLAG_SECRET:-12345678901234567890123456789012}"
AIR_BIN="${AIR_BIN:-air}"
AIR_CONFIG="${AIR_CONFIG:-${BACKEND_DIR}/.air.toml}"
MIGRATE_VERSION="${MIGRATE_VERSION:-v4.18.3}"
MIGRATE_TAGS="${MIGRATE_TAGS:-postgres}"
MIGRATE_GOPROXY="${MIGRATE_GOPROXY:-https://goproxy.cn,direct}"

WITH_MIGRATE=false
INFRA_MODE=""
RUN_MODE="run"

usage() {
  cat <<'EOF'
用法:
  ./scripts/dev-run.sh [--infra] [--infra-shared] [--migrate] [--hot]

说明:
  --infra         启动开发依赖容器
  --infra-shared  启动共享开发依赖容器
  --migrate       启动前执行数据库迁移
  --hot           使用 air 热重载启动后端

可覆盖环境变量:
  APP_ENV
  CTF_CONTAINER_FLAG_GLOBAL_SECRET
  CTF_FLAG_SECRET
  AIR_BIN
  AIR_CONFIG
EOF
}

resolve_compose_file() {
  local candidate

  candidate="${REPO_ROOT}/docker/ctf/docker-compose.dev.yml"
  if [[ -f "${candidate}" ]]; then
    printf '%s\n' "${candidate}"
    return 0
  fi

  while IFS= read -r worktree_path; do
    candidate="${worktree_path}/docker/ctf/docker-compose.dev.yml"
    if [[ -f "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done < <(git -C "${REPO_ROOT}" worktree list --porcelain | awk '/^worktree / {print substr($0, 10)}')

  return 1
}

start_infra() {
  local compose_file

  if ! compose_file="$(resolve_compose_file)"; then
    echo "未找到 docker/ctf/docker-compose.dev.yml，无法自动启动开发依赖容器。" >&2
    echo "如果你在主工作区维护了该文件，请在主工作区执行 docker compose，或先手动启动 PostgreSQL / Redis。" >&2
    exit 1
  fi

  docker compose -f "${compose_file}" up -d ctf-postgres ctf-redis
}

apply_runtime_defaults() {
  if [[ -n "${INFRA_MODE}" ]]; then
    export CTF_POSTGRES_HOST="${CTF_POSTGRES_HOST:-127.0.0.1}"
    export CTF_POSTGRES_PORT="${CTF_POSTGRES_PORT:-15432}"
    export CTF_REDIS_ADDR="${CTF_REDIS_ADDR:-127.0.0.1:16379}"
  fi

  if [[ -z "${CTF_HTTP_PORT:-}" ]] && ss -ltn 'sport = :8080' | grep -q LISTEN; then
    export CTF_HTTP_PORT=18080
    echo "检测到 8080 已被占用，自动改用 CTF_HTTP_PORT=${CTF_HTTP_PORT}"
  fi
}

run_migrations() {
  GOPROXY="${MIGRATE_GOPROXY}" \
    go run -tags "${MIGRATE_TAGS}" github.com/golang-migrate/migrate/v4/cmd/migrate@"${MIGRATE_VERSION}" \
    -path ./migrations \
    -database "${MIGRATE_DATABASE_URL:-postgres://postgres:postgres123456@127.0.0.1:15432/ctf?sslmode=disable}" \
    up
}

start_backend() {
  if [[ "${RUN_MODE}" == "hot" ]]; then
    if ! command -v "${AIR_BIN}" >/dev/null 2>&1; then
      echo "air 未安装，请先执行: go install github.com/air-verse/air@latest" >&2
      exit 1
    fi

    exec env \
      APP_ENV="${APP_ENV}" \
      CTF_CONTAINER_FLAG_GLOBAL_SECRET="${CTF_CONTAINER_FLAG_GLOBAL_SECRET}" \
      CTF_FLAG_SECRET="${CTF_FLAG_SECRET}" \
      "${AIR_BIN}" -c "${AIR_CONFIG}"
  fi

  exec env \
    APP_ENV="${APP_ENV}" \
    CTF_CONTAINER_FLAG_GLOBAL_SECRET="${CTF_CONTAINER_FLAG_GLOBAL_SECRET}" \
    CTF_FLAG_SECRET="${CTF_FLAG_SECRET}" \
    go run ./cmd/api
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --infra)
      INFRA_MODE="local"
      shift
      ;;
    --infra-shared)
      INFRA_MODE="shared"
      shift
      ;;
    --migrate)
      WITH_MIGRATE=true
      shift
      ;;
    --hot)
      RUN_MODE="hot"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

cd "${BACKEND_DIR}"

case "${INFRA_MODE}" in
  local)
    start_infra
    ;;
  shared)
    start_infra
    ;;
esac

if [[ "${WITH_MIGRATE}" == "true" ]]; then
  run_migrations
fi

apply_runtime_defaults

echo "APP_ENV=${APP_ENV}"
if [[ -n "${CTF_POSTGRES_PORT:-}" ]]; then
  echo "CTF_POSTGRES_PORT=${CTF_POSTGRES_PORT}"
fi
if [[ -n "${CTF_REDIS_ADDR:-}" ]]; then
  echo "CTF_REDIS_ADDR=${CTF_REDIS_ADDR}"
fi
if [[ -n "${CTF_HTTP_PORT:-}" ]]; then
  echo "CTF_HTTP_PORT=${CTF_HTTP_PORT}"
fi
echo "启动后端服务 (${RUN_MODE})..."

start_backend
