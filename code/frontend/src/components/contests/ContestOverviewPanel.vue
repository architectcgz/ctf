<script setup lang="ts">
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
    class="workspace-panel"
    role="tabpanel"
    aria-labelledby="contest-workspace-tab-overview"
  >
    <header class="contest-hero">
      <div class="contest-hero__main">
        <div class="workspace-overline">
          Contest
        </div>
        <h1 class="contest-hero__title workspace-page-title">
          {{ contest.title }}
        </h1>
        <p class="contest-hero__desc workspace-page-copy">
          {{ contest.description || '当前竞赛暂未提供描述。' }}
        </p>

        <div class="contest-meta-strip">
          <span class="contest-chip contest-chip--status">
            {{ getStatusLabel(contest.status) }}
          </span>
          <span class="contest-chip contest-chip--neutral">
            {{ getModeLabel(contest.mode) }}
          </span>
          <span class="contest-chip contest-chip--neutral">
            {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
          </span>
          <span
            v-if="countdown"
            class="contest-chip contest-chip--accent"
          >
            {{ countdown }}
          </span>
        </div>
      </div>

      <aside class="contest-score-rail metric-panel-card metric-panel-workspace-surface">
        <div class="contest-score-rail__label metric-panel-label">
          总分
        </div>
        <div class="contest-score-rail__value metric-panel-value">
          {{ totalPoints }} <small>分</small>
        </div>
        <div class="contest-score-rail__note metric-panel-helper">
          {{ challengeCount }} 题 · {{ solvedCount }} 已解 · {{ memberCount }} 人
        </div>
      </aside>
    </header>

    <div class="contest-divider" />

    <section class="contest-stat-grid metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label metric-panel-label">
          队伍成员
        </div>
        <div class="contest-stat__value metric-panel-value">
          {{ memberCount }}
        </div>
        <div class="contest-stat__hint metric-panel-helper">
          当前队伍人数
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label metric-panel-label">
          题目数量
        </div>
        <div class="contest-stat__value metric-panel-value">
          {{ challengeCount }}
        </div>
        <div class="contest-stat__hint metric-panel-helper">
          本场竞赛题目总数
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label metric-panel-label">
          已解题目
        </div>
        <div class="contest-stat__value metric-panel-value">
          {{ solvedCount }}
        </div>
        <div class="contest-stat__hint metric-panel-helper">
          当前账号已完成数量
        </div>
      </article>
      <article class="contest-stat progress-card metric-panel-card">
        <div class="contest-stat__label metric-panel-label">
          积分总览
        </div>
        <div class="contest-stat__value metric-panel-value">
          {{ totalPoints }}
        </div>
        <div class="contest-stat__hint metric-panel-helper">
          全部题目可获得积分
        </div>
      </article>
    </section>

    <div class="contest-divider" />

    <div class="contest-overview-grid">
      <section class="contest-section contest-section--flat">
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

    <div class="contest-divider" />

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
.contest-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 15rem;
  gap: var(--space-5);
}

.contest-hero__title {
  margin-top: var(--space-3-5);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-hero__desc {
  margin-top: var(--space-3);
  max-width: 60ch;
  color: var(--color-text-secondary);
}

.contest-meta-strip {
  margin-top: var(--space-4);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.contest-score-rail {
  --metric-panel-padding: var(--space-4);
  --metric-panel-value-size: var(--font-size-34);
  align-self: start;
}

.contest-score-rail__value small {
  font-size: var(--font-size-0-85);
  font-weight: 700;
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
  gap: var(--space-5);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-section--flat + .contest-section--flat {
  border-top: 0;
}

.contest-copy,
.contest-copy-list {
  margin-top: var(--space-4);
}

.contest-copy {
  white-space: pre-wrap;
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
  .contest-hero {
    grid-template-columns: 1fr;
  }

  .contest-score-rail {
    width: 100%;
  }

  .contest-stat-grid,
  .contest-overview-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
