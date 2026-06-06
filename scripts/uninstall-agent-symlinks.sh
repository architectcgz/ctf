#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

project_prefix="$(basename "$PWD")"
codex_home="${CODEX_HOME:-$HOME/.codex}"
codex_skills_dir="$codex_home/skills"
shared_skills_dir="$PWD/.agents/skills"

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
