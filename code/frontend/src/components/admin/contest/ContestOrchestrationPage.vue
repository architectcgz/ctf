<script setup lang="ts">
import { computed, watch } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'
import AWDOperationsPanel from '@/components/admin/contest/AWDOperationsPanel.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import type { ContestFieldLocks, ContestFormDraft } from '@/composables/useAdminContests'

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
  createDraft: ContestFormDraft
  createSaving: boolean
  createFieldLocks: ContestFieldLocks
  requestedPanel: ContestPanelKey | null
  requestedPanelVersion: number
}>()

const emit = defineEmits<{
  refresh: []
  prepareCreateContest: []
  saveCreateContest: [value: ContestFormDraft]
  updateStatusFilter: [value: StatusFilter]
  openEditDialog: [contest: ContestDetailData]
  exportContest: [contest: ContestDetailData]
  changePage: [page: number]
  'update:selectedAwdContestId': [value: string]
}>()

const panelTabs = [
  {
    key: 'overview',
    label: '总览',
    tabId: 'contest-tab-overview',
    panelId: 'contest-panel-overview',
  },
  {
    key: 'list',
    label: '赛事目录',
    tabId: 'contest-tab-list',
    panelId: 'contest-panel-list',
  },
  {
    key: 'create',
    label: '创建竞赛',
    tabId: 'contest-tab-create',
    panelId: 'contest-panel-create',
  },
  {
    key: 'operations',
    label: 'AWD 运维',
    tabId: 'contest-tab-operations',
    panelId: 'contest-panel-operations',
  },
] as const

const router = useRouter()

type ContestPanelKey = (typeof panelTabs)[number]['key']
const contestPanelOrder = panelTabs.map((tab) => tab.key) as ContestPanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: selectPanel,
  handleTabKeydown,
} = useUrlSyncedTabs<ContestPanelKey>({
  orderedTabs: contestPanelOrder,
  defaultTab: 'overview',
})

const registeringCount = computed(
  () => props.list.filter((item) => item.status === 'registering').length
)
const runningCount = computed(() => props.list.filter((item) => item.status === 'running').length)
const awdCount = computed(() => props.awdContests.length)
const listCount = computed(() => props.list.length)
const hasStatusFilter = computed(() => props.statusFilter !== 'all')

watch(
  () => props.requestedPanelVersion,
  () => {
    if (props.requestedPanel) {
      selectPanel(props.requestedPanel)
    }
  }
)

function openCreatePanel() {
  emit('prepareCreateContest')
  selectPanel('create')
}

