<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  FolderKanban,
  SearchCode,
  UsersRound,
  Calendar,
  ArrowDownWideNarrow,
  ArrowUpNarrowWide,
  SortAsc,
} from 'lucide-vue-next'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { resolveClassStudentsRouteName } from '@/utils/teachingWorkspaceRouting'

type ClassRosterFilter = 'all' | 'active' | 'empty'
type ClassSortOption = WorkspaceDirectorySortOption & {
  order: 'asc' | 'desc'
}

interface PlatformClassRow extends TeacherClassItem {
  code: string
  rosterStatus: ClassRosterFilter
}

const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const loading = ref(false)
const error = ref<string | null>(null)

const keyword = ref('')
const rosterFilter = ref<ClassRosterFilter>('all')
const sortOptions: ClassSortOption[] = [
  { key: 'name', order: 'asc', label: '班级名称 A-Z', icon: SortAsc },
  { key: 'student_count', order: 'desc', label: '学生数由高到低', icon: ArrowDownWideNarrow },
  { key: 'student_count', order: 'asc', label: '学生数由低到高', icon: ArrowUpNarrowWide },
  { key: 'code', order: 'asc', label: '班级编号', icon: Calendar },
]
const sortConfig = ref<ClassSortOption>(sortOptions[0]!)

const classTableColumns = [
  {
    key: 'code',
    label: '班级编号',
    widthClass: 'w-[16%] min-w-[7rem]',
    cellClass: 'admin-class-manage-table__mono',
  },
  {
    key: 'name',
    label: '班级名称',
    widthClass: 'w-[36%] min-w-[15rem]',
    cellClass: 'admin-class-manage-table__name',
  },
  {
    key: 'student_count',
    label: '学生数',
    align: 'center' as const,
    widthClass: 'w-[14%] min-w-[6rem]',
    cellClass: 'admin-class-manage-table__count',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center' as const,
    widthClass: 'w-[18%] min-w-[7rem]',
    cellClass: 'admin-class-manage-table__status',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[12rem]',
    cellClass: 'admin-class-manage-table__actions',
  },
]

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
const classRows = computed<PlatformClassRow[]>(() =>
  classes.value.map((item, index) => {
    const offset = (page.value - 1) * pageSize.value + index + 1
    const studentCount = item.student_count || 0

    return {
      ...item,
      code: `CL-${String(offset).padStart(2, '0')}`,
      rosterStatus: studentCount > 0 ? 'active' : 'empty',
    }
  })
)
const summaryStudentCount = computed(() =>
  classRows.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
)
const activeClassCount = computed(
  () => classRows.value.filter((item) => item.rosterStatus === 'active').length
)
const hasActiveFilters = computed(() => Boolean(keyword.value.trim() || rosterFilter.value !== 'all'))
const filteredRows = computed<PlatformClassRow[]>(() => {
  const normalizedKeyword = keyword.value.trim().toLowerCase()
  const nextRows = classRows.value.filter((item) => {
    const matchesKeyword =
      !normalizedKeyword ||
      item.name.toLowerCase().includes(normalizedKeyword) ||
      item.code.toLowerCase().includes(normalizedKeyword)
    const matchesRoster = rosterFilter.value === 'all' || item.rosterStatus === rosterFilter.value

    return matchesKeyword && matchesRoster
  })

  const sortedRows = [...nextRows]
  sortedRows.sort((left, right) => {
    if (sortConfig.value.key === 'student_count') {
      const delta = (left.student_count || 0) - (right.student_count || 0)
      return sortConfig.value.order === 'asc' ? delta : -delta
    }
    if (sortConfig.value.key === 'code') {
      return left.code.localeCompare(right.code, 'zh-CN', { numeric: true })
    }
    return left.name.localeCompare(right.name, 'zh-CN')
  })
  return sortedRows
})

