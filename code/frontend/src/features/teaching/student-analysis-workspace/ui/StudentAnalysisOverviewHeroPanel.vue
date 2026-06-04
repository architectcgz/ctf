<template>
  <header class="workspace-panel-header student-analysis-overview-head">
    <div class="workspace-panel-header__intro teacher-heading">
      <div class="workspace-overline">
        Student Insight
      </div>
      <h1 class="teacher-title workspace-page-title student-analysis-title">
        {{ selectedStudent?.name || selectedStudent?.username || '学员分析' }}
      </h1>
    </div>

    <div
      class="workspace-panel-header__actions header-actions"
      role="group"
      aria-label="学员分析快捷操作"
    >
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="emit('openClassStudents')"
      >
        返回学生列表
      </button>
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="emit('openReportExport')"
      >
        导出班级报告
      </button>
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="emit('openReviewArchive')"
      >
        完整复盘页
      </button>
      <button
        type="button"
        class="header-btn header-btn--primary"
        @click="emit('exportReviewArchive')"
      >
        导出复盘归档
      </button>
    </div>

    <div class="workspace-panel-header__summary summary-strip metric-panel-grid">
      <template v-if="loading">
        <article
          v-for="card in loadingMetricCards"
          :key="card.key"
          class="summary-card summary-card--loading progress-card metric-panel-card"
        >
          <div class="summary-card__label progress-card-label metric-panel-label">
            <span class="student-insight-skeleton-line summary-card-loading-label" />
            <span class="student-insight-skeleton-pill summary-card-loading-icon" />
          </div>
          <div class="summary-card__value progress-card-value metric-panel-value">
            <span :class="['student-insight-skeleton-line', 'summary-card-loading-value', card.valueClass]" />
          </div>
          <div class="summary-card__hint progress-card-hint metric-panel-helper">
            <span class="student-insight-skeleton-line summary-card-loading-hint" />
          </div>
        </article>
      </template>

      <template v-else>
        <article class="summary-card summary-card--solved progress-card metric-panel-card">
          <div class="summary-card__label progress-card-label metric-panel-label">
            <span>已做题目数</span>
            <CheckCircle class="h-4 w-4" />
          </div>
          <div class="summary-card__value progress-card-value metric-panel-value">
            {{ progress?.solved_challenges ?? 0 }}
          </div>
          <div class="summary-card__hint progress-card-hint metric-panel-helper">
            已成功完成的题目数量
          </div>
        </article>
        <article class="summary-card summary-card--completion progress-card metric-panel-card">
          <div class="summary-card__label progress-card-label metric-panel-label">
            <span>完成率</span>
            <Trophy class="h-4 w-4" />
          </div>
          <div class="summary-card__value progress-card-value metric-panel-value">
            {{ solvedRate }}%
          </div>
          <div class="summary-card__hint progress-card-hint metric-panel-helper">
            基于当前学员训练数据计算
          </div>
        </article>
        <article class="summary-card summary-card--weakness progress-card metric-panel-card">
          <div class="summary-card__label progress-card-label metric-panel-label">
            <span>薄弱维度</span>
            <AlertTriangle class="h-4 w-4" />
          </div>
          <div class="summary-card__value progress-card-value metric-panel-value">
            {{ weakDimensions.length > 0 ? weakDimensions.join('、') : '暂无' }}
          </div>
          <div class="summary-card__hint progress-card-hint metric-panel-helper">
            基于能力画像提炼的风险点
          </div>
        </article>
      </template>
    </div>
  </header>

</template>

<style src="../../student-analysis-shared/ui/studentInsightSurface.css"></style>

<style scoped>
.student-analysis-title {
  --workspace-page-title-margin-top: 0;
  max-width: min(100%, 38rem);
}

.summary-strip {
  --metric-panel-grid-gap: var(--space-2-5) var(--space-4);
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
  margin: var(--space-6) 0 0;
  padding: 0;
}

.summary-card {
  --summary-card-accent: var(--workspace-brand);
  min-width: 0;
  --metric-panel-border: color-mix(in srgb, var(--summary-card-accent) 18%, var(--teacher-card-border));
  --metric-panel-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--summary-card-accent) 15%, transparent),
      transparent 48%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--workspace-panel) 94%, var(--summary-card-accent)),
      color-mix(in srgb, var(--workspace-panel) 90%, var(--color-bg-base))
    );
  --metric-panel-shadow: var(--workspace-shadow-panel);
  --metric-panel-label-color: color-mix(in srgb, var(--summary-card-accent) 58%, var(--journal-muted));
  --metric-panel-value-color: color-mix(in srgb, var(--summary-card-accent) 76%, var(--journal-ink));
}

.summary-card--solved {
  --summary-card-accent: var(--color-primary);
}

.summary-card--completion {
  --summary-card-accent: var(--workspace-brand);
}

.summary-card--weakness {
  --summary-card-accent: var(--workspace-brand);
}

.summary-card .summary-card__label :is(svg, .lucide) {
  color: color-mix(in srgb, var(--summary-card-accent) 82%, var(--journal-ink));
}

.summary-card--loading {
  --summary-card-accent: var(--workspace-brand);
  pointer-events: none;
}

.summary-card-loading-label {
  width: min(6rem, 52%);
  height: var(--space-3);
}

.summary-card-loading-icon {
  width: var(--space-4);
  height: var(--space-4);
  flex: 0 0 auto;
}

.summary-card-loading-value {
  height: clamp(2rem, 4vw, 2.4rem);
  width: min(6rem, 58%);
}

.summary-card-loading-value--wide {
  width: min(10rem, 84%);
}

.summary-card-loading-hint {
  width: min(11rem, 88%);
  height: var(--space-3);
}

@media (max-width: 1023px) {
  .summary-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .summary-strip {
    grid-template-columns: 1fr;
  }

  .header-actions {
    width: 100%;
  }

  .header-btn {
    flex: 1 1 100%;
  }
}
</style>

<script setup lang="ts">
import { AlertTriangle, CheckCircle, Trophy } from 'lucide-vue-next'

import type { MyProgressData, StudentDirectoryItem } from '@/api/contracts'

defineProps<{
  selectedStudent: StudentDirectoryItem | null
  progress: MyProgressData | null
  solvedRate: number
  weakDimensions: string[]
  loading?: boolean
}>()

const emit = defineEmits<{
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
}>()

const loadingMetricCards = [
  { key: 'solved', valueClass: '' },
  { key: 'rate', valueClass: '' },
  { key: 'weakness', valueClass: 'summary-card-loading-value--wide' },
]
</script>
