<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getStudentsDirectory } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import StudentManagementPage from '@/components/teacher/student-management/StudentManagementPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useStudentDirectoryQuery } from '@/composables/useStudentDirectoryQuery'
import { useStudentFilters } from '@/composables/useStudentFilters'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassManagementRouteName,
  resolveStudentAnalysisRouteName,
} from '@/utils/teachingWorkspaceRouting'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

const router = useRouter()
const authStore = useAuthStore()

const classes = ref<TeacherClassItem[]>([])
const loadingClasses = ref(false)
const pageError = ref<string | null>(null)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const reportDialogVisible = ref(false)
const filters = useStudentFilters()
const studentDirectoryQuery = useStudentDirectoryQuery({
  debounceMs: 250,
  errorMessage: '加载学生列表失败，请稍后重试',
  request: getStudentsDirectory,
})

const { selectedClassName, searchQuery, studentNoQuery } = filters
const { students, total, loading: loadingStudents } = studentDirectoryQuery
const error = computed(() => pageError.value ?? studentDirectoryQuery.error.value)
const filteredTotal = computed(() => total.value)
const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredTotal.value / Math.max(pageSize.value, 1)))
)
const totalStudents = computed(() => {
  if (selectedClassName.value) {
    return (
      classes.value.find((item) => item.name === selectedClassName.value)?.student_count ??
      filteredTotal.value
    )
  }

  return classes.value.reduce((total, item) => total + (item.student_count || 0), 0)
})
const trimmedSearchQuery = computed(() => searchQuery.value.trim())
const searchLooksLikeStudentNo = computed(() =>
  Boolean(
    trimmedSearchQuery.value &&
    /\d/.test(trimmedSearchQuery.value) &&
    /^[A-Za-z0-9_-]+$/.test(trimmedSearchQuery.value)
  )
)
const directoryParams = computed(() => ({
  class_name: selectedClassName.value || undefined,
  keyword: searchLooksLikeStudentNo.value ? undefined : trimmedSearchQuery.value || undefined,
  student_no: searchLooksLikeStudentNo.value ? trimmedSearchQuery.value : undefined,
  sort_key: 'solved_count' as const,
  sort_order: 'desc' as const,
  page: page.value,
  page_size: pageSize.value,
}))

function resolvePreferredClass(): string {
  const teacherClassName = authStore.user?.class_name
  if (teacherClassName && classes.value.some((item) => item.name === teacherClassName)) {
    return teacherClassName
  }

  return ''
}

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
    const preferredClass = resolvePreferredClass()
    await selectClass(preferredClass)
  } catch (err) {
    console.error('初始化学生管理失败:', err)
    pageError.value = '加载学生管理失败，请稍后重试'
  }
}

function openStudent(studentId: string): void {
  const student = students.value.find((item) => item.id === studentId)
  router.push({
    name: resolveStudentAnalysisRouteName(authStore.user?.role),
    params: {
      className: selectedClassName.value || student?.class_name || '',
      studentId,
    },
  })
}

function updateSearchQuery(value: string): void {
  filters.updateSearchQuery(value)
  filters.updateStudentNoQuery('')
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

function openClassReportDialog(): void {
  reportDialogVisible.value = true
}

watch([searchQuery, studentNoQuery], () => {
  page.value = 1
  studentDirectoryQuery.scheduleLoadStudents({
    ...directoryParams.value,
    page: 1,
  })
})

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
  <section class="teacher-route-root">
    <StudentManagementPage
      :classes="classes"
      :selected-class-name="selectedClassName"
      :search-query="searchQuery"
      :student-no-query="studentNoQuery"
      :filtered-students="students"
      :filtered-total="filteredTotal"
      :total-students="totalStudents"
      :page="page"
      :total-pages="totalPages"
      :loading-classes="loadingClasses"
      :loading-students="loadingStudents"
      :error="error"
      @retry="initialize"
      @open-class-management="
        router.push({ name: resolveClassManagementRouteName(authStore.user?.role) })
      "
      @open-report-export="openClassReportDialog"
      @update-search-query="updateSearchQuery"
      @update-student-no-query="updateStudentNoQuery"
      @select-class="selectClass"
      @change-page="handlePageChange"
      @open-student="openStudent"
    />
    <TeacherClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="selectedClassName"
    />
  </section>
</template>
