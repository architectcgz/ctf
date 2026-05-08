#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
REGISTRY_INFRA_ROOT="${REPO_ROOT}/docker/ctf/infra/registry"
LEGACY_REGISTRY_ROOT="${HOME}/ctf-registry"
LEGACY_REGISTRY_DATA_DIR="${LEGACY_REGISTRY_ROOT}/data"
LEGACY_REGISTRY_AUTH_DIR="${LEGACY_REGISTRY_ROOT}/auth"
LEGACY_PLATFORM_ENV_FILE="${LEGACY_REGISTRY_AUTH_DIR}/ctf-platform-registry.env"
LEGACY_COMPOSE_REGISTRY_ENV_FILE="${REPO_ROOT}/docker/ctf/.env.registry"
CONFIG_FILE="${CONFIG_FILE:-${SCRIPT_DIR}/deploy-private-registry.conf}"
CONFIG_FILE_EXPLICIT=false
REGISTRY_NAME="${REGISTRY_NAME:-ctf-registry}"
REGISTRY_PORT="${REGISTRY_PORT:-5000}"
REGISTRY_SERVER="${REGISTRY_SERVER:-}"
REGISTRY_SCHEME="${REGISTRY_SCHEME:-http}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-ctf}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-}"
REGISTRY_DATA_DIR="${REGISTRY_DATA_DIR:-${REGISTRY_INFRA_ROOT}/runtime/data}"
REGISTRY_AUTH_DIR="${REGISTRY_AUTH_DIR:-${REGISTRY_INFRA_ROOT}/runtime/auth}"
CTF_COMPOSE_REGISTRY_ENV_FILE="${CTF_COMPOSE_REGISTRY_ENV_FILE:-${REGISTRY_INFRA_ROOT}/ctf-platform-registry.env}"
REGISTRY_IMAGE="${REGISTRY_IMAGE:-registry:2}"
HTPASSWD_IMAGE="${HTPASSWD_IMAGE:-httpd:2.4-alpine}"
REGISTRY_RESTART_POLICY="${REGISTRY_RESTART_POLICY:-always}"
FORCE_RECREATE=false
REGISTRY_COMPOSE_SERVICE="${REGISTRY_COMPOSE_SERVICE:-ctf-registry}"

usage() {
  cat <<'EOF'
用法:
  scripts/registry/deploy-private-registry.sh [选项]

说明:
  部署一个带 Basic Auth 的 Docker Registry，并生成平台后端可直接加载的环境变量文件。
  默认监听 127.0.0.1:5000/宿主机 5000 端口，适合单机部署或毕设演示。

选项:
  --config FILE          读取配置文件，默认 scripts/registry/deploy-private-registry.conf
  --name NAME             Registry 容器名，默认 ctf-registry
  --port PORT             宿主机端口，默认 5000
  --server HOST:PORT      平台配置中的 registry server，默认 127.0.0.1:<port>
  --scheme http|https     供构建脚本访问 registry API 使用的协议，默认 http
  --username USER         Registry 用户名，默认 ctf
  --password PASSWORD     Registry 密码；未提供时自动生成
  --data-dir DIR          镜像数据目录，默认 docker/ctf/infra/registry/runtime/data
  --auth-dir DIR          认证文件目录，默认 docker/ctf/infra/registry/runtime/auth
  --ctf-env-file FILE     平台与 ctf-api 共用的 registry env，默认 docker/ctf/infra/registry/ctf-platform-registry.env
  --no-ctf-env-file       不写入 ctf compose registry env
  --image IMAGE           Registry 镜像，默认 registry:2
  --htpasswd-image IMAGE  用于生成 htpasswd 的镜像，默认 httpd:2.4-alpine
  --force-recreate        若同名容器已存在，先删除后重建
  -h, --help              显示帮助

配置文件:
  cp scripts/registry/deploy-private-registry.conf.example scripts/registry/deploy-private-registry.conf
  编辑 scripts/registry/deploy-private-registry.conf 后直接运行:
  scripts/registry/deploy-private-registry.sh

也可通过同名环境变量或命令行参数覆盖默认值，例如:
  REGISTRY_PORT=15000 REGISTRY_PASSWORD='change-me' scripts/registry/deploy-private-registry.sh

输出:
  docker/ctf/infra/registry/ctf-platform-registry.env

该文件可供后端部署环境加载，内容形如:
  CTF_CONTAINER_REGISTRY_ENABLED=true
  CTF_CONTAINER_REGISTRY_BUILD_ENABLED=true
  CTF_CONTAINER_REGISTRY_SERVER=127.0.0.1:5000
  CTF_CONTAINER_REGISTRY_SCHEME=http
  CTF_CONTAINER_REGISTRY_USERNAME=ctf
  CTF_CONTAINER_REGISTRY_PASSWORD=...

当前实现会把 registry 作为 `docker/ctf/docker-compose.dev.yml` 里的 `ctf-registry`
service 启动到同一个 `ctf` Compose 项目里，不再长期依赖手工 `docker run`。
EOF
}

