#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DEFAULT_REGISTRY_ENV="${HOME}/ctf-registry/auth/ctf-platform-registry.env"
DEFAULT_REGISTRY_CONFIG="${SCRIPT_DIR}/deploy-private-registry.conf"
DEFAULT_REGISTRY_SERVER="${DEFAULT_REGISTRY_SERVER:-127.0.0.1:5000}"
DEFAULT_REGISTRY_SCHEME="${DEFAULT_REGISTRY_SCHEME:-http}"

PACK_DIR=""
TAG=""
REGISTRY_ENV="${REGISTRY_ENV:-${DEFAULT_REGISTRY_ENV}}"
REGISTRY_CONFIG="${REGISTRY_CONFIG:-${DEFAULT_REGISTRY_CONFIG}}"
REGISTRY_SERVER="${REGISTRY_SERVER:-}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-}"
REGISTRY_SCHEME="${REGISTRY_SCHEME:-}"
UPDATE_MANIFEST=false
NO_LOGIN=false

usage() {
  cat <<'EOF'
用法:
  scripts/registry/build-and-push-challenge-image.sh <challenge_pack_dir> [选项]

说明:
  面向题目作者/管理员的镜像交付脚本。
  传入题目包目录后，脚本会读取 challenge.yml 的 meta.slug，
  使用 docker/ 目录构建镜像，打上私有 registry 标签并 push。
  题目包目录既可以传当前目录相对路径，也可以传仓库根目录相对路径。

选项:
  --tag TAG              镜像标签；默认优先复用 challenge.yml 里的 runtime.image.ref 标签，否则使用当前日期时间
  --registry-env FILE    平台 registry 环境变量文件，默认 $HOME/ctf-registry/auth/ctf-platform-registry.env
  --registry-config FILE registry 部署配置文件，默认 scripts/registry/deploy-private-registry.conf
  --registry SERVER      registry 地址，不带 http/https，例如 192.168.1.10:5000
  --registry-scheme http|https
                         访问 registry API 使用的协议，默认 http
  --username USER        registry 用户名
  --password PASSWORD    registry 密码
  --update-manifest      push 成功后把 challenge.yml 的 runtime.image.ref 更新为最终镜像地址
  --no-login             跳过 docker login，适用于已登录或无认证 registry
  -h, --help             显示帮助

示例:
  scripts/registry/build-and-push-challenge-image.sh challenges/packs/web-notes-download

  scripts/registry/build-and-push-challenge-image.sh \
    challenges/packs/web-notes-download \
    --registry 192.168.1.10:5000 \
    --tag 20260409 \
    --update-manifest

配置优先级:
  命令行参数 > 当前 shell 环境变量 > registry 环境文件 > registry 配置文件 > 默认本机演示地址

输出:
  最后一行会输出可写入 challenge.yml 的镜像引用:
  IMAGE_REF=<registry>/ctf/<slug>:<tag>
EOF
}

die() {
  echo "错误: $*" >&2
  exit 1
}

log_info() {
  printf '[challenge-image] %s\n' "$*"
}

log_success() {
  printf '[challenge-image] 完成: %s\n' "$*"
}

require_command() {
  log_info "检查命令: $1"
  if ! command -v "$1" >/dev/null 2>&1; then
    die "缺少命令: $1"
  fi
}

resolve_pack_dir() {
  local input="$1"

  if [[ -d "${input}" ]]; then
    realpath "${input}"
    return 0
  fi

  if [[ -d "${REPO_ROOT}/${input}" ]]; then
    realpath "${REPO_ROOT}/${input}"
    return 0
  fi

  die "题目包目录不存在: ${input}；请传绝对路径、当前目录相对路径，或仓库根目录相对路径"
}

apply_registry_env_file() {
  local current_server current_username current_password current_scheme
  local file_server file_username file_password file_scheme

  current_server="${REGISTRY_SERVER}"
  current_username="${REGISTRY_USERNAME}"
  current_password="${REGISTRY_PASSWORD}"
  current_scheme="${REGISTRY_SCHEME}"

  # shellcheck source=/dev/null
  source "${REGISTRY_ENV}"

  file_server="${CTF_CONTAINER_REGISTRY_SERVER:-${REGISTRY_SERVER:-}}"
  file_username="${CTF_CONTAINER_REGISTRY_USERNAME:-${REGISTRY_USERNAME:-}}"
  file_password="${CTF_CONTAINER_REGISTRY_PASSWORD:-${REGISTRY_PASSWORD:-}}"
  file_scheme="${CTF_CONTAINER_REGISTRY_SCHEME:-${REGISTRY_SCHEME:-}}"

  REGISTRY_SERVER="${current_server:-${file_server}}"
  REGISTRY_USERNAME="${current_username:-${file_username}}"
  REGISTRY_PASSWORD="${current_password:-${file_password}}"
  REGISTRY_SCHEME="${current_scheme:-${file_scheme}}"
}