function openEditContest(contest: ContestDetailData) {
  void router.push({ name: 'ContestEdit', params: { id: contest.id } })
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero journal-eyebrow-text workspace-shell flex min-h-full flex-1 flex-col"
  >
    <nav class="top-tabs" role="tablist" aria-label="赛事管理视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="tab.key === 'create' ? openCreatePanel() : selectPanel(tab.key)"
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
        <header class="contest-overview-head">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Contest Workspace</div>
            <h1 class="workspace-page-title">赛事管理台</h1>
            <p class="workspace-page-copy">
              在同一套工作区里查看赛事窗口、切换目录筛选，并按需进入 AWD 运维视图。
            </p>
          </div>

          <div class="contest-panel-actions">
            <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button
              type="button"
              class="admin-btn admin-btn-primary"
              @click="openCreatePanel"
            >
              <Plus class="h-4 w-4" />
              创建竞赛
            </button>
          </div>
        </header>

        <div
          class="admin-summary-grid contest-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
        >
          <div class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">赛事总量</div>
            <div class="journal-note-value progress-card-value metric-panel-value">{{ total }}</div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前筛选条件下的赛事总数
            </div>
          </div>
          <div class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">报名中</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ registeringCount }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页开放报名的赛事
            </div>
          </div>
          <div class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">进行中</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ runningCount }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页正在进行的赛事
            </div>
          </div>
          <div class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">AWD</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ awdCount }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页可直接切换到攻防运维视图
            </div>
          </div>
        </div>

        <section class="workspace-directory-section contest-overview-section">
          <header class="list-heading">
            <div>
              <div class="journal-note-label">Contest Status</div>
              <h2 class="list-heading__title">当前赛事窗口</h2>
            </div>
            <div class="contest-section-meta">当前页 {{ listCount }} 场赛事</div>
          </header>

          <div class="contest-overview-rows">
            <article class="contest-overview-row">
              <div class="contest-overview-row__body">
                <h3 class="contest-overview-row__title">报名与开赛窗口</h3>
                <p class="contest-overview-row__copy">
                  当前页有 {{ registeringCount }} 场赛事开放报名，{{ runningCount }}
                  场赛事正在进行。
                </p>
              </div>
              <button type="button" class="contest-inline-link" @click="selectPanel('list')">
                查看赛事目录
              </button>
            </article>

            <article class="contest-overview-row">
              <div class="contest-overview-row__body">
                <h3 class="contest-overview-row__title">AWD 运维入口</h3>
                <p class="contest-overview-row__copy">
                  AWD 赛事会在这里汇总，便于直接进入攻防轮次、服务巡检和流量排查流程。
                </p>
              </div>
              <button type="button" class="contest-inline-link" @click="selectPanel('operations')">
                进入 AWD 运维
              </button>
            </article>
          </div>
        </section>
      </section>

      <section
        id="contest-panel-list"
        class="tab-panel contest-panel"
        :class="{ active: activePanel === 'list' }"
        role="tabpanel"
        aria-labelledby="contest-tab-list"
        :aria-hidden="activePanel === 'list' ? 'false' : 'true'"
      >
        <header class="list-heading contest-list-head">
          <div>
            <div class="workspace-overline">Contest Directory</div>
            <h2 class="list-heading__title">赛事目录</h2>
          </div>

          <div class="contest-list-actions">
            <div class="contest-section-meta">共 {{ total }} 场赛事</div>
            <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button
              type="button"
              class="admin-btn admin-btn-primary"
              @click="openCreatePanel"
            >
              <Plus class="h-4 w-4" />
              创建竞赛
            </button>
          </div>
        </header>

        <section class="workspace-directory-section contest-list-panel">
          <section class="contest-filter-strip" aria-label="赛事筛选">
            <div class="contest-filter-grid">
              <label class="contest-filter-field">
                <span class="contest-filter-label">状态筛选</span>
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

              <div class="contest-filter-field contest-filter-field--action">
                <span class="contest-filter-label contest-filter-label--ghost" aria-hidden="true">
                  操作
                </span>
                <button
                  type="button"
                  class="admin-btn admin-btn-ghost contest-filter-clear"
                  :disabled="!hasStatusFilter"
                  @click="emit('updateStatusFilter', 'all')"
                >
                  清空筛选
                </button>
              </div>
            </div>
          </section>

          <div
            v-if="loading && list.length === 0"
            class="workspace-directory-loading flex justify-center py-10"
          >
            <AppLoading>正在同步竞赛列表...</AppLoading>
          </div>

          <AppEmpty
            v-else-if="list.length === 0"
            class="workspace-directory-empty contest-empty-state"
            title="暂无竞赛"
            description="当前筛选条件下没有竞赛数据。"
            icon="Trophy"
          >
            <template #action>
              <button
                type="button"
                class="admin-btn admin-btn-primary"
                @click="openCreatePanel"
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
            @edit="openEditContest"
            @export="emit('exportContest', $event)"
            @change-page="emit('changePage', $event)"
          />
        </section>
      </section>

      <section
        id="contest-panel-create"
        class="tab-panel contest-panel"
        :class="{ active: activePanel === 'create' }"
        role="tabpanel"
        aria-labelledby="contest-tab-create"
        :aria-hidden="activePanel === 'create' ? 'false' : 'true'"
      >
        <section class="workspace-directory-section contest-create-panel">
          <header class="contest-overview-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Contest Setup</div>
              <h2 class="workspace-page-title">创建竞赛</h2>
              <p class="workspace-page-copy">
                在当前工作区里补齐竞赛基础信息和时间窗口，保存后直接回到赛事目录继续编排。
              </p>
            </div>
          </header>

          <AdminContestFormPanel
            :mode="'create'"
            :draft="createDraft"
            :saving="createSaving"
            :field-locks="createFieldLocks"
            :show-cancel="false"
            :note="'创建后可继续在赛事目录中编辑详情、挂载题目或进入 AWD 运维面板。'"
            @save="emit('saveCreateContest', $event)"
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
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-shadow-panel: 0 14px 34px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
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
  gap: var(--space-6);
}

