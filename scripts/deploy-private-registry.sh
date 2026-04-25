#!/usr/bin/env bash
set -euo pipefail

REGISTRY_NAME="${REGISTRY_NAME:-ctf-registry}"
REGISTRY_PORT="${REGISTRY_PORT:-5000}"
REGISTRY_SERVER="${REGISTRY_SERVER:-}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-ctf}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-}"
REGISTRY_DATA_DIR="${REGISTRY_DATA_DIR:-/data/ctf-registry}"
REGISTRY_AUTH_DIR="${REGISTRY_AUTH_DIR:-/data/ctf-registry-auth}"
REGISTRY_IMAGE="${REGISTRY_IMAGE:-registry:2}"
HTPASSWD_IMAGE="${HTPASSWD_IMAGE:-httpd:2.4-alpine}"
REGISTRY_RESTART_POLICY="${REGISTRY_RESTART_POLICY:-always}"
FORCE_RECREATE=false

usage() {
  cat <<'EOF'
用法:
  scripts/deploy-private-registry.sh [选项]

说明:
  部署一个带 Basic Auth 的 Docker Registry，并生成平台后端可直接加载的环境变量文件。
  默认监听 127.0.0.1:5000/宿主机 5000 端口，适合单机部署或毕设演示。

选项:
  --name NAME             Registry 容器名，默认 ctf-registry
  --port PORT             宿主机端口，默认 5000
  --server HOST:PORT      平台配置中的 registry server，默认 127.0.0.1:<port>
  --username USER         Registry 用户名，默认 ctf
  --password PASSWORD     Registry 密码；未提供时自动生成
  --data-dir DIR          镜像数据目录，默认 /data/ctf-registry
  --auth-dir DIR          认证文件目录，默认 /data/ctf-registry-auth
  --image IMAGE           Registry 镜像，默认 registry:2
  --htpasswd-image IMAGE  用于生成 htpasswd 的镜像，默认 httpd:2.4-alpine
  --force-recreate        若同名容器已存在，先删除后重建
  -h, --help              显示帮助

可通过同名环境变量覆盖默认值，例如:
  REGISTRY_PORT=15000 REGISTRY_PASSWORD='change-me' scripts/deploy-private-registry.sh

输出:
  <auth-dir>/ctf-platform-registry.env

该文件可供后端部署环境加载，内容形如:
  CTF_CONTAINER_REGISTRY_ENABLED=true
  CTF_CONTAINER_REGISTRY_SERVER=127.0.0.1:5000
  CTF_CONTAINER_REGISTRY_USERNAME=ctf
  CTF_CONTAINER_REGISTRY_PASSWORD=...
EOF
}

die() {
  echo "错误: $*" >&2
  exit 1
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    die "缺少命令: $1"
  fi
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
CTF_CONTAINER_REGISTRY_SERVER=${REGISTRY_SERVER}
CTF_CONTAINER_REGISTRY_USERNAME=${REGISTRY_USERNAME}
CTF_CONTAINER_REGISTRY_PASSWORD=${REGISTRY_PASSWORD}
EOF
  chmod 600 "${path}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
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

[[ -n "${REGISTRY_NAME}" ]] || die "--name 不能为空"
[[ -n "${REGISTRY_USERNAME}" ]] || die "--username 不能为空"
[[ -n "${REGISTRY_DATA_DIR}" ]] || die "--data-dir 不能为空"
[[ -n "${REGISTRY_AUTH_DIR}" ]] || die "--auth-dir 不能为空"
ensure_positive_port "${REGISTRY_PORT}"

if [[ -z "${REGISTRY_SERVER}" ]]; then
  REGISTRY_SERVER="127.0.0.1:${REGISTRY_PORT}"
fi

if [[ -z "${REGISTRY_PASSWORD}" ]]; then
  REGISTRY_PASSWORD="$(generate_password)"
fi

if container_exists "${REGISTRY_NAME}"; then
  if [[ "${FORCE_RECREATE}" != "true" ]]; then
    die "容器 ${REGISTRY_NAME} 已存在；如需重建请加 --force-recreate"
  fi
  docker rm -f "${REGISTRY_NAME}" >/dev/null
fi

mkdir -p "${REGISTRY_DATA_DIR}" "${REGISTRY_AUTH_DIR}"
chmod 700 "${REGISTRY_AUTH_DIR}"

docker run --rm "${HTPASSWD_IMAGE}" \
  htpasswd -Bbn "${REGISTRY_USERNAME}" "${REGISTRY_PASSWORD}" \
  >"${REGISTRY_AUTH_DIR}/htpasswd"
chmod 600 "${REGISTRY_AUTH_DIR}/htpasswd"

docker run -d \
  --name "${REGISTRY_NAME}" \
  --restart "${REGISTRY_RESTART_POLICY}" \
  -p "${REGISTRY_PORT}:5000" \
  -v "${REGISTRY_DATA_DIR}:/var/lib/registry" \
  -v "${REGISTRY_AUTH_DIR}:/auth:ro" \
  -e REGISTRY_AUTH=htpasswd \
  -e REGISTRY_AUTH_HTPASSWD_REALM=ctf-registry \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  "${REGISTRY_IMAGE}" >/dev/null

timeout 30 bash -c 'until curl -fsS -u "$0:$1" "http://127.0.0.1:$2/v2/" >/dev/null; do sleep 1; done' \
  "${REGISTRY_USERNAME}" "${REGISTRY_PASSWORD}" "${REGISTRY_PORT}"

PLATFORM_ENV_FILE="${REGISTRY_AUTH_DIR}/ctf-platform-registry.env"
write_platform_env "${PLATFORM_ENV_FILE}"

cat <<EOF
Registry 已部署:
  container: ${REGISTRY_NAME}
  listen:    127.0.0.1:${REGISTRY_PORT}
  server:    ${REGISTRY_SERVER}
  data_dir:  ${REGISTRY_DATA_DIR}
  auth_dir:  ${REGISTRY_AUTH_DIR}

平台后端环境变量已写入:
  ${PLATFORM_ENV_FILE}

后端配置等价于:
  container.registry.enabled=true
  container.registry.server=${REGISTRY_SERVER}
  container.registry.username=${REGISTRY_USERNAME}
  container.registry.password=<见 ${PLATFORM_ENV_FILE}>

镜像引用示例:
  ${REGISTRY_SERVER}/ctf/awd-supply-ticket:v1

注意:
  当前脚本部署的是 HTTP registry。若平台和 registry 不在同一台机器，
  运行节点 Docker daemon 需要配置 insecure-registries，或在 registry 前加 TLS 反向代理。
EOF
