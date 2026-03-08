<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getClasses, getClassStudents } from '@/api/teacher'
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import ClassStudentsPage from '@/components/teacher/class-management/ClassStudentsPage.vue'

const route = useRoute()
const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const students = ref<TeacherStudentItem[]>([])
const selectedClassName = ref('')
const loadingStudents = ref(false)
const error = ref<string | null>(null)
const studentNoQuery = ref('')

function classNameFromRoute(): string {
  return decodeURIComponent(String(route.params.className || ''))
}

async function loadClasses(): Promise<void> {
  try {
    classes.value = await getClasses()
  } catch (err) {
    console.error('加载班级列表失败:', err)
  }
}

async function loadStudents(className = classNameFromRoute()): Promise<void> {
  if (!className) {
    students.value = []
    selectedClassName.value = ''
    return
  }

  loadingStudents.value = true
  error.value = null
  selectedClassName.value = className

  try {
    students.value = await getClassStudents(className, {
      student_no: studentNoQuery.value.trim() || undefined,
    })
  } catch (err) {
    console.error('加载班级学生失败:', err)
    error.value = '加载班级学生失败，请稍后重试'
    students.value = []
  } finally {
    loadingStudents.value = false
  }
}

function updateStudentNoQuery(value: string): void {
  studentNoQuery.value = value
}

async function initialize(): Promise<void> {
  error.value = null

  try {
    await loadClasses()
    await loadStudents()
  } catch (err) {
    console.error('初始化班级学生页面失败:', err)
    error.value = '加载班级数据失败，请稍后重试'
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

watch(
  () => route.params.className,
  () => {
    studentNoQuery.value = ''
    loadStudents()
  }
)

watch(studentNoQuery, async () => {
  await loadStudents()
})

onMounted(() => {
  initialize()
})
</script>

<template>
  <ClassStudentsPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :students="students"
    :student-no-query="studentNoQuery"
    :loading-students="loadingStudents"
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @update-student-no-query="updateStudentNoQuery"
    @open-student="openStudent"
  />
</template>
