<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, FolderKanban, Search, Users } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

// PagePaginationControls is provided through WorkspaceDirectoryPagination.
interface StudentDirectoryTableRow {
  id: string
  student_no: string
  name: string
  username: string
  weak_dimension: string
  solved_count: number
  total_score: number
  actions: string
}

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  searchQuery: string
  studentNoQuery: string
  filteredStudents: TeacherStudentItem[]
  filteredTotal: number
  totalStudents: number
  page: number
  totalPages: number
  loadingClasses: boolean
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openReportExport: []
  updateSearchQuery: [value: string]
  updateStudentNoQuery: [value: string]
  selectClass: [className: string]
  changePage: [page: number]
  openStudent: [studentId: string]
}>()

const rows = computed<StudentDirectoryTableRow[]>(() =>
  props.filteredStudents.map((student) => ({
    id: student.id,
    student_no: student.student_no || '未设置学号',
    name: student.name || '未设置姓名',
    username: student.username,
    weak_dimension: student.weak_dimension || '暂无薄弱项',
    solved_count: student.solved_count ?? 0,
    total_score: student.total_score ?? 0,
    actions: 'open',
  }))
)

const hasActiveFilters = computed(() =>
  Boolean(props.selectedClassName || props.searchQuery.trim())
)

const columns = [
  { key: 'student_no', label: '学号', widthClass: 'w-[15%] min-w-[8rem]' },
  { key: 'name', label: '学生名称', widthClass: 'w-[20%] min-w-[11rem]' },
  { key: 'username', label: '昵称', widthClass: 'w-[16%] min-w-[10rem]' },
  { key: 'weak_dimension', label: '薄弱项', widthClass: 'w-[18%] min-w-[10rem]' },
  {
    key: 'solved_count',
    label: '做题数',
    widthClass: 'w-[10%] min-w-[6rem]',
    align: 'center' as const,
  },
  {
    key: 'total_score',
    label: '得分数',
    widthClass: 'w-[10%] min-w-[6rem]',
    align: 'center' as const,
  },
  { key: 'actions', label: '操作', widthClass: 'w-[9rem]', align: 'right' as const },
]

function resetFilters(): void {
  emit('selectClass', '')
  emit('updateSearchQuery', '')
}

function handleClassChange(event: Event): void {
  const target = event.target
  emit('selectClass', target instanceof HTMLSelectElement ? target.value : '')
}
</script>

