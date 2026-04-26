<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  Activity,
  Layers,
  Plus,
  RefreshCw,
  Trophy,
  Users,
} from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import PlatformContestFormPanel from '@/components/platform/contest/PlatformContestFormPanel.vue'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import PlatformContestTable from '@/components/platform/contest/PlatformContestTable.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import type { ContestFieldLocks, ContestFormDraft } from '@/composables/usePlatformContests'

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
  announce: [contest: ContestDetailData]
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
    name: 'ContestOperations',
    params: { id: contest.id },
  })
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <section
        v-show="activePanel === 'overview'"
        id="contest-panel-overview"
        class="contest-panel contest-panel--workspace"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <header class="list-heading contest-overview-head">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Contest Workspace
            </div>
            <h1 class="workspace-page-title">
              竞赛目录
            </h1>
          </div>

          <div class="ui-toolbar-actions contest-panel-actions">
            <button
              type="button"
              class="ui-btn ui-btn--ghost"
              @click="emit('refresh')"
            >
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

        <div class="admin-summary-grid contest-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">
              <span>赛事总量</span>
              <Trophy class="h-4 w-4" />
            </div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ total.toString().padStart(2, '0') }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前条件下的赛事总数
            </div>
          </article>

          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">
              <span>报名中</span>
              <Users class="h-4 w-4" />
            </div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ registeringCount.toString().padStart(2, '0') }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页开放报名的赛事
            </div>
          </article>

          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">
              <span>进行中</span>
              <Activity class="h-4 w-4" />
            </div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ runningCount.toString().padStart(2, '0') }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页正在进行的赛事
            </div>
          </article>

          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">
              <span>AWD 模式</span>
              <Layers class="h-4 w-4" />
            </div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ awdCount.toString().padStart(2, '0') }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              已接入运维链路的赛事
            </div>
          </article>
        </div>

        <section class="workspace-directory-section contest-directory-section">
          <header class="list-heading">
            <div>
              <div class="journal-note-label">
                Contest Directory
              </div>
              <h2 class="list-heading__title">
                竞赛列表
              </h2>
            </div>
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
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="openCreatePanel"
              >
                创建第一场竞赛
              </button>
            </template>
          </AppEmpty>

          <PlatformContestTable
            v-else
            :contests="list"
            :page="page"
            :page-size="pageSize"
            :total="total"
            @edit="openEditContest"
            @announce="emit('announce', $event)"
            @export="emit('exportContest', $event)"
            @workbench="openContestWorkbench"
            @change-page="emit('changePage', $event)"
          />
        </section>
      </section>

      <section
        v-show="activePanel === 'create'"
        id="contest-panel-create"
        class="contest-panel contest-panel--create"
        :aria-hidden="activePanel === 'create' ? 'false' : 'true'"
      >
        <section class="workspace-directory-section contest-create-panel">
          <header class="list-heading contest-create-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">
                Contest Setup
              </div>
              <h2 class="workspace-page-title">
                创建竞赛
              </h2>
              <p class="workspace-page-copy">
                在当前工作区里补齐竞赛基础信息和时间窗口，保存后直接回到赛事工作台继续编排。
              </p>
            </div>

            <div class="ui-toolbar-actions contest-panel-actions">
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="selectPanel('overview')"
              >
                返回工作台
              </button>
            </div>
          </header>

          <PlatformContestFormPanel
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
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: var(--color-primary-hover);
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 34%, var(--journal-border));
  --ui-btn-primary-hover-border: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  --ui-btn-primary-hover-shadow: 0 10px 24px
    color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-secondary-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-secondary-color: color-mix(in srgb, var(--journal-muted) 86%, var(--journal-ink));
  --ui-btn-secondary-border: var(--admin-control-border);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 5%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-border: color-mix(
    in srgb,
    var(--journal-accent) 22%,
    var(--admin-control-border)
  );
  --ui-btn-secondary-hover-color: var(--journal-ink);
  --ui-btn-ghost-color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 7%, transparent);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  --action-menu-accent: var(--journal-accent);
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
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.contest-panel-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.contest-directory-section,
.contest-create-panel {
  --workspace-directory-section-padding: var(--space-5) var(--space-5-5);
  background: transparent;
  border: none;
}

.contest-overview-summary.metric-panel-default-surface.metric-panel-workspace-surface {
  --metric-panel-columns: 4;
  --metric-panel-border: color-mix(in srgb, var(--workspace-brand) 16%, var(--workspace-line-soft));
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

@media (max-width: 767px) {
  .content-pane {
    gap: var(--space-5);
    padding: var(--space-5) var(--space-4) var(--space-6);
  }

  .contest-directory-section,
  .contest-create-panel {
    --workspace-directory-section-padding: var(--space-4-5) var(--space-4);
  }

  .contest-panel-actions {
    align-items: stretch;
  }
}
</style>
