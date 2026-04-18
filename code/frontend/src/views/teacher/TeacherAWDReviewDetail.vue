<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, Download, FileDown, Shield, Waypoints } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherAWDReviewTeamDrawer from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue'
import { useTeacherAwdReviewDetail } from '@/composables/useTeacherAwdReviewDetail'
import { formatDate } from '@/utils/format'

const {
  router,
  polling,
  loading,
  error,
  review,
  exporting,
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
  exportArchive,
  exportReport,
} = useTeacherAwdReviewDetail()

const activeTitle = computed(() => review.value?.contest.title || 'AWD复盘')
const activeSummaryTitle = computed(() =>
  selectedRoundNumber.value ? `第 ${selectedRoundNumber.value} 轮` : '整场总览'
)
const summaryStats = computed(() => {
  if (selectedRound.value) {
    return {
      roundCount: 1,
      teamCount: selectedRound.value.teams.length,
      serviceCount: selectedRound.value.round.service_count,
      attackCount: selectedRound.value.round.attack_count,
      trafficCount: selectedRound.value.round.traffic_count,
      latestEvidenceAt: undefined,
    }
  }

  return {
    roundCount: review.value?.overview?.round_count ?? 0,
    teamCount: review.value?.overview?.team_count ?? 0,
    serviceCount: review.value?.overview?.service_count ?? 0,
    attackCount: review.value?.overview?.attack_count ?? 0,
    trafficCount: review.value?.overview?.traffic_count ?? 0,
    latestEvidenceAt: review.value?.overview?.latest_evidence_at,
  }
})

function contestStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'ended':
      return '已结束'
    case 'frozen':
      return '冻结中'
    default:
      return status || '未开始'
  }
}

function roundStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '当前轮'
    case 'finished':
      return '已完成'
    default:
      return status || '待开始'
  }
}

function formatServiceRef(serviceId?: string): string {
  return `Service #${serviceId || '--'}`
}
</script>

