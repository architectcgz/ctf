<template>
  <div v-if="loading" class="insight-overview-layout">
    <div class="insight-overview-glass">
      <div class="insight-overview-glass__head">
        <span class="insight-overview-glass__title" />
        <span class="insight-overview-glass__subtitle" />
      </div>
      <div class="insight-overview-glass__radar" />
    </div>

    <div class="insight-overview-glass">
      <div class="insight-overview-glass__head">
        <span class="insight-overview-glass__title insight-overview-glass__title--short" />
        <span class="insight-overview-glass__subtitle" />
      </div>
      <div class="insight-overview-glass__bars">
        <div
          v-for="index in 6"
          :key="index"
          class="insight-overview-glass__bar-row"
        >
          <span class="insight-overview-glass__bar-label" />
          <span class="insight-overview-glass__bar-track" />
          <span class="insight-overview-glass__bar-value" />
        </div>
      </div>
    </div>
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
      <div v-if="rankedProfileDimensions.length > 0" class="insight-dimension-bars mt-4">
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
      <div v-else class="insight-dimension-empty mt-4">暂无画像维度数据</div>
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

/* ── Glass skeleton ── */

.insight-overview-glass {
  position: relative;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  border-radius: var(--workspace-radius-lg);
  padding: var(--space-4);
  background:
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
  box-shadow: var(--workspace-shadow-panel);
}

.insight-overview-glass::before {
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

.insight-overview-glass__head {
  display: grid;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.insight-overview-glass__title,
.insight-overview-glass__subtitle,
.insight-overview-glass__radar,
.insight-overview-glass__bar-label,
.insight-overview-glass__bar-track,
.insight-overview-glass__bar-value {
  display: block;
  overflow: hidden;
  border-radius: 999px;
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: insightOverviewSkeletonSweep 1.55s ease-in-out infinite;
}

.insight-overview-glass__title {
  width: min(16rem, 68%);
  height: var(--space-4);
}

.insight-overview-glass__title--short {
  width: min(12rem, 54%);
}

.insight-overview-glass__subtitle {
  width: min(24rem, 86%);
  height: var(--space-3);
}

.insight-overview-glass__radar {
  width: 100%;
  aspect-ratio: 1 / 1;
  max-height: 280px;
  border-radius: var(--workspace-radius-lg);
}

.insight-overview-glass__bars {
  display: grid;
  gap: var(--space-3-5);
}

.insight-overview-glass__bar-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) var(--space-12);
  align-items: center;
  gap: var(--space-3);
}

.insight-overview-glass__bar-label {
  width: min(8rem, 72%);
  height: var(--space-3);
}

.insight-overview-glass__bar-track {
  width: 100%;
  height: var(--space-2-5);
}

.insight-overview-glass__bar-value {
  width: var(--space-12);
  height: var(--space-3);
}

/* ── Real content ── */

.insight-dimension-bars {
  display: grid;
  gap: var(--space-3-5);
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

  .insight-overview-glass__bar-row {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .insight-overview-glass__bar-value {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .insight-overview-glass__title,
  .insight-overview-glass__subtitle,
  .insight-overview-glass__radar,
  .insight-overview-glass__bar-label,
  .insight-overview-glass__bar-track,
  .insight-overview-glass__bar-value {
    animation: none;
  }
}

@keyframes insightOverviewSkeletonSweep {
  0% {
    background-position: 120% 0;
  }

  100% {
    background-position: -120% 0;
  }
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

import type { SkillProfileData } from '@/api/contracts'
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
