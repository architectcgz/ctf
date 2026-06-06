#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mode="${1:---staged}"

case "$mode" in
  --staged)
    changed_files="$(git diff --cached --name-only --diff-filter=ACMR || true)"
    ;;
  --working)
    changed_files="$(git diff --name-only --diff-filter=ACMR || true)"
    ;;
  --all)
    changed_files="$(
      {
        git diff --cached --name-only --diff-filter=ACMR
        git diff --name-only --diff-filter=ACMR
      } | sort -u
    )"
    ;;
  -h|--help)
    cat <<'EOF'
Usage:
  scripts/check-skill-sync-reminder.sh [--staged|--working|--all]

Prints a non-blocking reminder when harness feedback,
prompts, policies, or templates changed and may need synchronization
back into the shared global harness or skills.
EOF
    exit 0
    ;;
  *)
    echo "Unknown mode: $mode" >&2
    exit 2
    ;;
esac

if [[ -z "${changed_files// }" ]]; then
  exit 0
fi

watch_pattern='(^feedback/.*\.md$|^harness/(prompts|policies|templates)/)'
matches="$(printf '%s\n' "$changed_files" | grep -E "$watch_pattern" || true)"

if [[ -z "${matches// }" ]]; then
  exit 0
fi

cat <<'EOF'
[skill-sync-reminder] Harness knowledge changed.

The following files may contain reusable agent lessons, anti-patterns, or workflow rules:
EOF

printf '%s\n' "$matches" | sed 's/^/  - /'

cat <<'EOF'

Please decide whether any durable, cross-project rule should be synchronized to the shared global agent layer:
  - skills: /home/azhi/.agents/skills/
  - harness prompts/workflows/tools: /home/azhi/.agents/harness/

Guideline:
  - feedback records are a sedimentation pool, not a behavior entrypoint by themselves.
  - If the change is only a project-local incident note or one-off context, it may stay in feedback without further projection.
  - Project fact or CTF-only path/policy -> keep in project harness.
  - Cross-project method, anti-pattern, checklist, or workflow -> update the relevant global skill or harness asset.
  - Current-task evidence -> keep in .harness/reuse-decisions/<task-slug>.md only.

This is a reminder only and does not block the commit.
EOF

exit 0
