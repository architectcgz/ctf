<script setup lang="ts">
import type { ChallengeDifficulty } from '@/api/contracts'
import { getChallengeDifficultyLabel } from '@/entities/challenge/model'

interface Props {
  difficulty: ChallengeDifficulty
  labelOverrides?: Partial<Record<ChallengeDifficulty, string>>
}

defineProps<Props>()
</script>

<template>
  <span class="challenge-difficulty-text">
    {{ labelOverrides?.[difficulty] ?? getChallengeDifficultyLabel(difficulty) }}
  </span>
</template>

<style scoped>
.challenge-difficulty-text {
  color: var(--challenge-page-muted, var(--journal-muted));
}
</style>
