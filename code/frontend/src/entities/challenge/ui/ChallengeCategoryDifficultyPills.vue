<script setup lang="ts">
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import { getChallengeDifficultyColor } from '@/entities/challenge/model'
import ChallengeCategoryPill from './ChallengeCategoryPill.vue'
import ChallengeDifficultyText from './ChallengeDifficultyText.vue'

interface Props {
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
}

defineProps<Props>()

function difficultyPillStyle(difficulty: ChallengeDifficulty): Record<string, string> {
  const color = getChallengeDifficultyColor(difficulty)
  return {
    '--challenge-table-pill-color': color,
    '--challenge-table-pill-bg': `color-mix(in srgb, ${color} 10%, transparent)`,
    '--challenge-table-pill-border': `color-mix(in srgb, ${color} 22%, transparent)`,
  }
}
</script>

<template>
  <div class="challenge-pill-row">
    <ChallengeCategoryPill :category="category" />
    <span
      class="challenge-table-pill challenge-table-pill--neutral"
      :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--muted']"
      :style="difficultyPillStyle(difficulty)"
    >
      <ChallengeDifficultyText :difficulty="difficulty" />
    </span>
  </div>
</template>

<style scoped>
.challenge-pill-row {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.challenge-table-pill {
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--neutral {
  border: 1px solid var(--challenge-table-pill-border);
  background: var(--challenge-table-pill-bg);
  color: var(--challenge-table-pill-color);
}
</style>
