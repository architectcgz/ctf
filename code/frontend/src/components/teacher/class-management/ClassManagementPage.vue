<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight, FolderKanban } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

interface ClassDirectoryTableRow {
  name: string
  code: string
  student_count: number
  status: 'ready' | 'empty'
  status_label: string
  actions: string
}

type ClassStatusFilter = ClassDirectoryTableRow['status'] | ''

const props = defineProps<{
  classes: TeacherClassItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  changePage: [page: number]
  openDashboard: []
  openReportExport: []
  openClass: [className: string]
}>()

const filterQuery = ref('')
const statusFilter = ref<ClassStatusFilter>('')

const classEntries = computed<ClassDirectoryTableRow[]>(() =>
  props.classes.map((item, index) => ({
    name: item.name,
    code: `CL-${String(index + 1).padStart(2, '0')}`,
    student_count: item.student_count || 0,
    status: (item.student_count || 0) > 0 ? 'ready' : 'empty',
    status_label: (item.student_count || 0) > 0 ? '可查看' : '待入班',
    actions: 'open',
  }))
)

const filteredClassEntries = computed(() => {
  const keyword = filterQuery.value.trim().toLowerCase()

  return classEntries.value.filter((row) => {
    const matchesKeyword =
      !keyword || row.code.toLowerCase().includes(keyword) || row.name.toLowerCase().includes(keyword)
    const matchesStatus = !statusFilter.value || row.status === statusFilter.value

    return matchesKeyword && matchesStatus
  })
})

const hasActiveFilters = computed(() => Boolean(filterQuery.value.trim() || statusFilter.value))

const columns = [
  { key: 'code', label: '班级编号', widthClass: 'w-[16%] min-w-[8rem]' },
  { key: 'name', label: '班级名称', widthClass: 'w-[34%] min-w-[14rem]' },
  { key: 'student_count', label: '学生数', widthClass: 'w-[14%] min-w-[7rem]', align: 'center' as const },
  { key: 'status', label: '状态', widthClass: 'w-[14%] min-w-[7rem]', align: 'center' as const },
  { key: 'actions', label: '操作', widthClass: 'w-[16%] min-w-[8rem]', align: 'right' as const },
]

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / Math.max(props.pageSize, 1))))
const currentPageStudentCount = computed(() =>
  props.classes.reduce((sum, item) => sum + (item.student_count || 0), 0)
)

function resetFilters(): void {
  filterQuery.value = ''
  statusFilter.value = ''
}

