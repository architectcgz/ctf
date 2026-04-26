<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Activity, Maximize2, Minimize2, MonitorUp, RefreshCw, ShieldAlert, Sword, Trophy, Users, Zap } from 'lucide-vue-next'

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
const projectorStageRef = ref<HTMLElement | null>(null)
const fullscreenActive = ref(false)

const awdContests = computed(() => contests.value.filter((item) => item.mode === 'awd'))
const projectorContests = computed(() =>
  awdContests.value
    .filter((item) => ['running', 'frozen', 'ended'].includes(item.status))
    .slice()
    .sort((left, right) => {
      const rightTime = new Date(right.starts_at ?? right.ends_at).getTime()
      const leftTime = new Date(left.starts_at ?? left.ends_at).getTime()
      return rightTime - leftTime
    })
)
const selectedContest = computed(
  () => projectorContests.value.find((item) => item.id === selectedContestId.value) ?? null
)
const scoreboardRows = computed(() => scoreboard.value?.scoreboard.list ?? [])
const topThreeRows = computed(() => scoreboardRows.value.slice(0, 3))
const leaderboardRows = computed(() => scoreboardRows.value.slice(0, 10))
const lastUpdatedLabel = ref('未同步')
const selectedRound = computed(() => rounds.value.find((item) => item.id === selectedRoundId.value) ?? null)
const firstBlood = computed(() =>
  attacks.value
    .filter((item) => item.is_success)
    .slice()
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())[0] ?? null
)
const latestAttackEvents = computed(() =>
  attacks.value
    .slice()
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 8)
)
const serviceStatusCounts = computed(() => ({
  up: services.value.filter((item) => item.service_status === 'up').length,
  down: services.value.filter((item) => item.service_status === 'down').length,
  compromised: services.value.filter((item) => item.service_status === 'compromised').length,
}))
const serviceHealthRate = computed(() => {
  const total = services.value.length
  if (total === 0) return 0
  return Math.round((serviceStatusCounts.value.up / total) * 100)
})
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
  return Array.from(teamMap.values()).slice(0, 10)
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
const trafficTrendBars = computed(() => {
  const buckets = (trafficSummary.value?.trend_buckets ?? []).slice(-12)
  const maxRequests = Math.max(...buckets.map((item) => item.request_count), 1)
  return buckets.map((item) => ({
    ...item,
    height: `${Math.max(10, Math.round((item.request_count / maxRequests) * 100))}%`,
    errorWidth: `${Math.min(100, Math.round((item.error_count / Math.max(item.request_count, 1)) * 100))}%`,
  }))
})
const hotVictims = computed(() => (trafficSummary.value?.top_victims ?? []).slice(0, 4))

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