<template>
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">AWD Review Workspace</div>
            <h1 class="teacher-title">{{ activeTitle }}</h1>
            <p class="teacher-copy">
              AWD复盘支持整场纵览、单轮聚焦和队伍下钻，当前视图可直接复用同一条导出链路。
            </p>
          </div>

          <div class="teacher-actions" role="group" aria-label="AWD 复盘操作">
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="router.push({ name: 'TeacherAWDReviewIndex' })"
            >
              <ArrowLeft class="h-4 w-4" />
              返回目录
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              data-testid="awd-review-export-archive"
              :disabled="loading || !review || exporting === 'archive'"
              @click="exportArchive"
            >
              <Download class="h-4 w-4" />
              导出复盘归档
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              data-testid="awd-review-export-report"
              :disabled="loading || !review || exporting === 'report' || !canExportReport"
              @click="exportReport"
            >
              <FileDown class="h-4 w-4" />
              导出教师报告
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <Shield class="h-4 w-4" />
            <span>Review Snapshot</span>
          </div>
          <div class="teacher-summary-grid progress-strip metric-panel-grid">
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">视图焦点</div>
              <div class="progress-card-value metric-panel-value">{{ activeSummaryTitle }}</div>
              <div class="progress-card-hint metric-panel-helper">
                当前查看整场总览或指定轮次摘要
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">服务 / 攻击 / 流量</div>
              <div class="progress-card-value metric-panel-value">
                {{ summaryStats.serviceCount }} / {{ summaryStats.attackCount }} /
                {{ summaryStats.trafficCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">当前视图下的关键证据量级</div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">轮次 / 队伍</div>
              <div class="progress-card-value metric-panel-value">
                {{ summaryStats.roundCount }} / {{ summaryStats.teamCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">轮次切换不会丢失目录上下文</div>
            </article>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section awd-review-round-section"
          aria-label="复盘轮次目录"
        >
          <header class="list-heading">
            <div>
              <div class="journal-note-label">Review Rounds</div>
              <h3 class="list-heading__title">轮次目录</h3>
            </div>
            <div class="awd-review-meta">
              <span class="teacher-directory-chip">
                {{ contestStatusLabel(review?.contest.status || '') }}
              </span>
              <span class="teacher-directory-chip teacher-directory-chip-muted">
                {{ review?.scope.snapshot_type === 'final' ? '赛后快照' : '实时快照' }}
              </span>
              <span class="teacher-directory-chip teacher-directory-chip-muted">
                {{ polling ? '导出轮询中' : '下载链路就绪' }}
              </span>
            </div>
          </header>

          <div class="awd-review-round-rail" role="tablist" aria-label="AWD 复盘视图">
            <button
              type="button"
              class="awd-review-round-pill"
              :class="{ 'awd-review-round-pill--active': !selectedRoundNumber }"
              @click="setRound(undefined)"
            >
              整场总览
            </button>
            <button
              v-for="round in review?.rounds || []"
              :key="round.id"
              type="button"
              class="awd-review-round-pill"
              :class="{
                'awd-review-round-pill--active': selectedRoundNumber === round.round_number,
              }"
              @click="setRound(round.round_number)"
            >
              第 {{ round.round_number }} 轮
            </button>
          </div>
        </section>

        <div v-if="loading" class="teacher-skeleton-list">
          <div
            v-for="index in 4"
            :key="index"
            class="h-28 animate-pulse rounded-[22px] bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]"
          />
        </div>

        <AppEmpty
          v-else-if="error"
          class="teacher-empty-state"
          icon="AlertTriangle"
          title="AWD复盘详情加载失败"
          :description="error"
        >
          <template #action>
            <button type="button" class="teacher-btn teacher-btn--primary" @click="loadReview">
              重新加载
            </button>
          </template>
        </AppEmpty>

        <AppEmpty
          v-else-if="!review"
          class="teacher-empty-state"
          icon="Waypoints"
          title="暂无复盘数据"
          description="当前赛事还没有可展示的 AWD 复盘内容。"
        />

        <div v-else class="awd-review-layout">
          <main class="awd-review-main">
            <section class="awd-review-panel">
              <div class="awd-review-panel__head">
                <div>
                  <div class="teacher-surface-eyebrow journal-eyebrow">Round Summary</div>
                  <h3>{{ activeSummaryTitle }}</h3>
                  <p>
                    {{
                      selectedRoundNumber
                        ? '聚焦当前轮次的队伍与事件，便于快速下钻。'
                        : '先看整场轮次变化，再决定进入哪一轮继续复盘。'
                    }}
                  </p>
                </div>
                <div class="awd-review-panel__meta">
                  <span>{{ review.contest.team_count }} 支队伍</span>
                  <span>{{ review.contest.round_count }} 轮</span>
                </div>
              </div>

              <div v-if="!selectedRound" class="awd-review-round-list">
                <article
                  v-for="round in review.rounds"
                  :key="round.id"
                  class="awd-review-round-card"
                >
                  <div class="awd-review-round-card__head">
                    <div>
                      <strong>第 {{ round.round_number }} 轮</strong>
                      <p>{{ roundStatusLabel(round.status) }}</p>
                    </div>
                    <button
                      type="button"
                      class="teacher-btn teacher-btn--ghost"
                      @click="setRound(round.round_number)"
                    >
                      进入单轮
                    </button>
                  </div>
                  <div class="awd-review-round-card__metrics">
                    <span>服务 {{ round.service_count }}</span>
                    <span>攻击 {{ round.attack_count }}</span>
                    <span>流量 {{ round.traffic_count }}</span>
                  </div>
                  <div class="awd-review-round-card__time">
                    <span>{{ round.started_at ? formatDate(round.started_at) : '待开始' }}</span>
                    <span>{{ round.ended_at ? formatDate(round.ended_at) : '进行中' }}</span>
                  </div>
                </article>
              </div>

              <div v-else class="teacher-directory" aria-label="轮次队伍目录">
                <div class="teacher-directory-top">
                  <h4 class="teacher-directory-title">
                    第 {{ selectedRound.round.round_number }} 轮队伍视图
                  </h4>
                  <div class="teacher-directory-meta">
                    共 {{ selectedRound.teams.length }} 支队伍
                  </div>
                </div>

                <div class="teacher-directory-head">
                  <span class="teacher-directory-head-cell teacher-directory-head-cell-code"
                    >队伍</span
                  >
                  <span class="teacher-directory-head-cell teacher-directory-head-cell-name"
                    >队伍概览</span
                  >
                  <span>成员</span>
                  <span>最近命中</span>
                  <span>操作</span>
                </div>

                <button
                  v-for="team in selectedRound.teams"
                  :key="team.team_id"
                  type="button"
                  class="teacher-directory-row"
                  @click="openTeam(team)"
                >
                  <div class="teacher-directory-cell teacher-directory-cell-code">
                    TEAM-{{ team.team_id }}
                  </div>
                  <div class="teacher-directory-cell teacher-directory-cell-name">
                    <h4 class="teacher-directory-row-title">{{ team.team_name }}</h4>
                    <p class="teacher-directory-row-copy">总分 {{ team.total_score }}</p>
                  </div>
                  <div class="teacher-directory-row-metrics">
                    <span>{{ team.member_count }} 人</span>
                    <span>队长 {{ team.captain_id }}</span>
                  </div>
                  <div class="teacher-directory-row-metrics">
                    <span>{{
                      team.last_solve_at ? formatDate(team.last_solve_at) : '暂无命中'
                    }}</span>
                  </div>
                  <div class="teacher-directory-row-cta">
                    <span>查看队伍</span>
                  </div>
                </button>
              </div>
            </section>

            <section v-if="selectedRound" class="awd-review-evidence-grid">
              <article class="awd-review-panel">
                <div class="awd-review-panel__head awd-review-panel__head--compact">
                  <div>
                    <div class="teacher-surface-eyebrow journal-eyebrow">Services</div>
                    <h3>服务状态</h3>
                  </div>
                  <span>{{ selectedRound.services.length }} 条</span>
                </div>

                <AppEmpty
                  v-if="selectedRound.services.length === 0"
                  icon="Server"
                  title="暂无服务数据"
                  description="当前轮次还没有可展示的服务状态。"
                />

                <div v-else class="awd-review-event-list">
                  <article
                    v-for="service in selectedRound.services"
                    :key="service.id"
                    class="awd-review-event-item"
                  >
                    <div class="awd-review-event-item__head">
                      <strong>{{ service.team_name }} · {{ service.challenge_title }}</strong>
                      <span
                        v-if="service.service_id"
                        class="awd-review-event-item__chip"
                        data-testid="awd-review-service-id"
                      >
                        {{ formatServiceRef(service.service_id) }}
                      </span>
                    </div>
                    <p>
                      {{ service.service_status }} · SLA {{ service.sla_score }} · Def
                      {{ service.defense_score }}
                    </p>
                  </article>
                </div>
              </article>

              <article class="awd-review-panel">
                <div class="awd-review-panel__head awd-review-panel__head--compact">
                  <div>
                    <div class="teacher-surface-eyebrow journal-eyebrow">Attacks</div>
                    <h3>攻击记录</h3>
                  </div>
                  <span>{{ selectedRound.attacks.length }} 条</span>
                </div>

                <AppEmpty
                  v-if="selectedRound.attacks.length === 0"
                  icon="Crosshair"
                  title="暂无攻击记录"
                  description="当前轮次还没有可展示的攻击事件。"
                />

                <div v-else class="awd-review-event-list">
                  <article
                    v-for="attack in selectedRound.attacks"
                    :key="attack.id"
                    class="awd-review-event-item"
                  >
                    <div class="awd-review-event-item__head">
                      <strong>{{ attack.attacker_team_name }} → {{ attack.victim_team_name }}</strong>
                      <span
                        v-if="attack.service_id"
                        class="awd-review-event-item__chip"
                        data-testid="awd-review-attack-service-id"
                      >
                        {{ formatServiceRef(attack.service_id) }}
                      </span>
                    </div>
                    <p>
                      {{ attack.challenge_title }} · {{ attack.attack_type }} · +{{
                        attack.score_gained
                      }}
                    </p>
                  </article>
                </div>
              </article>

              <article class="awd-review-panel">
                <div class="awd-review-panel__head awd-review-panel__head--compact">
                  <div>
                    <div class="teacher-surface-eyebrow journal-eyebrow">Traffic</div>
                    <h3>流量证据</h3>
                  </div>
                  <span>{{ selectedRound.traffic.length }} 条</span>
                </div>

                <AppEmpty
                  v-if="selectedRound.traffic.length === 0"
                  icon="Waypoints"
                  title="暂无流量证据"
                  description="当前轮次还没有可展示的流量事件。"
                />

                <div v-else class="awd-review-event-list">
                  <article
                    v-for="event in selectedRound.traffic"
                    :key="event.id"
                    class="awd-review-event-item"
                  >
                    <strong>{{ event.method }} {{ event.path }}</strong>
                    <p>
                      {{ event.attacker_team_name }} → {{ event.victim_team_name }} ·
                      {{ event.status_code }}
                    </p>
                  </article>
                </div>
              </article>
            </section>
          </main>

          <aside class="awd-review-side">
            <section class="awd-review-panel">
              <div class="awd-review-panel__head awd-review-panel__head--compact">
                <div>
                  <div class="teacher-surface-eyebrow journal-eyebrow">Contest Meta</div>
                  <h3>赛事态势</h3>
                </div>
              </div>
              <div class="awd-review-side__list">
                <div class="awd-review-side__row">
                  <span>赛事状态</span>
                  <strong>{{ contestStatusLabel(review.contest.status) }}</strong>
                </div>
                <div class="awd-review-side__row">
                  <span>快照类型</span>
                  <strong>{{
                    review.scope.snapshot_type === 'final' ? '赛后快照' : '实时快照'
                  }}</strong>
                </div>
                <div class="awd-review-side__row">
                  <span>最新证据</span>
                  <strong>{{
                    summaryStats.latestEvidenceAt
                      ? formatDate(summaryStats.latestEvidenceAt)
                      : '暂无'
                  }}</strong>
                </div>
                <div class="awd-review-side__row">
                  <span>教师报告</span>
                  <strong>{{ canExportReport ? '可导出' : '进行中禁用' }}</strong>
                </div>
              </div>
            </section>
          </aside>
        </div>
      </div>
    </section>

    <TeacherAWDReviewTeamDrawer
      :visible="Boolean(selectedTeam)"
      :team="selectedTeam"
      :services="selectedTeamServices"
      :attacks="selectedTeamAttacks"
      :traffic="selectedTeamTraffic"
      @close="closeTeam"
    />
  </div>
</template>

<style scoped>
.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
  --awd-review-team-columns: minmax(0, 7rem) minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1.1fr) auto;
}

.teacher-directory-section {
  margin-top: var(--space-6);
}

.awd-review-round-section {
  padding: var(--space-5);
  border: 1px solid var(--teacher-card-border);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 84%, transparent);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.awd-review-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.awd-review-round-rail {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-5);
}