apply_registry_config_file() {
  local current_server current_username current_password current_scheme
  local file_server file_username file_password file_scheme

  current_server="${REGISTRY_SERVER}"
  current_username="${REGISTRY_USERNAME}"
  current_password="${REGISTRY_PASSWORD}"
  current_scheme="${REGISTRY_SCHEME}"

  # shellcheck source=/dev/null
  source "${REGISTRY_CONFIG}"

  file_server="${REGISTRY_SERVER:-}"
  file_username="${REGISTRY_USERNAME:-}"
  file_password="${REGISTRY_PASSWORD:-}"
  file_scheme="${REGISTRY_SCHEME:-}"

  REGISTRY_SERVER="${current_server:-${file_server}}"
  REGISTRY_USERNAME="${current_username:-${file_username}}"
  REGISTRY_PASSWORD="${current_password:-${file_password}}"
  REGISTRY_SCHEME="${current_scheme:-${file_scheme}}"
}

load_registry_settings() {
  local loaded=false

  if [[ -f "${REGISTRY_ENV}" ]]; then
    log_info "读取 registry 环境文件: ${REGISTRY_ENV}"
    apply_registry_env_file
    loaded=true
  fi

  if [[ -f "${REGISTRY_CONFIG}" ]]; then
    log_info "读取 registry 配置文件: ${REGISTRY_CONFIG}"
    apply_registry_config_file
    loaded=true
  fi

  if [[ "${loaded}" != "true" ]]; then
    log_info "未找到 registry 配置文件，使用命令行参数、环境变量或默认 registry"
  fi
}

parse_yaml_with_python() {
  local file="$1"
  local expr="$2"

  python3 - "$file" "$expr" <<'PY'
import sys
from pathlib import Path

try:
    import yaml
except Exception:
    sys.exit(2)

path = Path(sys.argv[1])
expr = sys.argv[2].split(".")
data = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
value = data
for key in expr:
    if not isinstance(value, dict) or key not in value:
        value = ""
        break
    value = value[key]
print("" if value is None else value)
PY
}

parse_slug() {
  local manifest="$1"
  local slug
  slug="$(parse_yaml_with_python "${manifest}" "meta.slug" 2>/dev/null || true)"
  if [[ -n "${slug}" ]]; then
    printf '%s\n' "${slug}"
    return 0
  fi

  awk '
    /^meta:[[:space:]]*$/ { in_meta=1; next }
    /^[^[:space:]][^:]*:/ { in_meta=0 }
    in_meta && /^[[:space:]]+slug:[[:space:]]*/ {
      sub(/^[[:space:]]+slug:[[:space:]]*/, "")
      gsub(/^"|"$/, "")
      print
      exit
    }
  ' "${manifest}"
}

parse_runtime_image_ref() {
  local manifest="$1"
  parse_yaml_with_python "${manifest}" "runtime.image.ref" 2>/dev/null || true
}

extract_tag_from_ref() {
  local ref="$1"
  local last_slash last_colon
  last_slash="${ref##*/}"
  if [[ "${last_slash}" != *:* ]]; then
    return 0
  fi
  last_colon="${last_slash##*:}"
  printf '%s\n' "${last_colon}"
}

update_manifest_image_ref() {
  local manifest="$1"
  local image_ref="$2"

  python3 - "$manifest" "$image_ref" <<'PY'
import sys
from pathlib import Path

try:
    import yaml
except Exception as exc:
    raise SystemExit(f"PyYAML unavailable, cannot update manifest: {exc}")

path = Path(sys.argv[1])
image_ref = sys.argv[2]
data = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
runtime = data.setdefault("runtime", {})
runtime["type"] = "container"
image = runtime.setdefault("image", {})
image["ref"] = image_ref
path.write_text(yaml.safe_dump(data, allow_unicode=True, sort_keys=False), encoding="utf-8")
PY
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      [[ -n "${2:-}" ]] || die "--tag 需要值"
      TAG="$2"
      shift 2
      ;;
    --registry-env)
      [[ -n "${2:-}" ]] || die "--registry-env 需要文件路径"
      REGISTRY_ENV="$2"
      shift 2
      ;;
    --registry-config)
      [[ -n "${2:-}" ]] || die "--registry-config 需要文件路径"
      REGISTRY_CONFIG="$2"
      shift 2
      ;;
    --registry)
      [[ -n "${2:-}" ]] || die "--registry 需要地址"
      REGISTRY_SERVER="$2"
      shift 2
      ;;
    --registry-scheme)
      [[ -n "${2:-}" ]] || die "--registry-scheme 需要 http 或 https"
      REGISTRY_SCHEME="$2"
      shift 2
      ;;
    --username)
      [[ -n "${2:-}" ]] || die "--username 需要值"
      REGISTRY_USERNAME="$2"
      shift 2
      ;;
    --password)
      [[ -n "${2:-}" ]] || die "--password 需要值"
      REGISTRY_PASSWORD="$2"
      shift 2
      ;;
    --update-manifest)
      UPDATE_MANIFEST=true
      shift
      ;;
    --no-login)
      NO_LOGIN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      die "未知参数: $1"
      ;;
    *)
      if [[ -n "${PACK_DIR}" ]]; then
        die "只能传入一个题目包目录"
      fi
      PACK_DIR="$1"
      shift
      ;;
  esac
