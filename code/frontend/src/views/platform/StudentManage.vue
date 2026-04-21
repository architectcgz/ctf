<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { GraduationCap, UsersRound, SearchCode, Calendar, SortAsc } from 'lucide-vue-next'

import { getClasses, getStudentsDirectory } from '@/api/teacher'
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { useStudentDirectoryQuery } from '@/composables/useStudentDirectoryQuery'
import { useStudentFilters } from '@/composables/useStudentFilters'

type StudentSortKey = 'name' | 'student_no' | 'total_score' | 'solved_count'
type StudentSortOption = WorkspaceDirectorySortOption & {
  key: StudentSortKey
  order: 'asc' | 'desc'
}

interface PlatformStudentRow extends TeacherStudentItem {
  displayName: string
  classLabel: string
  studentCode: string
  score: number
  solved: number
}

const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const loadingClasses = ref(false)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const pageError = ref<string | null>(null)

const filters = useStudentFilters()
const { selectedClassName, searchQuery, studentNoQuery } = filters

const studentDirectoryQuery = useStudentDirectoryQuery({
  debounceMs: 250,
  errorMessage: '加载学生目录失败，请稍后重试',
  request: getStudentsDirectory,
})

const { students, total, loading } = studentDirectoryQuery

const sortOptions: StudentSortOption[] = [
  { key: 'name', order: 'asc', label: '学生姓名 A-Z', icon: SortAsc },
  { key: 'total_score', order: 'desc', label: '得分由高到低', icon: GraduationCap },
  { key: 'solved_count', order: 'desc', label: '做题数由高到低', icon: SearchCode },
  { key: 'student_no', order: 'asc', label: '学号顺序', icon: Calendar },
]
const sortConfig = ref<StudentSortOption>(sortOptions[0]!)

const error = computed(() => pageError.value ?? studentDirectoryQuery.error.value)
const hasActiveFilters = computed(
  () =>
    Boolean(
      searchQuery.value.trim() || studentNoQuery.value.trim() || selectedClassName.value.trim()
    )
)

const totalStudents = computed(() =>
  classes.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
)
const activeClassCount = computed(
  () => classes.value.filter((item) => (item.student_count || 0) > 0).length
)
const filteredTotal = computed(() => total.value)
const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredTotal.value / Math.max(pageSize.value, 1)))
)
const hasAnyStudents = computed(() => totalStudents.value > 0)
const directoryParams = computed(() => ({
  class_name: selectedClassName.value || undefined,
  keyword: searchQuery.value.trim() || undefined,
  student_no: studentNoQuery.value.trim() || undefined,
  sort_key: sortConfig.value.key,
  sort_order: sortConfig.value.order,
  page: page.value,
  page_size: pageSize.value,
}))

const studentRows = computed<PlatformStudentRow[]>(() =>
  students.value.map((student, index) => {
    const offset = (page.value - 1) * pageSize.value + index + 1
    return {
      ...student,
      displayName: student.name || student.username || '未设置姓名',
      classLabel: student.class_name || selectedClassName.value || '未分配班级',
      studentCode: student.student_no || `ST-${String(offset).padStart(3, '0')}`,
      score: student.total_score || 0,
      solved: student.solved_count || 0,
    }
  })
)

const studentTableColumns = [
  {
    key: 'studentCode',
    label: '学号',
    widthClass: 'w-[14%] min-w-[8rem]',
    cellClass: 'admin-student-manage-table__mono',
  },
  {
    key: 'displayName',
    label: '学生名称',
    widthClass: 'w-[24%] min-w-[12rem]',
    cellClass: 'admin-student-manage-table__name',
  },
  {
    key: 'username',
    label: '用户名',
    widthClass: 'w-[18%] min-w-[10rem]',
    cellClass: 'admin-student-manage-table__alias',
  },
  {
    key: 'classLabel',
    label: '班级',
    widthClass: 'w-[16%] min-w-[9rem]',
    cellClass: 'admin-student-manage-table__class',
  },
  {
    key: 'solved',
    label: '做题数',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'admin-student-manage-table__count',
  },
  {
    key: 'score',
    label: '得分',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'admin-student-manage-table__count',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[12rem]',
    cellClass: 'admin-student-manage-table__actions',
  },
]

async function loadClasses(): Promise<void> {
  loadingClasses.value = true
  try {
    classes.value = await getClasses()
  } finally {
    loadingClasses.value = false
  }
}

async function selectClass(className: string): Promise<void> {
  studentDirectoryQuery.cancelScheduledLoad()
  filters.updateSelectedClassName(className)
  page.value = 1
  await studentDirectoryQuery.loadStudents({
    ...directoryParams.value,
    class_name: className || undefined,
    page: 1,
  })
}

