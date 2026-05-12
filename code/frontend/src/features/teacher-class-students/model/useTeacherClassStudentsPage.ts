import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getClasses, getClassReview, getClassSummary, getClassTrend } from '@/api/teacher'
import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
} from '@/api/contracts'
import { useStudentFilters, useStudentListQuery } from '@/features/student-directory'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassManagementRouteName,
  resolveClassStudentsRouteName,
  resolveStudentAnalysisRouteName,
  resolveTeachingDashboardRouteName,
} from '@/utils/teachingWorkspaceRouting'

export function useTeacherClassStudentsPage() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()

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
      return { student_no }
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
      name: resolveClassStudentsRouteName(authStore.user?.role),
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
      name: resolveStudentAnalysisRouteName(authStore.user?.role),
      params: {
        className: selectedClassName.value,
        studentId,
      },
    })
  }

  function openClassManagement(): void {
    router.push({ name: resolveClassManagementRouteName(authStore.user?.role) })
  }

  function openDashboard(): void {
    router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })
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

  return {
    classes,
    selectedClassName,
    students,
    review,
    summary,
    trend,
    studentNoQuery,
    loadingStudents,
    error,
    reportDialogVisible,
    initialize,
    openClassManagement,
    openDashboard,
    openClassReportDialog,
    selectClass,
    updateStudentNoQuery,
    openStudent,
  }
}