die() {
  echo "错误: $*" >&2
  exit 1
}

log_info() {
  printf '[registry] %s\n' "$*"
}

log_success() {
  printf '[registry] 完成: %s\n' "$*"
}

log_warn() {
  printf '[registry] 注意: %s\n' "$*" >&2
}

require_command() {
  log_info "检查命令: $1"
  if ! command -v "$1" >/dev/null 2>&1; then
    die "缺少命令: $1"
  fi
}

load_config_file() {
  if [[ -f "${CONFIG_FILE}" ]]; then
    log_info "读取配置文件: ${CONFIG_FILE}"
    # shellcheck source=/dev/null
    source "${CONFIG_FILE}"
    return 0
  fi

  if [[ "${CONFIG_FILE_EXPLICIT}" == "true" ]]; then
    die "配置文件不存在: ${CONFIG_FILE}"
  fi

  log_info "未找到默认配置文件，使用脚本默认值和命令行参数"
}

generate_password() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 24
    return 0
  fi
  if [[ -r /dev/urandom ]]; then
    local password
    password="$(LC_ALL=C tr -dc 'A-Za-z0-9_@%+=:,.~-' </dev/urandom | head -c 32 || true)"
    [[ -n "${password}" ]] || die "随机密码生成失败"
    printf '%s\n' "${password}"
    return 0
  fi
  die "未提供 --password，且无法生成随机密码"
}

ensure_positive_port() {
  local port="$1"
  if ! [[ "${port}" =~ ^[0-9]+$ ]] || (( port < 1 || port > 65535 )); then
    die "端口必须在 1-65535 之间: ${port}"
  fi
}

container_exists() {
  docker ps -a --format '{{.Names}}' | grep -Fxq "$1"
}

write_platform_env() {
  local path="$1"

  umask 077
  cat >"${path}" <<EOF
CTF_CONTAINER_REGISTRY_ENABLED=true
CTF_CONTAINER_REGISTRY_BUILD_ENABLED=true
CTF_CONTAINER_REGISTRY_SERVER=${REGISTRY_SERVER}
CTF_CONTAINER_REGISTRY_SCHEME=${REGISTRY_SCHEME}
CTF_CONTAINER_REGISTRY_USERNAME=${REGISTRY_USERNAME}
CTF_CONTAINER_REGISTRY_PASSWORD=${REGISTRY_PASSWORD}
EOF
  chmod 600 "${path}"
}

