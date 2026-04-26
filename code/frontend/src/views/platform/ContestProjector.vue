<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Activity, MonitorUp, RefreshCw, ShieldAlert, Sword, Trophy, Users, Zap } from 'lucide-vue-next'

import {
  getAdminContestLiveScoreboard,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  getContests,
  listContestAWDRoundAttacks,
  listContestAWDRoundServices,
  listContestAWDRounds,
} from '@/api/admin'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
  ContestScoreboardData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
let scoreboardRequestToken = 0

const contests = ref<ContestDetailData[]>([])
const selectedContestId = ref('')
const scoreboard = ref<ContestScoreboardData | null>(null)
const rounds = ref<AWDRoundData[]>([])
const selectedRoundId = ref('')
const services = ref<AWDTeamServiceData[]>([])
const attacks = ref<AWDAttackLogData[]>([])
const roundSummary = ref<AWDRoundSummaryData | null>(null)
const trafficSummary = ref<AWDTrafficSummaryData | null>(null)
const loadingContests = ref(true)
const loadingScoreboard = ref(false)
const loadError = ref('')
const refreshTimer = ref<number | null>(null)

const awdContests = computed(() => contests.value.filter((item) => item.mode === 'awd'))
const projectorContests = computed(() =>
  awdContests.value.filter((item) => ['running', 'frozen', 'ended'].includes(item.status))
)
const selectedContest = computed(
  () => projectorContests.value.find((item) => item.id === selectedContestId.value) ?? null
)
const scoreboardRows = computed(() => scoreboard.value?.scoreboard.list ?? [])
const topThreeRows = computed(() => scoreboardRows.value.slice(0, 3))
const otherRows = computed(() => scoreboardRows.value.slice(3, 12))
const lastUpdatedLabel = ref('未同步')
const selectedRound = computed(() => rounds.value.find((item) => item.id === selectedRoundId.value) ?? null)
const firstBlood = computed(() =>
  attacks.value
    .filter((item) => item.is_success)
    .slice()
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())[0] ?? null
)
const serviceStatusCounts = computed(() => ({
  up: services.value.filter((item) => item.service_status === 'up').length,
  down: services.value.filter((item) => item.service_status === 'down').length,
  compromised: services.value.filter((item) => item.service_status === 'compromised').length,
}))
const serviceMatrixRows = computed(() => {
  const teamMap = new Map<string, { team_id: string; team_name: string; services: AWDTeamServiceData[] }>()
  for (const service of services.value) {
    const row = teamMap.get(service.team_id) ?? {
      team_id: service.team_id,
      team_name: service.team_name,
      services: [],
    }
    row.services.push(service)
    teamMap.set(service.team_id, row)
  }
  return Array.from(teamMap.values()).slice(0, 8)
})
const attackLeaders = computed(() => {
  const teamMap = new Map<string, { team_id: string; team_name: string; success: number; score: number }>()
  for (const attack of attacks.value) {
    const row = teamMap.get(attack.attacker_team_id) ?? {
      team_id: attack.attacker_team_id,
      team_name: attack.attacker_team,
      success: 0,
      score: 0,
    }
    if (attack.is_success) {
      row.success += 1
      row.score += attack.score_gained
    }
    teamMap.set(attack.attacker_team_id, row)
  }
  return Array.from(teamMap.values())
    .sort((a, b) => b.success - a.success || b.score - a.score)
    .slice(0, 5)
})

function getContestStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'frozen':
      return '已冻结'
    case 'ended':
      return '已结束'
    default:
      return '待同步'
  }
}

function formatScore(value: number): string {
  return new Intl.NumberFormat('zh-CN', {
    maximumFractionDigits: 2,
  }).format(value)
}

function formatTime(value?: string): string {
  if (!value) return '--'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getRoundStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'finished':
      return '已结束'
    case 'pending':
      return '待开始'
    default:
      return '未同步'
  }
}

function getServiceStatusLabel(status: AWDTeamServiceData['service_status']): string {
  switch (status) {
    case 'up':
      return 'UP'
    case 'down':
      return 'DOWN'
    case 'compromised':
      return 'PWN'
    default:
      return status
  }
}

function chooseInitialContest(): void {
  const preferred =
    projectorContests.value.find((item) => item.status === 'running') ??
    projectorContests.value.find((item) => item.status === 'frozen') ??
    projectorContests.value[0] ??
    null
  selectedContestId.value = preferred?.id ?? ''
}

