import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getClassReview, getClassSummary, getClassTrend } from '@/api/teacher'
import type {
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
} from '@/api/contracts'
import {
  buildTeacherClassInsightWindowQuery,
  describeTeacherClassInsightWindow,
  getTeacherClassInsightWindowError,
  isSameTeacherClassInsightWindow,
  parseTeacherClassInsightWindowQuery,
  hasTeacherClassInsightWindow,
} from '@/features/teacher-class-insight-window/model/window'
import { useStudentFilters, useStudentListQuery } from '@/features/student-directory'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassManagementRouteName,
  resolveStudentAnalysisRouteName,
  resolveTeachingDashboardRouteName,
} from '@/utils/teachingWorkspaceRouting'

export function useTeacherClassStudentsPage() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()

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
  const activeInsightWindow = computed(() => parseTeacherClassInsightWindowQuery(route.query))
  const insightWindowDraft = ref(parseTeacherClassInsightWindowQuery(route.query))
  const insightWindowError = computed(() =>
    getTeacherClassInsightWindowError(insightWindowDraft.value)
  )
  const insightWindowLabel = computed(() =>
    describeTeacherClassInsightWindow(activeInsightWindow.value)
  )
  const canApplyInsightWindow = computed(() => {
    if (insightWindowError.value) {
      return false
    }
    return !isSameTeacherClassInsightWindow(insightWindowDraft.value, activeInsightWindow.value)
  })
  const canResetInsightWindow = computed(
    () =>
      hasTeacherClassInsightWindow(insightWindowDraft.value) ||
      hasTeacherClassInsightWindow(activeInsightWindow.value)
  )
  let latestWorkspaceRequestID = 0

  function classNameFromRoute(): string {
    return String(route.params.className || '')
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
    const routeInsightWindow = parseTeacherClassInsightWindowQuery(route.query)
    const routeInsightWindowError = getTeacherClassInsightWindowError(routeInsightWindow)
    if (routeInsightWindowError) {
      workspaceError.value = routeInsightWindowError
      clearWorkspaceDetails()
      return
    }
    const insightWindowQuery = buildTeacherClassInsightWindowQuery(routeInsightWindow)

    try {
      const [nextReview, nextSummary, nextTrend] = await Promise.all([
        insightWindowQuery
          ? getClassReview(className, insightWindowQuery)
          : getClassReview(className),
        insightWindowQuery
          ? getClassSummary(className, insightWindowQuery)
          : getClassSummary(className),
        insightWindowQuery
          ? getClassTrend(className, insightWindowQuery)
          : getClassTrend(className),
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

  function updateInsightWindowFromDate(value: string): void {
    insightWindowDraft.value = {
      ...insightWindowDraft.value,
      fromDate: value.trim(),
    }
  }

  function updateInsightWindowToDate(value: string): void {
    insightWindowDraft.value = {
      ...insightWindowDraft.value,
      toDate: value.trim(),
    }
  }

  async function applyInsightWindow(): Promise<void> {
    if (insightWindowError.value) {
      return
    }

    const nextQuery = { ...route.query }
    const nextInsightWindow = buildTeacherClassInsightWindowQuery(insightWindowDraft.value)
    if (nextInsightWindow) {
      nextQuery.from_date = nextInsightWindow.from_date
      nextQuery.to_date = nextInsightWindow.to_date
    } else {
      delete nextQuery.from_date
      delete nextQuery.to_date
    }

    if (
      String(route.query.from_date || '') === String(nextQuery.from_date || '') &&
      String(route.query.to_date || '') === String(nextQuery.to_date || '')
    ) {
      return
    }

    await router.replace({ query: nextQuery })
  }

  async function resetInsightWindow(): Promise<void> {
    insightWindowDraft.value = {
      fromDate: '',
      toDate: '',
    }

    if (!hasTeacherClassInsightWindow(activeInsightWindow.value)) {
      return
    }

    const nextQuery = { ...route.query }
    delete nextQuery.from_date
    delete nextQuery.to_date
    await router.replace({ query: nextQuery })
  }

  watch(
    () => [route.params.className, route.query.from_date, route.query.to_date] as const,
    () => {
      insightWindowDraft.value = parseTeacherClassInsightWindowQuery(route.query)
      void loadClassWorkspace()
    },
    { immediate: true }
  )

  watch(studentNoQuery, () => {
    if (!selectedClassName.value) return
    studentListQuery.scheduleLoadStudents(selectedClassName.value)
  })

  return {
    selectedClassName,
    students,
    review,
    summary,
    trend,
    studentNoQuery,
    loadingStudents,
    error,
    reportDialogVisible,
    insightWindowFromDate: computed(() => insightWindowDraft.value.fromDate),
    insightWindowToDate: computed(() => insightWindowDraft.value.toDate),
    insightWindowError,
    insightWindowLabel,
    activeInsightWindowFromDate: computed(() => activeInsightWindow.value.fromDate),
    activeInsightWindowToDate: computed(() => activeInsightWindow.value.toDate),
    canApplyInsightWindow,
    canResetInsightWindow,
    initialize,
    openClassManagement,
    openDashboard,
    openClassReportDialog,
    updateStudentNoQuery,
    updateInsightWindowFromDate,
    updateInsightWindowToDate,
    applyInsightWindow,
    resetInsightWindow,
    openStudent,
  }
}
