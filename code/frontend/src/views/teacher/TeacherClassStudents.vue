<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getClasses, getClassReview, getClassSummary, getClassTrend } from '@/api/teacher'
import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
} from '@/api/contracts'
import ClassStudentsPage from '@/components/teacher/class-management/ClassStudentsPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useStudentFilters } from '@/composables/useStudentFilters'
import { useStudentListQuery } from '@/composables/useStudentListQuery'

const route = useRoute()
const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const review = ref<TeacherClassReviewData | null>(null)
const summary = ref<TeacherClassSummaryData | null>(null)
const trend = ref<TeacherClassTrendData | null>(null)
const workspaceError = ref<string | null>(null)
const reportDialogVisible = ref(false)
const filters = useStudentFilters()
const studentListQuery = useStudentListQuery({
  errorMessage: '加载班级学生失败，请稍后重试',
  getParams: () => {
    const { student_no } = filters.studentQueryParams.value
    return {
      student_no,
    }
  },
})

const { selectedClassName, studentNoQuery } = filters
const { students, loading: loadingStudents } = studentListQuery
const error = computed(() => workspaceError.value ?? studentListQuery.error.value)
let latestWorkspaceRequestID = 0

function classNameFromRoute(): string {
  return String(route.params.className || '')
}

async function loadClasses(): Promise<void> {
  try {
    classes.value = await getClasses()
  } catch (err) {
    console.error('加载班级列表失败:', err)
  }
}

function clearWorkspaceDetails(): void {
  review.value = null
  summary.value = null
  trend.value = null
}

async function loadWorkspaceDetails(className: string): Promise<void> {
  if (!className) {
    latestWorkspaceRequestID += 1
    clearWorkspaceDetails()
    return
  }

  const requestID = ++latestWorkspaceRequestID
  workspaceError.value = null

  try {
    const [nextReview, nextSummary, nextTrend] = await Promise.all([
      getClassReview(className),
      getClassSummary(className),
      getClassTrend(className),
    ])
    if (requestID !== latestWorkspaceRequestID) {
      return
    }
    review.value = nextReview
    summary.value = nextSummary
    trend.value = nextTrend
  } catch (err) {
    if (requestID !== latestWorkspaceRequestID) {
      return
    }
    console.error('加载班级详情失败:', err)
    workspaceError.value = '加载班级数据失败，请稍后重试'
    clearWorkspaceDetails()
  }
}

function updateStudentNoQuery(value: string): void {
  filters.updateStudentNoQuery(value)
}

function selectClass(className: string): void {
  if (!className || className === selectedClassName.value) {
    return
  }

  router.push({
    name: 'TeacherClassStudents',
    params: { className },
    query: route.query,
  })
}

async function loadClassWorkspace(className = classNameFromRoute()): Promise<void> {
  if (!className) {
    filters.updateSelectedClassName('')
    studentListQuery.cancelScheduledLoad()
    studentListQuery.clearStudents()
    clearWorkspaceDetails()
    return
  }

  filters.updateSelectedClassName(className)
  await Promise.all([studentListQuery.loadStudents(className), loadWorkspaceDetails(className)])
}

async function initialize(): Promise<void> {
  workspaceError.value = null

  try {
    await loadClasses()
    await loadClassWorkspace()
  } catch (err) {
    console.error('初始化班级学生页面失败:', err)
    workspaceError.value = '加载班级数据失败，请稍后重试'
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

function openClassReportDialog(): void {
  reportDialogVisible.value = true
}

watch(
  () => route.params.className,
  () => {
    void loadClassWorkspace()
  }
)

watch(studentNoQuery, () => {
  if (!selectedClassName.value) return
  studentListQuery.scheduleLoadStudents(selectedClassName.value)
})

onMounted(() => {
  void initialize()
})
</script>

<template>
  <ClassStudentsPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :students="students"
    :review="review"
    :summary="summary"
    :trend="trend"
    :student-no-query="studentNoQuery"
    :loading-students="loadingStudents"
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
    @open-report-export="openClassReportDialog"
    @select-class="selectClass"
    @update-student-no-query="updateStudentNoQuery"
    @open-student="openStudent"
  />
  <TeacherClassReportExportDialog
    v-model="reportDialogVisible"
    :default-class-name="selectedClassName"
  />
</template>