.awd-review-round-pill {
  min-height: 2.3rem;
  padding: 0.45rem 0.9rem;
  border: 1px solid var(--teacher-control-border);
  border-radius: 999px;
  background: var(--journal-surface);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  font-weight: 600;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.awd-review-round-pill:hover,
.awd-review-round-pill:focus-visible,
.awd-review-round-pill--active {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
  outline: none;
}

.teacher-skeleton-list {
  margin-top: var(--space-6);
  display: grid;
  gap: var(--space-3);
}

.teacher-empty-state {
  margin-top: var(--space-6);
}

.awd-review-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.9fr) minmax(18rem, 0.9fr);
  gap: var(--space-5);
  margin-top: var(--space-6);
}

.awd-review-main {
  display: grid;
  gap: var(--space-5);
}

.awd-review-side {
  display: grid;
  gap: var(--space-5);
  align-content: start;
}

.awd-review-panel {
  padding: var(--space-5);
  border: 1px solid var(--teacher-card-border);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 84%, transparent);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.awd-review-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-5);
}

.awd-review-panel__head--compact {
  margin-bottom: var(--space-4);
}

.awd-review-panel__head h3 {
  margin: var(--space-3) 0 var(--space-2);
  font-size: var(--font-size-1-08);
  font-weight: 700;
  color: var(--journal-ink);
}

