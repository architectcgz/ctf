<template>
  <SectionCard
    class="insight-tab-section-card"
    variant="teacher-flat"
    title="复盘工作台"
    subtitle="按攻击会话查看访问、请求、提交和复盘输出。"
  >
    <div
      v-if="summaryItems.length > 0"
      class="insight-kpi-grid teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface md:grid-cols-4"
    >
      <article
        v-for="item in summaryItems"
        :key="item.key"
        class="insight-kpi-card progress-card metric-panel-card"
      >
        <div class="insight-kpi-label progress-card-label metric-panel-label">
          <span>{{ item.label }}</span>
          <component :is="item.icon" class="h-4 w-4" />
        </div>
        <div class="insight-kpi-value progress-card-value metric-panel-value">
          {{ item.value }}
        </div>
        <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
          {{ item.hint }}
        </div>
      </article>
    </div>

    <StudentInsightStateSurface
      class="evidence-state-surface"
      :loading="reviewWorkspaceLoading && !attackSessions"
      :empty="!reviewWorkspaceLoading && (!attackSessions || attackSessions.sessions.length === 0)"
    >
      <template #loading>
        <div class="evidence-loading-filters">
          <span class="student-insight-skeleton-panel evidence-loading-filter" />
          <span class="student-insight-skeleton-panel evidence-loading-filter" />
          <span class="student-insight-skeleton-panel evidence-loading-filter" />
        </div>
        <div class="evidence-loading-summary">
          <span
            v-for="index in 4"
            :key="index"
            class="student-insight-skeleton-panel evidence-loading-summary-card"
          />
        </div>
        <div class="evidence-loading-sessions">
          <div
            v-for="index in 3"
            :key="index"
            class="evidence-loading-session"
          >
            <span class="student-insight-skeleton-panel evidence-loading-session-icon" />
            <span class="student-insight-skeleton-line evidence-loading-session-text" />
            <span class="student-insight-skeleton-line evidence-loading-session-meta" />
          </div>
        </div>
      </template>

      <template #empty>
        <AppEmpty
          class="student-insight-empty"
          title="暂无攻击会话"
          description="当前学员还没有可用于复盘的攻击过程记录。"
          icon="NotebookText"
        />
      </template>

      <StudentReviewWorkspace
        :evidence="evidence"
        :attack-sessions="attackSessions"
        :challenge-options="reviewChallengeOptions"
        :loading="reviewWorkspaceLoading"
        :query="reviewWorkspaceQuery"
        @update-filters="emit('updateReviewWorkspaceFilters', $event)"
      />
    </StudentInsightStateSurface>
  </SectionCard>
</template>

<style scoped>
.insight-tab-section-card {
  --section-card-border-top-width: 0;
  --section-card-header-border-bottom: 0;
}

.insight-kpi-grid {
  --teacher-summary-columns: repeat(4, minmax(0, 1fr));
  align-items: stretch;
}

.evidence-state-surface {
  --student-insight-state-gap: var(--space-4);
  margin-top: var(--space-5);
}

.evidence-loading-filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.evidence-loading-filter {
  height: var(--space-10);
}

.evidence-loading-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.evidence-loading-summary-card {
  height: var(--space-16);
}

.evidence-loading-sessions {
  display: grid;
  gap: var(--space-2);
  padding-top: var(--space-2);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 68%, transparent);
}

.evidence-loading-session {
  display: grid;
  grid-template-columns: var(--space-10) minmax(0, 1fr) var(--space-18);
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) 0;
}

.evidence-loading-session-icon {
  width: var(--space-10);
  height: var(--space-10);
}

.evidence-loading-session-text {
  width: min(28rem, 76%);
  height: var(--space-3);
}

.evidence-loading-session-meta {
  width: var(--space-18);
  height: var(--space-3);
}

@media (max-width: 767px) {
  .evidence-loading-filters,
  .evidence-loading-summary {
    grid-template-columns: 1fr;
  }

  .evidence-loading-session-meta {
    display: none;
  }
}
</style>

<script setup lang="ts">
import type { AttackSessionQuery, AttackSessionResponseData, StudentEvidenceData } from '@/api/contracts'
import { computed } from 'vue'
import { StudentInsightStateSurface } from '@/features/teaching/student-analysis-shared/ui'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import SectionCard from '@/shared/ui/common/SectionCard.vue'
import StudentReviewWorkspace from './StudentReviewWorkspace.vue'
import { buildReviewWorkspaceSummaryItems } from './studentInsightShared'

const props = defineProps<{
  attackSessions: AttackSessionResponseData | null
  evidence: StudentEvidenceData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: AttackSessionQuery
}>()

const emit = defineEmits<{
  updateReviewWorkspaceFilters: [payload: Partial<AttackSessionQuery>]
}>()

const summaryItems = computed(() =>
  props.attackSessions && props.attackSessions.sessions.length > 0
    ? buildReviewWorkspaceSummaryItems({
        sessionSummary: props.attackSessions.summary,
        evidenceSummary: props.evidence?.summary,
      })
    : []
)
</script>