async function loadClasses(nextPage = page.value): Promise<void> {
  loading.value = true
  error.value = null

  try {
    const payload = await getClasses({ page: nextPage, page_size: pageSize.value })
    classes.value = payload.list
    total.value = payload.total
    page.value = payload.page
    pageSize.value = payload.page_size
  } catch (err) {
    console.error('加载后台班级目录失败:', err)
    classes.value = []
    total.value = 0
    error.value = '加载班级目录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }
  void loadClasses(normalizedPage)
}

function setSort(option: WorkspaceDirectorySortOption): void {
  const matchedOption =
    sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
    sortOptions[0]

  if (!matchedOption) {
    return
  }

  sortConfig.value = matchedOption
}

function resetFilters(): void {
  keyword.value = ''
  rosterFilter.value = 'all'
}

function openClass(className: string): void {
  void router.push({
    name: resolveClassStudentsRouteName('admin'),
    params: { className },
  })
}

function getRosterStatusLabel(row: PlatformClassRow): string {
  return row.rosterStatus === 'active' ? '可查看' : '待入班'
}

onMounted(() => {
  void loadClasses()
})
</script>

<template>
  <section class="workspace-shell admin-class-manage-shell">
    <div class="admin-class-manage-shell__content">
      <header class="admin-class-manage-shell__hero">
        <div class="admin-class-manage-shell__hero-main">
          <div class="workspace-overline">Class Workspace</div>
          <h1 class="workspace-page-title">班级管理</h1>
          <p class="workspace-page-copy">在后台视角查看班级目录、学生规模，并快速进入班级详情。</p>
        </div>

        <div class="admin-class-manage-shell__hero-side">
          <button type="button" class="admin-class-manage-shell__hero-action" @click="loadClasses()">
            刷新目录
          </button>
        </div>
      </header>

      <div
        class="admin-class-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article class="journal-note progress-card metric-panel-card">
          <div class="admin-class-manage-shell__metric-head">
            <span class="journal-note-label progress-card-label metric-panel-label">班级总量</span>
            <FolderKanban class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">{{ total }}</div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">平台已接入班级</div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="admin-class-manage-shell__metric-head">
            <span class="journal-note-label progress-card-label metric-panel-label">当前页学生数</span>
            <UsersRound class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ summaryStudentCount }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            本页已加载班级学生汇总
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="admin-class-manage-shell__metric-head">
            <span class="journal-note-label progress-card-label metric-panel-label">可查看班级</span>
            <SearchCode class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ activeClassCount }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前页已有学生的班级
          </div>
        </article>
      </div>

      <section class="workspace-directory-section admin-class-manage-directory">
        <header class="list-heading">
          <div>
            <div class="workspace-overline">Class Directory</div>
            <h2 class="list-heading__title">班级目录</h2>
          </div>
        </header>

        <WorkspaceDirectoryToolbar
          v-model="keyword"
          :total="total"
          :selected-sort-label="sortConfig.label"
          :sort-options="sortOptions"
          search-placeholder="检索班级名称或编号..."
          :reset-disabled="!hasActiveFilters"
          total-suffix="个班级"
          @select-sort="setSort"
          @reset-filters="resetFilters"
        >
          <template #filter-panel>
            <div class="admin-class-manage-filter-grid">
              <label class="admin-class-manage-filter-field">
                <span class="admin-class-manage-filter-field__label">班级状态</span>
                <select v-model="rosterFilter" class="admin-class-manage-filter-field__control">
                  <option value="all">全部班级</option>
                  <option value="active">已有学生</option>
                  <option value="empty">待入班</option>
                </select>
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div
          v-if="loading"
          class="workspace-directory-loading admin-class-manage-directory__loading"
        >
          正在同步班级目录...
        </div>

        <AppEmpty
          v-else-if="classes.length === 0"
          class="workspace-directory-empty admin-class-manage-directory__empty"
          icon="Users"
          title="暂无班级"
          description="当前平台还没有可查看的班级数据。"
        />

        <AppEmpty
          v-else-if="filteredRows.length === 0"
          class="workspace-directory-empty admin-class-manage-directory__empty"
          icon="Search"
          title="没有匹配班级"
          description="调整关键词或筛选条件后再试。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-class-manage-table"
          :columns="classTableColumns"
          :rows="filteredRows"
          row-key="name"
          row-class="admin-class-manage-table__row"
        >
          <template #cell-code="{ row }">
            <span class="admin-class-manage-table__code">{{ (row as PlatformClassRow).code }}</span>
          </template>

          <template #cell-name="{ row }">
            <div class="admin-class-manage-table__name-wrap">
              <span class="admin-class-manage-table__name-text" :title="(row as PlatformClassRow).name">
                {{ (row as PlatformClassRow).name }}
              </span>
            </div>
          </template>

          <template #cell-student_count="{ row }">
            <span class="admin-class-manage-table__count-text">
              {{ (row as PlatformClassRow).student_count || 0 }}
            </span>
          </template>

          <template #cell-status="{ row }">
            <span
              class="admin-class-manage-table__status-pill"
              :class="`admin-class-manage-table__status-pill--${(row as PlatformClassRow).rosterStatus}`"
            >
              {{ getRosterStatusLabel(row as PlatformClassRow) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <button
              type="button"
              class="admin-class-manage-table__action"
              @click="openClass((row as PlatformClassRow).name)"
            >
              查看班级
            </button>
          </template>
        </WorkspaceDataTable>

        <WorkspaceDirectoryPagination
          v-if="total > 0 && filteredRows.length > 0"
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :total-label="`共 ${total} 个班级`"
          @change-page="handlePageChange"
        />
      </section>

      <div v-if="error" class="admin-class-manage-shell__error">
        {{ error }}
        <button type="button" class="admin-class-manage-shell__error-action" @click="loadClasses()">
          重试
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.admin-class-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base))
    ),
    radial-gradient(circle at top left, color-mix(in srgb, var(--color-primary) 10%, transparent), transparent 20rem);
}

