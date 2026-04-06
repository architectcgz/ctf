#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
用法：
  scripts/bootstrap-frontend-deps.sh [--source <frontend_dir>] [--target <frontend_dir>]

说明：
  - 默认 target 为当前 git worktree 下的 code/frontend
  - 默认 source 会在当前仓库的其他 worktree 中自动寻找
  - 只有 source 与 target 的 package-lock.json 哈希一致时才复用 node_modules
  - 找不到可复用依赖时，自动回退到 npm ci --prefer-offline
EOF
}

log() {
  printf '[bootstrap-frontend-deps] %s\n' "$*"
}

die() {
  printf '[bootstrap-frontend-deps] %s\n' "$*" >&2
  exit 1
}

sha256_file() {
  sha256sum "$1" | awk '{print $1}'
}

copy_node_modules() {
  local source_dir="$1"
  local target_dir="$2"

  mkdir -p "$target_dir"

  if cp -al "$source_dir/node_modules" "$target_dir/"; then
    log "已通过硬链接复用 node_modules"
    return 0
  fi

  log "硬链接失败，回退到普通复制"
  cp -a "$source_dir/node_modules" "$target_dir/"
}

discover_source() {
  local repo_root="$1"
  local target_dir="$2"
  local target_lock_hash="$3"
  local preferred_source=""

  while IFS= read -r line; do
    case "$line" in
      worktree\ *)
        local worktree_path="${line#worktree }"
        local candidate_dir="$worktree_path/code/frontend"

        if [[ "$candidate_dir" == "$target_dir" ]]; then
          continue
        fi

        if [[ ! -d "$candidate_dir/node_modules" || ! -f "$candidate_dir/package-lock.json" ]]; then
          continue
        fi

        local candidate_hash
        candidate_hash="$(sha256_file "$candidate_dir/package-lock.json")"
        if [[ "$candidate_hash" != "$target_lock_hash" ]]; then
          continue
        fi

        if [[ "$worktree_path" == "$repo_root" ]]; then
          printf '%s\n' "$candidate_dir"
          return 0
        fi

        if [[ -z "$preferred_source" ]]; then
          preferred_source="$candidate_dir"
        fi
        ;;
    esac
  done < <(git -C "$repo_root" worktree list --porcelain)

  if [[ -n "$preferred_source" ]]; then
    printf '%s\n' "$preferred_source"
  fi
}

run_npm_ci() {
  local target_dir="$1"

  log "未找到可复用的 node_modules，开始执行 npm ci --prefer-offline"
  (
    cd "$target_dir"
    npm ci --prefer-offline --no-audit --no-fund --progress=false
  )
}

main() {
  local repo_root=""
  local target_dir=""
  local source_dir=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --source)
        [[ $# -ge 2 ]] || die "--source 缺少路径"
        source_dir="$2"
        shift 2
        ;;
      --target)
        [[ $# -ge 2 ]] || die "--target 缺少路径"
        target_dir="$2"
        shift 2
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

  if [[ -z "$target_dir" ]]; then
    repo_root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
    [[ -n "$repo_root" ]] || die "未提供 --target，且当前目录不在 git 仓库内"
    target_dir="$repo_root/code/frontend"
  fi

  target_dir="$(realpath "$target_dir")"
  [[ -d "$target_dir" ]] || die "target 不存在: $target_dir"
  [[ -f "$target_dir/package-lock.json" ]] || die "target 缺少 package-lock.json: $target_dir"

  if [[ -d "$target_dir/node_modules" ]]; then
    log "target 已存在 node_modules，跳过"
    exit 0
  fi

  local target_lock_hash
  target_lock_hash="$(sha256_file "$target_dir/package-lock.json")"

  if [[ -n "$source_dir" ]]; then
    source_dir="$(realpath "$source_dir")"
    [[ -d "$source_dir/node_modules" ]] || die "source 缺少 node_modules: $source_dir"
    [[ -f "$source_dir/package-lock.json" ]] || die "source 缺少 package-lock.json: $source_dir"

    local source_lock_hash
    source_lock_hash="$(sha256_file "$source_dir/package-lock.json")"
    [[ "$source_lock_hash" == "$target_lock_hash" ]] || die "source 与 target 的 package-lock.json 不一致，拒绝复用"
  else
    if [[ -z "$repo_root" ]]; then
      repo_root="$(git -C "$target_dir" rev-parse --show-toplevel 2>/dev/null || true)"
    fi

    if [[ -n "$repo_root" ]]; then
      source_dir="$(discover_source "$repo_root" "$target_dir" "$target_lock_hash" || true)"
    fi
  fi

  if [[ -n "$source_dir" ]]; then
    log "复用来源: $source_dir"
    copy_node_modules "$source_dir" "$target_dir"
    exit 0
  fi

  run_npm_ci "$target_dir"
}

main "$@"