function handleStatusFilterChange(event: Event): void {
  const target = event.target
  statusFilter.value = target instanceof HTMLSelectElement ? target.value as ClassStatusFilter : ''
}
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline">
              Class Directory
            </div>
            <h1 class="teacher-title workspace-page-title">
              班级管理
            </h1>
            <p class="teacher-copy workspace-page-copy">
              查看当前可管理班级，并进入对应班级继续查看学生和训练表现。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openDashboard')"
            >
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="emit('openReportExport')"
            >
              导出班级报告
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <FolderKanban class="h-4 w-4" />
            <span>Directory Snapshot</span>
          </div>
          <div class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
            <div class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                班级数量
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ total }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前可管理班级总数
              </div>
            </div>
            <div class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                当前页学生数
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ currentPageStudentCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前分页已加载班级的学生数汇总
              </div>
            </div>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="班级目录"
        >
          <header class="list-heading">
            <div>
              <div class="workspace-overline">
                Class Directory
              </div>
              <h3 class="list-heading__title">
                班级目录
              </h3>
            </div>
          </header>

          <WorkspaceDirectoryToolbar
            v-model="filterQuery"
            :total="filteredClassEntries.length"
            selected-sort-label=""
            :sort-options="[]"
            search-placeholder="搜索班级编号或名称"
            total-suffix="个班级"
            :show-total="false"
            filter-panel-title="班级筛选"
            reset-label="清空筛选"
            :reset-disabled="!hasActiveFilters"
            @reset-filters="resetFilters"
          >
            <template #filter-panel>
              <label class="teacher-directory-filter-field">
                <span class="workspace-overline">班级状态</span>
                <select
                  :value="statusFilter"
                  class="teacher-directory-filter-control"
                  @change="handleStatusFilterChange"
                >
                  <option value="">
                    全部状态
                  </option>
                  <option value="ready">
                    可查看
                  </option>
                  <option value="empty">
                    待入班
                  </option>
                </select>
              </label>
            </template>
          </WorkspaceDirectoryToolbar>

          <div
            v-if="loading"
            class="workspace-directory-loading"
          >
            <AppLoading>同步班级目录...</AppLoading>
          </div>

          <AppEmpty
            v-else-if="classes.length === 0"
            class="teacher-empty-state workspace-directory-empty"
            icon="Users"
            title="暂无班级"
            description="当前教师账号下还没有可访问的班级。"
          />

          <div
            v-else
            class="teacher-directory"
          >
            <AppEmpty
              v-if="filteredClassEntries.length === 0"
              class="teacher-empty-state workspace-directory-empty"
              icon="Search"
              title="没有匹配班级"
              description="调整搜索关键词后再试。"
            />

            <WorkspaceDataTable
              v-else
              class="workspace-directory-list teacher-class-directory-table"
              :columns="columns"
              :rows="filteredClassEntries"
              row-key="name"
            >
              <template #cell-code="{ row }">
                <span class="teacher-directory-cell-class-code">
                  {{ (row as ClassDirectoryTableRow).code }}
                </span>
              </template>

              <template #cell-name="{ row }">
                <div class="teacher-directory-cell-class-name">
                  <h4
                    class="teacher-directory-row-title"
                    :title="(row as ClassDirectoryTableRow).name"
                  >
                    {{ (row as ClassDirectoryTableRow).name }}
                  </h4>
                </div>
              </template>

              <template #cell-student_count="{ row }">
                <span class="teacher-directory-row-points">
                  {{ (row as ClassDirectoryTableRow).student_count }}
                </span>
              </template>

              <template #cell-status="{ row }">
                <span
                  class="teacher-directory-state-chip"
                  :class="
                    (row as ClassDirectoryTableRow).status === 'ready'
                      ? 'teacher-directory-state-chip-ready'
                      : 'teacher-directory-state-chip-empty'
                  "
                >
                  {{ (row as ClassDirectoryTableRow).status_label }}
                </span>
              </template>

              <template #cell-actions="{ row }">
                <div class="teacher-directory-row-cta">
                  <button
                    type="button"
                    class="ui-btn ui-btn--primary ui-btn--xs"
                    :aria-label="`${(row as ClassDirectoryTableRow).name}，${(row as ClassDirectoryTableRow).student_count} 名学生，进入班级`"
                    @click="emit('openClass', (row as ClassDirectoryTableRow).name)"
                  >
                    进入班级
                    <ArrowRight class="h-4 w-4" />
                  </button>
                </div>
              </template>
            </WorkspaceDataTable>

            <WorkspaceDirectoryPagination
              v-if="total > 0 && filteredClassEntries.length > 0"
              class="teacher-directory-pagination"
              :page="page"
              :total-pages="totalPages"
              :total="total"
              :total-label="`共 ${total} 个班级`"
              @change-page="emit('changePage', $event)"
            />
          </div>
        </section>
        <div
          v-if="error"
          class="teacher-surface-error"
        >
          {{ error }}
          <button
            type="button"
            class="ml-3 font-medium underline"
            @click="emit('retry')"
          >
            重试
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --teacher-directory-columns: var(--teacher-class-directory-columns);
  --teacher-class-directory-columns: minmax(7rem, 0.7fr) minmax(11rem, 1.15fr) minmax(7rem, 0.7fr)
    minmax(7rem, 0.7fr) minmax(7rem, 0.75fr);
  font-family: var(--font-family-sans);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
}

.teacher-directory-section :deep(.workspace-directory-pagination-shell) {
  margin-top: var(--space-2);
}

.teacher-directory-filter-field {
  display: grid;
  gap: var(--space-2);
}

.teacher-directory-filter-control {
  min-height: 2.5rem;
  width: 100%;
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
  padding: 0 var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 900;
  color: var(--color-text-primary);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-cell-class-code,
.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-class-code {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-sans);
  font-size: var(--font-size-1-08);
  font-weight: 800;
  line-height: 1.35;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-class-directory-table :deep(.workspace-data-table__row:hover) .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  font-size: var(--font-size-1-00);
  font-weight: 900;
  color: var(--color-text-primary);
}

.teacher-directory-cell-class-code,
.teacher-directory-cell-class-name {
  width: 100%;
  min-width: 0;
}

.teacher-directory-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-3);
  border-radius: 0.5rem;
  font-size: var(--font-size-0-75);
  font-weight: 800;
}

.teacher-directory-state-chip-ready {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.teacher-directory-state-chip-empty {
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 960px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-directory-row-cta {
    justify-content: flex-start;
  }
}
</style>
