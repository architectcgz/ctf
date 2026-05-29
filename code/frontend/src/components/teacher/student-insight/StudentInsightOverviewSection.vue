<script setup lang="ts">
import { computed } from 'vue'

import type { SkillProfileData } from '@/api/contracts'
import SectionCard from '@/components/common/SectionCard.vue'
import SkillRadar from '@/components/common/SkillRadar.vue'
import { toRadarScores } from '@/entities/skill-profile'

const props = defineProps<{
  profile: SkillProfileData | null
}>()

const radarScores = computed(() => toRadarScores(props.profile))
const rankedProfileDimensions = computed(() =>
  [...(props.profile?.dimensions ?? [])].sort((left, right) => right.value - left.value)
)
</script>

<template>
  <div class="insight-overview-layout">
    <SectionCard title="六维能力分布" subtitle="雷达图展示当前六个能力维度的训练分布。">
      <div class="mt-4">
        <SkillRadar :scores="radarScores" />
      </div>
    </SectionCard>

    <SectionCard title="维度得分占比" subtitle="条状图展示各维度当前分值。">
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
  padding-top: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
}

.insight-overview-layout :deep(.section-card) {
  border-top: 0;
}

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
}
</style>
