<script setup lang="ts">
import AwdEventTimeline from '@/modules/awd/components/AwdEventTimeline.vue'
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

import { useStudentAwdWorkspacePage } from './useStudentAwdWorkspacePage'

const { pageModel, layoutProps } = useStudentAwdWorkspacePage('overview')
</script>

<template>
  <StudentAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-student-page">
      <section class="awd-student-section">
        <div class="awd-student-section__head">
          <div>
            <div class="workspace-overline">Scoreboard</div>
            <h2>当前榜单</h2>
          </div>
          <div class="awd-student-section__hint">
            {{ pageModel.overview.scoreboardRows.length }} 支队伍
          </div>
        </div>

        <div
          v-if="pageModel.overview.scoreboardRows.length === 0"
          class="awd-student-inline-note"
        >
          当前还没有可展示的实时榜单。
        </div>

        <div
          v-else
          class="awd-student-list"
        >
          <article
            v-for="row in pageModel.overview.scoreboardRows"
            :key="row.team_id"
            class="awd-student-list__row"
          >
            <div class="awd-student-list__main">
              <strong>#{{ row.rank }} {{ row.team_name }}</strong>
              <span>解题数 {{ row.solved_count }}</span>
            </div>
            <div class="awd-student-list__value">
              {{ row.score }}
            </div>
          </article>
        </div>
      </section>

      <section class="awd-student-section">
        <div class="awd-student-section__head">
          <div>
            <div class="workspace-overline">Defense</div>
            <h2>防守告警</h2>
          </div>
          <div class="awd-student-section__hint">
            {{ pageModel.overview.defenseAlerts.length }} 项
          </div>
        </div>

        <div
          v-if="pageModel.overview.defenseAlerts.length === 0"
          class="awd-student-inline-note"
        >
          当前轮次没有本队服务异常。
        </div>

        <div
          v-else
          class="awd-student-list"
        >
          <article
            v-for="alert in pageModel.overview.defenseAlerts"
            :key="alert.challengeId"
            class="awd-student-list__row"
          >
            <div class="awd-student-list__main">
              <strong>{{ alert.challengeTitle }}</strong>
              <span>{{ alert.issues.join('；') }}</span>
            </div>
            <div class="awd-student-list__badge">
              {{ alert.statusLabel }}
            </div>
          </article>
        </div>
      </section>

      <section class="awd-student-section">
        <div class="awd-student-section__head">
          <div>
            <div class="workspace-overline">Live Feed</div>
            <h2>战场事件流</h2>
          </div>
        </div>
        <AwdEventTimeline :items="pageModel.overview.recentEvents" empty-text="当前没有新的攻防事件。" />
      </section>
    </div>
  </StudentAwdWorkspaceLayout>
</template>

<style scoped>
.awd-student-page {
  display: grid;
  gap: 1rem;
}

.awd-student-section {
  display: grid;
  gap: 0.9rem;
}

.awd-student-section__head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
}

.awd-student-section__head h2 {
  margin: 0.3rem 0 0;
  font-size: var(--font-size-18);
}

.awd-student-section__hint {
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}

.awd-student-list {
  display: grid;
  gap: 0.75rem;
}

.awd-student-list__row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  padding: 0.95rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
}

.awd-student-list__main {
  display: grid;
  gap: 0.25rem;
}

.awd-student-list__main span {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}

.awd-student-list__value,
.awd-student-list__badge {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-14);
  font-weight: 600;
}

.awd-student-inline-note {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}
</style>
