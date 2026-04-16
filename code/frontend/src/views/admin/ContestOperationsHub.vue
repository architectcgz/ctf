<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, BarChart3, Cast, Radar, Server, Trophy } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getContests } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
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
        name: 'ContestEdit',
        params: { id: contestId },
        query: { panel: 'awd-config' },
      }),
    },
    secondaryAction: {
      type: 'contest-route',
      label: '打开赛前检查',
      buildLocation: (contestId) => ({
        name: 'ContestEdit',
        params: { id: contestId },
        query: { panel: 'preflight' },
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
        name: 'ContestEdit',
        params: { id: contestId },
        query: { panel: 'operations', opsPanel: 'inspector' },
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
    copy: '先把赛事大屏入口独立出来，便于现场切到可投屏视角，后续再继续补充全屏展示和轮播能力。',
    helper: '当前先提供竞赛排行榜和比赛详情入口，作为现场投屏的最小可用版本。',
    metricLabel: '投屏候选',
    metricHint: '排行榜 / 竞赛详情 / 运行态势',
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
        name: 'ContestEdit',
        params: { id: contestId },
        query: { panel: 'operations', opsPanel: 'inspector' },
      }),
    },
  },
}

const currentView = computed<ContestOpsViewKey>(() => {
  if (route.name === 'AdminContestOpsTraffic') return 'traffic'
  if (route.name === 'AdminContestOpsProjector') return 'projector'
  if (route.name === 'AdminContestOpsScoreboard') return 'scoreboard'
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
  if (currentView.value === 'scoreboard') {
    return String(runningCount.value)
  }
  return preferredContest.value ? getStatusLabel(preferredContest.value.status) : '待选择'
})

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
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
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero journal-eyebrow-text workspace-shell flex min-h-full flex-1 flex-col"
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
      class="metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"
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
      description="先在竞赛管理中创建或切换到 AWD 赛事，这里再承接环境、流量和榜单入口。"
      icon="Trophy"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--primary" @click="router.push({ name: 'ContestManage', query: { panel: 'create' } })">
          前往创建竞赛
        </button>
      </template>
    </AppEmpty>

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
</style>