.admin-class-manage-shell__content {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1.6rem 1.6rem 1.2rem;
}

.admin-class-manage-shell__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.admin-class-manage-shell__hero-main {
  max-width: 44rem;
}

.admin-class-manage-shell__hero-side {
  display: flex;
  align-items: center;
}

.admin-class-manage-shell__hero-action,
.admin-class-manage-table__action,
.admin-class-manage-shell__error-action {
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

.admin-class-manage-shell__hero-action:hover,
.admin-class-manage-table__action:hover,
.admin-class-manage-shell__error-action:hover {
  border-color: color-mix(in srgb, var(--color-primary) 32%, var(--journal-border));
  background: color-mix(in srgb, var(--color-primary) 12%, var(--journal-surface));
}

.admin-class-manage-shell__summary {
  --metric-panel-border: color-mix(in srgb, var(--workspace-brand) 16%, var(--workspace-line-soft));
  --metric-panel-background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--workspace-brand) 10%, transparent), transparent 15rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base))
    );
}

.admin-class-manage-shell__metric-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--journal-ink);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.9rem;
}

.list-heading__title {
  margin: 0.35rem 0 0;
  font-size: clamp(1.2rem, 1rem + 0.5vw, 1.45rem);
  font-weight: 700;
  line-height: 1.15;
  color: var(--journal-ink);
}

.admin-class-manage-directory {
  display: grid;
  gap: var(--space-4);
}

.admin-class-manage-directory > .list-heading {
  margin-bottom: 0;
}

.admin-class-manage-directory :deep(.workspace-directory-toolbar) {
  margin-bottom: 0;
}

.admin-class-manage-filter-grid {
  display: grid;
  gap: 0.85rem;
}

.admin-class-manage-filter-field {
  display: grid;
  gap: 0.45rem;
}

.admin-class-manage-filter-field__label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.admin-class-manage-filter-field__control {
  min-height: 2.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  padding: 0 0.9rem;
  color: var(--journal-ink);
}

.admin-class-manage-directory__loading,
.admin-class-manage-directory__empty {
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.25rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.admin-class-manage-directory__loading {
  padding: 1.4rem 1.2rem;
  font-size: 0.95rem;
  color: var(--journal-muted);
}

.admin-class-manage-table {
  border: 1px solid var(--workspace-line-soft);
  border-radius: 1.35rem;
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  padding: 0.25rem 0.9rem 0.4rem;
}

.admin-class-manage-table :deep(.workspace-data-table__head-cell) {
  border-bottom-color: var(--workspace-line-soft);
}

.admin-class-manage-table :deep(.workspace-data-table__row) {
  border-bottom-color: var(--workspace-line-soft);
}

.admin-class-manage-table :deep(.workspace-data-table__body tr:last-child) {
  border-bottom-color: transparent;
}

.admin-class-manage-table :deep(.workspace-data-table__row:hover) {
  background: color-mix(in srgb, var(--color-primary) 6%, transparent);
}

.admin-class-manage-table__code,
.admin-class-manage-table :deep(.admin-class-manage-table__mono) {
  font-family: var(--font-family-mono);
  font-size: 0.82rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.admin-class-manage-table__name-wrap {
  display: flex;
  min-width: 0;
  align-items: center;
}

.admin-class-manage-table__name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.admin-class-manage-table__count-text {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--journal-ink);
}

.admin-class-manage-table__status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2rem;
  min-width: 5.25rem;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0 0.75rem;
  font-size: 0.78rem;
  font-weight: 700;
}

.admin-class-manage-table__status-pill--active {
  border-color: color-mix(in srgb, var(--color-primary) 22%, transparent);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-primary);
}

.admin-class-manage-table__status-pill--empty {
  border-color: color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 78%, transparent);
  color: var(--journal-muted);
}

.admin-class-manage-shell__error {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--color-danger) 18%, var(--journal-border));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 6%, var(--journal-surface));
  padding: 0.95rem 1rem;
  color: color-mix(in srgb, var(--color-danger) 82%, var(--journal-ink));
}

@media (max-width: 768px) {
  .admin-class-manage-shell__content {
    padding: 1.15rem 1rem 0.9rem;
  }

  .admin-class-manage-shell__hero {
    flex-direction: column;
  }

  .admin-class-manage-shell__hero-side {
    width: 100%;
  }
}
</style>