merge_platform_env_file() {
  local env_file="$1"
  local current_server current_username current_password current_scheme
  local file_server file_username file_password file_scheme

  current_server="${REGISTRY_SERVER}"
  current_username="${REGISTRY_USERNAME}"
  current_password="${REGISTRY_PASSWORD}"
  current_scheme="${REGISTRY_SCHEME}"

  # shellcheck source=/dev/null
  source "${env_file}"

  file_server="${CTF_CONTAINER_REGISTRY_SERVER:-${REGISTRY_SERVER:-}}"
  file_username="${CTF_CONTAINER_REGISTRY_USERNAME:-${REGISTRY_USERNAME:-}}"
  file_password="${CTF_CONTAINER_REGISTRY_PASSWORD:-${REGISTRY_PASSWORD:-}}"
  file_scheme="${CTF_CONTAINER_REGISTRY_SCHEME:-${REGISTRY_SCHEME:-}}"

  REGISTRY_SERVER="${current_server:-${file_server}}"
  REGISTRY_USERNAME="${current_username:-${file_username}}"
  REGISTRY_PASSWORD="${current_password:-${file_password}}"
  REGISTRY_SCHEME="${current_scheme:-${file_scheme}}"
}

load_existing_platform_env() {
  local env_file

  for env_file in \
    "${CTF_COMPOSE_REGISTRY_ENV_FILE}" \
    "${LEGACY_COMPOSE_REGISTRY_ENV_FILE}" \
    "${REGISTRY_AUTH_DIR}/ctf-platform-registry.env" \
    "${LEGACY_PLATFORM_ENV_FILE}"; do
    if [[ -z "${env_file}" || ! -f "${env_file}" ]]; then
      continue
    fi
    log_info "复用现有 registry 环境变量文件: ${env_file}"
    merge_platform_env_file "${env_file}"
    return 0
  done
}

dir_has_entries() {
  local dir="$1"
  [[ -d "${dir}" ]] && [[ -n "$(ls -A "${dir}" 2>/dev/null)" ]]
}

migrate_legacy_registry_dir() {
  local label="$1"
  local legacy_dir="$2"
  local target_dir="$3"

  if [[ "${legacy_dir}" == "${target_dir}" ]]; then
    return 0
  fi
  if ! dir_has_entries "${legacy_dir}"; then
    return 0
  fi
  if dir_has_entries "${target_dir}"; then
    return 0
  fi

  log_info "迁移旧 ${label} 目录: ${legacy_dir} -> ${target_dir}"
  mkdir -p "${target_dir}"
  cp -R "${legacy_dir}/." "${target_dir}/"
  log_success "已迁移 ${label} 目录"
}

compose_registry() {
  CTF_HOST_ROOT="${REPO_ROOT}" \
  CTF_REGISTRY_CONTAINER_NAME="${REGISTRY_NAME}" \
  CTF_REGISTRY_PORT="${REGISTRY_PORT}" \
  CTF_REGISTRY_DATA_DIR="${REGISTRY_DATA_DIR}" \
  CTF_REGISTRY_AUTH_DIR="${REGISTRY_AUTH_DIR}" \
  CTF_REGISTRY_IMAGE="${REGISTRY_IMAGE}" \
  CTF_REGISTRY_RESTART_POLICY="${REGISTRY_RESTART_POLICY}" \
    docker compose \
      --profile registry \
      -f "${REPO_ROOT}/docker/ctf/docker-compose.dev.yml" \
      "$@"
}

ORIGINAL_ARGS=("$@")
while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      [[ -n "${2:-}" ]] || die "--config 需要文件路径"
      CONFIG_FILE="$2"
      CONFIG_FILE_EXPLICIT=true
      shift 2
      ;;
    --config=*)
      CONFIG_FILE="${1#--config=}"
      [[ -n "${CONFIG_FILE}" ]] || die "--config 需要文件路径"
      CONFIG_FILE_EXPLICIT=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      shift
      ;;
  esac
done

