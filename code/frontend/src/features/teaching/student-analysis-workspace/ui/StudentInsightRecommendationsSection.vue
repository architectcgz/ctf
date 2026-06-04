<template>
  <SectionCard
    class="insight-tab-section-card"
    variant="teacher-flat"
    title="推荐训练任务"
    subtitle="根据当前薄弱维度筛出的优先训练题目。"
  >
    <StudentInsightStateSurface
      class="insight-recommendation-list workspace-directory-list"
      :loading="loading"
      :empty="!loading && recommendations.length === 0"
      surface="plain"
    >
      <template #loading>
        <div class="insight-recommendation-skeleton-head">
          <span class="student-insight-skeleton-line insight-recommendation-skeleton-kicker" />
          <span class="student-insight-skeleton-line insight-recommendation-skeleton-count" />
        </div>
        <div
          v-for="index in 2"
          :key="index"
          class="insight-recommendation-skeleton-row"
        >
          <div class="insight-recommendation-skeleton-main">
            <span class="student-insight-skeleton-line insight-recommendation-skeleton-title" />
            <span class="student-insight-skeleton-line insight-recommendation-skeleton-copy" />
            <span class="student-insight-skeleton-line insight-recommendation-skeleton-evidence" />
          </div>
          <div class="insight-recommendation-skeleton-pills">
            <span class="student-insight-skeleton-pill" />
            <span class="student-insight-skeleton-pill" />
          </div>
          <span class="student-insight-skeleton-line insight-recommendation-skeleton-action" />
        </div>
      </template>

      <template #empty>
        <AppEmpty
          class="student-insight-empty"
          title="暂无推荐题目"
          description="当前画像还没有生成新的推荐训练任务。"
          icon="BookOpen"
        />
      </template>

      <template #default>
        <button
          v-for="item in recommendations"
          :key="item.challenge_id"
          type="button"
          class="insight-recommendation-row workspace-directory-grid-row"
          @click="emit('openChallenge', item.challenge_id)"
        >
          <div class="workspace-directory-cell insight-recommendation-main">
            <h5 class="workspace-directory-row-title">
              {{ item.title }}
            </h5>
            <p class="workspace-directory-row-subtitle">
              {{ item.summary }}
            </p>
            <p
              v-if="item.evidence"
              class="insight-recommendation-evidence"
            >
              {{ item.evidence }}
            </p>
          </div>
          <div class="insight-recommendation-pills">
            <ChallengeCategoryDifficultyPills
              :category="item.category"
              :difficulty="item.difficulty"
            />
          </div>
          <span class="workspace-directory-row-btn insight-recommendation-action">
            <span>查看题目</span>
            <ArrowRight class="h-4 w-4" />
          </span>
        </button>
      </template>
    </StudentInsightStateSurface>
  </SectionCard>
</template>

<style scoped>
.insight-tab-section-card {
  --section-card-border-top-width: 0;
  --section-card-header-border-bottom: 0;
}

.insight-recommendation-pills {
  display: inline-flex;
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.insight-recommendation-list {
  --workspace-directory-grid-columns: minmax(0, 1fr) auto auto;
  --workspace-directory-shell-border: color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      ellipse at top right,
      color-mix(in srgb, var(--journal-accent) 9%, transparent),
      transparent 46%
    ),
    radial-gradient(
      ellipse at bottom left,
      color-mix(in srgb, var(--color-bg-surface) 58%, transparent),
      transparent 52%
    ),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base))
    );
  --workspace-directory-shell-radius: var(--workspace-radius-lg);
  --workspace-directory-shell-padding: var(--space-3) var(--space-4);
  --workspace-directory-row-divider: color-mix(
    in srgb,
    var(--workspace-directory-shell-border) 82%,
    transparent
  );
  margin-top: var(--space-5);
  position: relative;
  overflow: hidden;
  box-shadow: var(--workspace-shadow-panel);
}

.insight-recommendation-list::before {
  position: absolute;
  inset: 1px;
  pointer-events: none;
  content: '';
  border-radius: calc(var(--workspace-radius-lg) - 1px);
  background:
    linear-gradient(
      115deg,
      transparent 0%,
      color-mix(in srgb, var(--color-bg-surface) 34%, transparent) 38%,
      transparent 72%
    );
  opacity: 0.54;
}

.insight-recommendation-row {
  position: relative;
  gap: var(--space-4);
}

.insight-recommendation-list.student-insight-state-surface--empty {
  display: grid;
  min-height: clamp(11rem, 24vh, 16rem);
  place-items: center;
}

.insight-recommendation-list.student-insight-state-surface--loading {
  display: grid;
  gap: var(--space-2-5);
  min-height: clamp(12rem, 24vh, 16rem);
  align-content: start;
}

.insight-recommendation-skeleton-row {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: center;
  gap: var(--space-4);
  min-height: var(--space-16);
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.insight-recommendation-skeleton-row:last-child {
  border-bottom-color: transparent;
}

.insight-recommendation-skeleton-head {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-1);
}

.insight-recommendation-skeleton-main,
.insight-recommendation-skeleton-pills {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.insight-recommendation-skeleton-pills {
  grid-template-columns: repeat(2, minmax(var(--space-12), 1fr));
}

.insight-recommendation-skeleton-kicker {
  width: min(12rem, 52%);
  height: var(--space-3);
}

.insight-recommendation-skeleton-count {
  width: var(--space-14);
  height: var(--space-3);
}

.insight-recommendation-skeleton-title {
  width: min(22rem, 74%);
  height: var(--space-4);
}

.insight-recommendation-skeleton-copy {
  width: min(34rem, 92%);
  height: var(--space-3);
}

.insight-recommendation-skeleton-evidence {
  width: min(26rem, 68%);
  height: var(--space-2);
}

.insight-recommendation-skeleton-pills span {
  height: var(--space-6);
}

.insight-recommendation-skeleton-action {
  width: var(--space-18);
  height: var(--space-7);
}

.insight-recommendation-main {
  display: grid;
  gap: var(--space-1);
  align-content: center;
}

.insight-recommendation-evidence {
  margin: var(--space-1) 0 0;
  color: color-mix(in srgb, var(--journal-muted) 86%, transparent);
  font-size: var(--font-size-12);
  line-height: 1.65;
}

.insight-recommendation-action {
  justify-self: end;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .insight-recommendation-list {
    --workspace-directory-grid-columns: minmax(0, 1fr);
  }

  .insight-recommendation-row {
    gap: var(--space-3);
  }

  .insight-recommendation-pills,
  .insight-recommendation-action {
    justify-self: start;
  }

  .insight-recommendation-skeleton-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .insight-recommendation-skeleton-pills,
  .insight-recommendation-skeleton-action {
    justify-self: start;
  }
}
</style>

<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import type { RecommendationItem } from '@/api/contracts'
import { StudentInsightStateSurface } from '@/features/teaching/student-analysis-shared/ui'
import { ChallengeCategoryDifficultyPills } from '@/entities/challenge'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import SectionCard from '@/shared/ui/common/SectionCard.vue'

defineProps<{
  recommendations: RecommendationItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
}>()
</script>
