<script setup lang="ts">
import type { ChallengeDifficulty } from '@/api/contracts'
import { getChallengeDifficultyLabel } from '@/entities/challenge/model'

interface Props {
  difficulty: ChallengeDifficulty
}

defineProps<Props>()
</script>

<template>
  <span class="challenge-difficulty-text">
    {{ getChallengeDifficultyLabel(difficulty) }}
  </span>
</template>

<style scoped>
.challenge-difficulty-text {
  color: var(--challenge-page-muted, var(--journal-muted));
}
</style>
