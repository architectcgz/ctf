import type { TeacherClassTrendData, TeacherStudentItem } from '@/api/contracts'

interface BuildStudentInsightRowsOptions {
  riskStudentCount: number
  topStudent: TeacherStudentItem | null
  dominantWeakDimension: string
  strongestDimensionCount: number
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
  const { riskStudentCount, topStudent, dominantWeakDimension, strongestDimensionCount } = options
  const rows: TeacherDashboardInsightRow[] = []

  rows.push({
    key: 'risk',
    title:
      riskStudentCount > 0
        ? `风险组: ${riskStudentCount} 名学生近 7 天无训练动作`
        : '风险组: 当前没有连续掉线学生',
    chips:
      riskStudentCount > 0
        ? ['活跃断层', '建议优先回访', '影响完成率']
        : ['训练稳定', '保持观察'],
    detail:
      riskStudentCount > 0
        ? '这部分学生此前具备基础训练记录，但最近一周基本掉出训练节奏，优先补一组低门槛题目会比直接推高难题更有效。'
        : '班级当前没有明显的活跃断层，可以把更多精力放在薄弱维度补强和中段学生提升上。',
    status: riskStudentCount > 0 ? '高优先级' : '稳定',
    tone: riskStudentCount > 0 ? 'warning' : 'ready',
  })

  rows.push({
    key: 'strong',
    title: topStudent
      ? `进步组: ${topStudent.name || topStudent.username} 当前保持领先`
      : '进步组: 暂无头部样本',
    chips: topStudent ? ['头部样本', '可转入更高阶题单'] : ['等待样本'],
    detail: topStudent
      ? `${topStudent.name || topStudent.username} 当前累计 ${topStudent.solved_count ?? 0} 题、${topStudent.total_score ?? 0} 分，可作为班级示范样本继续拉动训练氛围。`
      : '还没有足够的学生表现数据用于识别班级示范样本。',
    status: topStudent ? '可推进' : '待观察',
    tone: topStudent ? 'ready' : 'warning',
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

  return rows
}

interface BuildPortraitSummaryNotesOptions {
  strongestDimensionCount: number
  dominantWeakDimension: string
  firstReviewTitle?: string
}

export function buildPortraitSummaryNotes(options: BuildPortraitSummaryNotesOptions) {
  const { strongestDimensionCount, dominantWeakDimension, firstReviewTitle } = options
  return [
    {
      key: 'impact',
      label: '影响学生',
      value: `${strongestDimensionCount} 人`,
      copy:
        dominantWeakDimension === '待观察'
          ? '等待能力画像形成'
          : `当前最集中暴露在 ${dominantWeakDimension} 方向`,
    },
    {
      key: 'action',
      label: '优先动作',
      value: firstReviewTitle || '补训题单',
      copy: '优先用结构化题单把低活跃学生重新拉回训练链路。',
    },
    {
      key: 'window',
      label: '观察窗口',
      value: '近 7 天',
      copy: '先观察活跃率与薄弱维度是否回升，再决定是否安排线下复盘。',
    },
  ]
}

export function buildTrendSignals(points: TeacherClassTrendData['points']) {
  const totalEvents = points.reduce((sum, point) => sum + point.event_count, 0)
  const totalSolves = points.reduce((sum, point) => sum + point.solve_count, 0)
  const peakActive = points.reduce((max, point) => Math.max(max, point.active_student_count), 0)
  const peakDay = points.reduce<TeacherClassTrendData['points'][number] | null>((current, point) => {
    if (!current || point.event_count > current.event_count) return point
    return current
  }, null)

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
}
