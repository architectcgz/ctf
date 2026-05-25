<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import { formatDate } from '@/utils/format'

interface RoundSummary {
  id: string
  round_number: number
  service_count: number
  attack_count: number
  traffic_count: number
}

interface TeamSummary {
  team_id: string
  team_name: string
  captain_id: string
  total_score: number
  member_count: number
  last_solve_at?: string
}

interface SelectedRoundSummary {
  teams: TeamSummary[]
}

defineProps<{
  activeSummaryTitle: string
  rounds: RoundSummary[]
  selectedRound?: SelectedRoundSummary
  teamCount: number
}>()

const emit = defineEmits<{
  setRound: [roundNumber?: number]
  openTeam: [team: TeamSummary]
}>()
</script>

<template>
  <section class="workspace-directory-section teacher-directory-section awd-review-analysis-section">
    <header class="list-heading">
      <div>
        <div class="journal-note-label">
          Performance Analysis
        </div>
        <h3 class="list-heading__title">
          {{ activeSummaryTitle }} 表现分析
        </h3>
      </div>
      <div class="teacher-directory-meta">
        共 {{ teamCount }} 支参与队伍
      </div>
    </header>

    <div
      v-if="!selectedRound"
      class="awd-review-round-grid"
    >
      <article
        v-for="round in rounds"
        :key="round.id"
        class="metric-panel-card awd-review-round-card"
      >
        <div class="awd-review-round-card__head">
          <div>
            <div class="journal-note-label">
              Round {{ round.round_number }}
            </div>
            <strong class="awd-review-round-card__title">第 {{ round.round_number }} 轮</strong>
          </div>
          <button
            type="button"
            class="header-btn header-btn--ghost header-btn--compact"
            @click="emit('setRound', round.round_number)"
          >
            下钻分析
            <ChevronRight class="h-3.5 w-3.5" />
          </button>
        </div>
        <div class="awd-review-round-card__metrics">
          <span>服务 {{ round.service_count }}</span>
          <span>攻击 {{ round.attack_count }}</span>
          <span>流量 {{ round.traffic_count }}</span>
        </div>
      </article>
    </div>

    <AppEmpty
      v-else-if="selectedRound.teams.length === 0"
      class="teacher-empty-state workspace-directory-empty"
      icon="Users"
      title="当前轮次暂无队伍数据"
      description="该轮次还没有可展示的队伍表现。"
    />

    <section
      v-else
      class="teacher-directory"
    >
      <div class="teacher-directory-head">
        <span class="teacher-directory-head-cell teacher-directory-head-cell-name">队伍</span>
        <span>得分表现</span>
        <span>命中记录</span>
        <span>成员结构</span>
        <span>操作</span>
      </div>

      <button
        v-for="team in selectedRound.teams"
        :key="team.team_id"
        type="button"
        class="teacher-directory-row"
        @click="emit('openTeam', team)"
      >
        <div class="teacher-directory-cell teacher-directory-cell-name">
          <div class="awd-review-team-name">
            <span class="awd-review-team-dot" />
            <strong>{{ team.team_name }}</strong>
          </div>
        </div>
        <div class="teacher-directory-row-metrics awd-review-team-score">
          <strong>{{ team.total_score }}</strong>
          <span class="awd-review-team-score-suffix">pts</span>
        </div>
        <div class="teacher-directory-row-metrics">
          <span>{{ team.last_solve_at ? formatDate(team.last_solve_at) : '暂无命中' }}</span>
        </div>
        <div class="teacher-directory-row-metrics">
          <span>{{ team.member_count }} 成员</span>
          <span>UID: {{ team.captain_id }}</span>
        </div>
        <div class="teacher-directory-row-cta">
          <span class="teacher-directory-chip">调阅细节</span>
        </div>
      </button>
    </section>
  </section>
</template>

<style scoped>
.awd-review-analysis-section {
  background: color-mix(in srgb, var(--awd-review-surface) 98%, var(--color-bg-base));
}

.awd-review-round-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
}

.awd-review-round-card {
  display: grid;
  gap: var(--space-4);
}

.awd-review-round-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.awd-review-round-card__title {
  color: var(--awd-review-text);
}

.awd-review-round-card__metrics {
  display: grid;
  gap: var(--space-2);
  color: var(--awd-review-muted);
}

.awd-review-team-name {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  color: var(--awd-review-text);
}

.awd-review-team-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--awd-review-success) 88%, var(--awd-review-primary));
  flex-shrink: 0;
}

.awd-review-team-score {
  color: color-mix(in srgb, var(--awd-review-success) 90%, var(--awd-review-text));
}

.awd-review-team-score-suffix {
  color: var(--awd-review-faint);
  font-family: var(--font-family-sans);
}

.teacher-directory-row,
.teacher-directory-head {
  border-color: var(--awd-review-line);
}

.teacher-directory-row {
  background: color-mix(in srgb, var(--awd-review-surface) 98%, var(--color-bg-base));
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--awd-review-primary) 8%, transparent);
}

@media (max-width: 768px) {
  .awd-review-round-card__head {
    flex-direction: column;
  }
}
</style>
