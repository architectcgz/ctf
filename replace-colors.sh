#!/bin/bash

# 批量替换硬编码颜色为 CSS 变量

files=(
  "code/frontend/src/views/admin/ChallengeManage.vue"
  "code/frontend/src/views/admin/ImageManage.vue"
  "code/frontend/src/views/challenges/ChallengeDetail.vue"
  "code/frontend/src/views/challenges/ChallengeList.vue"
  "code/frontend/src/views/admin/ContestManage.vue"
  "code/frontend/src/views/admin/UserManage.vue"
  "code/frontend/src/views/contests/ContestDetail.vue"
  "code/frontend/src/views/contests/ContestList.vue"
  "code/frontend/src/views/instances/InstanceList.vue"
  "code/frontend/src/views/scoreboard/ScoreboardView.vue"
  "code/frontend/src/components/common/SkillRadar.vue"
)

for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    echo "处理: $file"

    # 背景色
    sed -i 's/bg-\[#0f1117\]/bg-[var(--color-bg-base)]/g' "$file"
    sed -i 's/bg-\[#0d1117\]/bg-[var(--color-bg-base)]/g' "$file"
    sed -i 's/bg-\[#161b22\]/bg-[var(--color-bg-surface)]/g' "$file"
    sed -i 's/bg-\[#1c2128\]/bg-[var(--color-bg-elevated)]/g' "$file"
    sed -i 's/bg-\[#111723\]/bg-[var(--color-bg-surface)]/g' "$file"

    # 边框色
    sed -i 's/border-\[#30363d\]/border-[var(--color-border-default)]/g' "$file"
    sed -i 's/border-\[#21262d\]/border-[var(--color-border-subtle)]/g' "$file"

    # 文字色
    sed -i 's/text-\[#e6edf3\]/text-[var(--color-text-primary)]/g' "$file"
    sed -i 's/text-\[#c9d1d9\]/text-[var(--color-text-primary)]/g' "$file"
    sed -i 's/text-\[#8b949e\]/text-[var(--color-text-secondary)]/g' "$file"
    sed -i 's/text-\[#6e7681\]/text-[var(--color-text-muted)]/g' "$file"

    # placeholder 色
    sed -i 's/placeholder-\[#6e7681\]/placeholder-[var(--color-text-muted)]/g' "$file"

    # 主题色
    sed -i 's/border-t-\[#0891b2\]/border-t-[var(--color-primary)]/g' "$file"
    sed -i 's/focus:border-\[#0891b2\]/focus:border-[var(--color-primary)]/g' "$file"
    sed -i 's/hover:border-\[#0891b2\]/hover:border-[var(--color-primary)]/g' "$file"
    sed -i 's/bg-\[#0891b2\]/bg-[var(--color-primary)]/g' "$file"
    sed -i 's/text-\[#0891b2\]/text-[var(--color-primary)]/g' "$file"

  fi
done

echo "替换完成！"
