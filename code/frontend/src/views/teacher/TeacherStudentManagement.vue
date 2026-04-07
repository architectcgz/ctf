<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getClassStudents } from '@/api/teacher'
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import StudentManagementPage from '@/components/teacher/student-management/StudentManagementPage.vue'
import { useStudentFilters } from '@/composables/useStudentFilters'
import { useStudentListQuery } from '@/composables/useStudentListQuery'
import { useAuthStore } from '@/stores/auth'

const ALL_CLASSES_KEY = '__all_classes__'

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
  request: loadStudentsByClassFilter,
})

const { selectedClassName, searchQuery, studentNoQuery } = filters
const { students, loading: loadingStudents } = studentListQuery
const error = computed(() => pageError.value ?? studentListQuery.error.value)
const totalStudents = computed(() => {
  if (selectedClassName.value) {
    return classes.value.find((item) => item.name === selectedClassName.value)?.student_count ?? students.value.length
  }

  return classes.value.reduce((total, item) => total + (item.student_count || 0), 0)
})

function resolveClassFilterKey(className: string): string {
  return className || ALL_CLASSES_KEY
}

function resolvePreferredClass(): string {
  const teacherClassName = authStore.user?.class_name
  if (teacherClassName && classes.value.some((item) => item.name === teacherClassName)) {
    return teacherClassName
  }

  return ''
}

async function loadStudentsByClassFilter(
  classFilterKey: string,
  params?: { keyword?: string; student_no?: string }
): Promise<TeacherStudentItem[]> {
  if (classFilterKey !== ALL_CLASSES_KEY) {
    return getClassStudents(classFilterKey, params)
  }

  if (classes.value.length === 0) {
    return []
  }

  const studentGroups = await Promise.all(
    classes.value.map(async (item) => {
      const classStudents = await getClassStudents(item.name, params)
      return classStudents.map((student) => ({
        ...student,
        class_name: item.name,
      }))
    })
  )

  return studentGroups.flat()
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
  studentListQuery.cancelScheduledLoad()
  filters.updateSelectedClassName(className)
  await studentListQuery.loadStudents(resolveClassFilterKey(className))
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
    name: 'TeacherStudentAnalysis',
    params: {
      className: selectedClassName.value || student?.class_name || '',
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
  studentListQuery.scheduleLoadStudents(resolveClassFilterKey(selectedClassName.value))
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
    :total-students="totalStudents"
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
