#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

APP_ENV="${APP_ENV:-dev}"
CTF_CONTAINER_FLAG_GLOBAL_SECRET="${CTF_CONTAINER_FLAG_GLOBAL_SECRET:-dev-integration-secret-123456789}"
CTF_FLAG_SECRET="${CTF_FLAG_SECRET:-12345678901234567890123456789012}"

WITH_MIGRATE=false
INFRA_MODE=""

usage() {
  cat <<'EOF'
用法:
  ./scripts/dev-run.sh [--infra] [--infra-shared] [--migrate]

说明:
  --infra         启动项目自带的 PostgreSQL/Redis
  --infra-shared  启动共享基础设施
  --migrate       启动前执行数据库迁移

可覆盖环境变量:
  APP_ENV
  CTF_CONTAINER_FLAG_GLOBAL_SECRET
  CTF_FLAG_SECRET
EOF
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
    make infra-up
    ;;
  shared)
    make infra-up-shared
    ;;
esac

if [[ "${WITH_MIGRATE}" == "true" ]]; then
  make migrate-up
fi

echo "APP_ENV=${APP_ENV}"
echo "启动后端服务..."

exec env \
  APP_ENV="${APP_ENV}" \
  CTF_CONTAINER_FLAG_GLOBAL_SECRET="${CTF_CONTAINER_FLAG_GLOBAL_SECRET}" \
  CTF_FLAG_SECRET="${CTF_FLAG_SECRET}" \
  make run
