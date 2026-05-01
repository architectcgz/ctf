<script setup lang="ts">
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import {
  getChallengeCategoryLabel,
  getChallengeDifficultyLabel,
} from '@/entities/challenge/model'

interface Props {
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
}

defineProps<Props>()
</script>

<template>
  <div class="challenge-pill-row">
    <span class="challenge-table-pill challenge-table-pill--category">
      {{ getChallengeCategoryLabel(category) }}
    </span>
    <span class="challenge-table-pill challenge-table-pill--neutral">
      {{ getChallengeDifficultyLabel(difficulty) }}
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
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  background: color-mix(in srgb, var(--workspace-brand) 10%, var(--challenge-page-surface));
  color: var(--challenge-page-accent);
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.challenge-table-pill--neutral {
  background: color-mix(in srgb, var(--challenge-page-line) 18%, var(--challenge-page-surface));
  color: var(--challenge-page-muted);
  border: 1px solid color-mix(in srgb, var(--challenge-page-line-strong) 78%, transparent);
}
</style>
