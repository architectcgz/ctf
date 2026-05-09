<script setup lang="ts">
import type { ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { formatDate } from '@/utils/format'

interface PlatformAwdReviewRow {
  id: string
  title: string
  status: string
  current_round?: number
  round_count: number
  team_count: number
  mode: string
  export_ready: boolean
  latest_evidence_at?: string | null
  contestCode: string
}

type AwdReviewStatusFilter = '' | ContestStatus

interface Props {
  loading: boolean
  error: string | null
  rows: PlatformAwdReviewRow[]
  total: number
  page: number
  totalPages: number
  hasContests: boolean
  keyword: string
  statusFilter: AwdReviewStatusFilter
  hasActiveFilters: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'update:statusFilter': [value: AwdReviewStatusFilter]
  'reset-filters': []
  retry: []
  'open-contest': [contestId: string]
  'change-page': [page: number]
}>()

const reviewTableColumns = [
  {
    key: 'contestCode',
    label: '代号',
    widthClass: 'w-[16%] min-w-[8rem]',
    cellClass: 'admin-awd-review-table__mono',
  },
  {
    key: 'title',
    label: '赛事',
    widthClass: 'w-[33%] min-w-[16rem]',
    cellClass: 'admin-awd-review-table__title',
  },
  {
    key: 'rounds',
    label: '轮次',
    widthClass: 'w-[14%] min-w-[7rem]',
    cellClass: 'admin-awd-review-table__meta',
  },
  {
    key: 'teams',
    label: '队伍',
    widthClass: 'w-[13%] min-w-[7rem]',
    cellClass: 'admin-awd-review-table__meta',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[14%] min-w-[7rem]',
    cellClass: 'admin-awd-review-table__status',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
    cellClass: 'admin-awd-review-table__actions',
  },
]

function contestStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'ended':
      return '已结束'
    case 'frozen':
      return '冻结中'
    case 'published':
      return '已发布'
    default:
      return status || '未开始'
  }
}

function formatEvidenceAt(value?: string | null): string {
  return value ? formatDate(value) : '暂无'
}

function updateStatusFilter(event: Event): void {
  emit('update:statusFilter', (event.target as HTMLSelectElement).value as AwdReviewStatusFilter)
}
</script>