async function initialize(): Promise<void> {
  pageError.value = null

  try {
    await loadClasses()
    await selectClass('')
  } catch (err) {
    console.error('初始化后台学生目录失败:', err)
    pageError.value = '加载学生目录失败，请稍后重试'
  }
}

function openStudent(row: PlatformStudentRow): void {
  void router.push({
    name: 'PlatformStudentAnalysis',
    params: {
      className: row.class_name || selectedClassName.value || '',
      studentId: row.id,
    },
  })
}

function updateSearchQuery(value: string): void {
  filters.updateSearchQuery(value)
}

function updateStudentNoQuery(value: string): void {
  filters.updateStudentNoQuery(value)
}

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }
  page.value = normalizedPage
  void studentDirectoryQuery.loadStudents(directoryParams.value)
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
  filters.updateSelectedClassName('')
  filters.resetTextFilters()
}

watch([searchQuery, studentNoQuery], () => {
  page.value = 1
  studentDirectoryQuery.scheduleLoadStudents({
    ...directoryParams.value,
    page: 1,
  })
})

watch(
  () => [sortConfig.value.key, sortConfig.value.order],
  () => {
    page.value = 1
    studentDirectoryQuery.scheduleLoadStudents({
      ...directoryParams.value,
      page: 1,
    })
  }
)

watch(filteredTotal, (nextTotal) => {
  if (nextTotal === 0 || page.value <= totalPages.value) {
    return
  }
  page.value = totalPages.value
  void studentDirectoryQuery.loadStudents(directoryParams.value)
})

