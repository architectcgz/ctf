<script setup lang="ts">
import type { ChallengeDetailData } from '@/api/contracts'
import {
  getChallengeCategoryColor,
  getChallengeCategoryLabel,
  getChallengeDifficultyColor,
  getChallengeDifficultyLabel,
} from '@/entities/challenge/model'

interface Props {
  challenge: ChallengeDetailData
}

defineProps<Props>()

function buildMetaPillStyle(color: string): Record<string, string> {
  return {
    '--brand-soft': `${color}18`,
    '--brand-ink': color,
    '--brand': color,
  }
}
</script>

<template>
  <div class="meta-strip">
    <span
      class="meta-pill meta-pill--brand"
      :style="buildMetaPillStyle(getChallengeCategoryColor(challenge.category))"
    >
      {{ getChallengeCategoryLabel(challenge.category) }}
    </span>
    <span
      class="meta-pill"
      :style="buildMetaPillStyle(getChallengeDifficultyColor(challenge.difficulty))"
    >
      {{ getChallengeDifficultyLabel(challenge.difficulty) }}
    </span>
    <span
      v-if="challenge.is_solved"
      class="meta-pill"
    >
      已解出
    </span>
    <span
      v-if="challenge.attachment_url"
      class="meta-pill"
    >
      附件可下载
    </span>
    <span
      v-for="tag in challenge.tags"
      :key="tag"
      class="meta-pill"
    >
      {{ tag }}
    </span>
  </div>
</template>

<style scoped>
.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--text-subtle);
}

.meta-pill--brand {
  border-color: color-mix(in srgb, var(--brand) 20%, transparent);
  background: var(--brand-soft);
  color: var(--brand-ink);
}
</style>
