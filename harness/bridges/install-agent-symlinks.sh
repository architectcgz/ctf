#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"

project_prefix="$(basename "$repo_root")"
codex_home="${CODEX_HOME:-$HOME/.codex}"
codex_skills_dir="$codex_home/skills"
shared_skills_dir="$repo_root/.agents/skills"

bash "$repo_root/harness/checks/check_shared_skills.sh"

mkdir -p "$codex_skills_dir" "$repo_root/.claude"
ln -sfn ../.agents/skills "$repo_root/.claude/skills"

while IFS= read -r source_dir; do
  skill="$(basename "$source_dir")"
  target_link="$codex_skills_dir/${project_prefix}-${skill}"

  if [[ -e "$target_link" && ! -L "$target_link" ]]; then
    echo "FAIL: existing non-symlink blocks install: $target_link" >&2
    exit 1
  fi

  ln -sfn "$source_dir" "$target_link"
  echo "linked: $target_link -> $source_dir"
done < <(find "$shared_skills_dir" -mindepth 1 -maxdepth 1 -type d | sort)

bash "$repo_root/harness/checks/check_agent_entrypoints.sh"
echo "Installed shared skills for Claude and Codex."
