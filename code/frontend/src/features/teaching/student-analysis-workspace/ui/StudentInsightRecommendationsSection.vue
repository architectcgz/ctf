<template>
  <SectionCard
    class="insight-tab-section-card"
    variant="teacher-flat"
    title="推荐训练任务"
    subtitle="根据当前薄弱维度筛出的优先训练题目。"
  >
    <AppEmpty
      v-if="recommendations.length === 0"
      title="暂无推荐题目"
      description="当前画像还没有生成新的推荐训练任务。"
      icon="BookOpen"
    />

    <div v-else class="insight-recommendation-list workspace-directory-list">
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
    </div>
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
  --workspace-directory-shell-padding: var(--space-2) var(--space-4);
  margin-top: var(--space-5);
}

.insight-recommendation-row {
  gap: var(--space-4);
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
}
</style>

<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import type { RecommendationItem } from '@/api/contracts'
import { ChallengeCategoryDifficultyPills } from '@/entities/challenge'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import SectionCard from '@/shared/ui/common/SectionCard.vue'

defineProps<{
  recommendations: RecommendationItem[]
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
}>()
</script>
