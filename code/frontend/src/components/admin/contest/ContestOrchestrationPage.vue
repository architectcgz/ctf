<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { Trophy } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'
import AWDOperationsPanel from '@/components/admin/contest/AWDOperationsPanel.vue'

type StatusFilter =
  | 'all'
  | Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>

const props = defineProps<{
  list: ContestDetailData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  statusFilter: StatusFilter
  awdContests: ContestDetailData[]
  selectedAwdContestId: string | null
}>()

const emit = defineEmits<{
  refresh: []
  openCreateDialog: []
  updateStatusFilter: [value: StatusFilter]
  openEditDialog: [contest: ContestDetailData]
  exportContest: [contest: ContestDetailData]
  changePage: [page: number]
  'update:selectedAwdContestId': [value: string]
}>()

const panelTabs = [
  {
    key: 'overview',
    label: '主窗口',
    tabId: 'contest-tab-overview',
    panelId: 'contest-panel-overview',
  },
  {
    key: 'list',
    label: '赛事列表',
    tabId: 'contest-tab-list',
    panelId: 'contest-panel-list',
  },
  {
    key: 'operations',
    label: 'AWD 运维视图',
    tabId: 'contest-tab-operations',
    panelId: 'contest-panel-operations',
  },
] as const

type ContestPanelKey = (typeof panelTabs)[number]['key']

const contestPanelSet = new Set<ContestPanelKey>(panelTabs.map((tab) => tab.key))

function resolvePanelFromLocation(): ContestPanelKey {
  if (typeof window === 'undefined') return 'overview'
  if (!window.location.pathname || window.location.pathname === '/') return 'overview'
  const panel = new URLSearchParams(window.location.search).get('panel')
  if (panel && contestPanelSet.has(panel as ContestPanelKey)) {
    return panel as ContestPanelKey
  }
  return 'overview'
}

