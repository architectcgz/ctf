<script setup lang="ts">
import { Activity, Target, Users } from 'lucide-vue-next'

import type { TeacherClassSummaryData, TeacherStudentItem } from '@/api/contracts'

const props = defineProps<{
  selectedClassName: string
  students: TeacherStudentItem[]
  summary: TeacherClassSummaryData | null
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
}>()

function averageSolvedText() {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
}

function activeRateText() {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
}
</script>

<template>
  <section>
    <header class="workspace-panel-header class-overview-topbar">
      <div class="workspace-panel-header__intro teacher-heading">
        <div class="workspace-overline">
          Class Snapshot
        </div>
        <h2 class="teacher-title workspace-page-title class-overview-title">
          {{ selectedClassName || '班级概览' }}
        </h2>
      </div>

      <div class="workspace-panel-header__actions header-actions">
        <button
          type="button"
          class="header-btn header-btn--ghost"
          @click="emit('openClassManagement')"
        >
          返回
        </button>
        <button
          type="button"
          class="header-btn header-btn--ghost"
          @click="emit('openDashboard')"
        >
          概览
        </button>
        <button
          type="button"
          class="header-btn header-btn--primary"
          @click="emit('openReportExport')"
        >
          导出班级报告
        </button>
      </div>

      <div
        class="workspace-panel-header__summary teacher-summary-grid class-overview-summary progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>班级人数</span>
            <Users class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ summary?.student_count ?? students.length }}
          </div>
          <div class="progress-card-hint metric-panel-helper">当前班级学生总数</div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>平均解题</span>
            <Target class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ averageSolvedText() }}
          </div>
          <div class="progress-card-hint metric-panel-helper">当前班级人均完成题目数</div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>当前窗口活跃率</span>
            <Activity class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ activeRateText() }}
          </div>
          <div class="progress-card-hint metric-panel-helper">
            当前时间段内至少产生训练事件的学生占比
          </div>
        </article>
      </div>
    </header>

    <div v-if="error" class="workspace-alert" role="alert">
      <div class="workspace-alert-title">加载失败</div>
      <div class="workspace-alert-copy">
        {{ error }}
      </div>
      <div class="workspace-alert-actions">
        <button type="button" class="ui-btn ui-btn--primary" @click="emit('retry')">
          重试加载
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.class-overview-title {
  max-width: min(100%, 38rem);
}

.class-overview-summary {
  padding: 0;
}
</style>
