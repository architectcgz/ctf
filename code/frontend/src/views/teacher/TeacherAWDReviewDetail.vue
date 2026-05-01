<script setup lang="ts">
import {
  Activity,
  ArrowLeft,
  ChevronRight,
  Clock,
  Download,
  FileDown,
  Shield,
  Target,
  TrendingUp,
  Trophy,
  Waypoints,
  Zap,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherAWDReviewTeamDrawer from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue'
import { useTeacherAwdReviewDetail } from '@/features/teacher-awd-review'
import { formatDate } from '@/utils/format'

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
                共 {{ summaryStats.teamCount }} 支参与队伍
              </div>
            </header>

            <div
              v-if="!selectedRound"
              class="awd-review-round-grid"
            >
              <article
                v-for="round in review.rounds"
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
                    class="teacher-btn teacher-btn--ghost teacher-btn--compact"
                    @click="setRound(round.round_number)"
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
                @click="openTeam(team)"
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

          <section
            v-if="selectedRound"
            class="awd-review-evidence-grid"
          >
            <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
              <header class="list-heading awd-review-evidence-head">
                <div>
                  <div class="journal-note-label">
                    Service Evidence
                  </div>
                  <h3 class="list-heading__title">
                    服务运行
                  </h3>
                </div>
                <Activity class="awd-review-evidence-icon awd-review-evidence-icon--service h-4 w-4" />
              </header>
              <div class="awd-review-evidence-list custom-scrollbar">
                <AppEmpty
                  v-if="selectedRound.services.length === 0"
                  icon="Shield"
                  title="无服务数据"
                  class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
                />
                <article
                  v-for="service in selectedRound.services"
                  v-else
                  :key="service.id"
                  class="awd-review-evidence-item"
                >
                  <div class="awd-review-evidence-item__head">
                    <strong>{{ service.team_name }} · {{ service.challenge_title }}</strong>
                    <span
                      v-if="service.service_id"
                      data-testid="awd-review-service-id"
                      class="teacher-directory-chip awd-review-service-chip"
                    >
                      {{ formatServiceRef(service.service_id) }}
                    </span>
                  </div>
                  <p>{{ service.service_status }} · SLA {{ service.sla_score }}</p>
                </article>
              </div>
            </article>

            <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
              <header class="list-heading awd-review-evidence-head">
                <div>
                  <div class="journal-note-label">
                    Attack Evidence
                  </div>
                  <h3 class="list-heading__title">
                    攻击记录
                  </h3>
                </div>
                <Target class="awd-review-evidence-icon awd-review-evidence-icon--attack h-4 w-4" />
              </header>
              <div class="awd-review-evidence-list custom-scrollbar">
                <AppEmpty
                  v-if="selectedRound.attacks.length === 0"
                  icon="Target"
                  title="无攻击记录"
                  class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
                />
                <article
                  v-for="attack in selectedRound.attacks"
                  v-else
                  :key="attack.id"
                  class="awd-review-evidence-item"
                >
                  <div class="awd-review-evidence-item__head">
                    <strong>{{ attack.attacker_team_name }} → {{ attack.victim_team_name }}</strong>
                    <span
                      v-if="attack.service_id"
                      data-testid="awd-review-attack-service-id"
                      class="teacher-directory-chip awd-review-service-chip"
                    >
                      {{ formatServiceRef(attack.service_id) }}
                    </span>
                  </div>
                  <p>{{ attack.challenge_title }} · {{ attack.attack_type }}</p>
                </article>
              </div>
            </article>

            <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
              <header class="list-heading awd-review-evidence-head">
                <div>
                  <div class="journal-note-label">
                    Traffic Evidence
                  </div>
                  <h3 class="list-heading__title">
                    流量审计
                  </h3>
                </div>
                <Waypoints class="awd-review-evidence-icon awd-review-evidence-icon--traffic h-4 w-4" />
              </header>
              <div class="awd-review-evidence-list custom-scrollbar">
                <AppEmpty
                  v-if="selectedRound.traffic.length === 0"
                  icon="Activity"
                  title="无流量证据"
                  class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
                />
                <article
                  v-for="event in selectedRound.traffic"
                  v-else
                  :key="event.id"
                  class="awd-review-evidence-item"
                >
                  <div class="awd-review-evidence-item__head">
                    <strong>{{ event.method }} {{ event.path }}</strong>
                    <span
                      v-if="event.service_id"
                      data-testid="awd-review-traffic-service-id"
                      class="teacher-directory-chip awd-review-service-chip"
                    >
                      {{ formatServiceRef(event.service_id) }}
                    </span>
                  </div>
                  <p>{{ event.attacker_team_name }} → {{ event.victim_team_name }} · {{ event.status_code }}</p>
                </article>
              </div>
            </article>
          </section>
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

.awd-review-analysis-section,
.awd-review-evidence-panel {
  background: color-mix(in srgb, var(--awd-review-surface) 98%, var(--color-bg-base));
}

.awd-review-round-grid,
.awd-review-evidence-grid {
  display: grid;
  gap: var(--space-4);
}

.awd-review-round-grid {
  grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
}

.awd-review-evidence-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
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

.awd-review-evidence-head {
  align-items: center;
}

.awd-review-evidence-icon {
  color: var(--awd-review-muted);
}

.awd-review-evidence-icon--service {
  color: color-mix(in srgb, var(--awd-review-success) 88%, var(--awd-review-text));
}

.awd-review-evidence-icon--attack {
  color: color-mix(in srgb, var(--awd-review-danger) 88%, var(--awd-review-text));
}

.awd-review-evidence-icon--traffic {
  color: color-mix(in srgb, var(--awd-review-blue) 88%, var(--awd-review-text));
}

.awd-review-evidence-list {
  display: grid;
  gap: var(--space-3);
  min-height: 14rem;
  max-height: 26rem;
  overflow-y: auto;
}

.awd-review-evidence-item {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--awd-review-line);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--awd-review-surface-subtle) 82%, transparent);
}

.awd-review-evidence-item__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.awd-review-evidence-item__head strong {
  color: var(--awd-review-text);
}

.awd-review-evidence-item p {
  margin: 0;
  color: var(--awd-review-muted);
  line-height: 1.6;
}

.awd-review-service-chip {
  flex-shrink: 0;
}

.awd-review-compact-empty {
  min-height: 100%;
}

.teacher-directory :deep(.teacher-directory-row),
.teacher-directory :deep(.teacher-directory-head) {
  border-color: var(--awd-review-line);
}

.teacher-directory :deep(.teacher-directory-row) {
  background: color-mix(in srgb, var(--awd-review-surface) 98%, var(--color-bg-base));
}

.teacher-directory :deep(.teacher-directory-row:hover),
.teacher-directory :deep(.teacher-directory-row:focus-visible) {
  background: color-mix(in srgb, var(--awd-review-primary) 8%, transparent);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1024px) {
  .awd-review-evidence-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .awd-review-round-card__head,
  .awd-review-evidence-item__head {
    flex-direction: column;
  }
}
</style>
