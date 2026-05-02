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

  const overviewDescription = computed(() => {
    if (!selectedClassName.value) return '当前还没有可展示的班级。'
    if (!summary.value) return '正在汇总班级近 7 天训练数据。'
    return `近 7 天活跃率 ${activeRateText.value}，人均解题 ${averageSolvedText.value}，重点关注 ${dominantWeakDimension.value} 方向与低活跃学生回流。`
  })

  const metaPills = computed(() => {
    const pills = [
      '学习进度主导',
      review.value?.items.length ? `${review.value.items.length} 条复盘结论` : '复盘结论待生成',
      weakDimensionStats.value.length
        ? `${weakDimensionStats.value.length} 个薄弱方向`
        : '能力画像待补全',
      riskStudentCount.value > 0 ? `${riskStudentCount.value} 名风险学生` : '班级趋势稳定',
    ]
    return pills
  })

  const overviewMetrics = computed(() => [
    {
      key: 'student-count',
      label: '班级人数',
      value: studentCountText.value,
      hint: '当前纳入教学分析视图的学生样本',
    },
    {
      key: 'active-student',
      label: '活跃学生',
      value: activeStudentValue.value,
      hint: '近 7 天至少有一次训练动作的学生',
    },
    {
      key: 'average-solved',
      label: '平均解题',
      value: averageSolvedText.value,
      hint: '班级当前人均解题数',
    },
    {
      key: 'recent-event',
      label: '训练事件',
      value: recentEventCountText.value,
      hint: '提交、实例启动和销毁等行为总量',
    },
  ])

  const interventionTips = computed(() => {
    const tips: string[] = []

    if (summary.value?.active_rate !== undefined) {
      tips.push(
        summary.value.active_rate < 65
          ? '近 7 天活跃率偏低，优先安排低活跃学生进行补训。'
          : '班级活跃率稳定，建议继续维持节奏并聚焦薄弱维度。'
      )
    }

    if (review.value?.items[0]?.title) {
      tips.push(`优先执行「${review.value.items[0].title}」对应的课堂动作。`)
    }

    if (dominantWeakDimension.value !== '待观察') {
      tips.push(`当前薄弱项集中在 ${dominantWeakDimension.value}，适合先布置该方向基础题。`)
    }

    if (tips.length === 0) {
      tips.push('完成本页排查后可直接导出报告，用于课后复盘与沟通。')
    }

    return tips.slice(0, 3)
  })

  const teachingAdvice = computed(() => {
    const fromReview =
      review.value?.items.slice(0, 3).map((item) => ({
        title: item.title,
        detail: item.detail,
      })) ?? []

    if (fromReview.length > 0) return fromReview

    return interventionTips.value.map((tip, index) => ({
      title: `教学建议 ${index + 1}`,
      detail: tip,
    }))
  })

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
