<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ArrowRight, BarChart3, Cast, Radar, Server, Trophy } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

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
  ScoreboardRow,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

type ContestOpsViewKey = 'environment' | 'traffic' | 'projector' | 'scoreboard'

type ContestOpsAction =
  | {
      type: 'route'
      label: string
      location:
        | {
            name: string
            params?: Record<string, string>
            query?: Record<string, string>
          }
        | {
            path: string
            query?: Record<string, string>
          }
    }
  | {
      type: 'contest-route'
      label: string
      buildLocation: (contestId: string) => {
        name: string
        params?: Record<string, string>
        query?: Record<string, string>
      }
    }

interface ContestOpsDefinition {
  overline: string
  title: string
  copy: string
  helper: string
  metricLabel: string
  metricHint: string
  icon: typeof Server
  primaryAction: ContestOpsAction
  secondaryAction: ContestOpsAction
}

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const awdContests = ref<ContestDetailData[]>([])
const projectorLoading = ref(false)
const projectorError = ref('')
const projectorRound = ref<AWDRoundData | null>(null)
const projectorSummary = ref<AWDRoundSummaryData | null>(null)
const projectorTrafficSummary = ref<AWDTrafficSummaryData | null>(null)
const projectorServices = ref<AWDTeamServiceData[]>([])
const projectorAttacks = ref<AWDAttackLogData[]>([])
const projectorScoreboardRows = ref<ScoreboardRow[]>([])

let projectorRequestToken = 0

const operationDefinitions: Record<ContestOpsViewKey, ContestOpsDefinition> = {
  environment: {
    overline: 'Event Operations',
    title: '环境管理',
    copy: '统一接入当前可操作 AWD 赛事的 checker、SLA、防守分和赛前环境准备，不再分散到平行后台里维护。',
    helper: '优先进入推荐赛事的 AWD 配置段，也可以回到竞赛目录切换其它赛事。',
    metricLabel: '配置入口',
    metricHint: 'AWD Checker / SLA / 防守分',
    icon: Server,
    primaryAction: {
      type: 'contest-route',
      label: '进入 AWD 配置',
      buildLocation: (contestId) => ({
        name: 'AdminAwdOverview',
        params: { id: contestId },
      }),
    },
    secondaryAction: {
      type: 'contest-route',
      label: '打开服务矩阵',
      buildLocation: (contestId) => ({
        name: 'AdminAwdServices',
        params: { id: contestId },
      }),
    },
  },
  traffic: {
    overline: 'Traffic Control',
    title: '流量监控',
    copy: '从管理员后台直接进入轮次态势、服务巡检、攻击流水和流量筛选，减少先翻到竞赛目录再找运行段的路径。',
    helper: '默认落到推荐赛事的运行段态势面板，继续看流量和当前轮状态。',
    metricLabel: '态势入口',
    metricHint: '回合态势 / 服务状态 / 攻击流水',
    icon: Radar,
    primaryAction: {
      type: 'contest-route',
      label: '进入流量态势',
      buildLocation: (contestId) => ({
        name: 'AdminAwdTraffic',
        params: { id: contestId },
      }),
    },
    secondaryAction: {
      type: 'route',
      label: '返回竞赛目录',
      location: {
        name: 'ContestManage',
        query: { panel: 'list' },
      },
    },
  },
  projector: {
    overline: 'Projection Desk',
    title: '大屏投射',
    copy: '把推荐 AWD 赛事的轮次进度、实时榜单、攻击反馈和流量热点收拢到同一页，便于现场直接投屏。',
    helper: '默认展示推荐赛事的当前轮次数据，也可以继续跳转到赛事详情或全站排行榜。',
    metricLabel: '投屏候选',
    metricHint: '当前轮次 / 实时榜单 / 最新攻击',
    icon: Cast,
    primaryAction: {
      type: 'route',
      label: '打开竞赛排行榜',
      location: {
        path: '/scoreboard',
        query: { tab: 'contest' },
      },
    },
    secondaryAction: {
      type: 'contest-route',
      label: '打开赛事详情',
      buildLocation: (contestId) => ({
        name: 'ContestDetail',
        params: { id: contestId },
      }),
    },
  },
  scoreboard: {
    overline: 'Scoreboard Desk',
    title: '排行榜',
    copy: '集中给出竞赛榜单入口和当前推荐赛事，便于在管理员后台快速切到榜单视图或继续回到工作台复核实时分数。',
    helper: '默认打开全站竞赛排行榜，也可以继续进入推荐赛事的运行段查看实时排行。',
    metricLabel: '榜单入口',
    metricHint: '竞赛排行 / 封榜状态 / 实时榜单',
    icon: Trophy,
    primaryAction: {
      type: 'route',
      label: '打开竞赛排行榜',
      location: {
        path: '/scoreboard',
        query: { tab: 'contest' },
      },
    },
    secondaryAction: {
      type: 'contest-route',
      label: '查看实时榜单',
      buildLocation: (contestId) => ({
        name: 'AdminAwdOverview',
        params: { id: contestId },
      }),
    },
  },
}

