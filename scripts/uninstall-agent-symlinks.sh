#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
cd "$repo_root"

project_prefix="$(basename "$repo_root")"
codex_home="${CODEX_HOME:-$HOME/.codex}"
codex_skills_dir="$codex_home/skills"
shared_skills_dir="$repo_root/.agents/skills"

while IFS= read -r source_dir; do
  skill="$(basename "$source_dir")"
  target_link="$codex_skills_dir/${project_prefix}-${skill}"
  if [[ -L "$target_link" && "$(readlink -f "$target_link")" == "$(readlink -f "$source_dir")" ]]; then
    rm "$target_link"
    echo "removed: $target_link"
  fi
done < <(find "$shared_skills_dir" -mindepth 1 -maxdepth 1 -type d | sort)

echo "Uninstalled shared Codex skill links for this project."
echo "Kept .claude/skills in place because it is part of the repository contract."
