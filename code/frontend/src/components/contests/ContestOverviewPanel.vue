<script setup lang="ts">
import { CheckCircle2, Target, Trophy, UsersRound } from 'lucide-vue-next'

import ContestAnnouncementsPanel from '@/components/contests/ContestAnnouncementsPanel.vue'
import type { ContestAnnouncement, ContestDetailData } from '@/api/contracts'
import { getModeLabel, getStatusLabel } from '@/utils/contest'
import { formatTime } from '@/utils/format'

interface Props {
  contest: ContestDetailData
  countdown: string
  totalPoints: number
  solvedCount: number
  memberCount: number
  challengeCount: number
  announcements: ContestAnnouncement[]
  announcementsError: string
}

defineProps<Props>()
</script>

<template>
  <section
    id="contest-workspace-panel-overview"
    class="workspace-panel contest-overview-panel"
    role="tabpanel"
    aria-labelledby="contest-workspace-tab-overview"
  >
    <header class="contest-hero workspace-panel-header">
      <div class="contest-hero__main workspace-panel-header__intro">
        <div class="workspace-overline">
          Contest
        </div>
        <h1 class="contest-hero__title workspace-page-title">
          {{ contest.title }}
        </h1>
        <p class="contest-hero__desc workspace-page-copy">
          {{ contest.description || '当前竞赛暂未提供描述。' }}
        </p>
      </div>
    </header>

    <div class="workspace-panel-divider" aria-hidden="true" />

    <section class="contest-stat-grid metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label progress-card-label metric-panel-label">
          <span>队伍成员</span>
          <UsersRound class="h-4 w-4" />
        </div>
        <div class="contest-stat__value progress-card-value metric-panel-value">
          {{ memberCount }}
        </div>
        <div class="contest-stat__hint progress-card-hint metric-panel-helper">
          当前队伍人数
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label progress-card-label metric-panel-label">
          <span>题目数量</span>
          <Target class="h-4 w-4" />
        </div>
        <div class="contest-stat__value progress-card-value metric-panel-value">
          {{ challengeCount }}
        </div>
        <div class="contest-stat__hint progress-card-hint metric-panel-helper">
          本场竞赛题目总数
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label progress-card-label metric-panel-label">
          <span>已解题目</span>
          <CheckCircle2 class="h-4 w-4" />
        </div>
        <div class="contest-stat__value progress-card-value metric-panel-value">
          {{ solvedCount }}
        </div>
        <div class="contest-stat__hint progress-card-hint metric-panel-helper">
          当前账号已完成数量
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label progress-card-label metric-panel-label">
          <span>积分总览</span>
          <Trophy class="h-4 w-4" />
        </div>
        <div class="contest-stat__value progress-card-value metric-panel-value">
          {{ totalPoints }}
        </div>
        <div class="contest-stat__hint progress-card-hint metric-panel-helper">
          全部题目可获得积分
        </div>
      </article>
    </section>

    <div class="workspace-panel-divider" aria-hidden="true" />

    <div class="contest-overview-grid">
      <section class="contest-section contest-section--flat contest-section--copy-tight">
        <div class="contest-section__head workspace-tab-heading">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Rules
            </div>
            <h2 class="contest-section__title workspace-tab-heading__title">
              竞赛规则
            </h2>
          </div>
        </div>
        <div class="contest-copy">
          {{ contest.rules || '当前竞赛暂无额外规则说明。' }}
        </div>
      </section>

      <section class="contest-section contest-section--flat">
        <div class="contest-section__head workspace-tab-heading">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Schedule
            </div>
            <h2 class="contest-section__title workspace-tab-heading__title">
              赛程信息
            </h2>
          </div>
        </div>
        <div class="contest-copy-list">
          <div class="contest-copy-row">
            <span>当前状态</span>
            <strong>{{ getStatusLabel(contest.status) }}</strong>
          </div>
          <div
            v-if="countdown"
            class="contest-copy-row"
          >
            <span>剩余时间</span>
            <strong>{{ countdown }}</strong>
          </div>
          <div class="contest-copy-row">
            <span>开始时间</span>
            <strong>{{ formatTime(contest.starts_at) }}</strong>
          </div>
          <div class="contest-copy-row">
            <span>结束时间</span>
            <strong>{{ formatTime(contest.ends_at) }}</strong>
          </div>
          <div class="contest-copy-row">
            <span>参赛模式</span>
            <strong>{{ getModeLabel(contest.mode) }}</strong>
          </div>
          <div class="contest-copy-row">
            <span>冻结榜单</span>
            <strong>{{ contest.scoreboard_frozen ? '是' : '否' }}</strong>
          </div>
        </div>
      </section>
    </div>

    <div class="workspace-panel-divider" aria-hidden="true" />

    <section class="contest-section contest-section--flat">
      <div class="contest-section__head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Announcements
          </div>
          <h2 class="contest-section__title workspace-tab-heading__title">
            公告预览
          </h2>
        </div>
        <div class="contest-section__hint">
          {{ announcements.length }} 条
        </div>
      </div>

      <ContestAnnouncementsPanel
        :announcements="announcements"
        :announcements-error="announcementsError"
        empty-variant="inline"
      />
    </section>
  </section>
</template>

<style scoped>
.contest-overview-panel {
  --workspace-panel-divider-gap: var(--space-divider-gap);
}

.contest-hero {
  min-width: 0;
}

.contest-hero__main {
  min-width: 0;
}

.contest-hero__title {
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-hero__desc {
  max-width: 60ch;
  color: var(--color-text-secondary);
}

.contest-stat-grid {
  --metric-panel-grid-gap: var(--space-3);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.contest-stat {
  min-width: 0;
}

.contest-overview-grid {
  display: grid;
  gap: var(--space-section-gap-compact);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-section--flat {
  display: grid;
  gap: var(--space-4);
}

.contest-section--copy-tight {
  align-content: start;
  gap: var(--space-2-5);
}

.contest-section--copy-tight > .contest-section__head {
  align-items: flex-start;
  justify-content: flex-start;
}

.contest-section--flat + .contest-section--flat {
  border-top: 0;
}

.contest-copy {
  width: 100%;
  justify-self: stretch;
  white-space: pre-wrap;
  text-align: start;
  font-size: var(--font-size-0-92);
  line-height: 1.8;
  color: var(--color-text-primary);
}

.contest-copy-list {
  display: grid;
  gap: var(--space-3);
}

.contest-copy-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2) var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  padding-bottom: var(--space-3);
  font-size: var(--font-size-0-88);
  color: var(--color-text-secondary);
}

.contest-copy-row strong {
  color: var(--color-text-primary);
}

@media (max-width: 1100px) {
  .contest-stat-grid,
  .contest-overview-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .contest-stat-grid,
  .contest-overview-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