done

[[ -n "${PACK_DIR}" ]] || die "缺少题目包目录"

require_command docker
require_command python3

PACK_DIR="$(resolve_pack_dir "${PACK_DIR}")"
MANIFEST="${PACK_DIR}/challenge.yml"
DOCKER_DIR="${PACK_DIR}/docker"
DOCKERFILE="${DOCKER_DIR}/Dockerfile"

[[ -f "${MANIFEST}" ]] || die "题目包缺少 challenge.yml: ${MANIFEST}"
[[ -f "${DOCKERFILE}" ]] || die "题目包缺少 docker/Dockerfile: ${DOCKERFILE}"

load_registry_settings
if [[ -z "${REGISTRY_SERVER}" ]]; then
  REGISTRY_SERVER="${DEFAULT_REGISTRY_SERVER}"
fi
if [[ -z "${REGISTRY_SCHEME}" ]]; then
  REGISTRY_SCHEME="${DEFAULT_REGISTRY_SCHEME}"
fi
[[ "${REGISTRY_SCHEME}" == "http" || "${REGISTRY_SCHEME}" == "https" ]] || die "--registry-scheme 只支持 http 或 https"

SLUG="$(parse_slug "${MANIFEST}")"
[[ -n "${SLUG}" ]] || die "challenge.yml meta.slug 不能为空"

if [[ -z "${TAG}" ]]; then
  EXISTING_REF="$(parse_runtime_image_ref "${MANIFEST}")"
  TAG="$(extract_tag_from_ref "${EXISTING_REF}")"
fi
if [[ -z "${TAG}" ]]; then
  TAG="$(date +%Y%m%d%H%M)"
fi

LOCAL_REF="ctf/${SLUG}:${TAG}"
IMAGE_REF="${REGISTRY_SERVER}/ctf/${SLUG}:${TAG}"

log_info "题目包: ${PACK_DIR}"
log_info "slug: ${SLUG}"
log_info "tag: ${TAG}"
log_info "registry: ${REGISTRY_SERVER}"
log_info "registry scheme: ${REGISTRY_SCHEME}"
log_info "local image: ${LOCAL_REF}"
log_info "target image: ${IMAGE_REF}"

if [[ "${NO_LOGIN}" != "true" ]]; then
  if [[ -n "${REGISTRY_USERNAME}" && -n "${REGISTRY_PASSWORD}" ]]; then
    log_info "验证 registry 认证接口"
    curl -fsS -u "${REGISTRY_USERNAME}:${REGISTRY_PASSWORD}" "${REGISTRY_SCHEME}://${REGISTRY_SERVER}/v2/" >/dev/null
    log_success "registry 认证接口可访问"

    log_info "docker login ${REGISTRY_SERVER}"
    printf '%s\n' "${REGISTRY_PASSWORD}" | docker login "${REGISTRY_SERVER}" -u "${REGISTRY_USERNAME}" --password-stdin >/dev/null
    log_success "docker login 完成"
  else
    log_info "未提供 registry 用户名/密码，跳过 docker login"
  fi
fi

log_info "docker build 开始"
docker build -t "${LOCAL_REF}" "${DOCKER_DIR}"
log_success "docker build 完成"

log_info "docker tag ${LOCAL_REF} -> ${IMAGE_REF}"
docker tag "${LOCAL_REF}" "${IMAGE_REF}"
log_success "docker tag 完成"

log_info "docker push 开始"
docker push "${IMAGE_REF}"
log_success "docker push 完成"

if [[ "${UPDATE_MANIFEST}" == "true" ]]; then
  log_info "更新 challenge.yml runtime.image.ref"
  update_manifest_image_ref "${MANIFEST}" "${IMAGE_REF}"
  log_success "challenge.yml 已更新"
  log_info "已设置 challenge.yml: runtime.image.ref=${IMAGE_REF}"
else
  log_info "请确认 challenge.yml 中已设置 runtime.image.ref=${IMAGE_REF}；也可以下次加 --update-manifest 自动写入"
fi

printf 'IMAGE_REF=%s\n' "${IMAGE_REF}"