async function loadScoreboard(contestId = selectedContestId.value): Promise<void> {
  if (!contestId) {
    return
  }

  const requestToken = ++scoreboardRequestToken
  loadingScoreboard.value = true
  try {
    const [nextScoreboard, nextRounds] = await Promise.all([
      getAdminContestLiveScoreboard(contestId, {
        page: 1,
        page_size: 20,
      }),
      listContestAWDRounds(contestId),
    ])
    if (requestToken !== scoreboardRequestToken) {
      return
    }
    scoreboard.value = nextScoreboard
    rounds.value = nextRounds
    const preferredRound =
      nextRounds.find((item) => item.status === 'running') ??
      nextRounds[nextRounds.length - 1] ??
      null
    selectedRoundId.value = preferredRound?.id ?? ''

    if (preferredRound) {
      const [nextServices, nextAttacks, nextRoundSummary, nextTrafficSummary] = await Promise.all([
        listContestAWDRoundServices(contestId, preferredRound.id),
        listContestAWDRoundAttacks(contestId, preferredRound.id),
        getContestAWDRoundSummary(contestId, preferredRound.id),
        getContestAWDRoundTrafficSummary(contestId, preferredRound.id),
      ])
      if (requestToken !== scoreboardRequestToken) {
        return
      }
      services.value = nextServices
      attacks.value = nextAttacks
      roundSummary.value = nextRoundSummary
      trafficSummary.value = nextTrafficSummary
    } else {
      services.value = []
      attacks.value = []
      roundSummary.value = null
      trafficSummary.value = null
    }
    lastUpdatedLabel.value = new Date().toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch (error) {
    if (requestToken !== scoreboardRequestToken) {
      return
    }
    scoreboard.value = null
    rounds.value = []
    selectedRoundId.value = ''
    services.value = []
    attacks.value = []
    roundSummary.value = null
    trafficSummary.value = null
    toast.error('同步大屏排行榜失败')
  } finally {
    if (requestToken === scoreboardRequestToken) {
      loadingScoreboard.value = false
    }
  }
}

async function loadContests(): Promise<void> {
  loadingContests.value = true
  loadError.value = ''
  try {
    const response = await getContests({
      page: 1,
      page_size: 100,
    })
    contests.value = response.list
    if (!selectedContestId.value || !projectorContests.value.some((item) => item.id === selectedContestId.value)) {
      chooseInitialContest()
    }
    await loadScoreboard()
  } catch (error) {
    contests.value = []
    scoreboard.value = null
    loadError.value = error instanceof Error ? error.message : '大屏赛事加载失败'
  } finally {
    loadingContests.value = false
  }
}

async function selectContest(contestId: string): Promise<void> {
  if (contestId === selectedContestId.value) {
    return
  }
  selectedContestId.value = contestId
  await loadScoreboard(contestId)
}

function startAutoRefresh(): void {
  if (refreshTimer.value !== null) {
    window.clearInterval(refreshTimer.value)
  }
  refreshTimer.value = window.setInterval(() => {
    void loadScoreboard()
  }, 15000)
}

onMounted(() => {
  void loadContests()
  startAutoRefresh()
})

onUnmounted(() => {
  if (refreshTimer.value !== null) {
    window.clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
})
</script>

<template>
  <section class="contest-projector-shell workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col">
    <div class="workspace-grid">
      <main class="content-pane contest-projector-content">
        <section class="projector-stage workspace-directory-section">
          <header class="projector-header">
            <div>
              <div class="projector-overline">Contest Projector</div>
              <h1 class="projector-title">大屏展示</h1>
            </div>
            <div class="projector-actions">
              <span class="projector-sync">同步于 {{ lastUpdatedLabel }}</span>
              <button
                type="button"
                class="ops-btn ops-btn--neutral"
                :disabled="loadingContests || loadingScoreboard"
                @click="void loadContests()"
              >
                <RefreshCw
                  class="btn-icon"
                  :class="{ 'animate-spin': loadingContests || loadingScoreboard }"
                />
                <span>刷新</span>
              </button>
            </div>
          </header>

          <AppLoading v-if="loadingContests">正在同步大屏赛事...</AppLoading>

          <AppEmpty
            v-else-if="loadError"
            title="大屏展示暂时不可用"
            :description="loadError"
            icon="AlertTriangle"
            class="py-20"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="void loadContests()"
              >
                重试加载
              </button>
            </template>
          </AppEmpty>

          <AppEmpty
            v-else-if="projectorContests.length === 0"
            title="暂无可展示的 AWD 赛事"
            description="大屏展示只面向进行中、冻结或已结束的 AWD 赛事。"
            icon="FileChartColumnIncreasing"
            class="py-20"
          />

          <div
            v-else
            class="projector-layout"
          >
            <aside class="contest-rail">
              <button
                v-for="contest in projectorContests"
                :key="contest.id"
                type="button"
                class="contest-switch"
                :class="{ active: contest.id === selectedContestId }"
                @click="void selectContest(contest.id)"
              >
                <span class="contest-switch__title">{{ contest.title }}</span>
                <span class="contest-switch__meta">{{ getContestStatusLabel(contest.status) }}</span>
              </button>
            </aside>

            <section class="scoreboard-projector">
              <header class="scoreboard-projector__head">
                <div>
                  <div class="projector-overline">Live Scoreboard</div>
                  <h2 class="scoreboard-projector__title">
                    {{ selectedContest?.title ?? scoreboard?.contest.title ?? '未选择赛事' }}
                  </h2>
                </div>
                <div class="projector-status">
                  <MonitorUp class="status-icon" />
                  <span>{{ getContestStatusLabel(selectedContest?.status ?? scoreboard?.contest.status) }}</span>
                </div>
              </header>

              <div class="projector-metrics">
                <article class="projector-metric">
                  <Users class="metric-icon" />
                  <span>队伍</span>
                  <strong>{{ scoreboardRows.length }}</strong>
                </article>
                <article class="projector-metric">
                  <Trophy class="metric-icon metric-icon--rank" />
                  <span>榜首</span>
                  <strong>{{ topThreeRows[0]?.team_name ?? '--' }}</strong>
                </article>
                <article class="projector-metric">
                  <span>结束</span>
                  <strong>{{ formatTime(selectedContest?.ends_at ?? scoreboard?.contest.ends_at) }}</strong>
                </article>
                <article class="projector-metric">
                  <Activity class="metric-icon metric-icon--live" />
                  <span>轮次</span>
                  <strong>R{{ selectedRound?.round_number ?? '--' }} · {{ getRoundStatusLabel(selectedRound?.status) }}</strong>
                </article>
                <article class="projector-metric">
                  <Sword class="metric-icon metric-icon--attack" />
                  <span>成功攻击</span>
                  <strong>{{ roundSummary?.metrics?.successful_attack_count ?? 0 }}</strong>
                </article>
                <article class="projector-metric">
                  <ShieldAlert class="metric-icon metric-icon--danger" />
                  <span>异常服务</span>
                  <strong>{{ serviceStatusCounts.down + serviceStatusCounts.compromised }}</strong>
                </article>
              </div>

              <div
                v-if="loadingScoreboard"
                class="scoreboard-loading"
              >
                <AppLoading>正在刷新榜单...</AppLoading>
              </div>

              <div
                v-else
                class="projector-board"
              >
                <div class="podium-grid">
                  <article
                    v-for="row in topThreeRows"
                    :key="row.team_id"
                    class="podium-card"
                    :class="`rank-${row.rank}`"
                  >
                    <span class="podium-rank">#{{ row.rank }}</span>
                    <strong>{{ row.team_name }}</strong>
                    <span>{{ formatScore(row.score) }} pts</span>
                  </article>
                </div>

                <div class="battle-grid">
                  <article class="battle-panel first-blood-panel">
                    <header class="panel-head">
                      <div>
                        <div class="projector-overline">First Blood</div>
                        <h3>首血</h3>
                      </div>
                      <Zap class="panel-icon panel-icon--accent" />
                    </header>
                    <div
                      v-if="firstBlood"
                      class="first-blood-body"
                    >
                      <strong>{{ firstBlood.attacker_team }}</strong>
                      <span>攻破 {{ firstBlood.victim_team }}</span>
                      <small>{{ formatScore(firstBlood.score_gained) }} pts · {{ formatTime(firstBlood.created_at) }}</small>
                    </div>
                    <div
                      v-else
                      class="panel-empty"
                    >
                      暂无首血记录
                    </div>
                  </article>

                  <article class="battle-panel attack-panel">
                    <header class="panel-head">
                      <div>
                        <div class="projector-overline">Attack Board</div>
                        <h3>攻击展示</h3>
                      </div>
                      <Sword class="panel-icon panel-icon--attack" />
                    </header>
                    <div class="attack-list">
                      <div
                        v-for="leader in attackLeaders"
                        :key="leader.team_id"
                        class="attack-row"
                      >
                        <span>{{ leader.team_name }}</span>
                        <strong>{{ leader.success }} HIT</strong>
                        <small>{{ formatScore(leader.score) }} pts</small>
                      </div>
                      <div
                        v-if="attackLeaders.length === 0"
                        class="panel-empty"
                      >
                        暂无成功攻击
                      </div>
                    </div>
                  </article>

                  <article class="battle-panel traffic-panel">
                    <header class="panel-head">
                      <div>
                        <div class="projector-overline">Traffic Pulse</div>
                        <h3>流量热点</h3>
                      </div>
                      <Activity class="panel-icon panel-icon--live" />
                    </header>
                    <div class="traffic-strip">
                      <span>请求 {{ trafficSummary?.total_request_count ?? 0 }}</span>
                      <span>攻击方 {{ trafficSummary?.active_attacker_team_count ?? 0 }}</span>
                      <span>错误 {{ trafficSummary?.error_request_count ?? 0 }}</span>
                    </div>
                    <div class="traffic-list">
                      <div
                        v-for="item in (trafficSummary?.top_attackers ?? []).slice(0, 3)"
                        :key="item.team_id"
                        class="attack-row"
                      >
                        <span>{{ item.team_name }}</span>
                        <strong>{{ item.request_count }} REQ</strong>
                        <small>{{ item.error_count }} ERR</small>
                      </div>
                      <div
                        v-if="(trafficSummary?.top_attackers ?? []).length === 0"
                        class="panel-empty"
                      >
                        暂无代理流量
                      </div>
                    </div>
                  </article>
                </div>

                <section class="service-matrix-panel">
                  <header class="panel-head">
                    <div>
                      <div class="projector-overline">Service Status</div>
                      <h3>队伍服务状态</h3>
                    </div>
                    <div class="service-counts">
                      <span class="service-chip service-chip--up">UP {{ serviceStatusCounts.up }}</span>
                      <span class="service-chip service-chip--down">DOWN {{ serviceStatusCounts.down }}</span>
                      <span class="service-chip service-chip--compromised">PWN {{ serviceStatusCounts.compromised }}</span>
                    </div>
                  </header>
                  <div class="service-matrix">
                    <div
                      v-for="row in serviceMatrixRows"
                      :key="row.team_id"
                      class="service-team-row"
                    >
                      <strong>{{ row.team_name }}</strong>
                      <div class="service-cell-list">
                        <span
                          v-for="service in row.services.slice(0, 8)"
                          :key="service.id"
                          class="service-dot"
                          :class="`service-dot--${service.service_status}`"
                          :title="`${service.challenge_id}: ${getServiceStatusLabel(service.service_status)}`"
                        >
                          {{ getServiceStatusLabel(service.service_status) }}
                        </span>
                      </div>
                    </div>
                    <div
                      v-if="serviceMatrixRows.length === 0"
                      class="panel-empty"
                    >
                      暂无服务状态
                    </div>
                  </div>
                </section>

                <div class="projector-table-wrap">
                  <table class="projector-table">
                    <thead>
                      <tr>
                        <th>排名</th>
                        <th>队伍</th>
                        <th class="text-right">总分</th>
                        <th class="text-right">解题</th>
                        <th class="text-right">最后提交</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="row in otherRows"
                        :key="row.team_id"
                      >
                        <td>#{{ row.rank }}</td>
                        <td>{{ row.team_name }}</td>
                        <td class="text-right font-mono">{{ formatScore(row.score) }}</td>
                        <td class="text-right font-mono">{{ row.solved_count }}</td>
                        <td class="text-right">{{ formatTime(row.last_submission_at) }}</td>
                      </tr>
                      <tr v-if="scoreboardRows.length === 0">
                        <td
                          colspan="5"
                          class="empty-row"
                        >
                          暂无得分记录
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </section>
          </div>
        </section>
      </main>
    </div>
  </section>
</template>

<style scoped>
.contest-projector-shell {
  --workspace-shell-bg: var(--journal-surface);
  --workspace-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.contest-projector-content,
.projector-stage,
.projector-layout,
.scoreboard-projector,
.projector-board {
  display: flex;
  flex-direction: column;
}

.contest-projector-content {
  padding: 0;
}

.projector-stage {
  --workspace-directory-section-padding: var(--space-5) var(--space-5-5);
  gap: var(--space-5);
  background: transparent;
}

.projector-header,
.scoreboard-projector__head,
.projector-actions,
.projector-status,
.projector-metric,
.podium-card {
  display: flex;
  align-items: center;
}

.projector-header,
.scoreboard-projector__head {
  justify-content: space-between;
  gap: var(--space-4);
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.projector-title,
.scoreboard-projector__title {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-45);
  font-weight: 900;
}

.scoreboard-projector__title {
  font-size: var(--font-size-1-25);
}

.projector-actions {
  gap: var(--space-3);
}

.projector-sync {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border-radius: 0.5rem;
  font-size: var(--font-size-13);
  font-weight: 800;
  transition: all 0.2s ease;
}

.ops-btn--neutral {
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
}

.ops-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.btn-icon,
.status-icon,
.metric-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.projector-layout {
  display: grid;
  grid-template-columns: minmax(14rem, 18rem) minmax(0, 1fr);
  gap: var(--space-5);
}

.contest-rail,
.scoreboard-projector {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background: var(--color-bg-surface);
}

.contest-rail {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-3);
}

.contest-switch {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-1);
  width: 100%;
  border: 1px solid transparent;
  border-radius: 0.5rem;
  padding: var(--space-3);
  color: var(--color-text-secondary);
  text-align: left;
  transition: all 0.2s ease;
}

.contest-switch:hover,
.contest-switch.active {
  border-color: color-mix(in srgb, var(--journal-accent) 40%, var(--color-border-default));
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-ink);
}

