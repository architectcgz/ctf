<script setup lang="ts">
import type { ChallengeCategory } from '@/api/contracts'
import { getChallengeCategoryLabel } from '@/entities/challenge/model'

interface Props {
  category: ChallengeCategory
}

defineProps<Props>()
</script>

<template>
  <span class="challenge-table-pill challenge-table-pill--category">
    {{ getChallengeCategoryLabel(category) }}
  </span>
</template>

<style scoped>
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
</style>