const currentView = computed<ContestOpsViewKey>(() => {
  if (route.name === 'PlatformContestOpsTraffic') return 'traffic'
  if (route.name === 'PlatformContestOpsProjector') return 'projector'
  if (route.name === 'PlatformContestOpsScoreboard') return 'scoreboard'
  return 'environment'
})

const currentDefinition = computed(() => operationDefinitions[currentView.value])
const preferredContest = computed(
  () =>
    awdContests.value.find((item) => item.status === 'running' || item.status === 'frozen') ||
    awdContests.value.find((item) => item.status === 'registering') ||
    awdContests.value[0] ||
    null
)
const runningCount = computed(
  () =>
    awdContests.value.filter((item) => item.status === 'running' || item.status === 'frozen').length
)
const currentMetricValue = computed(() => {
  if (currentView.value === 'projector') {
    return projectorRound.value ? `Round ${projectorRound.value.round_number}` : '待同步'
  }
  if (currentView.value === 'scoreboard') {
    return String(runningCount.value)
  }
  return preferredContest.value ? getStatusLabel(preferredContest.value.status) : '待选择'
})

const projectorMetrics = computed(() => {
  const metrics = projectorSummary.value?.metrics
  if (!metrics) {
    return []
  }
  return [
    {
      label: '在线服务',
      value: String(metrics.service_up_count),
      helper: `共 ${metrics.total_service_count} 个服务单元`,
    },
    {
      label: '失陷服务',
      value: String(metrics.service_compromised_count),
      helper: `离线 ${metrics.service_down_count} 个，受攻击 ${metrics.attacked_service_count} 个`,
    },
    {
      label: '成功攻击',
      value: String(metrics.successful_attack_count),
      helper: `总攻击 ${metrics.total_attack_count} 次，失败 ${metrics.failed_attack_count} 次`,
    },
    {
      label: '实时流量',
      value: String(projectorTrafficSummary.value?.total_request_count ?? 0),
      helper: `活跃攻击队 ${projectorTrafficSummary.value?.active_attacker_team_count ?? 0} 支`,
    },
  ]
})

const projectorLatestAttacks = computed(() => projectorAttacks.value.slice(0, 6))
const projectorHotChallenges = computed(() => projectorTrafficSummary.value?.top_challenges.slice(0, 3) || [])
const projectorTopAttackers = computed(() => projectorTrafficSummary.value?.top_attackers.slice(0, 3) || [])
const projectorCompromisedServices = computed(() =>
  projectorServices.value.filter((item) => item.service_status === 'compromised').slice(0, 5)
)
const projectorTopTeam = computed(() => projectorScoreboardRows.value[0] || null)

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatCompactTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatAttackResultLabel(success: boolean): string {
  return success ? '命中' : '未命中'
}

function formatServiceStatusLabel(status: string): string {
  switch (status) {
    case 'up':
      return '在线'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待同步'
  }
}

function pickProjectorRound(rounds: AWDRoundData[]): AWDRoundData | null {
  return (
    rounds.find((item) => item.status === 'running') ||
    rounds.find((item) => item.status === 'finished') ||
    rounds[rounds.length - 1] ||
    null
  )
}

async function loadContests() {
  loading.value = true
  loadError.value = ''

  try {
    const response = await getContests({
      page: 1,
      page_size: 100,
    })
    awdContests.value = response.list.filter((item) => item.mode === 'awd')
  } catch (error) {
    awdContests.value = []
    loadError.value = error instanceof Error ? error.message : '赛事运维入口加载失败'
  } finally {
    loading.value = false
  }
}

