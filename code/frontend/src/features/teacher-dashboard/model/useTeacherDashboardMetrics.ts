import { computed, type Ref } from 'vue'

import type { TeacherOverviewData, TeacherStudentItem } from '@/api/contracts'
import {
  buildPortraitSummaryNotes,
  buildStudentInsightRows,
  buildTrendSignals,
} from './teacherDashboardInsightBuilders'
import {
  buildInterventionTargets,
  buildMetaPills,
  buildOverviewDescription,
  buildOverviewMetrics,
  buildReviewHighlights,
} from './teacherDashboardOverviewBuilders'

interface UseTeacherDashboardMetricsOptions {
  overview: Ref<TeacherOverviewData | null>
}

interface WeakDimensionStat {
  dimension: string
  count: number
  width: string
}

export function useTeacherDashboardMetrics({ overview }: UseTeacherDashboardMetricsOptions) {
  const summary = computed(() => overview.value?.summary ?? null)
  const focusClasses = computed(() => overview.value?.focus_classes ?? [])
  const focusStudents = computed(() => overview.value?.focus_students ?? [])
  const spotlightStudent = computed(() => overview.value?.spotlight_student ?? null)

  const averageSolvedText = computed(() =>
    summary.value ? summary.value.average_solved.toFixed(1) : '--'
  )

  const activeRateText = computed(() =>
    summary.value ? `${Math.round(summary.value.active_rate)}%` : '--'
  )

  const classCountText = computed(() => summary.value?.class_count ?? '--')
  const studentCountText = computed(() => summary.value?.student_count ?? '--')
  const activeStudentCountText = computed(() => summary.value?.active_student_count ?? '--')
  const recentEventCountText = computed(() => summary.value?.recent_event_count ?? '--')
  const recentTrendPoints = computed(() => overview.value?.trend.points ?? [])

  const weakDimensionStats = computed<WeakDimensionStat[]>(() => {
    const source = overview.value?.weak_dimensions ?? []
    const maxCount = Math.max(...source.map((item) => item.student_count), 0)
    return source.map((item) => ({
      dimension: item.dimension,
      count: item.student_count,
      width: maxCount > 0 ? `${Math.round((item.student_count / maxCount) * 100)}%` : '0%',
    }))
  })

  const dominantWeakDimension = computed(() => weakDimensionStats.value[0]?.dimension ?? '待观察')
  const strongestDimensionCount = computed(() => weakDimensionStats.value[0]?.count ?? 0)
  const riskStudentCount = computed(
    () => summary.value?.risk_student_count ?? focusStudents.value.length
  )

  const activeStudentValue = computed(() => {
    if (activeStudentCountText.value !== '--') return String(activeStudentCountText.value)
    return '--'
  })

  const topStudent = computed<TeacherStudentItem | null>(() => spotlightStudent.value ?? null)

  const overviewDescription = computed(() =>
    buildOverviewDescription({
      classCount: summary.value?.class_count ?? 0,
      hasSummary: Boolean(summary.value),
      activeRateText: activeRateText.value,
      averageSolvedText: averageSolvedText.value,
      dominantWeakDimension: dominantWeakDimension.value,
      riskStudentCount: riskStudentCount.value,
    })
  )

  const metaPills = computed(() =>
    buildMetaPills({
      classCount: summary.value?.class_count ?? 0,
      weakDimensionCount: weakDimensionStats.value.length,
      riskStudentCount: riskStudentCount.value,
      focusStudentCount: focusStudents.value.length,
    })
  )

  const overviewMetrics = computed(() =>
    buildOverviewMetrics({
      classCountText: classCountText.value,
      studentCountText: studentCountText.value,
      activeStudentValue: activeStudentValue.value,
      recentEventCountText: recentEventCountText.value,
    })
  )

  const reviewHighlights = computed(() =>
    buildReviewHighlights(focusClasses.value, overview.value?.weak_dimensions ?? [])
  )

  const interventionTargets = computed(() => buildInterventionTargets(focusStudents.value))

  const studentInsightRows = computed(() =>
    buildStudentInsightRows({
      riskStudentCount: riskStudentCount.value,
      spotlightStudent: topStudent.value,
      dominantWeakDimension: dominantWeakDimension.value,
      strongestDimensionCount: strongestDimensionCount.value,
      focusClass: focusClasses.value[0] ?? null,
    })
  )

  const portraitSummaryNotes = computed(() =>
    buildPortraitSummaryNotes({
      classCount: summary.value?.class_count ?? 0,
      strongestDimensionCount: strongestDimensionCount.value,
      dominantWeakDimension: dominantWeakDimension.value,
      focusClassName: focusClasses.value[0]?.class_name,
      riskStudentCount: riskStudentCount.value,
    })
  )

  const trendSignals = computed(() => buildTrendSignals(recentTrendPoints.value))

  return {
    averageSolvedText,
    activeRateText,
    classCountText,
    studentCountText,
    activeStudentCountText,
    recentEventCountText,
    recentTrendPoints,
    weakDimensionStats,
    dominantWeakDimension,
    riskStudentCount,
    activeStudentValue,
    topStudent,
    strongestDimensionCount,
    focusClasses,
    focusStudents,
    overviewDescription,
    metaPills,
    overviewMetrics,
    reviewHighlights,
    interventionTargets,
    studentInsightRows,
    portraitSummaryNotes,
    trendSignals,
  }
}
