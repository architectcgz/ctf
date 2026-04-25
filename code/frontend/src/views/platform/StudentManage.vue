<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getStudentsDirectory } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import StudentManageHeroPanel from '@/components/platform/student/StudentManageHeroPanel.vue'
import StudentManageWorkspacePanel from '@/components/platform/student/StudentManageWorkspacePanel.vue'
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
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-student-manage-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <StudentManageHeroPanel
          :total="total"
          :active-students="activeStudents"
          :assigned-class-count="assignedClassCount"
          @refresh="void initialize()"
        />

        <StudentManageWorkspacePanel
          :classes="classes"
          :loading="loading"
          :loading-classes="loadingClasses"
          :error="error"
          :keyword="keyword"
          :class-filter="classFilter"
          :total="total"
          :has-active-filters="hasActiveFilters"
          :rows="rows"
          :page="page"
          :total-pages="totalPages"
          @update:keyword="handleKeywordChange"
          @change:class-filter="handleClassFilterChange"
          @reset-filters="resetFilters"
          @change-page="handlePageChange"
          @open-student="openStudent"
        />
      </main>
    </div>
  </div>
</template>

<style scoped>
</style>