.contest-panel {
  gap: var(--space-5);
}

.workspace-shell .tab-panel.contest-panel {
  display: none;
}

.workspace-shell .tab-panel.contest-panel.active {
  display: grid;
  gap: var(--space-5);
}

.contest-overview-head {
  display: grid;
  gap: var(--space-4);
}

.contest-panel-actions,
.contest-list-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
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

.contest-section-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-overview-summary {
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
}

.contest-overview-summary.metric-panel-default-surface.metric-panel-workspace-surface {
  --metric-panel-border: color-mix(in srgb, var(--workspace-brand) 16%, var(--workspace-line-soft));
  --metric-panel-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--workspace-brand) 16%, transparent),
      transparent 42%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--workspace-panel) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--workspace-panel-soft) 82%, transparent)
    );
  --metric-panel-shadow: var(--workspace-shadow-panel);
}

.contest-overview-section,
.contest-list-panel,
.contest-create-panel {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-5-5);
}

.contest-overview-rows {
  display: grid;
  gap: var(--space-3);
}

.contest-overview-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 68%, transparent);
  padding-top: var(--space-4);
}

.contest-overview-row:first-child {
  border-top: none;
  padding-top: 0;
}

.contest-overview-row__body {
  display: grid;
  gap: var(--space-1-5);
  max-width: 42rem;
}

.contest-overview-row__title {
  margin: 0;
  font-size: var(--font-size-0-98);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-overview-row__copy {
  margin: 0;
  line-height: 1.7;
  color: var(--journal-muted);
}

.contest-inline-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease;
}

.contest-inline-link:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 36%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 14%, var(--journal-surface));
}

.contest-list-head {
  align-items: flex-end;
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
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
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
}

.admin-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.contest-filter-strip {
  display: grid;
  gap: var(--space-3);
}

.contest-filter-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(4, minmax(0, 1fr));
  align-items: end;
}

.contest-filter-field {
  display: grid;
  gap: var(--space-2);
}

.contest-filter-field--action {
  justify-self: end;
  min-width: 0;
}

.contest-filter-label {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.contest-filter-label--ghost {
  opacity: 0;
}

.contest-filter-clear {
  min-width: 7rem;
}

.contest-empty-state {
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 68%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 68%, transparent);
}

@media (max-width: 1023px) {
  .contest-overview-summary {
    --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .content-pane {
    gap: var(--space-5);
    padding: var(--space-5) var(--space-4) var(--space-6);
  }

  .contest-overview-section,
  .contest-list-panel,
  .contest-create-panel {
    padding: var(--space-4-5) var(--space-4);
  }

  .contest-overview-row,
  .contest-panel-actions,
  .contest-list-actions,
  .contest-filter-grid {
    align-items: stretch;
  }

  .contest-filter-field,
  .contest-filter-field--action,
  .contest-inline-link {
    width: 100%;
    max-width: none;
  }

  .contest-filter-grid {
    grid-template-columns: 1fr;
  }

  .contest-filter-field--action {
    justify-self: stretch;
  }
}

@media (max-width: 640px) {
  .contest-overview-summary {
    --admin-summary-grid-columns: 1fr;
  }
}
</style>
