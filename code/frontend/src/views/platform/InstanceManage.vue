<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Server,
  Activity,
  AlertTriangle,
  Calendar,
  ArrowDownWideNarrow,
  SortAsc,
} from 'lucide-vue-next'

import type { TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useTeacherInstances } from '@/composables/useTeacherInstances'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

type InstanceSortKey = 'student_name' | 'challenge_title' | 'created_at' | 'remaining_time'
type InstanceSortOption = WorkspaceDirectorySortOption & {
  key: InstanceSortKey
  order: 'asc' | 'desc'
}

interface PlatformInstanceRow extends TeacherInstanceItem {
  displayStudentName: string
  displayStudentCode: string
  displayUrl: string
}

const router = useRouter()
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)

const {
  classes,
  instances,
  filters,
  loadingClasses,
  loadingInstances,
  destroyingId,
  error,
  totalCount,
  runningCount,
  expiringSoonCount,
  initialize,
  updateFilter,
  removeInstance,
} = useTeacherInstances()

const sortOptions: InstanceSortOption[] = [
  { key: 'created_at', order: 'desc', label: '最近创建', icon: Calendar },
  { key: 'remaining_time', order: 'asc', label: '即将到期优先', icon: AlertTriangle },
  { key: 'student_name', order: 'asc', label: '学生姓名 A-Z', icon: SortAsc },
  { key: 'challenge_title', order: 'asc', label: '题目名称 A-Z', icon: ArrowDownWideNarrow },
]
const sortConfig = ref<InstanceSortOption>(sortOptions[0]!)

const hasActiveFilters = computed(
  () => Boolean(filters.className.trim() || filters.keyword.trim() || filters.studentNo.trim())
)
const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / Math.max(pageSize.value, 1)))
)

const instanceRows = computed<PlatformInstanceRow[]>(() =>
  instances.value.map((item) => ({
    ...item,
    displayStudentName: item.student_name || item.student_username,
    displayStudentCode: item.student_no || `@${item.student_username}`,
    displayUrl: item.access_url || '暂未分配访问地址',
  }))
)

const filteredRows = computed<PlatformInstanceRow[]>(() => {
  const nextRows = [...instanceRows.value]
  nextRows.sort((left, right) => {
    switch (sortConfig.value.key) {
      case 'remaining_time':
        return sortConfig.value.order === 'asc'
          ? left.remaining_time - right.remaining_time
          : right.remaining_time - left.remaining_time
      case 'student_name':
        return left.displayStudentName.localeCompare(right.displayStudentName, 'zh-CN')
      case 'challenge_title':
        return left.challenge_title.localeCompare(right.challenge_title, 'zh-CN')
      case 'created_at':
      default: {
        const leftTime = new Date(left.created_at).getTime()
        const rightTime = new Date(right.created_at).getTime()
        return sortConfig.value.order === 'asc' ? leftTime - rightTime : rightTime - leftTime
      }
    }
  })

  return nextRows
})

const paginatedRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredRows.value.slice(start, start + pageSize.value)
})

const instanceTableColumns = [
  {
    key: 'displayStudentCode',
    label: '学生',
    widthClass: 'w-[14%] min-w-[8rem]',
    cellClass: 'admin-instance-manage-table__mono',
  },
  {
    key: 'displayStudentName',
    label: '姓名',
    widthClass: 'w-[14%] min-w-[8rem]',
    cellClass: 'admin-instance-manage-table__student',
  },
  {
    key: 'challenge_title',
    label: '题目',
    widthClass: 'w-[18%] min-w-[11rem]',
    cellClass: 'admin-instance-manage-table__challenge',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'admin-instance-manage-table__status',
  },
  {
    key: 'created_at',
    label: '创建时间',
    widthClass: 'w-[12%] min-w-[8rem]',
    cellClass: 'admin-instance-manage-table__time',
  },
  {
    key: 'expires_at',
    label: '到期时间',
    widthClass: 'w-[12%] min-w-[8rem]',
    cellClass: 'admin-instance-manage-table__time',
  },
  {
    key: 'remaining_time',
    label: '剩余时间',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[7rem]',
    cellClass: 'admin-instance-manage-table__count',
  },
  {
    key: 'displayUrl',
    label: '访问地址',
    widthClass: 'w-[18%] min-w-[12rem]',
    cellClass: 'admin-instance-manage-table__url',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
    cellClass: 'admin-instance-manage-table__actions',
  },
]