function resetProjectorData() {
  projectorError.value = ''
  projectorRound.value = null
  projectorSummary.value = null
  projectorTrafficSummary.value = null
  projectorServices.value = []
  projectorAttacks.value = []
  projectorScoreboardRows.value = []
}

async function loadProjectorData(contestId: string) {
  const requestToken = ++projectorRequestToken
  projectorLoading.value = true
  projectorError.value = ''

  try {
    const rounds = await listContestAWDRounds(contestId)
    if (requestToken !== projectorRequestToken) {
      return
    }

    const nextRound = pickProjectorRound(rounds)
    projectorRound.value = nextRound
    if (!nextRound) {
      projectorSummary.value = null
      projectorTrafficSummary.value = null
      projectorServices.value = []
      projectorAttacks.value = []
      projectorScoreboardRows.value = []
      projectorError.value = '当前赛事还没有生成可投屏轮次'
      return
    }

    const [summary, trafficSummary, services, attacks, scoreboard] = await Promise.all([
      getContestAWDRoundSummary(contestId, nextRound.id),
      getContestAWDRoundTrafficSummary(contestId, nextRound.id),
      listContestAWDRoundServices(contestId, nextRound.id),
      listContestAWDRoundAttacks(contestId, nextRound.id),
      getAdminContestLiveScoreboard(contestId, { page: 1, page_size: 10 }),
    ])

    if (requestToken !== projectorRequestToken) {
      return
    }

    projectorSummary.value = summary
    projectorTrafficSummary.value = trafficSummary
    projectorServices.value = services
    projectorAttacks.value = attacks
    projectorScoreboardRows.value = scoreboard.scoreboard.list
  } catch (error) {
    if (requestToken !== projectorRequestToken) {
      return
    }
    resetProjectorData()
    projectorError.value = error instanceof Error ? error.message : '投屏数据加载失败'
  } finally {
    if (requestToken === projectorRequestToken) {
      projectorLoading.value = false
    }
  }
}

function resolveLocation(action: ContestOpsAction) {
  if (action.type === 'contest-route') {
    const contest = preferredContest.value
    if (!contest) {
      return {
        name: 'ContestManage',
        query: { panel: 'list' },
      }
    }
    return action.buildLocation(contest.id)
  }

  return action.location
}

async function executeAction(action: ContestOpsAction) {
  await router.push(resolveLocation(action))
}

onMounted(() => {
  void loadContests()
})

