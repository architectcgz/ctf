<script setup lang="ts">
import {
  ArrowLeft,
  Clock,
  Download,
  FileDown,
  TrendingUp,
  Trophy,
  Zap,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherAWDReviewAnalysisSection from '@/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue'
import TeacherAWDReviewEvidenceGrid from '@/components/teacher/awd-review/TeacherAWDReviewEvidenceGrid.vue'
import TeacherAWDReviewTeamDrawer from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue'
import { useTeacherAwdReviewDetail } from '@/features/teacher-awd-review'

const {
  router,
  polling,
  loading,
  error,
  review,
  exporting,
  activeContestTitle,
  activeSummaryTitle,
  summaryStats,
  timelineRounds,
  selectedRoundNumber,
  selectedRound,
  selectedTeam,
  selectedTeamServices,
  selectedTeamAttacks,
  selectedTeamTraffic,
  canExportReport,
  loadReview,
  setRound,
  openTeam,
  closeTeam,
  contestStatusLabel,
  formatServiceRef,
  exportArchive,
  exportReport,
} = useTeacherAwdReviewDetail()
</script>

<template>
  <div class="teacher-management-shell teacher-surface workspace-shell flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero teacher-review-workspace flex min-h-full flex-1 flex-col border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar workspace-tab-heading awd-review-detail-header">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline awd-review-detail-overline">
              AWD Review
            </div>
            <h1 class="teacher-title workspace-page-title">
              AWD复盘
            </h1>
            <p class="teacher-copy workspace-page-copy">
              <span class="awd-review-detail-contest-title">{{ activeContestTitle }}</span>
              <span> · </span>
              多维复盘攻防实战过程。通过轮次下钻与流量回溯，协助教师评估学生的防御加固能力与漏洞挖掘表现。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="router.push({ name: 'TeacherAWDReviewIndex' })"
            >
              <ArrowLeft class="h-4 w-4" />
              返回列表
            </button>
            <button
              data-testid="awd-review-export-archive"
              type="button"
              class="teacher-btn teacher-btn--ghost"
              :disabled="loading || !review || exporting === 'archive'"
              @click="exportArchive"
            >
              <Download class="h-4 w-4" />
              归档导出
            </button>
            <button
              data-testid="awd-review-export-report"
              type="button"
              class="teacher-btn teacher-btn--primary"
              :disabled="loading || !review || exporting === 'report' || !canExportReport"
              @click="exportReport"
            >
              <FileDown class="h-4 w-4" />
              生成评估报告
            </button>
          </div>
        </header>

        <section class="teacher-summary teacher-summary--flat metric-panel-default-surface awd-review-summary">
          <div class="teacher-summary-title">
            <Trophy class="h-4 w-4" />
            <span>{{ activeSummaryTitle }}</span>
            <span
              v-if="review"
              class="awd-review-status-chip"
              :class="`awd-review-status-chip--${review.contest.status || 'idle'}`"
            >
              {{ contestStatusLabel(review.contest.status || '') }}
            </span>
          </div>

          <div class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                轮次范围
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ summaryStats.roundCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前视图覆盖的轮次数量
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                参与队伍
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ summaryStats.teamCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前视图包含的队伍数量
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                服务 / 攻击 / 流量
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ summaryStats.serviceCount }} / {{ summaryStats.attackCount }} / {{ summaryStats.trafficCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                证据总量与服务运行信号
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                导出状态
              </div>
              <div class="progress-card-value metric-panel-value awd-review-status-text">
                {{ polling ? '后台处理中...' : '链路就绪' }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                归档与教师报告导出链路状态
              </div>
            </article>
          </div>
        </section>

        <section class="workspace-directory-section teacher-directory-section awd-review-round-section">
          <header class="list-heading">
            <div>
              <div class="journal-note-label">
                Review Scope
              </div>
              <h3 class="list-heading__title">
                轮次切换
              </h3>
            </div>
            <div class="teacher-directory-meta">
              默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。
            </div>
          </header>

          <div class="awd-review-round-list custom-scrollbar">
            <button
              type="button"
              class="teacher-directory-chip awd-review-round-chip"
              :class="{ 'awd-review-round-chip--active': !selectedRoundNumber }"
              @click="setRound(undefined)"
            >
              整场总览
            </button>
            <button
              v-for="round in timelineRounds"
              :key="round.id"
              type="button"
              class="teacher-directory-chip awd-review-round-chip"
              :class="{ 'awd-review-round-chip--active': selectedRoundNumber === round.round_number }"
              @click="setRound(round.round_number)"
            >
              R{{ round.round_number }}
            </button>
          </div>
        </section>

        <div
          v-if="loading"
          class="teacher-empty-state workspace-directory-empty awd-review-loading"
        >
          <div class="academy-spinner" />
          <p>正在载入复盘分析数据...</p>
        </div>

        <AppEmpty
          v-else-if="error"
          title="复盘详情加载失败"
          :description="error"
          icon="AlertCircle"
          class="teacher-empty-state workspace-directory-empty"
        >
          <template #action>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="loadReview"
            >
              重新加载
            </button>
          </template>
        </AppEmpty>

        <template v-else-if="review">
          <TeacherAWDReviewAnalysisSection
            :active-summary-title="activeSummaryTitle"
            :rounds="review.rounds"
            :selected-round="selectedRound"
            :team-count="summaryStats.teamCount"
            @set-round="setRound"
            @open-team="openTeam"
          />

          <TeacherAWDReviewEvidenceGrid
            v-if="selectedRound"
            :selected-round="selectedRound"
            :format-service-ref="formatServiceRef"
          />
        </template>
      </div>

      <TeacherAWDReviewTeamDrawer
        :visible="Boolean(selectedTeam)"
        :team="selectedTeam"
        :services="selectedTeamServices"
        :attacks="selectedTeamAttacks"
        :traffic="selectedTeamTraffic"
        @close="closeTeam"
      />
    </section>
  </div>
</template>

<style scoped>
.teacher-review-workspace {
  --awd-review-surface: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base));
  --awd-review-surface-subtle: color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 92%, var(--color-bg-base));
  --awd-review-line: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 74%, transparent);
  --awd-review-line-strong: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 84%, transparent);
  --awd-review-text: var(--journal-ink, var(--color-text-primary));
  --awd-review-muted: var(--journal-muted, var(--color-text-secondary));
  --awd-review-faint: color-mix(in srgb, var(--color-text-muted) 88%, var(--awd-review-muted));
  --awd-review-primary: var(--journal-accent, var(--color-primary));
  --awd-review-primary-strong: var(--journal-accent-strong, var(--color-primary-hover));
  --awd-review-success: var(--color-success);
  --awd-review-warning: var(--color-warning);
  --awd-review-danger: var(--color-danger);
  --awd-review-blue: var(--color-brand-swatch-blue);
  gap: var(--space-6);
}

