<script setup lang="ts">
import type { ChallengeCategory } from '@/api/contracts'
import { getChallengeCategoryLabel } from '@/entities/challenge/model'

interface Props {
  category: ChallengeCategory
}

defineProps<Props>()
</script>

<template>
  <span class="challenge-category-text">
    {{ getChallengeCategoryLabel(category) }}
  </span>
</template>

<style scoped>
.challenge-category-text {
  color: inherit;
}
</style>