function getAttackTypeLabel(type: AWDAttackLogData['attack_type']): string {
  switch (type) {
    case 'flag_capture':
      return 'Flag'
    case 'service_exploit':
      return 'Exploit'
    default:
      return type
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

function handleContestSelect(event: Event): void {
  const target = event.target as HTMLSelectElement
  void selectContest(target.value)
}

function syncFullscreenState(): void {
  fullscreenActive.value = document.fullscreenElement === projectorStageRef.value
}

async function toggleFullscreen(): Promise<void> {
  try {
    if (fullscreenActive.value) {
      await document.exitFullscreen()
      return
    }

    const target = projectorStageRef.value
    if (!target?.requestFullscreen) {
      toast.error('当前浏览器不支持全屏展示')
      return
    }
    await target.requestFullscreen()
  } catch (error) {
    toast.error('切换全屏失败')
  }
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
  document.addEventListener('fullscreenchange', syncFullscreenState)
  void loadContests()
  startAutoRefresh()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', syncFullscreenState)
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
        <section
          ref="projectorStageRef"
          class="projector-stage workspace-directory-section"
        >
          <header class="projector-header">
            <div>
              <div class="projector-overline">
                Contest Projector
              </div>
              <h1 class="projector-title">
                大屏展示
              </h1>
            </div>
            <div class="projector-actions">
              <span class="projector-sync">同步于 {{ lastUpdatedLabel }}</span>
              <button
                type="button"
                class="ops-btn ops-btn--neutral"
                @click="void toggleFullscreen()"
              >
                <Minimize2
                  v-if="fullscreenActive"
                  class="btn-icon"
                />
                <Maximize2
                  v-else
                  class="btn-icon"
                />
                <span>{{ fullscreenActive ? '退出全屏' : '全屏' }}</span>
              </button>
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

          <AppLoading v-if="loadingContests">
            正在同步大屏赛事...
          </AppLoading>

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
            <div class="contest-selector">
              <label
                class="contest-selector__label"
                for="projector-contest-select"
              >
                竞赛
              </label>
              <select
                id="projector-contest-select"
                class="contest-selector__control"
                :value="selectedContestId"
                :disabled="loadingScoreboard"
                @change="handleContestSelect"
              >
                <option
                  v-for="contest in projectorContests"
                  :key="contest.id"
                  :value="contest.id"
                >
                  {{ contest.title }} · {{ getContestStatusLabel(contest.status) }} · {{ formatTime(contest.starts_at) }}
                </option>
              </select>
            </div>

            <section class="scoreboard-projector">
              <header class="projector-hero">
                <div class="projector-hero__title">
                  <div class="projector-overline">
                    实时态势
                  </div>
                  <h2 class="scoreboard-projector__title">
                    {{ selectedContest?.title ?? scoreboard?.contest.title ?? '未选择赛事' }}
                  </h2>
                </div>
                <div class="projector-hero__status">
                  <div class="projector-status">
                    <MonitorUp class="status-icon" />
                    <span>{{ getContestStatusLabel(selectedContest?.status ?? scoreboard?.contest.status) }}</span>
                  </div>
                  <div class="projector-round-pill">
                    R{{ selectedRound?.round_number ?? '--' }} · {{ getRoundStatusLabel(selectedRound?.status) }}
                  </div>
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
                  <Activity class="metric-icon metric-icon--live" />
                  <span>服务健康</span>
                  <strong>{{ serviceHealthRate }}%</strong>
                </article>
                <article class="projector-metric">
                  <Sword class="metric-icon metric-icon--attack" />
                  <span>成功攻击</span>
                  <strong>{{ roundSummary?.metrics?.successful_attack_count ?? 0 }}</strong>
                </article>
                <article class="projector-metric">
                  <Activity class="metric-icon metric-icon--live" />
                  <span>代理流量</span>
                  <strong>{{ trafficSummary?.total_request_count ?? 0 }}</strong>
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
                <div class="arena-grid">
                  <section class="leaderboard-panel">
                    <header class="panel-head">
                      <div>
                        <div class="projector-overline">
                          排行榜
                        </div>
                        <h3>实时排名</h3>
                      </div>
                      <Trophy class="panel-icon panel-icon--accent" />
                    </header>

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

                    <div class="leaderboard-list">
                      <div
                        v-for="row in leaderboardRows"
                        :key="row.team_id"
                        class="leaderboard-row"
                        :class="{ 'leaderboard-row--top': row.rank <= 3 }"
                      >
                        <span class="leaderboard-rank">#{{ row.rank }}</span>
                        <strong>{{ row.team_name }}</strong>
                        <span class="leaderboard-score">{{ formatScore(row.score) }}</span>
                      </div>
                      <div
                        v-if="scoreboardRows.length === 0"
                        class="panel-empty"
                      >
                        暂无得分记录
                      </div>
                    </div>
                  </section>

                  <section class="battlefield-panel">
                    <div class="service-matrix-panel">
                      <header class="panel-head">
                        <div>
                          <div class="projector-overline">
                            服务墙
                          </div>
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
                              v-for="service in row.services.slice(0, 10)"
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
                    </div>

                    <div class="traffic-panel">
                      <header class="panel-head">
                        <div>
                          <div class="projector-overline">
                            流量态势
                          </div>
                          <h3>代理流量</h3>
                        </div>
                        <Activity class="panel-icon panel-icon--live" />
                      </header>
                      <div class="traffic-strip">
                        <span>请求 {{ trafficSummary?.total_request_count ?? 0 }}</span>
                        <span>攻击方 {{ trafficSummary?.active_attacker_team_count ?? 0 }}</span>
                        <span>目标 {{ trafficSummary?.victim_team_count ?? 0 }}</span>
                        <span>错误 {{ trafficSummary?.error_request_count ?? 0 }}</span>
                      </div>
                      <div class="traffic-trend">
                        <span
                          v-for="bucket in trafficTrendBars"
                          :key="bucket.bucket_start_at"
                          class="traffic-bar"
                          :style="{ height: bucket.height }"
                          :title="`${formatTime(bucket.bucket_start_at)} · ${bucket.request_count} req`"
                        >
                          <i :style="{ height: bucket.errorWidth }" />
                        </span>
                      </div>
                      <div class="traffic-columns">
                        <div class="traffic-list">
                          <div class="traffic-list__title">
                            活跃攻击方
                          </div>
                          <div
                            v-for="item in (trafficSummary?.top_attackers ?? []).slice(0, 4)"
                            :key="item.team_id"
                            class="attack-row"
                          >
                            <span>{{ item.team_name }}</span>
                            <strong>{{ item.request_count }} REQ</strong>
                            <small>{{ item.error_count }} ERR</small>
                          </div>
                        </div>
                        <div class="traffic-list">
                          <div class="traffic-list__title">
                            高压目标
                          </div>
                          <div
                            v-for="item in hotVictims"
                            :key="item.team_id"
                            class="attack-row"
                          >
                            <span>{{ item.team_name }}</span>
                            <strong>{{ item.request_count }} REQ</strong>
                            <small>{{ item.error_count }} ERR</small>
                          </div>
                        </div>
                      </div>
                      <div
                        v-if="(trafficSummary?.top_attackers ?? []).length === 0 && hotVictims.length === 0"
                        class="panel-empty"
                      >
                        暂无代理流量
                      </div>
                    </div>
                  </section>

                  <section class="event-panel">
                    <article class="first-blood-panel">
                      <header class="panel-head">
                        <div>
                          <div class="projector-overline">
                            首血
                          </div>
                          <h3>First Blood</h3>
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

                    <article class="attack-panel">
                      <header class="panel-head">
                        <div>
                          <div class="projector-overline">
                            攻击榜
                          </div>
                          <h3>命中队伍</h3>
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

                    <article class="attack-feed-panel">
                      <header class="panel-head">
                        <div>
                          <div class="projector-overline">
                            攻击流水
                          </div>
                          <h3>实时事件</h3>
                        </div>
                      </header>
                      <div class="attack-feed">
                        <div
                          v-for="event in latestAttackEvents"
                          :key="event.id"
                          class="attack-event"
                          :class="{ 'attack-event--success': event.is_success }"
                        >
                          <div class="attack-event__line">
                            <strong>{{ event.attacker_team }}</strong>
                            <span>→</span>
                            <strong>{{ event.victim_team }}</strong>
                          </div>
                          <div class="attack-event__meta">
                            <span>{{ getAttackTypeLabel(event.attack_type) }}</span>
                            <span>{{ event.is_success ? '成功' : '未命中' }}</span>
                            <span>{{ formatTime(event.created_at) }}</span>
                          </div>
                        </div>
                        <div
                          v-if="latestAttackEvents.length === 0"
                          class="panel-empty"
                        >
                          暂无攻击流水
                        </div>
                      </div>
                    </article>
                  </section>
                </div>

                <div class="projector-footer">
                  <span>结束 {{ formatTime(selectedContest?.ends_at ?? scoreboard?.contest.ends_at) }}</span>
                  <span>最新流量 {{ formatTime(trafficSummary?.latest_event_at) }}</span>
                  <span>轮次创建 {{ formatTime(selectedRound?.created_at) }}</span>
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
.projector-board,
.battlefield-panel,
.event-panel {
  display: flex;
  flex-direction: column;
}

.contest-projector-content {
  padding: 0;
}

.projector-stage {
  --workspace-directory-section-padding: var(--space-4) var(--space-5);
  gap: var(--space-4);
  background: transparent;
}

.projector-stage:fullscreen {
  min-height: 100vh;
  overflow: auto;
  background: var(--journal-surface);
  padding: var(--space-5);
}

.projector-header,
.projector-actions,
.projector-status,
.projector-hero,
.projector-hero__status,
.projector-metric,
.podium-card,
.projector-footer {
  display: flex;
  align-items: center;
}

.projector-header,
.projector-hero {
  justify-content: space-between;
  gap: var(--space-4);
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.14em;
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
  font-size: var(--font-size-1-60);
  line-height: 1.05;
}

.projector-actions {
  gap: var(--space-3);
}

.projector-sync {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
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
  gap: var(--space-4);
}

.contest-selector {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
  width: 100%;
  max-width: 32rem;
}

.contest-selector__label {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.contest-selector__control {
  width: min(100%, 26rem);
  flex: 1 1 16rem;
  min-height: var(--ui-control-height-md);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 76%, transparent);
  padding: 0 var(--space-3);
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 800;
}

.contest-selector__control:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.contest-selector__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 56%, var(--color-border-default));
  outline: none;
  box-shadow: 0 0 0 0.1875rem color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.contest-selector__control option {
  background: var(--color-bg-surface);
  color: var(--journal-ink);
}

.contest-selector__control,
.contest-selector__control option {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scoreboard-projector {
  position: relative;
  gap: var(--space-4);
  min-height: 38rem;
}

.projector-hero__title {
  min-width: 0;
}

.projector-hero__status {
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.projector-status,
.projector-round-pill {
  gap: var(--space-2);
  border: 1px solid var(--color-border-subtle);
  border-radius: 999rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 70%, transparent);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.projector-metrics {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--space-3);
}

.projector-metric {
  justify-content: space-between;
  gap: var(--space-3);
  min-width: 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 62%, transparent);
  padding: var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.projector-metric strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-00);
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
  gap: var(--space-4);
}

.arena-grid {
  display: grid;
  grid-template-columns: minmax(18rem, 1.05fr) minmax(26rem, 1.65fr) minmax(18rem, 1fr);
  gap: var(--space-4);
  align-items: stretch;
}

.leaderboard-panel,
.service-matrix-panel,
.traffic-panel,
.first-blood-panel,
.attack-panel,
.attack-feed-panel {
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 56%, transparent);
  padding: var(--space-4);
}