.awd-review-panel__head p {
  margin: 0;
  color: var(--journal-muted);
  line-height: 1.7;
}

.awd-review-panel__meta {
  display: grid;
  gap: var(--space-1);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  text-align: right;
}

.awd-review-round-list {
  display: grid;
  gap: var(--space-3);
}

.awd-review-round-card {
  padding: var(--space-4);
  border: 1px solid var(--teacher-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.awd-review-round-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.awd-review-round-card__head strong {
  display: block;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
}

.awd-review-round-card__head p {
  margin: var(--space-1-5) 0 0;
  color: var(--journal-muted);
}

.awd-review-round-card__metrics,
.awd-review-round-card__time {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-3);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.teacher-directory-title {
  margin: 0;
  font-size: var(--font-size-1-02);
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-directory-meta {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--awd-review-team-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px dashed var(--teacher-divider);
  color: var(--journal-muted);
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--awd-review-team-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4-5) 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 62%, transparent);
  outline: none;
}

.teacher-directory-cell {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
  align-content: center;
}

.teacher-directory-cell-code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-78);
  font-weight: 700;
  color: var(--journal-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-96);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.teacher-directory-row-copy {
  margin: 0;
  color: var(--journal-muted);
  font-size: var(--font-size-0-84);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: var(--space-1);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.7rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  color: var(--journal-accent-strong);
  font-size: var(--font-size-0-82);
  font-weight: 700;
}

.awd-review-evidence-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.awd-review-event-list {
  display: grid;
  gap: var(--space-3);
}

.awd-review-event-item {
  padding: var(--space-4);
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.awd-review-event-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.awd-review-event-item strong {
  display: block;
  color: var(--journal-ink);
}

.awd-review-event-item__chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.6rem;
  padding: 0 var(--space-2-5);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-74);
  font-weight: 700;
  white-space: nowrap;
}

.awd-review-event-item p {
  margin: var(--space-2) 0 0;
  color: var(--journal-muted);
  line-height: 1.6;
}

.awd-review-side__list {
  display: grid;
  gap: var(--space-3);
}

.awd-review-side__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px dashed var(--teacher-divider);
}

.awd-review-side__row:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.awd-review-side__row span {
  color: var(--journal-muted);
}

.awd-review-side__row strong {
  color: var(--journal-ink);
  text-align: right;
}

@media (max-width: 1200px) {
  .awd-review-layout,
  .awd-review-evidence-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid {
    grid-template-columns: 1fr;
  }

  .teacher-directory-head {
    display: none;
  }

  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
}
</style>
