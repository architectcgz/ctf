import { computed, type Ref } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'

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
    () => summary.value?.student_count || selectedClass.value?.student_count || students.value.length
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

  const overviewPulseCards = computed(() => [
    {
      key: 'active-rate',
      label: '活跃率',
      value: activeRateText.value,
      copy: summary.value ? '近 7 天班级训练参与强度' : '等待班级训练汇总',
    },
    {
      key: 'risk',
      label: '风险学生',
      value: `${riskStudentCount.value}`,
      copy: riskStudentCount.value > 0 ? '连续 7 天无训练动作' : '暂无明显训练断层',
    },
    {
      key: 'weak',
      label: '薄弱方向',
      value: dominantWeakDimension.value,
      copy:
        strongestDimensionCount.value > 0
          ? `${strongestDimensionCount.value} 名学生集中暴露该薄弱项`
          : '等待能力画像形成',
    },
    {
      key: 'review',
      label: '建议动作',
      value: review.value?.items[0]?.title || '继续观察',
      copy: review.value?.items[0]?.detail || '可先查看趋势与学生洞察，再决定课堂动作。',
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

  const studentInsightRows = computed(() => {
    const rows: Array<{
      key: string
      title: string
      chips: string[]
      detail: string
      status: string
      tone: 'ready' | 'warning' | 'danger'
    }> = []

    rows.push({
      key: 'risk',
      title:
        riskStudentCount.value > 0
          ? `风险组: ${riskStudentCount.value} 名学生近 7 天无训练动作`
          : '风险组: 当前没有连续掉线学生',
      chips:
        riskStudentCount.value > 0
          ? ['活跃断层', '建议优先回访', '影响完成率']
          : ['训练稳定', '保持观察'],
      detail:
        riskStudentCount.value > 0
          ? '这部分学生此前具备基础训练记录，但最近一周基本掉出训练节奏，优先补一组低门槛题目会比直接推高难题更有效。'
          : '班级当前没有明显的活跃断层，可以把更多精力放在薄弱维度补强和中段学生提升上。',
      status: riskStudentCount.value > 0 ? '高优先级' : '稳定',
      tone: riskStudentCount.value > 0 ? 'warning' : 'ready',
    })

    rows.push({
      key: 'strong',
      title: topStudent.value
        ? `进步组: ${topStudent.value.name || topStudent.value.username} 当前保持领先`
        : '进步组: 暂无头部样本',
      chips: topStudent.value ? ['头部样本', '可转入更高阶题单'] : ['等待样本'],
      detail: topStudent.value
        ? `${topStudent.value.name || topStudent.value.username} 当前累计 ${topStudent.value.solved_count ?? 0} 题、${topStudent.value.total_score ?? 0} 分，可作为班级示范样本继续拉动训练氛围。`
        : '还没有足够的学生表现数据用于识别班级示范样本。',
      status: topStudent.value ? '可推进' : '待观察',
      tone: topStudent.value ? 'ready' : 'warning',
    })

    rows.push({
      key: 'weak',
      title:
        dominantWeakDimension.value !== '待观察'
          ? `${dominantWeakDimension.value} 方向仍是当前薄弱维度`
          : '薄弱维度尚未形成稳定分布',
      chips:
        dominantWeakDimension.value !== '待观察'
          ? ['薄弱维度', '建议统一讲解']
          : ['画像不足', '继续采样'],
      detail:
        dominantWeakDimension.value !== '待观察'
          ? `当前已有 ${strongestDimensionCount.value} 名学生在 ${dominantWeakDimension.value} 方向暴露薄弱项，说明问题更接近方法迁移断层，不只是练习数量不够。`
          : '当前班级还没有足够的弱项样本，暂不建议过早做统一干预。',
      status: dominantWeakDimension.value !== '待观察' ? '需介入' : '待观察',
      tone: dominantWeakDimension.value !== '待观察' ? 'danger' : 'warning',
    })

    return rows
  })

  const portraitSummaryNotes = computed(() => [
    {
      key: 'impact',
      label: '影响学生',
      value: `${strongestDimensionCount.value} 人`,
      copy:
        dominantWeakDimension.value === '待观察'
          ? '等待能力画像形成'
          : `当前最集中暴露在 ${dominantWeakDimension.value} 方向`,
    },
    {
      key: 'action',
      label: '优先动作',
      value: review.value?.items[0]?.title || '补训题单',
      copy: '优先用结构化题单把低活跃学生重新拉回训练链路。',
    },
    {
      key: 'window',
      label: '观察窗口',
      value: '近 7 天',
      copy: '先观察活跃率与薄弱维度是否回升，再决定是否安排线下复盘。',
    },
  ])

  const trendSignals = computed(() => {
    const points = recentTrendPoints.value
    const totalEvents = points.reduce((sum, point) => sum + point.event_count, 0)
    const totalSolves = points.reduce((sum, point) => sum + point.solve_count, 0)
    const peakActive = points.reduce((max, point) => Math.max(max, point.active_student_count), 0)
    const peakDay = points.reduce<TeacherClassTrendData['points'][number] | null>(
      (current, point) => {
        if (!current || point.event_count > current.event_count) return point
        return current
      },
      null
    )

    return [
      {
        key: 'events',
        label: '事件总量',
        value: totalEvents ? `${totalEvents}` : '--',
        copy: peakDay
          ? `峰值出现在 ${peakDay.date.slice(5)}，训练事件达到 ${peakDay.event_count}。`
          : '当前还没有趋势数据。',
      },
      {
        key: 'solves',
        label: '成功解题',
        value: totalSolves ? `${totalSolves}` : '--',
        copy: totalSolves
          ? '把训练事件和解题转化放在同一条时间轴上观察更容易定位断层。'
          : '等待成功解题数据形成趋势。',
      },
      {
        key: 'active',
        label: '活跃波动',
        value: peakActive ? `${peakActive} 人` : '--',
        copy: peakActive
          ? '班级没有明显断崖，但尾部学生的参与深度仍然偏弱。'
          : '当前还无法判断活跃波动。',
      },
    ]
  })

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
    overviewPulseCards,
    interventionTips,
    teachingAdvice,
    studentInsightRows,
    portraitSummaryNotes,
    trendSignals,
  }
}