load_config_file
set -- "${ORIGINAL_ARGS[@]}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      shift 2
      ;;
    --config=*)
      shift
      ;;
    --name)
      REGISTRY_NAME="${2:-}"
      shift 2
      ;;
    --port)
      REGISTRY_PORT="${2:-}"
      shift 2
      ;;
    --server)
      REGISTRY_SERVER="${2:-}"
      shift 2
      ;;
    --scheme)
      REGISTRY_SCHEME="${2:-}"
      shift 2
      ;;
    --username)
      REGISTRY_USERNAME="${2:-}"
      shift 2
      ;;
    --password)
      REGISTRY_PASSWORD="${2:-}"
      shift 2
      ;;
    --data-dir)
      REGISTRY_DATA_DIR="${2:-}"
      shift 2
      ;;
    --auth-dir)
      REGISTRY_AUTH_DIR="${2:-}"
      shift 2
      ;;
    --ctf-env-file)
      CTF_COMPOSE_REGISTRY_ENV_FILE="${2:-}"
      shift 2
      ;;
    --no-ctf-env-file)
      CTF_COMPOSE_REGISTRY_ENV_FILE=""
      shift
      ;;
    --image)
      REGISTRY_IMAGE="${2:-}"
      shift 2
      ;;
    --htpasswd-image)
      HTPASSWD_IMAGE="${2:-}"
      shift 2
      ;;
    --force-recreate)
      FORCE_RECREATE=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "未知参数: $1"
      ;;
  esac
done

require_command docker
require_command curl

load_existing_platform_env

[[ -n "${REGISTRY_NAME}" ]] || die "--name 不能为空"
[[ -n "${REGISTRY_USERNAME}" ]] || die "--username 不能为空"
[[ -n "${REGISTRY_DATA_DIR}" ]] || die "--data-dir 不能为空"
[[ -n "${REGISTRY_AUTH_DIR}" ]] || die "--auth-dir 不能为空"
ensure_positive_port "${REGISTRY_PORT}"
[[ "${REGISTRY_SCHEME}" == "http" || "${REGISTRY_SCHEME}" == "https" ]] || die "--scheme 只支持 http 或 https"

if [[ -z "${REGISTRY_SERVER}" ]]; then
  REGISTRY_SERVER="127.0.0.1:${REGISTRY_PORT}"
fi

if [[ -z "${REGISTRY_PASSWORD}" ]]; then
  log_info "未提供 registry 密码，自动生成随机密码"
  REGISTRY_PASSWORD="$(generate_password)"
fi

log_info "部署参数:"
log_info "  container: ${REGISTRY_NAME}"
log_info "  compose:   ctf/${REGISTRY_COMPOSE_SERVICE}"
log_info "  listen:    127.0.0.1:${REGISTRY_PORT}"
log_info "  server:    ${REGISTRY_SERVER}"
log_info "  scheme:    ${REGISTRY_SCHEME}"
log_info "  username:  ${REGISTRY_USERNAME}"
log_info "  data_dir:  ${REGISTRY_DATA_DIR}"
log_info "  auth_dir:  ${REGISTRY_AUTH_DIR}"
log_info "  image:     ${REGISTRY_IMAGE}"

if container_exists "${REGISTRY_NAME}"; then
  if [[ "${FORCE_RECREATE}" != "true" ]]; then
    die "容器 ${REGISTRY_NAME} 已存在；如需重建请加 --force-recreate"
  fi
  log_warn "容器 ${REGISTRY_NAME} 已存在，按 --force-recreate 删除后重建"
  docker rm -f "${REGISTRY_NAME}" >/dev/null
  log_success "已删除旧容器 ${REGISTRY_NAME}"
fi

log_info "创建数据目录和认证目录"
migrate_legacy_registry_dir "registry 数据" "${LEGACY_REGISTRY_DATA_DIR}" "${REGISTRY_DATA_DIR}"
migrate_legacy_registry_dir "registry 认证" "${LEGACY_REGISTRY_AUTH_DIR}" "${REGISTRY_AUTH_DIR}"
mkdir -p "${REGISTRY_DATA_DIR}" "${REGISTRY_AUTH_DIR}"
chmod 700 "${REGISTRY_AUTH_DIR}"
log_success "目录已准备"

