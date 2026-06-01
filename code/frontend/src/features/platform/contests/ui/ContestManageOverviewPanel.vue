<template>
  <section
    id="contest-panel-overview"
    class="contest-panel contest-panel--workspace"
    :aria-hidden="active ? 'false' : 'true'"
  >
    <header class="workspace-panel-header contest-overview-head">
      <div class="workspace-panel-header__intro">
        <div class="workspace-overline">
          Contest Workspace
        </div>
        <h1 class="workspace-page-title">竞赛目录</h1>
      </div>

      <div class="workspace-panel-header__actions ui-toolbar-actions contest-panel-actions">
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
          @click="emit('openCreate')"
        >
          <Plus class="h-4 w-4" />
          创建竞赛
        </button>
      </div>
      <div class="workspace-panel-header__summary admin-summary-grid contest-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
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
            当前筛选条件下开放报名的赛事
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
            当前筛选条件下正在进行的赛事
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
            当前页已接入运维链路的赛事
          </div>
        </article>
      </div>
    </header>

    <div class="workspace-panel-divider" aria-hidden="true" />

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
                      ($event.target as HTMLSelectElement).value as ContestManageStatusFilter
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
            @click="emit('openCreate')"
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
        :build-edit-route="buildEditRoute"
        :build-workbench-route="buildWorkbenchRoute"
        @announce="emit('announce', $event)"
        @change-page="emit('changePage', $event)"
      />
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Activity,
  Layers,
  Plus,
  RefreshCw,
  Trophy,
  Users,
} from 'lucide-vue-next'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import WorkspaceDirectoryToolbar from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'
import type { ContestDetailData, ContestListSummaryData } from '@/api/contracts'
import type {
  ContestEditRouteTarget,
  ContestOperationsRouteTarget,
} from '../model'
import PlatformContestTable from './PlatformContestTable.vue'
import type { ContestManageStatusFilter } from './contestOrchestrationPage.types'

const props = defineProps<{
  active: boolean
  list: ContestDetailData[]
  total: number
  summary: ContestListSummaryData
  page: number
  pageSize: number
  loading: boolean
  statusFilter: ContestManageStatusFilter
  awdContests: ContestDetailData[]
  buildEditRoute: (contestId: string) => ContestEditRouteTarget
  buildWorkbenchRoute: (contestId: string) => ContestOperationsRouteTarget
}>()

const emit = defineEmits<{
  refresh: []
  openCreate: []
  updateStatusFilter: [value: ContestManageStatusFilter]
  announce: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const registeringCount = computed(() => props.summary.registering_count)
const runningCount = computed(() => props.summary.running_count)
const awdCount = computed(() => props.awdContests.length)
const hasStatusFilter = computed(() => props.statusFilter !== 'all')
</script>
