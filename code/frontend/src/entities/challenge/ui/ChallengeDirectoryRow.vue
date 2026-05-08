<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import type { ChallengeListItem } from '@/api/contracts'
import {
  getChallengeCategoryColor,
  getChallengeCategoryLabel,
  getChallengeDifficultyColor,
  getChallengeDifficultyLabel,
} from '@/entities/challenge/model'

interface Props {
  challenge: ChallengeListItem
}

defineProps<Props>()

defineEmits<{
  open: [challengeId: string]
}>()
</script>

<template>
  <button
    type="button"
    class="challenge-row"
    :style="{ '--challenge-row-accent': getChallengeCategoryColor(challenge.category) }"
    :aria-label="`${challenge.title}，${getChallengeCategoryLabel(challenge.category)}，${getChallengeDifficultyLabel(challenge.difficulty)}，${challenge.is_solved ? '已解出' : '待攻克'}`"
    @click="$emit('open', challenge.id)"
  >
    <div class="challenge-row-main">
      <div class="challenge-row-title-group">
        <h2 class="challenge-row-title" :title="challenge.title">
          {{ challenge.title }}
        </h2>
      </div>
    </div>

    <div class="challenge-row-points">
      <span class="challenge-row-points-value">{{ challenge.points }}</span>
      <span class="challenge-row-points-unit"> pts</span>
    </div>

    <div class="challenge-row-category">
      <span
        class="challenge-chip"
        :style="{
          '--challenge-chip-bg': `${getChallengeCategoryColor(challenge.category)}18`,
          '--challenge-chip-color': getChallengeCategoryColor(challenge.category),
        }"
      >
        {{ getChallengeCategoryLabel(challenge.category) }}
      </span>
    </div>

    <div class="challenge-row-difficulty">
      <span
        class="challenge-chip"
        :style="{
          '--challenge-chip-bg': `${getChallengeDifficultyColor(challenge.difficulty)}18`,
          '--challenge-chip-color': getChallengeDifficultyColor(challenge.difficulty),
        }"
      >
        {{ getChallengeDifficultyLabel(challenge.difficulty) }}
      </span>
    </div>

    <div class="challenge-row-tags">
      <template v-if="challenge.tags.length > 0">
        <span
          v-for="tag in challenge.tags.slice(0, 2)"
          :key="tag"
          class="challenge-chip challenge-chip-muted"
        >
          {{ tag }}
        </span>
      </template>
      <span v-else class="challenge-chip challenge-chip-muted challenge-chip-muted--placeholder">
        -
      </span>
    </div>

    <div class="challenge-row-status">
      <span
        class="challenge-state-chip"
        :class="challenge.is_solved ? 'challenge-state-chip-solved' : 'challenge-state-chip-ready'"
      >
        {{ challenge.is_solved ? '已解出' : '待攻克' }}
      </span>
    </div>

    <div class="challenge-row-solved">{{ challenge.solved_count }} 人解出</div>

    <div class="challenge-row-attempts">尝试 {{ challenge.total_attempts }} 次</div>

    <div class="challenge-row-cta">
      <span
        class="challenge-row-cta-pill"
        :class="
          challenge.is_solved ? 'challenge-row-cta-pill--solved' : 'challenge-row-cta-pill--ready'
        "
      >
        <span>{{ challenge.is_solved ? '继续查看' : '开始做题' }}</span>
        <ArrowRight class="h-4 w-4" />
      </span>
    </div>
  </button>
</template>

<style scoped>
.challenge-row {
  display: grid;
  grid-template-columns: var(--challenge-directory-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4) 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease,
    box-shadow 160ms ease;
}

.challenge-row:hover,
.challenge-row:focus-visible {
  background: color-mix(
    in srgb,
    var(--challenge-row-accent, var(--journal-accent)) 5%,
    transparent
  );
  box-shadow: inset 2px 0 0
    color-mix(in srgb, var(--challenge-row-accent, var(--journal-accent)) 58%, transparent);
}

.challenge-row:focus-visible {
  outline: 2px solid
    color-mix(in srgb, var(--challenge-row-accent, var(--journal-accent)) 36%, transparent);
  outline-offset: -2px;
}

.challenge-row-main {
  display: flex;
  align-items: center;
  min-width: 0;
}

.challenge-row-title-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1-5);
  min-width: 0;
}

.challenge-row-title {
  margin: 0;
  overflow: hidden;
  font-size: var(--font-size-1-06);
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--journal-ink);
}

.challenge-row-points {
  display: inline-flex;
  align-items: baseline;
  gap: var(--space-1);
  font-variant-numeric: tabular-nums;
}

.challenge-row-points-value {
  font-size: var(--font-size-0-98);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-row-points-unit {
  font-size: var(--font-size-12);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-row-category,
.challenge-row-difficulty {
  display: flex;
  align-items: center;
  min-width: 0;
}

.challenge-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.challenge-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  max-width: 100%;
  padding: 0 var(--space-3);
  border-radius: 999px;
  background: var(
    --challenge-chip-bg,
    color-mix(in srgb, var(--journal-surface-subtle) 88%, transparent)
  );
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--challenge-chip-color, var(--journal-ink));
}

.challenge-chip-muted {
  --challenge-chip-bg: color-mix(in srgb, var(--journal-surface-subtle) 88%, transparent);
  --challenge-chip-color: var(--journal-muted);
}

.challenge-chip-muted--placeholder {
  justify-content: center;
  min-width: calc(var(--space-6) + var(--space-5));
}

.challenge-row-status {
  display: flex;
  align-items: center;
}

.challenge-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 var(--space-3);
  border-radius: 999px;
  font-size: var(--font-size-12);
  font-weight: 700;
}

.challenge-state-chip-solved {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-success);
}

.challenge-state-chip-ready {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: var(--color-warning);
}

.challenge-row-solved,
.challenge-row-attempts {
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.challenge-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
}

.challenge-row-cta-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-sm);
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--challenge-row-accent) 24%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--challenge-row-accent) 8%, transparent);
  font-size: var(--font-size-13);
  font-weight: 700;
}

.challenge-row-cta-pill--ready {
  color: color-mix(in srgb, var(--challenge-row-accent) 82%, var(--journal-ink));
}

.challenge-row-cta-pill--solved {
  border-color: color-mix(in srgb, var(--color-success) 26%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: color-mix(in srgb, var(--color-success) 84%, var(--journal-ink));
}

@media (max-width: 960px) {
  .challenge-row {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-2-5);
    padding: var(--space-4) 0;
  }

  .challenge-row-cta {
    justify-content: flex-start;
  }
}
</style>
