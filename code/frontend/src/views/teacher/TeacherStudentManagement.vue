<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getClassStudents } from '@/api/teacher'
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import StudentManagementPage from '@/components/teacher/student-management/StudentManagementPage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const classes = ref<TeacherClassItem[]>([])
const students = ref<TeacherStudentItem[]>([])
const selectedClassName = ref('')
const searchQuery = ref('')
const studentNoQuery = ref('')
const loadingClasses = ref(false)
const loadingStudents = ref(false)
const error = ref<string | null>(null)
let latestStudentRequestID = 0

async function loadClasses(): Promise<void> {
  loadingClasses.value = true
  try {
    classes.value = await getClasses()
  } finally {
    loadingClasses.value = false
  }
}

async function loadStudents(className: string): Promise<void> {
  if (!className) {
    latestStudentRequestID += 1
    students.value = []
    selectedClassName.value = ''
    loadingStudents.value = false
    return
  }

  const requestID = ++latestStudentRequestID
  loadingStudents.value = true
  error.value = null
  selectedClassName.value = className

  try {
    const nextStudents = await getClassStudents(className, {
      keyword: searchQuery.value.trim() || undefined,
      student_no: studentNoQuery.value.trim() || undefined,
    })
    if (requestID !== latestStudentRequestID) {
      return
    }
    students.value = nextStudents
  } catch (err) {
    if (requestID !== latestStudentRequestID) {
      return
    }
    console.error('加载学生列表失败:', err)
    error.value = '加载学生列表失败，请稍后重试'
    students.value = []
  } finally {
    if (requestID === latestStudentRequestID) {
      loadingStudents.value = false
    }
  }
}

async function initialize(): Promise<void> {
  error.value = null

  try {
    await loadClasses()
    const preferredClass = authStore.user?.class_name || classes.value[0]?.name || ''
    await loadStudents(preferredClass)
  } catch (err) {
    console.error('初始化学生管理失败:', err)
    error.value = '加载学生管理失败，请稍后重试'
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
  searchQuery.value = value
}

function updateStudentNoQuery(value: string): void {
  studentNoQuery.value = value
}

watch([searchQuery, studentNoQuery], async () => {
  if (!selectedClassName.value) return
  await loadStudents(selectedClassName.value)
})

onMounted(() => {
  initialize()
})
</script>

<template>
  <StudentManagementPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :search-query="searchQuery"
    :student-no-query="studentNoQuery"
    :filtered-students="students"
    :total-students="classes.find((item) => item.name === selectedClassName)?.student_count || students.length"
    :loading-classes="loadingClasses"
    :loading-students="loadingStudents"
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @update-search-query="updateSearchQuery"
    @update-student-no-query="updateStudentNoQuery"
    @select-class="loadStudents"
    @open-student="openStudent"
  />
</template>