onMounted(() => {
  void initialize()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-student-manage-shell flex min-h-full flex-1 flex-col"
  >
    <header class="admin-workbench-header">
      <div class="admin-workbench-header__top">
        <div class="admin-workbench-header__identity">
          <div class="workspace-overline">
            Student Workspace
          </div>
          <h1 class="admin-workbench-header__title">
            学生管理
          </h1>
          <p class="admin-workbench-header__description">
            在后台视角查看学生目录、班级归属与学习表现，并快速进入学员分析。
          </p>
        </div>

        <div class="admin-workbench-header__actions">
          <button
            type="button"
            class="ui-btn ui-btn--ghost ui-btn--sm"
            @click="initialize()"
          >
            <RefreshCw class="h-3.5 w-3.5" />
            刷新目录
          </button>
        </div>
      </div>
    </header>

    <main class="admin-workbench-content flex-1 p-[2rem_3rem] space-y-6">
      <div class="metric-panel-grid metric-panel-grid--premium cols-3">
        <article class="metric-panel-card metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>学生总量</span>
            <GraduationCap class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ totalStudents.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            平台已纳入学生
          </div>
        </article>

        <article class="metric-panel-card metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>可查看班级</span>
            <UsersRound class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ activeClassCount.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            已有学生数据的班级数
          </div>
        </article>

        <article class="metric-panel-card metric-panel-card--premium">
          <div class="metric-panel-label">
            <span>当前结果</span>
            <SearchCode class="h-4 w-4" />
          </div>
          <div class="metric-panel-value">
            {{ filteredTotal.toString().padStart(2, '0') }}
          </div>
          <div class="metric-panel-helper">
            当前条件下的学生数量
          </div>
        </article>
      </div>

      <section class="workspace-directory-section admin-student-manage-directory">
        <header class="list-heading">
          <div>
            <div class="workspace-overline">
              Student Directory
            </div>
            <h2 class="list-heading__title">
              学生目录
            </h2>
          </div>
        </header>

        <WorkspaceDirectoryToolbar
          v-model="searchQuery"
          :total="filteredTotal"
          :selected-sort-label="sortConfig.label"
          :sort-options="sortOptions"
          search-placeholder="检索姓名、用户名或学号..."
          :reset-disabled="!hasActiveFilters"
          total-suffix="名学生"
          @select-sort="setSort"
          @reset-filters="resetFilters"
        >
          <template #filter-panel>
            <div class="admin-student-manage-filter-grid">
              <label class="admin-student-manage-filter-field">
                <span class="admin-student-manage-filter-field__label">班级范围</span>
                <select
                  :value="selectedClassName"
                  class="admin-student-manage-filter-field__control"
                  :disabled="loadingClasses"
                  @change="selectClass(($event.target as HTMLSelectElement).value)"
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

              <label class="admin-student-manage-filter-field">
                <span class="admin-student-manage-filter-field__label">按学号查询</span>
                <input
                  :value="studentNoQuery"
                  type="text"
                  class="admin-student-manage-filter-field__control"
                  placeholder="输入学号精确查询"
                  @input="updateStudentNoQuery(($event.target as HTMLInputElement).value)"
                >
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div
          v-if="loading"
          class="workspace-directory-loading"
        >
          正在同步学生目录...
        </div>

        <AppEmpty
          v-else-if="students.length === 0 && !hasActiveFilters && !hasAnyStudents"
          class="workspace-directory-empty"
          icon="Users"
          title="暂无学生"
          description="当前平台还没有可查看的学生数据。"
        />

        <AppEmpty
          v-else-if="students.length === 0"
          class="workspace-directory-empty"
          icon="Search"
          title="没有匹配学生"
          description="调整关键词、班级或学号后再试。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-student-manage-table"
          :columns="studentTableColumns"
          :rows="studentRows"
          row-key="id"
          row-class="admin-student-manage-table__row"
        >
          <template #cell-studentCode="{ row }">
            <span class="admin-student-manage-table__code">
              {{ (row as PlatformStudentRow).studentCode }}
            </span>
          </template>

          <template #cell-displayName="{ row }">
            <div class="admin-student-manage-table__name-wrap">
              <span
                class="admin-student-manage-table__name-text"
                :title="(row as PlatformStudentRow).displayName"
              >
                {{ (row as PlatformStudentRow).displayName }}
              </span>
            </div>
          </template>

          <template #cell-username="{ row }">
            <span
              class="admin-student-manage-table__alias-text"
              :title="(row as PlatformStudentRow).username"
            >
              {{ (row as PlatformStudentRow).username }}
            </span>
          </template>

          <template #cell-classLabel="{ row }">
            <span class="admin-student-manage-table__class-pill">
              {{ (row as PlatformStudentRow).classLabel }}
            </span>
          </template>

          <template #cell-solved="{ row }">
            <span class="admin-student-manage-table__count-text">
              {{ (row as PlatformStudentRow).solved }}
            </span>
          </template>

          <template #cell-score="{ row }">
            <span class="admin-student-manage-table__count-text">
              {{ (row as PlatformStudentRow).score }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <button
              type="button"
              class="admin-student-manage-table__action"
              @click="openStudent(row as PlatformStudentRow)"
            >
              查看学员
            </button>
          </template>
        </WorkspaceDataTable>

        <WorkspaceDirectoryPagination
          v-if="filteredTotal > 0 && studentRows.length > 0"
          :page="page"
          :total-pages="totalPages"
          :total="filteredTotal"
          :total-label="`共 ${filteredTotal} 名学生`"
          @change-page="handlePageChange"
        />
      </section>

      <div
        v-if="error"
        class="admin-student-manage-shell__error"
      >
        {{ error }}
        <button
          type="button"
          class="admin-student-manage-shell__error-action"
          @click="initialize()"
        >
          重试
        </button>
      </div>
    </main>
  </section>
</template>

<style scoped>
.admin-student-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --student-directory-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --student-directory-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
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

.admin-student-manage-shell__content {
  display: grid;
  gap: var(--workspace-directory-page-block-gap);
}

.admin-student-manage-shell__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.admin-student-manage-shell__hero-main {
  max-width: 46rem;
}

.admin-student-manage-shell__hero-side {
  display: flex;
  align-items: center;
}

.admin-student-manage-shell__hero-action,
.admin-student-manage-table__action,
.admin-student-manage-shell__error-action {
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

.admin-student-manage-shell__hero-action:hover,
.admin-student-manage-table__action:hover,
.admin-student-manage-shell__error-action:hover {
  border-color: color-mix(in srgb, var(--color-primary) 32%, var(--journal-border));
  background: color-mix(in srgb, var(--color-primary) 12%, var(--journal-surface));
}

.admin-student-manage-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.admin-student-manage-filter-field {
  display: grid;
  gap: var(--space-2);
}

.admin-student-manage-filter-field__label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.admin-student-manage-filter-field__control {
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

.admin-student-manage-filter-field__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.admin-student-manage-table {
  --workspace-directory-shell-border: var(--student-directory-border);
  --workspace-directory-head-divider: var(--student-directory-border);
  --workspace-directory-row-divider: var(--student-directory-row-divider);
}

.admin-student-manage-table :deep(.workspace-data-table__row:hover) {
  background: color-mix(in srgb, var(--color-primary) 5%, transparent);
}

.admin-student-manage-table__code,
.admin-student-manage-table__count-text {
  font-family: var(--font-family-mono);
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-student-manage-table__code {
  letter-spacing: 0.04em;
  color: var(--journal-muted);
}

.admin-student-manage-table__name-wrap {
  display: flex;
  min-width: 0;
  align-items: center;
}

.admin-student-manage-table__name-text,
.admin-student-manage-table__alias-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-student-manage-table__name-text {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-student-manage-table__alias-text {
  font-size: 0.94rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.admin-student-manage-table__class-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 14%, var(--journal-border));
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 7%, var(--journal-surface));
  padding: 0 0.8rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-primary);
}

.admin-student-manage-shell__error {
  padding: 1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--color-danger) 18%, var(--journal-border));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 8%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 76%, var(--journal-ink));
}
</style>