.leaderboard-panel,
.service-matrix-panel,
.traffic-panel,
.first-blood-panel,
.attack-panel,
.attack-feed-panel,
.leaderboard-list,
.attack-feed,
.attack-list,
.traffic-list,
.service-matrix {
  display: flex;
  flex-direction: column;
}

.leaderboard-panel,
.battlefield-panel,
.event-panel {
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

.podium-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.podium-card {
  min-height: 8rem;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--journal-accent) 9%, transparent);
  color: var(--color-text-secondary);
  text-align: center;
}

.podium-card.rank-1 {
  min-height: 9rem;
  background: color-mix(in srgb, var(--color-warning) 18%, transparent);
}

.podium-card strong {
  max-width: 90%;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.podium-rank {
  color: var(--color-warning);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-35);
  font-weight: 900;
}

.leaderboard-list {
  gap: var(--space-2);
  margin-top: var(--space-4);
}

.leaderboard-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  border: 1px solid transparent;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 44%, transparent);
  padding: var(--space-2-5) var(--space-3);
  color: var(--color-text-secondary);
}

.leaderboard-row--top {
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.leaderboard-rank,
.leaderboard-score {
  font-family: var(--font-family-mono);
  font-weight: 900;
}

.leaderboard-rank {
  color: var(--color-warning);
}

.leaderboard-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-score {
  color: var(--journal-ink);
}

.service-matrix-panel {
  gap: var(--space-4);
}

.service-matrix {
  gap: var(--space-2);
}

.service-team-row {
  display: grid;
  grid-template-columns: minmax(8rem, 12rem) minmax(0, 1fr);
  align-items: center;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
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
  justify-content: center;
  min-height: var(--ui-control-height-sm);
  border-radius: 0.375rem;
  padding: 0 var(--space-2-5);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.service-dot {
  min-width: 3.6rem;
  font-family: var(--font-family-mono);
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

.traffic-panel {
  gap: var(--space-4);
}

.traffic-strip span {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-text-secondary);
}

.traffic-trend {
  display: flex;
  align-items: flex-end;
  gap: var(--space-1-5);
  height: 8rem;
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-surface) 48%, transparent);
  padding: var(--space-3);
}