watch(
  () => [currentView.value, preferredContest.value?.id || ''] as const,
  ([view, contestId]) => {
    if (view !== 'projector') {
      projectorRequestToken++
      projectorLoading.value = false
      resetProjectorData()
      return
    }
    if (!contestId) {
      projectorLoading.value = false
      resetProjectorData()
      return
    }
    void loadProjectorData(contestId)
  },
  { immediate: true }
)
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <header class="list-heading contest-ops-hero workspace-directory-section">
      <div class="contest-ops-hero__main">
        <div class="workspace-overline">{{ currentDefinition.overline }}</div>
        <h1 class="workspace-page-title">{{ currentDefinition.title }}</h1>
        <p class="workspace-page-copy">{{ currentDefinition.copy }}</p>
      </div>

      <div class="contest-ops-hero__actions">
        <button
          id="contest-ops-primary-action"
          type="button"
          class="ui-btn ui-btn--primary"
          @click="executeAction(currentDefinition.primaryAction)"
        >
          <ArrowRight class="h-4 w-4" />
          {{ currentDefinition.primaryAction.label }}
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="executeAction(currentDefinition.secondaryAction)"
        >
          {{ currentDefinition.secondaryAction.label }}
        </button>
      </div>
    </header>

    <div
      class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"
    >
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">AWD 赛事</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ awdContests.length }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">当前可纳入赛事运维的 AWD 赛事总数</div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">推荐赛事</div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ preferredContest ? preferredContest.title : '暂无' }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          优先选择进行中，其次报名中 AWD 赛事
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          {{ currentDefinition.metricLabel }}
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ currentMetricValue }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          {{ currentDefinition.metricHint }}
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">进行中</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ runningCount }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          可直接承接流量与榜单运维的赛事数量
        </div>
      </article>
    </div>

    <section v-if="loading" class="workspace-directory-section contest-ops-section">
      <AppLoading>正在同步赛事运维入口...</AppLoading>
    </section>

    <AppEmpty
      v-else-if="loadError"
      class="workspace-directory-section contest-ops-section"
      title="赛事运维入口暂时不可用"
      :description="loadError"
      icon="AlertTriangle"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--ghost" @click="loadContests">重试加载</button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="awdContests.length === 0"
      class="workspace-directory-section contest-ops-section"
      title="当前还没有可运维的 AWD 赛事"
      description="先在竞赛目录中创建或切换到 AWD 赛事，这里再承接环境、流量和榜单入口。"
      icon="Trophy"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--primary" @click="router.push({ name: 'ContestManage', query: { panel: 'create' } })">
          前往创建竞赛
        </button>
      </template>
    </AppEmpty>

    <template v-else-if="preferredContest && currentView === 'projector'">
      <section class="workspace-directory-section contest-ops-section contest-projector-stage">
        <header class="list-heading">
          <div>
            <div class="workspace-overline">Projector Live</div>
            <h2 class="list-heading__title">{{ preferredContest.title }}</h2>
          </div>
          <div class="contest-section-meta">
            {{ getStatusLabel(preferredContest.status) }} ·
            {{ projectorRound ? `Round ${projectorRound.round_number}` : '待同步' }}
          </div>
        </header>

        <div v-if="projectorLoading" class="contest-projector-loading">
          <AppLoading>正在同步投屏数据...</AppLoading>
        </div>

        <AppEmpty
          v-else-if="projectorError"
          class="contest-projector-empty"
          title="投屏数据暂时不可用"
          :description="projectorError"
          icon="Cast"
        />

        <div v-else class="contest-projector-layout">
          <div class="contest-projector-spotlight">
            <article class="contest-projector-hero-card">
              <div class="contest-projector-hero-card__head">
                <div>
                  <div class="journal-note-label">当前轮次</div>
                  <h3>{{ projectorRound ? `Round ${projectorRound.round_number}` : '待同步' }}</h3>
                </div>
                <div class="contest-projector-chip">
                  {{ projectorRound ? projectorRound.status : 'pending' }}
                </div>
              </div>
              <div class="contest-projector-hero-card__meta">
                <span>比赛窗口 {{ formatDateTime(preferredContest.starts_at) }} - {{ formatDateTime(preferredContest.ends_at) }}</span>
                <span>最近事件 {{ formatCompactTime(projectorTrafficSummary?.latest_event_at) }}</span>
                <span>榜首 {{ projectorTopTeam ? `${projectorTopTeam.team_name} · ${projectorTopTeam.score}` : '待同步' }}</span>
              </div>
            </article>

            <div class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-projector-metrics">
              <article
                v-for="item in projectorMetrics"
                :key="item.label"
                class="journal-note progress-card metric-panel-card"
              >
                <div class="journal-note-label progress-card-label metric-panel-label">{{ item.label }}</div>
                <div class="journal-note-value progress-card-value metric-panel-value">{{ item.value }}</div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">{{ item.helper }}</div>
              </article>
            </div>
          </div>

          <div class="contest-projector-board">
            <section class="contest-projector-panel">
              <header class="contest-projector-panel__head">
                <div class="workspace-overline">Scoreboard</div>
                <h3>实时榜单</h3>
              </header>
              <div v-if="projectorScoreboardRows.length === 0" class="contest-projector-note">当前还没有榜单数据。</div>
              <div v-else class="contest-projector-scoreboard">
                <div
                  v-for="item in projectorScoreboardRows.slice(0, 6)"
                  :key="item.team_id"
                  class="contest-projector-scoreboard__row"
                >
                  <span class="contest-projector-rank">{{ item.rank }}</span>
                  <span class="contest-projector-team">{{ item.team_name }}</span>
                  <strong class="contest-projector-score">{{ item.score }}</strong>
                </div>
              </div>
            </section>

            <section class="contest-projector-panel">
              <header class="contest-projector-panel__head">
                <div class="workspace-overline">Attack Feed</div>
                <h3>最新攻击</h3>
              </header>
              <div v-if="projectorLatestAttacks.length === 0" class="contest-projector-note">当前轮次还没有攻击记录。</div>
              <div v-else class="contest-projector-feed">
                <article
                  v-for="item in projectorLatestAttacks"
                  :key="item.id"
                  class="contest-projector-feed__row"
                >
                  <div class="contest-projector-feed__title">
                    <span>{{ item.attacker_team }}</span>
                    <span>→</span>
                    <span>{{ item.victim_team }}</span>
                  </div>
                  <div class="contest-projector-feed__meta">
                    <span>{{ formatAttackResultLabel(item.is_success) }}</span>
                    <span>{{ item.score_gained }} 分</span>
                    <span>{{ formatCompactTime(item.created_at) }}</span>
                  </div>
                </article>
              </div>
            </section>
          </div>

          <div class="contest-projector-board contest-projector-board--lower">
            <section class="contest-projector-panel">
              <header class="contest-projector-panel__head">
                <div class="workspace-overline">Service Heat</div>
                <h3>热点服务</h3>
              </header>
              <div v-if="projectorHotChallenges.length === 0" class="contest-projector-note">当前轮次还没有服务热点。</div>
              <div v-else class="contest-projector-list">
                <article
                  v-for="item in projectorHotChallenges"
                  :key="item.challenge_id"
                  class="contest-projector-list__row"
                >
                  <div>
                    <strong>{{ item.challenge_title || item.challenge_id }}</strong>
                    <div class="contest-projector-list__meta">请求 {{ item.request_count }} · 错误 {{ item.error_count }}</div>
                  </div>
                </article>
              </div>
            </section>

            <section class="contest-projector-panel">
              <header class="contest-projector-panel__head">
                <div class="workspace-overline">Compromised</div>
                <h3>服务状态</h3>
              </header>
              <div v-if="projectorCompromisedServices.length === 0" class="contest-projector-note">当前没有失陷服务。</div>
              <div v-else class="contest-projector-list">
                <article
                  v-for="item in projectorCompromisedServices"
                  :key="item.id"
                  class="contest-projector-list__row"
                >
                  <div>
                    <strong>{{ item.team_name }}</strong>
                    <div class="contest-projector-list__meta">
                      {{ formatServiceStatusLabel(item.service_status) }} · 收到攻击 {{ item.attack_received }} 次
                    </div>
                  </div>
                </article>
              </div>
            </section>

            <section class="contest-projector-panel">
              <header class="contest-projector-panel__head">
                <div class="workspace-overline">Traffic Focus</div>
                <h3>流量焦点</h3>
              </header>
              <div v-if="projectorTopAttackers.length === 0" class="contest-projector-note">当前轮次还没有活跃流量。</div>
              <div v-else class="contest-projector-list">
                <article
                  v-for="item in projectorTopAttackers"
                  :key="item.team_id"
                  class="contest-projector-list__row"
                >
                  <div>
                    <strong>{{ item.team_name }}</strong>
                    <div class="contest-projector-list__meta">
                      请求 {{ item.request_count }} · 错误 {{ item.error_count }}
                    </div>
                  </div>
                </article>
              </div>
            </section>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="preferredContest">
      <section class="workspace-directory-section contest-ops-section">
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Recommended Contest</div>
            <h2 class="list-heading__title">{{ preferredContest.title }}</h2>
          </div>
          <div class="contest-section-meta">
            {{ getStatusLabel(preferredContest.status) }} · {{ getModeLabel(preferredContest.mode) }}
          </div>
        </header>

        <div class="contest-ops-grid">
          <article class="contest-ops-card">
            <div class="contest-ops-card__icon">
              <component :is="currentDefinition.icon" class="h-4 w-4" />
            </div>
            <div class="contest-ops-card__body">
              <h3 class="contest-ops-card__title">{{ currentDefinition.title }}</h3>
              <p class="contest-ops-card__copy">{{ currentDefinition.helper }}</p>
            </div>
          </article>

          <article class="contest-ops-card">
            <div class="contest-ops-card__icon">
              <BarChart3 class="h-4 w-4" />
            </div>
            <div class="contest-ops-card__body">
              <h3 class="contest-ops-card__title">赛事窗口</h3>
              <p class="contest-ops-card__copy">
                {{ formatDateTime(preferredContest.starts_at) }} 至
                {{ formatDateTime(preferredContest.ends_at) }}
              </p>
            </div>
          </article>
        </div>
      </section>

      <section class="workspace-directory-section contest-ops-section">
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Next Step</div>
            <h2 class="list-heading__title">继续处理当前赛事</h2>
          </div>
          <div class="contest-section-meta">优先承接正在运行或最近可操作的一场赛事</div>
        </header>

        <div class="contest-ops-actions">
          <button
            id="contest-ops-inline-primary"
            type="button"
            class="ui-btn ui-btn--primary"
            @click="executeAction(currentDefinition.primaryAction)"
          >
            <ArrowRight class="h-4 w-4" />
            {{ currentDefinition.primaryAction.label }}
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="executeAction(currentDefinition.secondaryAction)"
          >
            {{ currentDefinition.secondaryAction.label }}
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="router.push({ name: 'ContestManage', query: { panel: 'list' } })"
          >
            返回竞赛目录
          </button>
        </div>
      </section>
    </template>
  </section>
