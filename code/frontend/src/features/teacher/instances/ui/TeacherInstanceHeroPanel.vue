<script setup lang="ts">
import { Activity, Clock3, Eye } from 'lucide-vue-next'

import AppRouteLink from '@/components/navigation/AppRouteLink.vue'

interface DashboardRouteTarget {
  name: string
}

defineProps<{
  totalCount: number
  runningCount: number
  expiringSoonCount: number
  dashboardRoute: DashboardRouteTarget
}>()
</script>

<template>
  <div>
    <header class="workspace-page-header teacher-topbar">
      <div class="teacher-heading workspace-tab-heading__main">
        <div class="workspace-overline">Teacher Instance Ops</div>
        <h1 class="teacher-title workspace-page-title">实例管理</h1>
        <p class="teacher-copy workspace-page-copy">先筛班级与学员，再快速定位异常或即将到期的训练实例。</p>
      </div>

      <div class="header-actions">
        <AppRouteLink :to="dashboardRoute" class="header-btn header-btn--primary">
          返回教学概览
        </AppRouteLink>
      </div>
    </header>

    <section class="teacher-summary metric-panel-default-surface">
      <div class="teacher-summary-title">
        <span>Instance Snapshot</span>
      </div>
      <div class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>当前可见</span>
            <Eye class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ totalCount }}
          </div>
          <div class="progress-card-hint metric-panel-helper">符合当前筛选条件的实例数量</div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>运行中</span>
            <Activity class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ runningCount }}
          </div>
          <div class="progress-card-hint metric-panel-helper">仍在占用环境资源的实例数量</div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            <span>即将到期</span>
            <Clock3 class="h-4 w-4" />
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ expiringSoonCount }}
          </div>
          <div class="progress-card-hint metric-panel-helper">剩余时间不足 10 分钟的实例数量</div>
        </article>
      </div>
    </section>
  </div>
</template>