log_info "生成 htpasswd 认证文件"
docker run --rm "${HTPASSWD_IMAGE}" \
  htpasswd -Bbn "${REGISTRY_USERNAME}" "${REGISTRY_PASSWORD}" \
  >"${REGISTRY_AUTH_DIR}/htpasswd"
chmod 600 "${REGISTRY_AUTH_DIR}/htpasswd"
log_success "认证文件已写入 ${REGISTRY_AUTH_DIR}/htpasswd"

log_info "启动 ctf Compose 项目下的 registry service"
compose_registry up -d --force-recreate "${REGISTRY_COMPOSE_SERVICE}" >/dev/null
log_success "registry service 已启动: ${REGISTRY_NAME}"

log_info "等待 registry 健康检查通过: http://127.0.0.1:${REGISTRY_PORT}/v2/"
timeout 30 bash -c 'until curl -fsS -u "$0:$1" "http://127.0.0.1:$2/v2/" >/dev/null; do sleep 1; done' \
  "${REGISTRY_USERNAME}" "${REGISTRY_PASSWORD}" "${REGISTRY_PORT}"
log_success "registry 已可访问"

PLATFORM_ENV_FILE=""
if [[ -n "${CTF_COMPOSE_REGISTRY_ENV_FILE}" ]]; then
  PLATFORM_ENV_FILE="${CTF_COMPOSE_REGISTRY_ENV_FILE}"
  log_info "写入平台后端与 ctf-api 共用的 registry 环境变量文件"
  mkdir -p "$(dirname "${CTF_COMPOSE_REGISTRY_ENV_FILE}")"
  write_platform_env "${CTF_COMPOSE_REGISTRY_ENV_FILE}"
  log_success "registry 环境变量文件已写入 ${CTF_COMPOSE_REGISTRY_ENV_FILE}"
else
  log_warn "未写入 registry 环境变量文件；ctf-api 需要手工注入 container.registry 配置"
fi

LEGACY_RUNTIME_ENV_FILE="${REGISTRY_AUTH_DIR}/ctf-platform-registry.env"
if [[ -n "${PLATFORM_ENV_FILE}" && "${LEGACY_RUNTIME_ENV_FILE}" != "${PLATFORM_ENV_FILE}" && -f "${LEGACY_RUNTIME_ENV_FILE}" ]]; then
  log_info "移除旧的重复 registry env 副本: ${LEGACY_RUNTIME_ENV_FILE}"
  rm -f "${LEGACY_RUNTIME_ENV_FILE}"
  log_success "已移除重复 registry env 副本"
fi

cat <<EOF
Registry 已部署:
  container: ${REGISTRY_NAME}
  compose:   ctf/${REGISTRY_COMPOSE_SERVICE}
  listen:    127.0.0.1:${REGISTRY_PORT}
  server:    ${REGISTRY_SERVER}
  scheme:    ${REGISTRY_SCHEME}
  data_dir:  ${REGISTRY_DATA_DIR}
  auth_dir:  ${REGISTRY_AUTH_DIR}
$(if [[ -n "${PLATFORM_ENV_FILE}" ]]; then printf '\n平台后端 registry 环境变量已写入:\n  %s\n' "${PLATFORM_ENV_FILE}"; fi)

后端配置等价于:
  container.registry.enabled=true
  container.registry.build_enabled=true
  container.registry.server=${REGISTRY_SERVER}
  container.registry.username=${REGISTRY_USERNAME}
  container.registry.password=<见 ${PLATFORM_ENV_FILE:-自定义注入}>

镜像引用示例:
  ${REGISTRY_SERVER}/awd/awd-supply-ticket:v1

注意:
  当前脚本部署的是 HTTP registry。若平台和 registry 不在同一台机器，
  运行节点 Docker daemon 需要配置 insecure-registries，或在 registry 前加 TLS 反向代理。
EOF