</template>

<style scoped>
.contest-ops-hero,
.contest-ops-section {
  padding: 1.5rem;
}

.contest-ops-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.contest-ops-hero__main {
  display: grid;
  gap: 0.75rem;
  max-width: 52rem;
}

.contest-ops-hero__actions,
.contest-ops-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.contest-ops-summary {
  margin-top: 1.5rem;
}

.contest-projector-stage {
  display: grid;
  gap: 1.25rem;
}

.contest-projector-loading,
.contest-projector-empty {
  margin-top: 0.5rem;
}

.contest-projector-layout,
.contest-projector-spotlight,
.contest-projector-board {
  display: grid;
  gap: 1rem;
}

.contest-projector-board {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-projector-board--lower {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.contest-projector-hero-card,
.contest-projector-panel {
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.2rem;
  background: color-mix(in srgb, var(--journal-surface) 95%, transparent);
}

.contest-projector-hero-card {
  display: grid;
  gap: 0.85rem;
  padding: 1.15rem 1.25rem;
}

.contest-projector-hero-card__head,
.contest-projector-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.contest-projector-hero-card__head h3,
.contest-projector-panel__head h3 {
  margin: 0.2rem 0 0;
  color: var(--journal-ink);
}

.contest-projector-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.28rem 0.7rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  font-size: 0.78rem;
  font-weight: 600;
  text-transform: uppercase;
}

.contest-projector-hero-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  color: var(--color-text-secondary);
  font-size: 0.88rem;
}

