<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useRoute, useRouter, type RouteLocationRaw } from 'vue-router'

import { getContests } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

type ContestOpsViewKey = 'environment' | 'traffic' | 'projector' | 'scoreboard'

interface ContestOpsDefinition {
  overline: string
  title: string
  copy: string
  directoryMeta: string
  metricLabel: string
  metricHint: string
  metricValue: (context: {
    awdContests: ContestDetailData[]
    preferredContest: ContestDetailData | null
    activeCount: number
    frozenCount: number
  }) => string | number
  primaryActionLabel: string
  secondaryActionLabel: string
  getPrimaryLocation: (contest: ContestDetailData) => RouteLocationRaw
  getSecondaryLocation: (contest: ContestDetailData) => RouteLocationRaw
}

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const awdContests = ref<ContestDetailData[]>([])

const operationDefinitions: Record<ContestOpsViewKey, ContestOpsDefinition> = {
  environment: {
    overline: 'Contest Environment',
    title: '环境管理',
    copy: '这里直接承接可运维的 AWD 赛事，用统一目录处理 checker、SLA、防守分和赛前准备，不再通过漂浮入口反复跳转。',
    directoryMeta: '按开始时间查看全部可操作 AWD 赛事',
    metricLabel: '待配置赛事',
    metricHint: '优先处理进行中与已冻结赛事的环境准备',
    metricValue: ({ activeCount }) => activeCount,
    primaryActionLabel: '进入 AWD 配置',
    secondaryActionLabel: '打开赛前检查',
    getPrimaryLocation: (contest) => ({
      name: 'ContestEdit',
      params: { id: contest.id },
      query: { panel: 'awd-config' },
    }),
    getSecondaryLocation: (contest) => ({
      name: 'ContestEdit',
      params: { id: contest.id },
      query: { panel: 'preflight' },
    }),
  },
  traffic: {
    overline: 'Traffic Control',
    title: '流量监控',
    copy: '将回合态势、服务状态和攻击流水的入口收敛到具体赛事行内，操作时始终保留对象上下文。',
    directoryMeta: '优先关注正在运行和已冻结的 AWD 赛事',
    metricLabel: '监控焦点',
    metricHint: '运行中与冻结赛事会持续产生流量与对抗数据',
    metricValue: ({ activeCount }) => activeCount,
    primaryActionLabel: '进入流量态势',
    secondaryActionLabel: '查看赛事详情',
    getPrimaryLocation: (contest) => ({
      name: 'ContestEdit',
      params: { id: contest.id },
      query: { panel: 'operations', opsPanel: 'inspector' },
    }),
    getSecondaryLocation: (contest) => ({
      name: 'ContestDetail',
      params: { id: contest.id },
    }),
  },
  projector: {
    overline: 'Projection Desk',
    title: '大屏投射',
    copy: '面向现场展示的操作保持轻量，但仍以赛事对象为中心，让投屏切换和详情查看都贴着具体比赛展开。',
    directoryMeta: '从目录中选择需要投射的 AWD 赛事',
    metricLabel: '推荐投屏',
    metricHint: '默认优先展示进行中，其次展示最近可操作的 AWD 赛事',
    metricValue: ({ preferredContest }) => preferredContest?.title ?? '暂无',
    primaryActionLabel: '打开赛事详情',
    secondaryActionLabel: '打开竞赛排行榜',
    getPrimaryLocation: (contest) => ({
      name: 'ContestDetail',
      params: { id: contest.id },
    }),
    getSecondaryLocation: (contest) => ({
      path: '/scoreboard',
      query: { tab: 'contest', contest: contest.id },
    }),
  },
  scoreboard: {
    overline: 'Scoreboard Desk',
    title: '排行榜',
    copy: '榜单入口不再独立悬空，直接围绕赛事目录查看实时排行和赛事详情，减少在工作台之间来回切换。',
    directoryMeta: '从具体 AWD 赛事进入榜单与详情视角',
    metricLabel: '榜单关注',
    metricHint: '进行中与冻结赛事仍然是当前最需要关注的分数对象',
    metricValue: ({ activeCount, frozenCount }) => `${activeCount} / ${frozenCount}`,
    primaryActionLabel: '查看实时榜单',
    secondaryActionLabel: '打开赛事详情',
    getPrimaryLocation: (contest) => ({
      name: 'ContestEdit',
      params: { id: contest.id },
      query: { panel: 'operations', opsPanel: 'inspector' },
    }),
    getSecondaryLocation: (contest) => ({
      name: 'ContestDetail',
      params: { id: contest.id },
    }),
  },
}

const currentView = computed<ContestOpsViewKey>(() => {
  if (route.name === 'AdminContestOpsTraffic') return 'traffic'
  if (route.name === 'AdminContestOpsProjector') return 'projector'
  if (route.name === 'AdminContestOpsScoreboard') return 'scoreboard'
  return 'environment'
})