<template>
  <section class="workspace-directory-section admin-awd-review-directory">
    <header class="list-heading">
      <div>
        <div class="workspace-overline">Review Directory</div>
        <h2 class="list-heading__title">赛事目录</h2>
      </div>
      <div class="admin-awd-review-directory__meta">共 {{ total }} 场赛事</div>
    </header>

    <WorkspaceDirectoryToolbar
      :model-value="keyword"
      :total="total"
      selected-sort-label=""
      :sort-options="[]"
      search-placeholder="搜索赛事标题"
      filter-panel-title="赛事筛选"
      total-suffix="场赛事"
      :reset-disabled="!hasActiveFilters"
      @update:model-value="emit('update:keyword', $event)"
      @reset-filters="emit('reset-filters')"
    >
      <template #filter-panel>
        <div class="admin-awd-review-filter-grid">
          <label class="admin-awd-review-filter-field">
            <span class="admin-awd-review-filter-field__label">赛事状态</span>
            <select
              :value="statusFilter"
              class="workspace-directory-filter-control admin-awd-review-filter-field__control"
              @change="updateStatusFilter"
            >
              <option value="">全部状态</option>
              <option value="running">进行中</option>
              <option value="ended">已结束</option>
              <option value="frozen">冻结中</option>
            </select>
          </label>
        </div>
      </template>
    </WorkspaceDirectoryToolbar>

    <div v-if="loading" class="workspace-directory-loading">正在同步 AWD 复盘目录...</div>

    <AppEmpty
      v-else-if="error"
      class="workspace-directory-empty"
      icon="AlertTriangle"
      title="AWD复盘目录加载失败"
      :description="error"
    >
      <template #action>
        <button type="button" class="ui-btn ui-btn--primary" @click="emit('retry')">
          重新加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="!hasContests"
      class="workspace-directory-empty"
      icon="BookOpen"
      title="暂无 AWD 赛事"
      description="当前筛选条件下没有可进入平台复盘的 AWD 赛事。"
    />

    <template v-else>
      <WorkspaceDataTable
        class="workspace-directory-list admin-awd-review-table"
        :columns="reviewTableColumns"
        :rows="rows"
        row-key="id"
        row-class="admin-awd-review-table__row"
      >
        <template #cell-contestCode="{ row }">
          <span class="admin-awd-review-table__code">
            {{ (row as PlatformAwdReviewRow).contestCode }}
          </span>
        </template>

        <template #cell-title="{ row }">
          <div class="admin-awd-review-table__title-wrap">
            <span
              class="admin-awd-review-table__title-text"
              :title="(row as PlatformAwdReviewRow).title"
            >
              {{ (row as PlatformAwdReviewRow).title }}
            </span>
            <span class="admin-awd-review-table__hint">
              最近信号
              {{ formatEvidenceAt((row as PlatformAwdReviewRow).latest_evidence_at) }}
            </span>
          </div>
        </template>

        <template #cell-rounds="{ row }">
          <div class="admin-awd-review-table__meta-block">
            <span>
              {{
                (row as PlatformAwdReviewRow).current_round
                  ? `第 ${(row as PlatformAwdReviewRow).current_round} 轮`
                  : '未开始'
              }}
            </span>
            <span>共 {{ (row as PlatformAwdReviewRow).round_count }} 轮</span>
          </div>
        </template>

        <template #cell-teams="{ row }">
          <div class="admin-awd-review-table__meta-block">
            <span>{{ (row as PlatformAwdReviewRow).team_count }} 支队伍</span>
            <span>{{ (row as PlatformAwdReviewRow).mode.toUpperCase() }}</span>
          </div>
        </template>

        <template #cell-status="{ row }">
          <div class="admin-awd-review-table__status-wrap">
            <span
              class="admin-awd-review-table__status-pill"
              :class="[
                'workspace-directory-status-pill',
                'workspace-directory-status-pill--primary',
              ]"
            >
              {{ contestStatusLabel((row as PlatformAwdReviewRow).status) }}
            </span>
            <span
              class="admin-awd-review-table__status-pill admin-awd-review-table__status-pill--muted"
              :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--muted']"
            >
              {{ (row as PlatformAwdReviewRow).export_ready ? '可导出' : '实时复盘' }}
            </span>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="workspace-directory-row-actions admin-awd-review-table__actions">
            <button
              type="button"
              class="ui-btn ui-btn--primary ui-btn--xs admin-awd-review-table__action"
              @click="emit('open-contest', (row as PlatformAwdReviewRow).id)"
            >
              进入复盘
            </button>
          </div>
        </template>
      </WorkspaceDataTable>

      <WorkspaceDirectoryPagination
        v-if="total > 0"
        class="admin-pagination"
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :disabled="loading"
        :total-label="`共 ${total} 场赛事`"
        @change-page="emit('change-page', $event)"
      />
    </template>
  </section>
</template>

<style scoped>
.admin-awd-review-directory__meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.admin-awd-review-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.admin-awd-review-filter-field {
  display: grid;
  gap: var(--space-2);
}

.admin-awd-review-filter-field__label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.admin-awd-review-table {
  --workspace-directory-shell-border: var(--awd-review-directory-border);
  --workspace-directory-head-divider: var(--awd-review-directory-border);
  --workspace-directory-row-divider: var(--awd-review-directory-row-divider);
}

.admin-awd-review-table :deep(.workspace-data-table__row:hover) {
  background: color-mix(in srgb, var(--color-primary) 5%, transparent);
}

.admin-awd-review-table__code,
.admin-awd-review-table :deep(.admin-awd-review-table__mono) {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-82);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.admin-awd-review-table__title-wrap,
.admin-awd-review-table__meta-block {
  display: grid;
  gap: var(--space-1);
}

.admin-awd-review-table__title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.98rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-awd-review-table__hint,
.admin-awd-review-table__meta-block {
  font-size: var(--font-size-0-84);
  color: var(--journal-muted);
}

.admin-awd-review-table__status-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

</style>