.contest-projector-metrics {
  margin-top: 0;
}

.contest-projector-panel {
  display: grid;
  gap: 0.9rem;
  padding: 1rem 1.05rem;
}

.contest-projector-note {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.contest-projector-scoreboard,
.contest-projector-feed,
.contest-projector-list {
  display: grid;
  gap: 0.7rem;
}

.contest-projector-scoreboard__row,
.contest-projector-feed__row,
.contest-projector-list__row {
  display: grid;
  gap: 0.35rem;
  padding: 0.8rem 0.9rem;
  border-radius: 0.95rem;
  background: color-mix(in srgb, var(--journal-surface-muted) 82%, transparent);
}

.contest-projector-scoreboard__row {
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 0.75rem;
}

.contest-projector-rank,
.contest-projector-score {
  font-family: var(--font-family-mono, 'JetBrains Mono', monospace);
}

.contest-projector-rank {
  color: var(--journal-accent);
  font-weight: 700;
}

.contest-projector-team {
  color: var(--journal-ink);
  font-weight: 600;
}

.contest-projector-score {
  color: var(--journal-ink);
}

.contest-projector-feed__title,
.contest-projector-feed__meta,
.contest-projector-list__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem 0.75rem;
}

.contest-projector-feed__title {
  color: var(--journal-ink);
  font-weight: 600;
}

.contest-projector-feed__meta,
.contest-projector-list__meta {
  color: var(--color-text-secondary);
  font-size: 0.86rem;
}

.contest-ops-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
}

.contest-ops-card {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.9rem;
  padding: 1.1rem 1.2rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.1rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.contest-ops-card__icon {
  display: inline-flex;
  width: 2.1rem;
  height: 2.1rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.8rem;
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent);
}

.contest-ops-card__body {
  display: grid;
  gap: 0.45rem;
}

.contest-ops-card__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-ops-card__copy {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

@media (max-width: 1100px) {
  .contest-projector-board,
  .contest-projector-board--lower {
    grid-template-columns: 1fr;
  }
}
</style>
