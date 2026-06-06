#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

project_prefix="$(basename "$PWD")"
codex_home="${CODEX_HOME:-$HOME/.codex}"
codex_skills_dir="$codex_home/skills"
shared_skills_dir="$PWD/.agents/skills"

if [[ ! -d "$shared_skills_dir" ]]; then
  echo "FAIL: missing shared skills dir: $shared_skills_dir" >&2
  exit 1
fi

bash scripts/check-shared-skills.sh

mkdir -p "$codex_skills_dir" .claude
ln -sfn ../.agents/skills .claude/skills

while IFS= read -r source_dir; do
  skill="$(basename "$source_dir")"
  target_link="$codex_skills_dir/${project_prefix}-${skill}"

  if [[ ! -f "$source_dir/SKILL.md" ]]; then
    echo "FAIL: missing skill source: $source_dir" >&2
    exit 1
  fi

  if [[ -e "$target_link" && ! -L "$target_link" ]]; then
    echo "FAIL: existing non-symlink blocks install: $target_link" >&2
    exit 1
  fi

  ln -sfn "$source_dir" "$target_link"
  echo "linked: $target_link -> $source_dir"
done < <(find "$shared_skills_dir" -mindepth 1 -maxdepth 1 -type d | sort)

bash scripts/check-agent-entrypoints.sh
echo "Installed shared skills for Claude and Codex."
