<script setup lang="ts">
import { computed, watch } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import type { ContestFieldLocks, ContestFormDraft } from '@/composables/useAdminContests'

type RequestedContestPanelKey = 'overview' | 'list' | 'create'
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
  createDraft: ContestFormDraft
  createSaving: boolean
  createFieldLocks: ContestFieldLocks
  requestedPanel: RequestedContestPanelKey | null
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
}>()

const router = useRouter()

type ContestPanelKey = 'overview' | 'create'
const {
  activeTab: activePanel,
  selectTab: selectPanel,
} = useUrlSyncedTabs<ContestPanelKey>({
  orderedTabs: ['overview', 'create'],
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
      selectPanel(props.requestedPanel === 'create' ? 'create' : 'overview')
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

function openContestWorkbench(contest: ContestDetailData) {
  void router.push({
    name: 'ContestEdit',
    params: { id: contest.id },
    query: { panel: 'operations', opsPanel: 'inspector' },
  })
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <section
        id="contest-panel-overview"
        v-show="activePanel === 'overview'"
        class="contest-panel contest-panel--workspace"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <header class="list-heading contest-overview-head">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Contest Workspace</div>
            <h1 class="workspace-page-title">竞赛目录</h1>
            <p class="workspace-page-copy">
              上面直接查看关键赛事指标，下面围绕具体竞赛对象完成筛选、编辑、导出和进入攻防运维。
            </p>
          </div>

          <div class="ui-toolbar-actions contest-panel-actions">
            <button type="button" class="ui-btn ui-btn--ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button
              id="contest-open-create"
              type="button"
              class="ui-btn ui-btn--primary"
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
              当前页已接入攻防运维链路的 AWD 赛事
            </div>
          </div>
        </div>

        <section class="workspace-directory-section contest-directory-section">
          <header class="list-heading">
            <div>
              <div class="journal-note-label">Contest Directory</div>
              <h2 class="list-heading__title">竞赛列表</h2>
            </div>
            <div class="contest-section-meta">当前页 {{ listCount }} 场赛事</div>
          </header>
          <WorkspaceDirectoryToolbar
            model-value=""
            selected-sort-label=""
            :sort-options="[]"
            :total="total"
            :show-search="false"
            filter-panel-title="赛事筛选"
            total-suffix="场赛事"
            reset-label="清空筛选"
            :reset-disabled="!hasStatusFilter"
            @reset-filters="emit('updateStatusFilter', 'all')"
          >
            <template #filter-panel>
              <div class="contest-filter-stack">
                <label class="ui-field contest-filter-field">
                  <span class="ui-field__label contest-filter-label">状态筛选</span>
                  <span class="ui-control-wrap">
                    <select
                      :value="statusFilter"
                      class="ui-control contest-filter-control"
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
                  </span>
                </label>
              </div>
            </template>
          </WorkspaceDirectoryToolbar>

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
              <button type="button" class="ui-btn ui-btn--primary" @click="openCreatePanel">
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
            @workbench="openContestWorkbench"
            @change-page="emit('changePage', $event)"
          />
        </section>
      </section>

      <section
        id="contest-panel-create"
        v-show="activePanel === 'create'"
        class="contest-panel contest-panel--create"
        :aria-hidden="activePanel === 'create' ? 'false' : 'true'"
      >
        <section class="workspace-directory-section contest-create-panel">
          <header class="list-heading contest-create-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Contest Setup</div>
              <h2 class="workspace-page-title">创建竞赛</h2>
              <p class="workspace-page-copy">
                在当前工作区里补齐竞赛基础信息和时间窗口，保存后直接回到赛事工作台继续编排。
              </p>
            </div>

            <div class="ui-toolbar-actions contest-panel-actions">
              <button type="button" class="ui-btn ui-btn--ghost" @click="selectPanel('overview')">
                返回工作台
              </button>
            </div>
          </header>

          <AdminContestFormPanel
            :mode="'create'"
            :draft="createDraft"
            :saving="createSaving"
            :field-locks="createFieldLocks"
            :show-cancel="true"
            :note="'创建后可继续在赛事工作台中筛选目录、编辑详情或进入具体 AWD 赛区。'"
            @cancel="selectPanel('overview')"
            @save="emit('saveCreateContest', $event)"
          />
        </section>
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
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
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
  display: grid;
  gap: var(--space-5);
}

.contest-overview-head {
  display: grid;
  gap: var(--space-4);
}

.contest-panel-actions {
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
  margin: 0.35rem 0 0;
  font-size: clamp(1.2rem, 1rem + 0.5vw, 1.45rem);
  font-weight: 700;
  line-height: 1.15;
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

.contest-directory-section,
.contest-create-panel {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-5-5);
}

.contest-directory-section :deep(.workspace-directory-toolbar) {
  margin-bottom: 0;
}

.contest-create-head {
  align-items: flex-start;
}

.contest-filter-stack {
  display: grid;
  gap: var(--space-3);
}

.contest-filter-field {
  display: grid;
  gap: var(--space-2);
}

.contest-filter-label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-filter-control {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 0.95rem;
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: 0 var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.contest-filter-control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
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

  .contest-directory-section,
  .contest-create-panel {
    padding: var(--space-4-5) var(--space-4);
  }

  .contest-panel-actions {
    align-items: stretch;
  }
}

@media (max-width: 640px) {
  .contest-overview-summary {
    --admin-summary-grid-columns: 1fr;
  }
}
</style>
