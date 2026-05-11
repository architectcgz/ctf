<script setup lang="ts">
import type { ChallengeDifficulty } from '@/api/contracts'
import { getChallengeDifficultyColor, getChallengeDifficultyLabel } from '@/entities/challenge/model'

interface Props {
  difficulty: ChallengeDifficulty
  labelOverrides?: Partial<Record<ChallengeDifficulty, string>>
}

defineProps<Props>()

function difficultyTextStyle(difficulty: ChallengeDifficulty): Record<string, string> {
  return {
    '--challenge-difficulty-text-color': getChallengeDifficultyColor(difficulty),
  }
}
</script>

<template>
  <span class="challenge-difficulty-text" :style="difficultyTextStyle(difficulty)">
    {{ labelOverrides?.[difficulty] ?? getChallengeDifficultyLabel(difficulty) }}
  </span>
</template>

<style scoped>
.challenge-difficulty-text {
  color: var(--challenge-difficulty-text-color, var(--challenge-page-muted, var(--journal-muted)));
}
</style>