.traffic-bar {
  position: relative;
  display: block;
  flex: 1;
  min-height: var(--space-3);
  border-radius: 999rem 999rem 0 0;
  background: color-mix(in srgb, var(--color-primary) 58%, transparent);
  overflow: hidden;
}

.traffic-bar i {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  display: block;
  background: color-mix(in srgb, var(--color-danger) 66%, transparent);
}

.traffic-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.traffic-list,
.attack-list,
.attack-feed {
  gap: var(--space-2-5);
}

.traffic-list__title {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.first-blood-body {
  display: flex;
  min-height: 6rem;
  flex: 1;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-1);
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

.attack-event {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 50%, transparent);
  padding: var(--space-2-5);
}

.attack-event--success {
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 9%, transparent);
}

.attack-event__line,
.attack-event__meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-2);
}

.attack-event__line strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-12);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-event__line span {
  color: var(--color-danger);
  font-weight: 900;
}

.attack-event__meta {
  flex-wrap: wrap;
  margin-top: var(--space-1);
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
  font-weight: 800;
}

.panel-empty {
  padding: var(--space-4) 0;
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}

.projector-footer {
  justify-content: space-between;
  gap: var(--space-3);
  border-top: 1px solid var(--color-border-subtle);
  padding-top: var(--space-3);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

@media (max-width: 1280px) {
  .projector-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .arena-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .projector-stage {
    --workspace-directory-section-padding: var(--space-4);
  }

  .projector-header,
  .projector-hero,
  .projector-footer {
    align-items: flex-start;
    flex-direction: column;
  }

  .projector-hero__status {
    justify-content: flex-start;
  }

  .projector-metrics,
  .podium-grid,
  .traffic-columns {
    grid-template-columns: 1fr;
  }

  .scoreboard-projector {
    padding: var(--space-4);
  }

  .service-team-row {
    grid-template-columns: 1fr;
  }
}
</style>
