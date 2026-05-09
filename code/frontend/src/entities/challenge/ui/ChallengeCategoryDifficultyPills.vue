<script setup lang="ts">
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import ChallengeCategoryPill from './ChallengeCategoryPill.vue'
import ChallengeDifficultyText from './ChallengeDifficultyText.vue'

interface Props {
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
}

defineProps<Props>()
</script>

<template>
  <div class="challenge-pill-row">
    <ChallengeCategoryPill :category="category" />
    <span
      class="challenge-table-pill challenge-table-pill--neutral"
      :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--muted']"
    >
      <ChallengeDifficultyText :difficulty="difficulty" />
    </span>
  </div>
</template>

<style scoped>
.challenge-pill-row {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.challenge-table-pill {
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--neutral {
  background: color-mix(in srgb, var(--challenge-page-line) 18%, var(--challenge-page-surface));
  border: 1px solid color-mix(in srgb, var(--challenge-page-line-strong) 78%, transparent);
}
</style>
