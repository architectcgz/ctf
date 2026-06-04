<template>
  <div v-if="loading" class="insight-overview-layout">
    <StudentInsightLoadingSurface class="insight-overview-loading-surface">
      <div class="insight-overview-loading-head">
        <span class="student-insight-skeleton-line insight-overview-loading-title" />
        <span class="student-insight-skeleton-line insight-overview-loading-subtitle" />
      </div>
      <span class="student-insight-skeleton-panel insight-overview-loading-radar" />
    </StudentInsightLoadingSurface>

    <StudentInsightLoadingSurface class="insight-overview-loading-surface">
      <div class="insight-overview-loading-head">
        <span class="student-insight-skeleton-line insight-overview-loading-title insight-overview-loading-title--short" />
        <span class="student-insight-skeleton-line insight-overview-loading-subtitle" />
      </div>
      <div class="insight-overview-loading-bars">
        <div
          v-for="index in 6"
          :key="index"
          class="insight-overview-loading-bar-row"
        >
          <span class="student-insight-skeleton-line insight-overview-loading-bar-label" />
          <span class="student-insight-skeleton-line insight-overview-loading-bar-track" />
          <span class="student-insight-skeleton-line insight-overview-loading-bar-value" />
        </div>
      </div>
    </StudentInsightLoadingSurface>
  </div>

  <div v-else class="insight-overview-layout">
    <SectionCard
      class="insight-overview-card"
      variant="teacher-flat"
      title="六维能力分布"
      subtitle="雷达图展示当前六个能力维度的训练分布。"
    >
      <div class="mt-4">
        <SkillRadar :scores="radarScores" />
      </div>
    </SectionCard>

    <SectionCard
      class="insight-overview-card"
      variant="teacher-flat"
      title="维度得分占比"
      subtitle="条状图展示各维度当前分值。"
    >
      <div class="insight-dimension-frame mt-4">
        <div v-if="rankedProfileDimensions.length > 0" class="insight-dimension-bars">
          <article
            v-for="item in rankedProfileDimensions"
            :key="item.key"
            class="insight-dimension-item"
          >
            <div class="insight-dimension-item__head">
              <strong>{{ item.name }}</strong>
              <span>{{ item.value }}%</span>
            </div>
            <div class="insight-dimension-item__track">
              <div class="insight-dimension-item__fill" :style="{ width: `${item.value}%` }" />
            </div>
          </article>
        </div>
        <div v-else class="insight-dimension-empty">暂无画像维度数据</div>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
.insight-overview-layout {
  display: grid;
  gap: var(--space-6);
  grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
  align-items: start;
  padding-top: var(--space-5);
}

.insight-overview-card {
  --section-card-border-top-width: 0;
}

.insight-overview-loading-head {
  display: grid;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.insight-overview-loading-title {
  width: min(16rem, 68%);
  height: var(--space-4);
}

.insight-overview-loading-title--short {
  width: min(12rem, 54%);
}

.insight-overview-loading-subtitle {
  width: min(24rem, 86%);
  height: var(--space-3);
}

.insight-overview-loading-radar {
  width: 100%;
  aspect-ratio: 1 / 1;
  max-height: 280px;
}

.insight-overview-loading-bars {
  display: grid;
  gap: var(--space-3-5);
}

.insight-overview-loading-bar-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) var(--space-12);
  align-items: center;
  gap: var(--space-3);
}

.insight-overview-loading-bar-label {
  width: min(8rem, 72%);
  height: var(--space-3);
}

.insight-overview-loading-bar-track {
  width: 100%;
  height: var(--space-2-5);
}

.insight-overview-loading-bar-value {
  width: var(--space-12);
  height: var(--space-3);
}

/* ── Real content ── */

.insight-dimension-bars {
  display: grid;
  gap: var(--space-3-5);
}

.insight-dimension-frame {
  display: grid;
  gap: var(--space-3-5);
  min-height: 100%;
  padding: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  border-radius: var(--workspace-radius-lg);
  background:
    radial-gradient(
      ellipse at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 44%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 82%, var(--color-bg-base))
    );
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 78%, transparent),
    0 18px 32px -28px color-mix(in srgb, var(--journal-shadow) 26%, transparent);
}

.insight-dimension-item {
  display: grid;
  gap: var(--space-2);
}

.insight-dimension-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  color: var(--journal-ink);
}

.insight-dimension-item__head span {
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.insight-dimension-item__track {
  height: 10px;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-border) 36%, transparent);
}

.insight-dimension-item__fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink)),
    color-mix(in srgb, var(--journal-accent) 48%, white) 58%,
    color-mix(in srgb, var(--color-warning) 84%, var(--journal-accent))
  );
}

.insight-dimension-empty {
  font-size: var(--font-size-0-84);
  color: var(--journal-muted);
}

@media (max-width: 767px) {
  .insight-overview-layout {
    grid-template-columns: 1fr;
  }

  .insight-overview-loading-bar-row {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .insight-overview-loading-bar-value {
    display: none;
  }
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

import type { SkillProfileData } from '@/api/contracts'
import { StudentInsightLoadingSurface } from '@/features/teaching/student-analysis-shared/ui'
import { toRadarScores } from '@/entities/skill-profile'
import SectionCard from '@/shared/ui/common/SectionCard.vue'
import SkillRadar from '@/shared/ui/common/SkillRadar.vue'

const props = defineProps<{
  profile: SkillProfileData | null
  loading?: boolean
}>()

const radarScores = computed(() => toRadarScores(props.profile))
const rankedProfileDimensions = computed(() =>
  [...(props.profile?.dimensions ?? [])].sort((left, right) => right.value - left.value)
)
</script>