function syncPanelToLocation(panelKey: ContestPanelKey): void {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  url.searchParams.set('panel', panelKey)
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`)
}

const activePanel = ref<ContestPanelKey>(resolvePanelFromLocation())
const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])

const registeringCount = computed(
  () => props.list.filter((item) => item.status === 'registering').length
)
const runningCount = computed(() => props.list.filter((item) => item.status === 'running').length)
const awdCount = computed(() => props.awdContests.length)
const listCount = computed(() => props.list.length)

function setTabButtonRef(index: number, element: HTMLButtonElement | null): void {
  tabButtonRefs.value[index] = element
}

function selectPanel(panelKey: ContestPanelKey): void {
  if (activePanel.value === panelKey) return
  activePanel.value = panelKey
  syncPanelToLocation(panelKey)
}

function focusTabByIndex(index: number): void {
  nextTick(() => {
    tabButtonRefs.value[index]?.focus()
  })
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return

  event.preventDefault()

  if (event.key === 'Home') {
    selectPanel(panelTabs[0].key)
    focusTabByIndex(0)
    return
  }

  if (event.key === 'End') {
    const endIndex = panelTabs.length - 1
    selectPanel(panelTabs[endIndex].key)
    focusTabByIndex(endIndex)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + panelTabs.length) % panelTabs.length
  selectPanel(panelTabs[nextIndex].key)
  focusTabByIndex(nextIndex)
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero journal-eyebrow-text workspace-shell flex min-h-full flex-1 flex-col"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Contest Workspace</span>
        <span class="class-chip">赛事管理</span>
      </div>
      <div class="top-note">
        <span>当前页 {{ listCount }} 场</span>
        <span>进行中 {{ runningCount }} 场</span>
        <span>AWD {{ awdCount }} 场</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="赛事管理视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(index, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="selectPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
      <section
        id="contest-panel-overview"
        class="tab-panel contest-panel"
        :class="{ active: activePanel === 'overview' }"
        role="tabpanel"
        aria-labelledby="contest-tab-overview"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <header class="panel-head panel-head--overview">
          <div class="panel-copy workspace-tab-heading__main">
            <div class="journal-eyebrow">Contest Orchestration</div>
            <h1 class="hero-title workspace-tab-heading__title">赛事编排台</h1>
            <p class="admin-page-copy">在这里查看赛事窗口、状态筛选和 AWD 运维入口。</p>
          </div>

          <article class="journal-brief panel-brief rounded-[24px] border px-5 py-5">
            <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
              <Trophy class="h-5 w-5 text-[var(--journal-accent)]" />
              当前赛事概况
            </div>
            <div class="mt-5 grid gap-3 sm:grid-cols-2">
              <div class="journal-note">
                <div class="journal-note-label">赛事总量</div>
                <div class="journal-note-value">{{ total }}</div>
                <div class="journal-note-helper">当前筛选条件下的赛事总数</div>
              </div>
              <div class="journal-note">
                <div class="journal-note-label">报名中</div>
                <div class="journal-note-value">{{ registeringCount }}</div>
                <div class="journal-note-helper">当前页开放报名的赛事</div>
              </div>
              <div class="journal-note">
                <div class="journal-note-label">进行中</div>
                <div class="journal-note-value">{{ runningCount }}</div>
                <div class="journal-note-helper">当前页正在进行的赛事</div>
              </div>
              <div class="journal-note">
                <div class="journal-note-label">AWD</div>
                <div class="journal-note-value">{{ awdCount }}</div>
                <div class="journal-note-helper">当前页可直接切换到攻防运维视图</div>
              </div>
            </div>
          </article>
        </header>

      </section>

      <section
        id="contest-panel-list"
        class="tab-panel contest-panel"
        :class="{ active: activePanel === 'list' }"
        role="tabpanel"
        aria-labelledby="contest-tab-list"
        :aria-hidden="activePanel === 'list' ? 'false' : 'true'"
      >
        <section class="contest-list-filters">
          <label class="space-y-2">
            <span class="text-sm text-[var(--journal-muted)]">状态筛选</span>
            <select
              :value="statusFilter"
              class="admin-input"
              @change="
                emit(
                  'updateStatusFilter',
                  ($event.target as HTMLSelectElement).value as StatusFilter
                )
              "
            >
              <option value="all">全部状态</option>
              <option value="draft">草稿</option>
              <option value="registering">报名中</option>
              <option value="running">进行中</option>
              <option value="frozen">已冻结</option>
              <option value="ended">已结束</option>
            </select>
          </label>
        </section>

        <section class="workspace-directory-section contest-list-panel">
          <header class="list-heading workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="journal-note-label">Contests</div>
              <h3 class="list-heading__title workspace-tab-heading__title">当前筛选结果</h3>
            </div>
          </header>

          <div v-if="loading && list.length === 0" class="flex justify-center py-10">
            <AppLoading>正在同步竞赛列表...</AppLoading>
          </div>

          <AppEmpty
            v-else-if="list.length === 0"
            class="contest-empty-state"
            title="暂无竞赛"
            description="当前筛选条件下没有竞赛数据。"
            icon="Trophy"
          >
            <template #action>
              <button
                type="button"
                class="admin-btn admin-btn-primary"
                @click="emit('openCreateDialog')"
              >
                创建第一场竞赛
              </button>
            </template>
          </AppEmpty>

          <AdminContestTable
            v-else
            :contests="list"
            :page="page"
            :page-size="pageSize"
            :total="total"
            @edit="emit('openEditDialog', $event)"
            @export="emit('exportContest', $event)"
            @change-page="emit('changePage', $event)"
          />
        </section>
      </section>

      <section
        id="contest-panel-operations"
        class="tab-panel contest-panel"
        :class="{ active: activePanel === 'operations' }"
        role="tabpanel"
        aria-labelledby="contest-tab-operations"
        :aria-hidden="activePanel === 'operations' ? 'false' : 'true'"
      >
        <AWDOperationsPanel
          :contests="awdContests"
          :selected-contest-id="selectedAwdContestId"
          @update:selected-contest-id="emit('update:selectedAwdContestId', $event)"
        />
      </section>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --workspace-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-shell-bg: var(--journal-surface);
  --workspace-shell-shadow: 0 22px 50px var(--color-shadow-soft);
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-faint: var(--journal-muted);
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --journal-note-label-weight: 600;
  --journal-note-label-spacing: 0.15em;
  --journal-note-label-color: var(--journal-muted);
  --journal-divider-border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 1.5rem;
}

.contest-panel {
  gap: 1rem;
}

.workspace-shell .tab-panel.contest-panel {
  display: none;
}

.workspace-shell .tab-panel.contest-panel.active {
  display: grid;
  gap: 1rem;
}

.contest-list-panel {
  display: grid;
  gap: 1rem;
}

.panel-head {
  display: grid;
  gap: 1.5rem;
}

.panel-head--overview {
  grid-template-columns: minmax(0, 1.08fr) minmax(18rem, 0.92fr);
  align-items: start;
}

.panel-copy {
  max-width: 42rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.panel-title {
  margin: 0.35rem 0 0;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.panel-copy > p {
  max-width: 48rem;
}

.panel-copy > p {
  margin-top: 0.85rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.contest-list-filters {
  display: grid;
  max-width: 20rem;
  gap: 0.75rem;
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-section-head-intro {
  position: relative;
  padding: 1rem 1.1rem 1rem 1.35rem;
  border: 1px dashed color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 18px;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 10%, transparent),
    transparent 72%
  );
}

.admin-section-head-intro::before {
  content: '';
  position: absolute;
  left: 0.82rem;
  top: 0.95rem;
  bottom: 0.95rem;
  width: 3px;
  border-radius: 999px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink)),
    color-mix(in srgb, var(--journal-accent) 20%, transparent)
  );
}

.admin-section-head-intro .journal-note-label {
  color: var(--journal-accent);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.8rem;
}

.list-heading__title:not(.workspace-tab-heading__title) {
  margin: 0.3rem 0 0;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-primary:hover {
  background: var(--color-primary-hover);
}

.admin-btn-ghost {
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--admin-control-border);
  background: var(--journal-surface);
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
}

.admin-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.contest-empty-state {
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 68%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 68%, transparent);
}

:global([data-theme='dark']) .admin-section-head-intro {
  border-color: color-mix(in srgb, var(--journal-accent) 22%, var(--journal-border));
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 14%, transparent),
    transparent 72%
  );
}

@media (max-width: 767px) {
  .content-pane {
    gap: 1.25rem;
    padding: 1.25rem 1rem 1.5rem;
  }

  .panel-head--overview {
    grid-template-columns: 1fr;
  }
}
</style>