.awd-review-detail-header {
  gap: var(--space-4);
}

.awd-review-summary {
  --metric-panel-border: color-mix(in srgb, var(--awd-review-primary) 14%, var(--awd-review-line));
}

.awd-review-status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  padding: 0 var(--space-3);
  border-radius: 999px;
  border: 1px solid var(--awd-review-line);
  background: color-mix(in srgb, var(--awd-review-line) 14%, var(--awd-review-surface));
  color: var(--awd-review-muted);
  font-size: var(--font-size-0-74);
  font-weight: 700;
}

.awd-review-status-chip--running {
  border-color: color-mix(in srgb, var(--awd-review-primary) 24%, transparent);
  background: color-mix(in srgb, var(--awd-review-primary) 10%, transparent);
  color: var(--awd-review-primary-strong);
}

.awd-review-status-chip--ended,
.awd-review-status-chip--frozen {
  border-color: color-mix(in srgb, var(--awd-review-line-strong) 86%, transparent);
  background: color-mix(in srgb, var(--awd-review-line) 18%, var(--awd-review-surface-subtle));
  color: var(--awd-review-muted);
}

.awd-review-status-text {
  color: var(--awd-review-muted);
}

.awd-review-round-list {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--space-3);
  overflow-x: auto;
}

.awd-review-round-chip {
  border: 1px solid var(--awd-review-line);
  background: color-mix(in srgb, var(--awd-review-surface-subtle) 92%, transparent);
  color: var(--awd-review-muted);
}

.awd-review-round-chip--active {
  border-color: color-mix(in srgb, var(--awd-review-primary) 28%, transparent);
  background: color-mix(in srgb, var(--awd-review-primary) 12%, transparent);
  color: var(--awd-review-primary-strong);
}

.awd-review-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  min-height: 14rem;
  color: var(--awd-review-muted);
}

.academy-spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid color-mix(in srgb, var(--awd-review-line) 88%, transparent);
  border-top-color: var(--awd-review-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
