import type {
  TeacherOverviewClassFocusData,
  TeacherOverviewTrendData,
  TeacherStudentItem,
} from '@/api/contracts'

interface BuildStudentInsightRowsOptions {
  riskStudentCount: number
  spotlightStudent: TeacherStudentItem | null
  dominantWeakDimension: string
  strongestDimensionCount: number
  focusClass: TeacherOverviewClassFocusData | null
}

export interface TeacherDashboardInsightRow {
  key: string
  title: string
  chips: string[]
  detail: string
  status: string
  tone: 'ready' | 'warning' | 'danger'
}

export function buildStudentInsightRows(
  options: BuildStudentInsightRowsOptions
): TeacherDashboardInsightRow[] {
  const {
    riskStudentCount,
    spotlightStudent,
    dominantWeakDimension,
    strongestDimensionCount,
    focusClass,
  } = options
  const rows: TeacherDashboardInsightRow[] = []

  rows.push({
    key: 'risk',
    title:
      riskStudentCount > 0
        ? `风险组: ${riskStudentCount} 名学生处于低活跃或低进度区间`
        : '风险组: 当前没有明显掉队学生',
    chips:
      riskStudentCount > 0
        ? ['优先回访', '训练回流', '影响完成率']
        : ['训练稳定', '保持观察'],
    detail:
      riskStudentCount > 0
        ? '这部分学生需要先拉回训练链路，再决定是否继续推进难度；先补一组低门槛题目通常比直接加压更有效。'
        : '当前教学范围内没有明显掉队样本，可以把更多注意力放在薄弱维度补强和重点班级观察上。',
    status: riskStudentCount > 0 ? '高优先级' : '稳定',
    tone: riskStudentCount > 0 ? 'warning' : 'ready',
  })

  rows.push({
    key: 'strong',
    title: spotlightStudent
      ? `头部样本: ${spotlightStudent.name || spotlightStudent.username} 当前保持领先`
      : '进步组: 暂无头部样本',
    chips: spotlightStudent ? ['头部样本', '可转入更高阶题单'] : ['等待样本'],
    detail: spotlightStudent
      ? `${spotlightStudent.name || spotlightStudent.username} 当前累计 ${spotlightStudent.solved_count ?? 0} 题、${spotlightStudent.total_score ?? 0} 分，可作为课堂示范样本继续拉动训练氛围。`
      : '还没有足够的学生表现数据用于识别班级示范样本。',
    status: spotlightStudent ? '可推进' : '待观察',
    tone: spotlightStudent ? 'ready' : 'warning',
  })

  rows.push({
    key: 'weak',
    title:
      dominantWeakDimension !== '待观察'
        ? `${dominantWeakDimension} 方向仍是当前薄弱维度`
        : '薄弱维度尚未形成稳定分布',
    chips:
      dominantWeakDimension !== '待观察'
        ? ['薄弱维度', '建议统一讲解']
        : ['画像不足', '继续采样'],
    detail:
      dominantWeakDimension !== '待观察'
        ? `当前已有 ${strongestDimensionCount} 名学生在 ${dominantWeakDimension} 方向暴露薄弱项，说明问题更接近方法迁移断层，不只是练习数量不够。`
        : '当前班级还没有足够的弱项样本，暂不建议过早做统一干预。',
    status: dominantWeakDimension !== '待观察' ? '需介入' : '待观察',
    tone: dominantWeakDimension !== '待观察' ? 'danger' : 'warning',
  })

  rows.push({
    key: 'class-focus',
    title: focusClass
      ? `${focusClass.class_name} 是当前最值得观察的班级`
      : '重点班级仍在形成中',
    chips: focusClass
      ? [
          `${focusClass.risk_student_count} 名待跟进`,
          `活跃率 ${Math.round(focusClass.active_rate)}%`,
        ]
      : ['等待样本'],
    detail: focusClass
      ? `${focusClass.class_name} 近 7 天共有 ${focusClass.recent_event_count} 次训练事件，${focusClass.dominant_weak_dimension || '薄弱维度尚未集中'} 方向最值得继续下钻。`
      : '当前还没有足够的范围数据用于锁定重点班级。',
    status: focusClass ? '需观察' : '待观察',
    tone: focusClass ? 'warning' : 'ready',
  })

  return rows
}

interface BuildPortraitSummaryNotesOptions {
  classCount: number
  strongestDimensionCount: number
  dominantWeakDimension: string
  focusClassName?: string
  riskStudentCount: number
}

export function buildPortraitSummaryNotes(options: BuildPortraitSummaryNotesOptions) {
  const { classCount, strongestDimensionCount, dominantWeakDimension, focusClassName, riskStudentCount } =
    options
  return [
    {
      key: 'scope',
      label: '覆盖班级',
      value: `${classCount} 个`,
      copy: focusClassName ? `当前最值得下钻的是 ${focusClassName}` : '等待重点班级浮现',
    },
    {
      key: 'impact',
      label: '影响学生',
      value: `${strongestDimensionCount} 人`,
      copy:
        dominantWeakDimension === '待观察'
          ? '等待薄弱维度形成'
          : `当前最集中暴露在 ${dominantWeakDimension} 方向`,
    },
    {
      key: 'risk',
      label: '待跟进学生',
      value: `${riskStudentCount} 人`,
      copy: '先把掉队学生拉回训练链路，再安排统一讲解或专项题单。',
    },
  ]
}

export function buildTrendSignals(points: TeacherOverviewTrendData['points']) {
  const totalEvents = points.reduce((sum, point) => sum + point.event_count, 0)
  const totalSolves = points.reduce((sum, point) => sum + point.solve_count, 0)
  const peakActive = points.reduce((max, point) => Math.max(max, point.active_student_count), 0)
  const peakDay = points.reduce<TeacherOverviewTrendData['points'][number] | null>(
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
        ? '把训练事件和解题转化放在同一条时间轴上观察，更容易判断课堂动作是否真正起效。'
        : '等待成功解题数据形成趋势。',
    },
    {
      key: 'active',
      label: '活跃波动',
      value: peakActive ? `${peakActive} 人` : '--',
      copy: peakActive
        ? '当前范围内活跃人数没有明显断崖，但尾部学生的参与深度仍然偏弱。'
        : '当前还无法判断活跃波动。',
    },
  ]
}
