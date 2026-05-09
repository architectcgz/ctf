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
CTF_BACKEND_LOG="${CTF_BACKEND_LOG:-/tmp/ctf-backend.log}"
CTF_BACKEND_LOG_TAIL_LINES="${CTF_BACKEND_LOG_TAIL_LINES:-80}"
CTF_BACKEND_SIGINT_GRACE_TICKS="${CTF_BACKEND_SIGINT_GRACE_TICKS:-120}"
CTF_BACKEND_SIGTERM_GRACE_TICKS="${CTF_BACKEND_SIGTERM_GRACE_TICKS:-30}"
REGISTRY_ENV_FILE_PATH=""
FOREGROUND_BACKEND_PID=""
FOREGROUND_STOP_REQUESTED=false
LAUNCHED_BACKEND_PID=""

WITH_MIGRATE=false
INFRA_MODE=""
RUN_MODE="run"
BACKGROUND=false

usage() {
  cat <<'EOF'
用法:
  ./scripts/dev-run.sh [--infra] [--infra-shared] [--migrate] [--hot] [--background]

说明:
  --infra         启动开发依赖容器
  --infra-shared  启动共享开发依赖容器
  --migrate       启动前执行数据库迁移
  --hot           使用 air 热重载启动后端
  --background    后台启动后端，并把输出写入日志文件

可覆盖环境变量:
  APP_ENV
  CTF_CONTAINER_FLAG_GLOBAL_SECRET
  CTF_FLAG_SECRET
  AIR_BIN
  AIR_CONFIG
  CTF_BACKEND_LOG
  CTF_BACKEND_LOG_TAIL_LINES
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

resolve_registry_env_file() {
  local candidate

  candidate="${REPO_ROOT}/docker/ctf/infra/registry/ctf-platform-registry.env"
  if [[ -f "${candidate}" ]]; then
    printf '%s\n' "${candidate}"
    return 0
  fi

  while IFS= read -r worktree_path; do
    candidate="${worktree_path}/docker/ctf/infra/registry/ctf-platform-registry.env"
    if [[ -f "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done < <(git -C "${REPO_ROOT}" worktree list --porcelain | awk '/^worktree / {print substr($0, 10)}')

  return 1
}

trim_whitespace() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "${value}"
}

load_registry_env_if_present() {
  local env_file line key value

  if ! env_file="$(resolve_registry_env_file)"; then
    return 0
  fi

  while IFS= read -r line || [[ -n "${line}" ]]; do
    line="${line%$'\r'}"
    if [[ -z "$(trim_whitespace "${line}")" ]]; then
      continue
    fi
    if [[ "${line}" =~ ^[[:space:]]*# ]]; then
      continue
    fi
    if [[ "${line}" != *"="* ]]; then
      continue
    fi

    key="$(trim_whitespace "${line%%=*}")"
    value="${line#*=}"
    if [[ "${key}" == export\ * ]]; then
      key="$(trim_whitespace "${key#export }")"
    fi
    if [[ -z "${key}" ]]; then
      continue
    fi
    if [[ -z "${!key:-}" ]]; then
      export "${key}=${value}"
    fi
  done <"${env_file}"

  REGISTRY_ENV_FILE_PATH="${env_file}"
}

start_infra() {
  local compose_file

  if ! compose_file="$(resolve_compose_file)"; then
    echo "未找到 docker/ctf/docker-compose.dev.yml，无法自动启动开发依赖容器。" >&2
    echo "如果你在主工作区维护了该文件，请在主工作区执行 docker compose，或先手动启动 PostgreSQL / Redis。" >&2
    exit 1
  fi

  CTF_HOST_ROOT="${REPO_ROOT}" docker compose -f "${compose_file}" up -d ctf-postgres ctf-redis
}

apply_runtime_defaults() {
  if [[ -n "${INFRA_MODE}" ]]; then
    export CTF_POSTGRES_HOST="${CTF_POSTGRES_HOST:-127.0.0.1}"
    export CTF_POSTGRES_PORT="${CTF_POSTGRES_PORT:-15432}"
    export CTF_POSTGRES_PASSWORD="${CTF_POSTGRES_PASSWORD:-postgres123456}"
    export CTF_REDIS_ADDR="${CTF_REDIS_ADDR:-127.0.0.1:16379}"
    export CTF_REDIS_PASSWORD="${CTF_REDIS_PASSWORD:-redis123456}"
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

prepare_log_file() {
  mkdir -p "$(dirname "${CTF_BACKEND_LOG}")"
  touch "${CTF_BACKEND_LOG}"
}

write_log_banner() {
  {
    printf '\n===== ctf backend start %s mode=%s background=%s =====\n' \
      "$(date '+%Y-%m-%d %H:%M:%S %z')" "${RUN_MODE}" "${BACKGROUND}"
  } >>"${CTF_BACKEND_LOG}"
}

launch_backend_session() {
  local output_target="$1"
  local detach_stdin="${2:-false}"

  if [[ "${detach_stdin}" == "true" ]]; then
    setsid bash -c '
      set -euo pipefail
      backend_dir="$1"
      output_target="$2"
      run_mode="$3"
      app_env="$4"
      flag_global_secret="$5"
      flag_secret="$6"
      air_bin="$7"
      air_config="$8"

      cd "${backend_dir}"
      if [[ "${run_mode}" == "hot" ]]; then
        if ! command -v "${air_bin}" >/dev/null 2>&1; then
          echo "air 未安装，请先执行: go install github.com/air-verse/air@latest" >&2
          exit 1
        fi

        exec env \
          APP_ENV="${app_env}" \
          CTF_CONTAINER_FLAG_GLOBAL_SECRET="${flag_global_secret}" \
          CTF_FLAG_SECRET="${flag_secret}" \
          "${air_bin}" -c "${air_config}" >>"${output_target}" 2>&1
      fi

      exec env \
        APP_ENV="${app_env}" \
        CTF_CONTAINER_FLAG_GLOBAL_SECRET="${flag_global_secret}" \
        CTF_FLAG_SECRET="${flag_secret}" \
        go run ./cmd/api >>"${output_target}" 2>&1
    ' bash "${BACKEND_DIR}" "${output_target}" "${RUN_MODE}" "${APP_ENV}" "${CTF_CONTAINER_FLAG_GLOBAL_SECRET}" "${CTF_FLAG_SECRET}" "${AIR_BIN}" "${AIR_CONFIG}" </dev/null &
  else
    setsid bash -c '
      set -euo pipefail
      backend_dir="$1"
      output_target="$2"
      run_mode="$3"
      app_env="$4"
      flag_global_secret="$5"
      flag_secret="$6"
      air_bin="$7"
      air_config="$8"

      cd "${backend_dir}"
      if [[ "${run_mode}" == "hot" ]]; then
        if ! command -v "${air_bin}" >/dev/null 2>&1; then
          echo "air 未安装，请先执行: go install github.com/air-verse/air@latest" >&2
          exit 1
        fi

        exec env \
          APP_ENV="${app_env}" \
          CTF_CONTAINER_FLAG_GLOBAL_SECRET="${flag_global_secret}" \
          CTF_FLAG_SECRET="${flag_secret}" \
          "${air_bin}" -c "${air_config}" >>"${output_target}" 2>&1
      fi

      exec env \
        APP_ENV="${app_env}" \
        CTF_CONTAINER_FLAG_GLOBAL_SECRET="${flag_global_secret}" \
        CTF_FLAG_SECRET="${flag_secret}" \
        go run ./cmd/api >>"${output_target}" 2>&1
    ' bash "${BACKEND_DIR}" "${output_target}" "${RUN_MODE}" "${APP_ENV}" "${CTF_CONTAINER_FLAG_GLOBAL_SECRET}" "${CTF_FLAG_SECRET}" "${AIR_BIN}" "${AIR_CONFIG}" &
  fi

  LAUNCHED_BACKEND_PID="$!"
}

list_child_pids() {
  local parent_pid="$1"
  ps -o pid= --ppid "${parent_pid}" 2>/dev/null | awk '{print $1}'
}

collect_descendant_pids() {
  local parent_pid="$1"
  local child_pid

  while IFS= read -r child_pid; do
    child_pid="$(trim_whitespace "${child_pid}")"
    if [[ -z "${child_pid}" ]]; then
      continue
    fi
    printf '%s\n' "${child_pid}"
    collect_descendant_pids "${child_pid}"
  done < <(list_child_pids "${parent_pid}")
}

collect_leaf_pids() {
  local parent_pid="$1"
  local child_pid
  local child_pids=()

  while IFS= read -r child_pid; do
    child_pid="$(trim_whitespace "${child_pid}")"
    if [[ -n "${child_pid}" ]]; then
      child_pids+=("${child_pid}")
    fi
  done < <(list_child_pids "${parent_pid}")

  if (( ${#child_pids[@]} == 0 )); then
    if process_exists "${parent_pid}"; then
      printf '%s\n' "${parent_pid}"
    fi
    return 0
  fi

  for child_pid in "${child_pids[@]}"; do
    collect_leaf_pids "${child_pid}"
  done
}

process_exists() {
  local pid="$1"
  [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null
}

collect_backend_tree_snapshot() {
  local root_pid="$1"
  local descendant_pid

  if process_exists "${root_pid}"; then
    printf '%s\n' "${root_pid}"
  fi

  while IFS= read -r descendant_pid; do
    descendant_pid="$(trim_whitespace "${descendant_pid}")"
    if [[ -n "${descendant_pid}" ]]; then
      printf '%s\n' "${descendant_pid}"
    fi
  done < <(collect_descendant_pids "${root_pid}")
}

pid_list_alive() {
  local pid

  for pid in "$@"; do
    if process_exists "${pid}"; then
      return 0
    fi
  done

  return 1
}

signal_pid_list() {
  local signal_name="$1"
  shift

  local pid
  local index
  local pids=("$@")

  for (( index=${#pids[@]}-1; index>=0; index-- )); do
    pid="${pids[index]}"
    if process_exists "${pid}"; then
      kill "-${signal_name}" "${pid}" 2>/dev/null || true
    fi
  done
}

stop_backend_tree() {
  local root_pid="$1"
  local delay
  local leaf_pid
  local tracked_pid
  local leaf_pids=()
  local tracked_pids=()

  if [[ -z "${root_pid}" ]]; then
    return 0
  fi

  while IFS= read -r tracked_pid; do
    tracked_pid="$(trim_whitespace "${tracked_pid}")"
    if [[ -n "${tracked_pid}" ]]; then
      tracked_pids+=("${tracked_pid}")
    fi
  done < <(collect_backend_tree_snapshot "${root_pid}")

  if (( ${#tracked_pids[@]} == 0 )); then
    return 0
  fi

  while IFS= read -r leaf_pid; do
    leaf_pid="$(trim_whitespace "${leaf_pid}")"
    if [[ -n "${leaf_pid}" ]]; then
      leaf_pids+=("${leaf_pid}")
    fi
  done < <(collect_leaf_pids "${root_pid}")

  if (( ${#leaf_pids[@]} == 0 )); then
    leaf_pids=("${tracked_pids[@]}")
  fi

  # 先让真正承载业务的叶子进程自行优雅退出，避免在 hot-reload 场景里和 air 的清理逻辑互相抢杀。
  signal_pid_list INT "${leaf_pids[@]}"
  for delay in $(seq 1 "${CTF_BACKEND_SIGINT_GRACE_TICKS}"); do
    if ! pid_list_alive "${tracked_pids[@]}"; then
      return 0
    fi
    if ! pid_list_alive "${leaf_pids[@]}"; then
      break
    fi
    sleep 0.1
  done

  signal_pid_list INT "${tracked_pids[@]}"
  for delay in $(seq 1 "${CTF_BACKEND_SIGTERM_GRACE_TICKS}"); do
    if ! pid_list_alive "${tracked_pids[@]}"; then
      return 0
    fi
    sleep 0.1
  done

  signal_pid_list TERM "${tracked_pids[@]}"
  for delay in $(seq 1 "${CTF_BACKEND_SIGTERM_GRACE_TICKS}"); do
    if ! pid_list_alive "${tracked_pids[@]}"; then
      return 0
    fi
    sleep 0.1
  done

  signal_pid_list KILL "${tracked_pids[@]}"
}

cleanup_foreground_log_pipe() {
  local fifo_path="$1"
  local tee_pid="${2:-}"

  if [[ -n "${tee_pid}" ]] && process_exists "${tee_pid}"; then
    kill "${tee_pid}" 2>/dev/null || true
    wait "${tee_pid}" 2>/dev/null || true
  fi

  if [[ -n "${fifo_path}" && -e "${fifo_path}" ]]; then
    rm -f "${fifo_path}"
  fi
}

handle_foreground_stop_signal() {
  if [[ "${FOREGROUND_STOP_REQUESTED}" == "true" ]]; then
    return 0
  fi
  FOREGROUND_STOP_REQUESTED=true

  if [[ -n "${FOREGROUND_BACKEND_PID}" ]]; then
    echo "收到停止信号，正在关闭后端及其子进程..."
    stop_backend_tree "${FOREGROUND_BACKEND_PID}"
  fi
}

start_backend() {
  prepare_log_file
  write_log_banner
  echo "日志文件: ${CTF_BACKEND_LOG}"
  echo "查看日志: tail -f ${CTF_BACKEND_LOG}"

  if [[ "${BACKGROUND}" == "true" ]]; then
    launch_backend_session "${CTF_BACKEND_LOG}" true
    backend_pid="${LAUNCHED_BACKEND_PID}"
    echo "后端已后台启动，launcher_pid=${backend_pid}"
    tail -n "${CTF_BACKEND_LOG_TAIL_LINES}" "${CTF_BACKEND_LOG}"
    return
  fi

  local log_pipe
  local tee_pid
  local backend_pid
  local exit_code=0

  log_pipe="$(mktemp -u "${TMPDIR:-/tmp}/ctf-backend-log.XXXXXX.pipe")"
  mkfifo "${log_pipe}"
  tee -a "${CTF_BACKEND_LOG}" <"${log_pipe}" &
  tee_pid=$!

  FOREGROUND_BACKEND_PID=""
  FOREGROUND_STOP_REQUESTED=false
  trap 'handle_foreground_stop_signal' INT TERM

  launch_backend_session "${log_pipe}"
  backend_pid="${LAUNCHED_BACKEND_PID}"
  FOREGROUND_BACKEND_PID="${backend_pid}"

  wait "${backend_pid}" || exit_code=$?

  trap - INT TERM
  FOREGROUND_BACKEND_PID=""
  cleanup_foreground_log_pipe "${log_pipe}" "${tee_pid}"

  return "${exit_code}"
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
    --background)
      BACKGROUND=true
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

load_registry_env_if_present

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
if [[ -n "${REGISTRY_ENV_FILE_PATH}" ]]; then
  echo "CTF_CONTAINER_REGISTRY_ENV=${REGISTRY_ENV_FILE_PATH}"
fi
if [[ -n "${CTF_CONTAINER_REGISTRY_ENABLED:-}" ]]; then
  echo "CTF_CONTAINER_REGISTRY_ENABLED=${CTF_CONTAINER_REGISTRY_ENABLED}"
fi
if [[ -n "${CTF_CONTAINER_REGISTRY_SERVER:-}" ]]; then
  echo "CTF_CONTAINER_REGISTRY_SERVER=${CTF_CONTAINER_REGISTRY_SERVER}"
fi
echo "CTF_BACKEND_LOG=${CTF_BACKEND_LOG}"
echo "启动后端服务 (${RUN_MODE})..."

start_backend
