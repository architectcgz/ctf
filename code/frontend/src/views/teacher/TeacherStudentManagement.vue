<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import StudentManagementPage from '@/components/teacher/student-management/StudentManagementPage.vue'
import { useStudentFilters } from '@/composables/useStudentFilters'
import { useStudentListQuery } from '@/composables/useStudentListQuery'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const classes = ref<TeacherClassItem[]>([])
const loadingClasses = ref(false)
const pageError = ref<string | null>(null)
const filters = useStudentFilters()
const studentListQuery = useStudentListQuery({
  debounceMs: 250,
  errorMessage: '加载学生列表失败，请稍后重试',
  getParams: () => filters.studentQueryParams.value,
})

const { selectedClassName, searchQuery, studentNoQuery } = filters
const { students, loading: loadingStudents } = studentListQuery
const error = computed(() => pageError.value ?? studentListQuery.error.value)

async function loadClasses(): Promise<void> {
  loadingClasses.value = true
  try {
    classes.value = await getClasses()
  } finally {
    loadingClasses.value = false
  }
}

async function selectClass(className: string): Promise<void> {
  studentListQuery.cancelScheduledLoad()
  filters.updateSelectedClassName(className)
  await studentListQuery.loadStudents(className)
}

async function initialize(): Promise<void> {
  pageError.value = null

  try {
    await loadClasses()
    const preferredClass = authStore.user?.class_name || classes.value[0]?.name || ''
    await selectClass(preferredClass)
  } catch (err) {
    console.error('初始化学生管理失败:', err)
    pageError.value = '加载学生管理失败，请稍后重试'
  }
}

function openStudent(studentId: string): void {
  router.push({
    name: 'TeacherStudentAnalysis',
    params: {
      className: selectedClassName.value,
      studentId,
    },
  })
}

function updateSearchQuery(value: string): void {
  filters.updateSearchQuery(value)
}

function updateStudentNoQuery(value: string): void {
  filters.updateStudentNoQuery(value)
}

watch([searchQuery, studentNoQuery], () => {
  if (!selectedClassName.value) return
  studentListQuery.scheduleLoadStudents(selectedClassName.value)
})

onMounted(() => {
  void initialize()
})
</script>

<template>
  <StudentManagementPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :search-query="searchQuery"
    :student-no-query="studentNoQuery"
    :filtered-students="students"
    :total-students="
      classes.find((item) => item.name === selectedClassName)?.student_count || students.length
    "
    :loading-classes="loadingClasses"
    :loading-students="loadingStudents"
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @update-search-query="updateSearchQuery"
    @update-student-no-query="updateStudentNoQuery"
    @select-class="selectClass"
    @open-student="openStudent"
  />
</template>