<template>
  <div
    class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline">Student Directory</div>
            <h1 class="teacher-title workspace-page-title">学生管理</h1>
            <p class="teacher-copy workspace-page-copy">按班级筛选、搜索并进入学员分析。</p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('openClassManagement')"
            >
              班级管理
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('openReportExport')"
            >
              导出班级报告
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <span>Directory Snapshot</span>
          </div>
          <div
            class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"
          >
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>可访问班级</span>
                <FolderKanban class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ classes.length }}
              </div>
              <div class="progress-card-hint metric-panel-helper">当前教师可切换的班级数量</div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>当前班级学生</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ totalStudents }}
              </div>
              <div class="progress-card-hint metric-panel-helper">当前选中班级的学生总数</div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>搜索结果</span>
                <Search class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ filteredStudents.length }}
              </div>
              <div class="progress-card-hint metric-panel-helper">当前搜索条件下匹配的学生数量</div>
            </article>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="学生目录"
        >
          <section class="teacher-directory-shell workspace-directory-list">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">Student Directory</div>
                <h3 class="list-heading__title">学生目录</h3>
              </div>
            </header>

            <WorkspaceDirectoryToolbar
              :model-value="searchQuery"
              :total="filteredTotal"
              selected-sort-label=""
              :sort-options="[]"
              search-placeholder="搜索姓名、用户名或学号"
              total-suffix="名学生"
              :show-total="false"
              filter-panel-title="学生筛选"
              reset-label="清空筛选"
              :reset-disabled="!hasActiveFilters"
              @update:model-value="emit('updateSearchQuery', $event)"
              @reset-filters="resetFilters"
            >
              <template #filter-panel>
                <label class="teacher-directory-filter-field">
                  <span class="workspace-overline">班级</span>
                  <select
                    :value="selectedClassName"
                    class="teacher-directory-filter-control"
                    :disabled="loadingClasses"
                    @change="handleClassChange"
                  >
                    <option value="">全部班级</option>
                    <option v-for="item in classes" :key="item.name" :value="item.name">
                      {{ item.name }} · {{ item.student_count || 0 }}
                    </option>
                  </select>
                </label>
              </template>
            </WorkspaceDirectoryToolbar>

            <div v-if="loadingStudents" class="workspace-directory-loading">
              <AppLoading>同步学生目录...</AppLoading>
            </div>

            <AppEmpty
              v-else-if="filteredStudents.length === 0"
              class="teacher-empty-state workspace-directory-empty"
              icon="Users"
              title="没有匹配学生"
              description="调整搜索词或切换班级后再试。"
            />

            <div v-else class="teacher-directory">
              <WorkspaceDataTable
                class="teacher-student-directory-table"
                :columns="columns"
                :rows="rows"
                row-key="id"
              >
                <template #cell-student_no="{ row }">
                  <span class="teacher-directory-cell-student-no">
                    {{ (row as StudentDirectoryTableRow).student_no }}
                  </span>
                </template>

                <template #cell-name="{ row }">
                  <div class="teacher-directory-cell-name">
                    <h4
                      class="teacher-directory-row-title"
                      :title="(row as StudentDirectoryTableRow).name"
                    >
                      {{ (row as StudentDirectoryTableRow).name }}
                    </h4>
                  </div>
                </template>

                <template #cell-username="{ row }">
                  <span
                    class="teacher-directory-row-points"
                    :title="(row as StudentDirectoryTableRow).username"
                  >
                    {{ (row as StudentDirectoryTableRow).username }}
                  </span>
                </template>

                <template #cell-weak_dimension="{ row }">
                  <span
                    class="teacher-directory-chip teacher-directory-chip-muted"
                    :class="'workspace-directory-status-pill workspace-directory-status-pill--muted'"
                  >
                    {{ (row as StudentDirectoryTableRow).weak_dimension }}
                  </span>
                </template>

                <template #cell-solved_count="{ row }">
                  <span class="teacher-directory-row-solved">
                    {{ (row as StudentDirectoryTableRow).solved_count }}
                  </span>
                </template>

                <template #cell-total_score="{ row }">
                  <span class="teacher-directory-row-score">
                    {{ (row as StudentDirectoryTableRow).total_score }}
                  </span>
                </template>

                <template #cell-actions="{ row }">
                  <div class="workspace-directory-row-actions teacher-directory-row-cta">
                    <button
                      type="button"
                      class="ui-btn ui-btn--primary ui-btn--xs"
                      :aria-label="`${(row as StudentDirectoryTableRow).name || (row as StudentDirectoryTableRow).username}，${(row as StudentDirectoryTableRow).solved_count} 题，${(row as StudentDirectoryTableRow).total_score} 分，查看学员分析`"
                      @click="emit('openStudent', (row as StudentDirectoryTableRow).id)"
                    >
                      学员分析
                      <ArrowRight class="h-4 w-4" />
                    </button>
                  </div>
                </template>
              </WorkspaceDataTable>

              <WorkspaceDirectoryPagination
                v-if="filteredTotal > 0"
                class="teacher-directory-pagination"
                :page="page"
                :total-pages="totalPages"
                :total="filteredTotal"
                :total-label="`共 ${filteredTotal} 名学生`"
                @change-page="emit('changePage', $event)"
              />
            </div>
          </section>
        </section>
        <div v-if="error" class="teacher-surface-error">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">
            重试
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  font-family: var(--font-family-sans);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
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

.teacher-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.teacher-directory-section :deep(.workspace-directory-pagination-shell) {
  margin-top: var(--space-2);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 900;
  color: var(--color-text-primary);
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

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-student-no {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-0-98);
  font-weight: 800;
  line-height: 1.35;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-student-directory-table
  :deep(.workspace-data-table__row:hover)
  .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  font-size: var(--font-size-0-80);
  font-weight: 800;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-solved,
.teacher-directory-row-score {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-81);
  font-weight: 800;
  color: var(--color-text-primary);
}

.teacher-directory-row-cta {
  justify-content: flex-end;
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid {
    grid-template-columns: 1fr;
  }

  .teacher-directory-row-cta {
    justify-content: flex-start;
  }
}
</style>