const currentDefinition = computed(() => operationDefinitions[currentView.value])
const preferredContest = computed<ContestDetailData | null>(
  () =>
    awdContests.value.find((item) => item.status === 'running' || item.status === 'frozen') ||
    awdContests.value.find((item) => item.status === 'registering' || item.status === 'published') ||
    awdContests.value[0] ||
    null
)
const activeCount = computed(
  () =>
    awdContests.value.filter((item) => item.status === 'running' || item.status === 'frozen').length
)
const frozenCount = computed(() => awdContests.value.filter((item) => item.status === 'frozen').length)
const directoryLabel = computed(() => `${awdContests.value.length} 场 AWD 赛事`)
const currentMetricValue = computed(() =>
  currentDefinition.value.metricValue({
    awdContests: awdContests.value,
    preferredContest: preferredContest.value,
    activeCount: activeCount.value,
    frozenCount: frozenCount.value,
  })
)

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStatusPillClass(status: ContestDetailData['status']): string {
  if (status === 'running') return 'contest-ops-status-pill--running'
  if (status === 'registering' || status === 'published') return 'contest-ops-status-pill--registering'
  if (status === 'draft') return 'contest-ops-status-pill--draft'
  if (status === 'frozen') return 'contest-ops-status-pill--frozen'
  if (status === 'ended' || status === 'archived') return 'contest-ops-status-pill--ended'
  if (status === 'cancelled') return 'contest-ops-status-pill--cancelled'
  return 'contest-ops-status-pill--neutral'
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

async function openContestManage() {
  await router.push({ name: 'ContestManage', query: { panel: 'overview' } })
}

async function openPrimaryAction(contest: ContestDetailData) {
  await router.push(currentDefinition.value.getPrimaryLocation(contest))
}

async function openSecondaryAction(contest: ContestDetailData) {
  await router.push(currentDefinition.value.getSecondaryLocation(contest))
}

onMounted(() => {
  void loadContests()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <header class="list-heading contest-ops-workbench-head">
      <div class="contest-ops-workbench-head__main">
        <div class="workspace-overline">{{ currentDefinition.overline }}</div>
        <h1 class="workspace-page-title">{{ currentDefinition.title }}</h1>
        <p class="workspace-page-copy">{{ currentDefinition.copy }}</p>
      </div>

      <div class="contest-ops-workbench-head__actions">
        <button type="button" class="ui-btn ui-btn--ghost" @click="loadContests">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button type="button" class="ui-btn ui-btn--primary" @click="openContestManage">
          返回竞赛管理
        </button>
      </div>
    </header>

    <div
      class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"
    >
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">AWD 赛事</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ awdContests.length }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">当前已接入赛事运维链路的 AWD 赛事总数</div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">进行中</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ activeCount }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">当前仍可继续处理流量、榜单与环境动作的赛事数量</div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">推荐赛事</div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ preferredContest ? preferredContest.title : '暂无' }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">默认优先选择进行中，其次已冻结和最近可操作的 AWD 赛事</div>
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
    </div>

    <section v-if="loading" class="workspace-directory-section contest-ops-directory-section">
      <AppLoading>正在同步 AWD 赛事目录...</AppLoading>
    </section>

    <AppEmpty
      v-else-if="loadError"
      class="workspace-directory-section contest-ops-directory-section"
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
      class="workspace-directory-section contest-ops-directory-section"
      title="当前还没有可运维的 AWD 赛事"
      description="先在竞赛管理中创建 AWD 赛事，这里再接入环境、流量、投屏和排行榜运维。"
      icon="Trophy"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          @click="router.push({ name: 'ContestManage', query: { panel: 'create' } })"
        >
          前往创建竞赛
        </button>
      </template>
    </AppEmpty>

    <section v-else class="workspace-directory-section contest-ops-directory-section">
      <header class="list-heading">
        <div>
          <div class="workspace-overline">AWD Directory</div>
          <h2 class="list-heading__title">全部 AWD 赛事</h2>
        </div>
        <div class="contest-ops-directory-meta">
          <span>{{ currentDefinition.directoryMeta }}</span>
          <span>{{ directoryLabel }}</span>
        </div>
      </header>

      <div class="contest-ops-directory workspace-directory-list">
        <div class="contest-ops-directory__head" aria-hidden="true">
          <span>赛事</span>
          <span>模式</span>
          <span>状态</span>
          <span>开始时间</span>
          <span>结束时间</span>
          <span class="contest-ops-directory__head-actions">操作</span>
        </div>

        <article v-for="contest in awdContests" :key="contest.id" class="contest-ops-row">
          <div class="contest-ops-row__identity">
            <h3 class="contest-ops-row__title" :title="contest.title">{{ contest.title }}</h3>
            <p class="contest-ops-row__description">
              {{ contest.description || '当前未填写赛事说明。' }}
            </p>
          </div>

          <div class="contest-ops-row__mode">{{ getModeLabel(contest.mode) }}</div>

          <div class="contest-ops-row__status">
            <span class="ui-badge contest-ops-status-pill" :class="getStatusPillClass(contest.status)">
              {{ getStatusLabel(contest.status) }}
            </span>
          </div>

          <div class="contest-ops-row__starts-at">
            <span>{{ formatDateTime(contest.starts_at) }}</span>
          </div>

          <div class="contest-ops-row__ends-at">
            <span>{{ formatDateTime(contest.ends_at) }}</span>
          </div>

          <div class="ui-row-actions contest-ops-row__actions" role="group" aria-label="赛事运维操作">
            <button
              :id="`contest-ops-row-primary-${contest.id}`"
              type="button"
              class="ui-btn ui-btn--sm ui-btn--primary"
              @click="openPrimaryAction(contest)"
            >
              {{ currentDefinition.primaryActionLabel }}
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--sm ui-btn--secondary"
              @click="openSecondaryAction(contest)"
            >
              {{ currentDefinition.secondaryActionLabel }}
            </button>
          </div>
        </article>
      </div>
    </section>
  </section>
</template>

<style scoped>
.contest-ops-workbench-head,
.contest-ops-directory-section {
  padding: 1.5rem;
}

.contest-ops-workbench-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.contest-ops-workbench-head__main {
  display: grid;
  gap: 0.75rem;
  max-width: 58rem;
}

.contest-ops-workbench-head__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.contest-ops-summary {
  margin-top: 1.5rem;
}

.contest-ops-directory-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.75rem 1.25rem;
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-ops-directory {
  --contest-ops-directory-columns: minmax(18rem, 1.58fr) minmax(6rem, 0.5fr) minmax(7rem, 0.62fr)
    minmax(9.75rem, 0.84fr) minmax(9.75rem, 0.84fr) minmax(13rem, 13rem);
  display: grid;
  gap: 0;
}

.contest-ops-directory__head,
.contest-ops-row {
  display: grid;
  grid-template-columns: var(--contest-ops-directory-columns);
  gap: var(--space-4);
}

.contest-ops-directory__head {
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-ops-directory__head > span,
.contest-ops-row > div {
  min-width: 0;
}

.contest-ops-directory__head-actions {
  text-align: right;
}

.contest-ops-row {
  align-items: start;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.contest-ops-row__identity {
  display: grid;
  gap: var(--space-1-5);
}

.contest-ops-row__title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-1-00);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-ops-row__description {
  margin: 0;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--font-size-0-875);
  line-height: 1.55;
  color: var(--journal-muted);
}

.contest-ops-row__mode,
.contest-ops-row__starts-at,
.contest-ops-row__ends-at {
  font-size: var(--font-size-0-90);
  color: var(--journal-muted);
}

.contest-ops-row__starts-at span,
.contest-ops-row__ends-at span {
  display: inline-block;
  line-height: 1.5;
}

.contest-ops-row__starts-at span {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.contest-ops-row__ends-at span {
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-ops-status-pill {
  --ui-badge-radius: 999px;
  --ui-badge-padding: 0.35rem 0.75rem;
  --ui-badge-size: var(--font-size-0-78);
  --ui-badge-spacing: 0.02em;
  line-height: 1;
}

.contest-ops-status-pill--running {
  --ui-badge-border: color-mix(in srgb, #0ea5e9 30%, transparent);
  --ui-badge-background: color-mix(in srgb, #0ea5e9 12%, var(--journal-surface));
  --ui-badge-color: #0369a1;
}

.contest-ops-status-pill--registering {
  --ui-badge-border: color-mix(in srgb, #f59e0b 34%, transparent);
  --ui-badge-background: color-mix(in srgb, #f59e0b 12%, var(--journal-surface));
  --ui-badge-color: #b45309;
}

.contest-ops-status-pill--draft {
  --ui-badge-border: color-mix(in srgb, #a78bfa 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #a78bfa 10%, var(--journal-surface));
  --ui-badge-color: #6d28d9;
}

.contest-ops-status-pill--frozen {
  --ui-badge-border: color-mix(in srgb, #60a5fa 30%, transparent);
  --ui-badge-background: color-mix(in srgb, #60a5fa 10%, var(--journal-surface));
  --ui-badge-color: #1d4ed8;
}

.contest-ops-status-pill--ended {
  --ui-badge-border: color-mix(in srgb, #34d399 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #34d399 10%, var(--journal-surface));
  --ui-badge-color: #047857;
}

.contest-ops-status-pill--cancelled,
.contest-ops-status-pill--neutral {
  --ui-badge-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  --ui-badge-color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-ops-row__actions {
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .contest-ops-directory {
    --contest-ops-directory-columns: minmax(16rem, 1.2fr) minmax(5.5rem, 0.48fr) minmax(6.5rem, 0.58fr)
      minmax(8.5rem, 0.75fr) minmax(8.5rem, 0.75fr) minmax(11rem, 11rem);
  }
}

@media (max-width: 900px) {
  .contest-ops-directory__head {
    display: none;
  }

  .contest-ops-row {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .contest-ops-row__actions {
    justify-content: flex-start;
  }

  .contest-ops-directory-meta {
    justify-content: flex-start;
  }
}
</style>
