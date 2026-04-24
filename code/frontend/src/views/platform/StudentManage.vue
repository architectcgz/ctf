<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getStudentsDirectory } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import StudentManageHeroPanel from '@/components/platform/student/StudentManageHeroPanel.vue'
import { useStudentDirectoryQuery } from '@/composables/useStudentDirectoryQuery'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

const router = useRouter()
const classes = ref<TeacherClassItem[]>([])
const loadingClasses = ref(false)
const pageError = ref<string | null>(null)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const keyword = ref('')
const classFilter = ref('')
const studentDirectoryQuery = useStudentDirectoryQuery({
  debounceMs: 250,
  errorMessage: '加载学生目录失败，请稍后重试',
  request: getStudentsDirectory,
})

const list = computed(() => studentDirectoryQuery.students.value)
const total = computed(() => studentDirectoryQuery.total.value)
const loading = computed(() => studentDirectoryQuery.loading.value)
const error = computed(() => pageError.value ?? studentDirectoryQuery.error.value)
const hasActiveFilters = computed(() => Boolean(keyword.value.trim() || classFilter.value))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
const activeStudents = computed(() =>
  list.value.filter((item) => (item.recent_event_count ?? 0) > 0).length
)
const assignedClassCount = computed(() =>
  classes.value.filter((item) => (item.student_count ?? 0) > 0).length
)
const directoryParams = computed(() => ({
  class_name: classFilter.value || undefined,
  keyword: keyword.value.trim() || undefined,
  student_no: undefined,
  sort_key: 'name' as const,
  sort_order: 'asc' as const,
  page: page.value,
  page_size: pageSize.value,
}))
const rows = computed(() =>
  list.value.map((item) => ({
    id: item.id,
    name: item.name?.trim() || '未设置姓名',
    username: item.username,
    student_no: item.student_no?.trim() || '未设置学号',
    class_name: item.class_name || '未分班',
    total_score: item.total_score ?? 0,
    actions: '查看学员',
  }))
)

async function loadClasses(): Promise<void> {
  loadingClasses.value = true
  try {
    classes.value = await getClasses()
  } finally {
    loadingClasses.value = false
  }
}

async function loadStudents(): Promise<void> {
  await studentDirectoryQuery.loadStudents(directoryParams.value)
}

async function initialize(): Promise<void> {
  pageError.value = null
  studentDirectoryQuery.cancelScheduledLoad()

  try {
    await loadClasses()
    await loadStudents()
  } catch (err) {
    console.error('初始化学生管理失败:', err)
    pageError.value = '加载学生管理失败，请稍后重试'
  }
}

function handleKeywordChange(value: string): void {
  keyword.value = value
  page.value = 1
  studentDirectoryQuery.scheduleLoadStudents({
    ...directoryParams.value,
    keyword: value.trim() || undefined,
    page: 1,
  })
}

function handleClassFilterChange(value: string): void {
  classFilter.value = value
  page.value = 1
  studentDirectoryQuery.cancelScheduledLoad()
  void studentDirectoryQuery.loadStudents({
    ...directoryParams.value,
    class_name: value || undefined,
    page: 1,
  })
}

function resetFilters(): void {
  keyword.value = ''
  classFilter.value = ''
  page.value = 1
  studentDirectoryQuery.cancelScheduledLoad()
  void studentDirectoryQuery.loadStudents({
    ...directoryParams.value,
    class_name: undefined,
    keyword: undefined,
    page: 1,
  })
}

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }

  page.value = normalizedPage
  void loadStudents()
}

function openStudent(studentId: string): void {
  const student = list.value.find((item) => item.id === studentId)
  void router.push({
    name: 'PlatformStudentAnalysis',
    params: {
      className: student?.class_name || classFilter.value || '',
      studentId,
    },
  })
}

onMounted(() => {
  void initialize()
})

const columns = [
  { key: 'name', label: '学生姓名', widthClass: 'w-[20%] min-w-[12rem]' },
  { key: 'username', label: '用户名', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'student_no', label: '学号', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'class_name', label: '班级', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'total_score', label: '得分', widthClass: 'w-[12%] min-w-[8rem]', align: 'center' as const },
  { key: 'actions', label: '操作', widthClass: 'w-[12%] min-w-[8rem]', align: 'right' as const },
]
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <StudentManageHeroPanel
          :total="total"
          :active-students="activeStudents"
          :assigned-class-count="assignedClassCount"
          @refresh="void initialize()"
        />

        <div class="admin-student-manage-shell__content">
          <section class="workspace-directory-section admin-student-manage-directory">
            <WorkspaceDirectoryToolbar
              :model-value="keyword"
              :total="total"
              selected-sort-label=""
              :sort-options="[]"
              search-placeholder="检索姓名、用户名或学号..."
              filter-panel-title="学生筛选"
              total-suffix="名学生"
              reset-label="清空筛选"
              :reset-disabled="!hasActiveFilters"
              @update:model-value="handleKeywordChange"
              @reset-filters="resetFilters"
            >
              <template #filter-panel>
                <label class="workspace-student-filter-field">
                  <span class="workspace-overline">班级范围</span>
                  <select
                    :value="classFilter"
                    class="admin-input workspace-student-filter-control"
                    @change="handleClassFilterChange(($event.target as HTMLSelectElement).value)"
                  >
                    <option value="">
                      全部班级
                    </option>
                    <option
                      v-for="item in classes"
                      :key="item.name"
                      :value="item.name"
                    >
                      {{ item.name }}
                    </option>
                  </select>
                </label>
              </template>
            </WorkspaceDirectoryToolbar>

            <div
              v-if="(loading || loadingClasses) && list.length === 0"
              class="py-12 flex justify-center"
            >
              <AppLoading>同步学员目录...</AppLoading>
            </div>

            <template v-else>
              <AppEmpty
                v-if="list.length === 0"
                class="workspace-directory-empty"
                icon="Users"
                title="暂无学生数据"
                description="当前平台上没有任何学生账号。"
              />

              <WorkspaceDataTable
                v-else
                class="workspace-directory-list admin-student-manage-table"
                :columns="columns"
                :rows="rows"
                row-key="id"
              >
                <template #cell-actions="{ row }">
                  <button
                    type="button"
                    class="ui-btn ui-btn--ghost"
                    @click="openStudent(String((row as { id: string }).id))"
                  >
                    查看学员
                  </button>
                </template>
              </WorkspaceDataTable>

              <div class="workspace-directory-pagination">
                <WorkspaceDirectoryPagination
                  :page="page"
                  :total-pages="totalPages"
                  :total="total"
                  total-label="名学生"
                  @change-page="handlePageChange"
                />
              </div>
            </template>
          </section>

          <div
            v-if="error"
            class="teacher-surface-error"
          >
            {{ error }}
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-student-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}

</style>
