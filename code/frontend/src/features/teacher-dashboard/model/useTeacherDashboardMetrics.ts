import { computed, type Ref } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import {
  buildPortraitSummaryNotes,
  buildStudentInsightRows,
  buildTrendSignals,
} from './teacherDashboardInsightBuilders'
import {
  buildInterventionTips,
  buildMetaPills,
  buildOverviewDescription,
  buildOverviewMetrics,
  buildTeachingAdvice,
} from './teacherDashboardOverviewBuilders'

interface UseTeacherDashboardMetricsOptions {
  students: Ref<TeacherStudentItem[]>
  selectedClassName: Ref<string>
  selectedClass: Ref<TeacherClassItem | null>
  review: Ref<TeacherClassReviewData | null>
  summary: Ref<TeacherClassSummaryData | null>
  trend: Ref<TeacherClassTrendData | null>
}

interface WeakDimensionStat {
  dimension: string
  count: number
  width: string
}

export function useTeacherDashboardMetrics({
  students,
  selectedClassName,
  selectedClass,
  review,
  summary,
  trend,
}: UseTeacherDashboardMetricsOptions) {
  const averageSolvedText = computed(() => {
    if (!summary.value) return '--'
    return summary.value.average_solved.toFixed(1)
  })

  const activeRateText = computed(() => {
    if (!summary.value) return '--'
    return `${Math.round(summary.value.active_rate)}%`
  })

  const studentCountText = computed(
    () =>
      summary.value?.student_count || selectedClass.value?.student_count || students.value.length
  )

  const activeStudentCountText = computed(() => summary.value?.active_student_count ?? '--')
  const recentEventCountText = computed(() => summary.value?.recent_event_count ?? '--')
  const recentTrendPoints = computed(() => trend.value?.points ?? [])

  const weakDimensionStats = computed<WeakDimensionStat[]>(() => {
    const counter = new Map<string, number>()
    for (const student of students.value) {
      const key = student.weak_dimension?.trim()
      if (!key) continue
      counter.set(key, (counter.get(key) ?? 0) + 1)
    }

    const maxCount = Math.max(...counter.values(), 0)
    return Array.from(counter.entries())
      .map(([dimension, count]) => ({
        dimension,
        count,
        width: maxCount > 0 ? `${Math.round((count / maxCount) * 100)}%` : '0%',
      }))
      .sort((left, right) => {
        const countGap = right.count - left.count
        if (countGap !== 0) return countGap
        return left.dimension.localeCompare(right.dimension)
      })
  })

  const dominantWeakDimension = computed(() => weakDimensionStats.value[0]?.dimension ?? '待观察')

  const riskStudentCount = computed(
    () => students.value.filter((student) => (student.recent_event_count ?? 0) === 0).length
  )

  const activeStudentValue = computed(() => {
    if (activeStudentCountText.value !== '--') return String(activeStudentCountText.value)
    return `${students.value.filter((student) => (student.recent_event_count ?? 0) > 0).length}`
  })

  const topStudent = computed(
    () =>
      [...students.value]
        .sort((left, right) => {
          const solvedGap = (right.solved_count ?? 0) - (left.solved_count ?? 0)
          if (solvedGap !== 0) return solvedGap
          const scoreGap = (right.total_score ?? 0) - (left.total_score ?? 0)
          if (scoreGap !== 0) return scoreGap
          return (left.username || '').localeCompare(right.username || '')
        })
        .at(0) ?? null
  )

  const strongestDimensionCount = computed(() => weakDimensionStats.value[0]?.count ?? 0)

  const overviewDescription = computed(() =>
    buildOverviewDescription({
      selectedClassName: selectedClassName.value,
      hasSummary: Boolean(summary.value),
      activeRateText: activeRateText.value,
      averageSolvedText: averageSolvedText.value,
      dominantWeakDimension: dominantWeakDimension.value,
    })
  )

  const metaPills = computed(() =>
    buildMetaPills({
      reviewItemCount: review.value?.items.length ?? 0,
      weakDimensionCount: weakDimensionStats.value.length,
      riskStudentCount: riskStudentCount.value,
    })
  )

  const overviewMetrics = computed(() =>
    buildOverviewMetrics({
      studentCountText: studentCountText.value,
      activeStudentValue: activeStudentValue.value,
      averageSolvedText: averageSolvedText.value,
      recentEventCountText: recentEventCountText.value,
    })
  )

  const interventionTips = computed(() =>
    buildInterventionTips({
      activeRate: summary.value?.active_rate,
      firstReviewTitle: review.value?.items[0]?.title,
      dominantWeakDimension: dominantWeakDimension.value,
    })
  )

  const teachingAdvice = computed(() =>
    buildTeachingAdvice(review.value?.items || [], interventionTips.value)
  )

  const studentInsightRows = computed(() =>
    buildStudentInsightRows({
      riskStudentCount: riskStudentCount.value,
      topStudent: topStudent.value,
      dominantWeakDimension: dominantWeakDimension.value,
      strongestDimensionCount: strongestDimensionCount.value,
    })
  )

  const portraitSummaryNotes = computed(() =>
    buildPortraitSummaryNotes({
      strongestDimensionCount: strongestDimensionCount.value,
      dominantWeakDimension: dominantWeakDimension.value,
      firstReviewTitle: review.value?.items[0]?.title,
    })
  )

  const trendSignals = computed(() => buildTrendSignals(recentTrendPoints.value))

  return {
    averageSolvedText,
    activeRateText,
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
    overviewDescription,
    metaPills,
    overviewMetrics,
    interventionTips,
    teachingAdvice,
    studentInsightRows,
    portraitSummaryNotes,
    trendSignals,
  }
}
