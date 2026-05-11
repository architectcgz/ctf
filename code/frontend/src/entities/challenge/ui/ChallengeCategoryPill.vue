<script setup lang="ts">
import type { ChallengeCategory } from '@/api/contracts'
import { getChallengeCategoryColor, getChallengeCategoryLabel } from '@/entities/challenge/model'

interface Props {
  category: ChallengeCategory
}

defineProps<Props>()

function categoryPillStyle(category: ChallengeCategory): Record<string, string> {
  const color = getChallengeCategoryColor(category)
  return {
    '--challenge-table-pill-color': color,
    '--challenge-table-pill-bg': `color-mix(in srgb, ${color} 10%, transparent)`,
    '--challenge-table-pill-border': `color-mix(in srgb, ${color} 22%, transparent)`,
  }
}
</script>

<template>
  <span
    class="challenge-table-pill challenge-table-pill--category"
    :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--primary']"
    :style="categoryPillStyle(category)"
  >
    {{ getChallengeCategoryLabel(category) }}
  </span>
</template>

<style scoped>
.challenge-table-pill {
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  border: 1px solid var(--challenge-table-pill-border);
  background: var(--challenge-table-pill-bg);
  color: var(--challenge-table-pill-color);
}
</style>