.contest-switch__title {
  font-size: var(--font-size-13);
  font-weight: 800;
}

.contest-switch__meta {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.scoreboard-projector {
  position: relative;
  gap: var(--space-5);
  min-height: 34rem;
  padding: var(--space-5);
}

.projector-status {
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  font-weight: 800;
}

.projector-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.projector-metric {
  justify-content: space-between;
  gap: var(--space-3);
  min-width: 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  padding: var(--space-3) var(--space-4);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.projector-metric strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-icon--rank {
  color: var(--color-warning);
}

.metric-icon--live {
  color: var(--color-primary);
}

.metric-icon--attack,
.panel-icon--attack {
  color: var(--color-danger);
}

.metric-icon--danger {
  color: var(--color-warning);
}

.scoreboard-loading {
  display: flex;
  min-height: 20rem;
  align-items: center;
  justify-content: center;
}

.projector-board {
  gap: var(--space-5);
}

.podium-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-4);
}

.podium-card {
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
  min-height: 11rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--color-text-secondary);
  text-align: center;
}

.podium-card.rank-1 {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
}

.podium-card strong {
  max-width: 90%;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-1-125);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.podium-rank {
  color: var(--color-warning);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-45);
  font-weight: 900;
}

.battle-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-4);
}

.battle-panel,
.service-matrix-panel {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 58%, transparent);
  padding: var(--space-4);
}