function setSort(option: WorkspaceDirectorySortOption): void {
  const matchedOption =
    sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
    sortOptions[0]

  if (!matchedOption) {
    return
  }

  sortConfig.value = matchedOption
}

function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatRemainingTime(seconds: number): string {
  if (seconds <= 0) return '已到期'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const restSeconds = seconds % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(restSeconds).padStart(2, '0')}`
}

function statusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'creating':
      return '创建中'
    case 'expired':
      return '已过期'
    case 'failed':
      return '异常'
    default:
      return status
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'running':
      return 'admin-instance-manage-table__status-pill--running'
    case 'creating':
      return 'admin-instance-manage-table__status-pill--creating'
    case 'expired':
      return 'admin-instance-manage-table__status-pill--expired'
    case 'failed':
      return 'admin-instance-manage-table__status-pill--failed'
    default:
      return ''
  }
}

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }
  page.value = normalizedPage
}

function resetFilters(): void {
  updateFilter('className', '')
  updateFilter('keyword', '')
  updateFilter('studentNo', '')
}

async function handleDestroy(id: string): Promise<void> {
  const confirmed = await confirmDestructiveAction({
    title: '确认销毁实例',
    message: '确定要销毁该实例吗？此操作不可恢复。',
    confirmButtonText: '确认销毁',
    cancelButtonText: '取消',
  })
  if (!confirmed) {
    return
  }
  await removeInstance(id)
}

watch(
  () => [filters.className, filters.keyword, filters.studentNo],
  () => {
    page.value = 1
  }
)

watch(totalCount, () => {
  if (page.value > totalPages.value) {
    page.value = totalPages.value
  }
})

onMounted(() => {
  initialize()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-instance-manage-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane admin-instance-manage-shell__content">
      <header class="admin-instance-manage-shell__hero">
        <div class="admin-instance-manage-shell__hero-main">
          <div class="workspace-overline">
            Instance Workspace
          </div>
          <h1 class="workspace-page-title">
            实例管理
          </h1>
          <p class="workspace-page-copy">
            在后台视角查看实例状态、到期节奏与访问地址，并快速销毁异常环境。
          </p>
        </div>

        <div class="admin-instance-manage-shell__hero-side">
          <button
            type="button"
            class="admin-instance-manage-shell__hero-action"
            @click="router.push({ name: 'PlatformOverview' })"
          >
            返回后台概览
          </button>
        </div>
      </header>

      <div class="metric-panel-grid--premium cols-3">
        <article class="metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>当前可见</span>
            <Server class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ totalCount.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            符合当前筛选条件的实例数量
          </div>
        </article>

        <article class="metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>运行中</span>
            <Activity class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ runningCount.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            仍在占用环境资源的实例数量
          </div>
        </article>

        <article class="metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>即将到期</span>
            <AlertTriangle class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ expiringSoonCount.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            剩余时间不足 10 分钟的实例
          </div>
        </article>
      </div>

      <section class="workspace-directory-section admin-instance-manage-directory">
        <header class="list-heading">
          <div>
            <div class="workspace-overline">
              Instance Directory
            </div>
            <h2 class="list-heading__title">
              实例目录
            </h2>
          </div>
        </header>

        <WorkspaceDirectoryToolbar
          :model-value="filters.keyword"
          :total="totalCount"
          :selected-sort-label="sortConfig.label"
          :sort-options="sortOptions"
          search-placeholder="检索学生、用户名或题目名称..."
          :reset-disabled="!hasActiveFilters"
          total-suffix="条实例"
          @update:model-value="updateFilter('keyword', $event)"
          @select-sort="setSort"
          @reset-filters="resetFilters"
        >
          <template #filter-panel>
            <div class="admin-instance-manage-filter-grid">
              <label class="admin-instance-manage-filter-field">
                <span class="admin-instance-manage-filter-field__label">班级范围</span>
                <select
                  :value="filters.className"
                  class="admin-instance-manage-filter-field__control"
                  :disabled="loadingClasses"
                  @change="updateFilter('className', ($event.target as HTMLSelectElement).value)"
                >
                  <option value="">全部班级</option>
                  <option
                    v-for="item in classes"
                    :key="item.name"
                    :value="item.name"
                  >
                    {{ item.name }} · {{ item.student_count || 0 }}
                  </option>
                </select>
              </label>

              <label class="admin-instance-manage-filter-field">
                <span class="admin-instance-manage-filter-field__label">按学号查询</span>
                <input
                  :value="filters.studentNo"
                  type="text"
                  class="admin-instance-manage-filter-field__control"
                  placeholder="输入学号精确查询"
                  @input="updateFilter('studentNo', ($event.target as HTMLInputElement).value)"
                >
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div
          v-if="loadingInstances"
          class="workspace-directory-loading"
        >
          正在同步实例目录...
        </div>

        <AppEmpty
          v-else-if="instances.length === 0"
          class="workspace-directory-empty"
          icon="Inbox"
          title="当前没有匹配到实例"
          description="调整筛选条件，或等待学员创建新的训练环境后再查看。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-instance-manage-table"
          :columns="instanceTableColumns"
          :rows="paginatedRows"
          row-key="id"
          row-class="admin-instance-manage-table__row"
        >
          <template #cell-displayStudentCode="{ row }">
            <span class="admin-instance-manage-table__code">
              {{ (row as PlatformInstanceRow).displayStudentCode }}
            </span>
          </template>

          <template #cell-displayStudentName="{ row }">
            <div class="admin-instance-manage-table__student-wrap">
              <span
                class="admin-instance-manage-table__student-text"
                :title="(row as PlatformInstanceRow).displayStudentName"
              >
                {{ (row as PlatformInstanceRow).displayStudentName }}
              </span>
            </div>
          </template>

          <template #cell-challenge_title="{ row }">
            <span
              class="admin-instance-manage-table__challenge-text"
              :title="(row as PlatformInstanceRow).challenge_title"
            >
              {{ (row as PlatformInstanceRow).challenge_title }}
            </span>
          </template>

          <template #cell-status="{ row }">
            <span
              class="admin-instance-manage-table__status-pill"
              :class="statusClass((row as PlatformInstanceRow).status)"
            >
              {{ statusLabel((row as PlatformInstanceRow).status) }}
            </span>
          </template>

          <template #cell-created_at="{ row }">
            <span class="admin-instance-manage-table__time-text">
              {{ formatDateTime((row as PlatformInstanceRow).created_at) }}
            </span>
          </template>

          <template #cell-expires_at="{ row }">
            <span class="admin-instance-manage-table__time-text">
              {{ formatDateTime((row as PlatformInstanceRow).expires_at) }}
            </span>
          </template>

          <template #cell-remaining_time="{ row }">
            <span class="admin-instance-manage-table__count-text">
              {{ formatRemainingTime((row as PlatformInstanceRow).remaining_time) }}
            </span>
          </template>

          <template #cell-displayUrl="{ row }">
            <span
              class="admin-instance-manage-table__url-text"
              :title="(row as PlatformInstanceRow).displayUrl"
            >
              {{ (row as PlatformInstanceRow).displayUrl }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <button
              type="button"
              class="admin-instance-manage-table__action admin-instance-manage-table__action--danger"
              :disabled="destroyingId === (row as PlatformInstanceRow).id"
              :data-instance-id="(row as PlatformInstanceRow).id"
              @click="handleDestroy((row as PlatformInstanceRow).id)"
            >
              {{ destroyingId === (row as PlatformInstanceRow).id ? '销毁中...' : '销毁实例' }}
            </button>
          </template>
        </WorkspaceDataTable>

        <WorkspaceDirectoryPagination
          v-if="totalCount > 0"
          :page="page"
          :total-pages="totalPages"
          :total="totalCount"
          :total-label="`共 ${totalCount} 条实例`"
          @change-page="handlePageChange"
        />
      </section>

      <div
        v-if="error"
        class="admin-instance-manage-shell__error"
      >
        {{ error }}
        <button
          type="button"
          class="admin-instance-manage-shell__error-action"
          @click="initialize()"
        >
          重试
        </button>
      </div>
    </main>
  </section>
</template>

<style scoped>
.admin-instance-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --instance-directory-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --instance-directory-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 10%, transparent),
      transparent 20rem
    );
}

.admin-instance-manage-shell__content {
  display: grid;
  gap: var(--workspace-directory-page-block-gap);
}

.admin-instance-manage-shell__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.admin-instance-manage-shell__hero-main {
  max-width: 46rem;
}

.admin-instance-manage-shell__hero-side {
  display: flex;
  align-items: center;
}

.admin-instance-manage-shell__hero-action,
.admin-instance-manage-shell__error-action,
.admin-instance-manage-table__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.4rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 18%, var(--journal-border));
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 8%, var(--journal-surface));
  padding: 0 0.95rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-primary);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.admin-instance-manage-shell__hero-action:hover,
.admin-instance-manage-shell__error-action:hover,
.admin-instance-manage-table__action:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--color-primary) 32%, var(--journal-border));
  background: color-mix(in srgb, var(--color-primary) 12%, var(--journal-surface));
}

.admin-instance-manage-table__action--danger {
  border-color: color-mix(in srgb, var(--color-danger) 20%, var(--journal-border));
  background: color-mix(in srgb, var(--color-danger) 7%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.admin-instance-manage-table__action--danger:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--color-danger) 36%, var(--journal-border));
  background: color-mix(in srgb, var(--color-danger) 11%, var(--journal-surface));
}

.admin-instance-manage-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.admin-instance-manage-filter-field {
  display: grid;
  gap: var(--space-2);
}

.admin-instance-manage-filter-field__label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.admin-instance-manage-filter-field__control {
  min-height: 2.75rem;
  border: 1px solid var(--admin-control-border);
  border-radius: 0.95rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: 0 var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.admin-instance-manage-filter-field__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.admin-instance-manage-table {
  --workspace-directory-shell-border: var(--instance-directory-border);
  --workspace-directory-head-divider: var(--instance-directory-border);
  --workspace-directory-row-divider: var(--instance-directory-row-divider);
}

.admin-instance-manage-table :deep(.workspace-data-table__row:hover) {
  background: color-mix(in srgb, var(--color-primary) 5%, transparent);
}

.admin-instance-manage-table__code,
.admin-instance-manage-table__count-text,
.admin-instance-manage-table__time-text {
  font-family: var(--font-family-mono);
  font-size: 0.84rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-instance-manage-table__code {
  letter-spacing: 0.04em;
  color: var(--journal-muted);
}

.admin-instance-manage-table__student-wrap {
  display: flex;
  min-width: 0;
  align-items: center;
}

.admin-instance-manage-table__student-text,
.admin-instance-manage-table__challenge-text,
.admin-instance-manage-table__url-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-instance-manage-table__student-text,
.admin-instance-manage-table__challenge-text {
  font-size: 0.96rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-instance-manage-table__url-text {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.admin-instance-manage-table__status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2rem;
  min-width: 5rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 75%, transparent);
  border-radius: 999px;
  padding: 0 0.75rem;
  font-size: 0.8rem;
  font-weight: 600;
}

.admin-instance-manage-table__status-pill--running {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 9%, var(--journal-surface));
  color: var(--color-success);
}

.admin-instance-manage-table__status-pill--creating {
  border-color: color-mix(in srgb, var(--color-primary) 28%, transparent);
  background: color-mix(in srgb, var(--color-primary) 9%, var(--journal-surface));
  color: var(--color-primary);
}

.admin-instance-manage-table__status-pill--expired {
  border-color: color-mix(in srgb, var(--color-warning) 28%, transparent);
  background: color-mix(in srgb, var(--color-warning) 9%, var(--journal-surface));
  color: var(--color-warning);
}

.admin-instance-manage-table__status-pill--failed {
  border-color: color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 9%, var(--journal-surface));
  color: var(--color-danger);
}

.admin-instance-manage-shell__error {
  padding: 1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--color-danger) 18%, var(--journal-border));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 8%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 76%, var(--journal-ink));
}
</style>