.battle-panel {
  display: flex;
  min-height: 13rem;
  flex-direction: column;
  gap: var(--space-4);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.panel-head h3 {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
  font-weight: 900;
}

.panel-icon {
  width: var(--space-5);
  height: var(--space-5);
}

.panel-icon--accent {
  color: var(--color-warning);
}

.first-blood-body,
.attack-list,
.traffic-list,
.service-matrix {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.first-blood-body {
  justify-content: center;
  flex: 1;
}

.first-blood-body strong {
  color: var(--journal-ink);
  font-size: var(--font-size-1-25);
  font-weight: 900;
}

.first-blood-body span,
.first-blood-body small,
.attack-row small,
.panel-empty {
  color: var(--color-text-muted);
}

.attack-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: center;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.attack-row span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-row strong {
  color: var(--journal-ink);
  font-family: var(--font-family-mono);
}

.traffic-strip,
.service-counts,
.service-cell-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.traffic-strip span,
.service-chip,
.service-dot {
  display: inline-flex;
  align-items: center;
  min-height: var(--ui-control-height-sm);
  border-radius: 0.375rem;
  padding: 0 var(--space-2-5);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.traffic-strip span {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-text-secondary);
}

.service-matrix-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.service-team-row {
  display: grid;
  grid-template-columns: minmax(10rem, 14rem) minmax(0, 1fr);
  align-items: center;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2-5);
}

.service-team-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-chip--up,
.service-dot--up {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-success);
}

.service-chip--down,
.service-dot--down {
  background: color-mix(in srgb, var(--color-danger) 14%, transparent);
  color: var(--color-danger);
}

.service-chip--compromised,
.service-dot--compromised {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: var(--color-warning);
}

.projector-table-wrap {
  overflow-x: auto;
}

.projector-table {
  width: 100%;
  min-width: 42rem;
  border-collapse: collapse;
}

.projector-table th,
.projector-table td {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--space-3);
}

.projector-table th {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
  text-align: left;
}

.projector-table td {
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  font-weight: 700;
}

.text-right {
  text-align: right;
}

.font-mono {
  font-family: var(--font-family-mono);
}

.empty-row {
  padding: var(--space-8) var(--space-3);
  text-align: center;
}

@media (max-width: 900px) {
  .projector-layout,
  .projector-metrics,
  .podium-grid,
  .battle-grid {
    grid-template-columns: 1fr;
  }

  .service-team-row {
    grid-template-columns: 1fr;
  }

  .projector-header,
  .scoreboard-projector__head {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
