import type {
  TeacherOverviewClassFocusData,
  TeacherOverviewWeakDimensionData,
  TeacherStudentItem,
} from '@/api/contracts'

interface BuildOverviewDescriptionOptions {
  classCount: number
  hasSummary: boolean
  activeRateText: string
  averageSolvedText: string
  dominantWeakDimension: string
  riskStudentCount: number
}

export function buildOverviewDescription(options: BuildOverviewDescriptionOptions): string {
  const {
    classCount,
    hasSummary,
    activeRateText,
    averageSolvedText,
    dominantWeakDimension,
    riskStudentCount,
  } = options
  if (classCount === 0) return '当前还没有可展示的教学范围。'
  if (!hasSummary) return '正在汇总当前教学范围近 7 天的训练数据。'

  const weakDimensionCopy =
    dominantWeakDimension === '待观察' ? '薄弱维度仍在形成中' : `重点关注 ${dominantWeakDimension} 方向`
  const riskCopy =
    riskStudentCount > 0 ? `${riskStudentCount} 名学生需要优先跟进` : '当前没有明显掉队学生'

  return `当前覆盖 ${classCount} 个班级，近 7 天活跃率 ${activeRateText}，人均解题 ${averageSolvedText}；${weakDimensionCopy}，${riskCopy}。`
}

interface BuildMetaPillsOptions {
  classCount: number
  weakDimensionCount: number
  riskStudentCount: number
  focusStudentCount: number
}

export function buildMetaPills(options: BuildMetaPillsOptions): string[] {
  const { classCount, weakDimensionCount, riskStudentCount, focusStudentCount } = options
  return [
    classCount ? `${classCount} 个班级纳入总览` : '教学范围待建立',
    weakDimensionCount ? `${weakDimensionCount} 个薄弱方向` : '能力画像待补全',
    riskStudentCount > 0 ? `${riskStudentCount} 名待跟进学生` : '当前范围整体稳定',
    focusStudentCount > 0 ? `${focusStudentCount} 名优先介入对象` : '介入名单待生成',
  ]
}

interface BuildOverviewMetricsOptions {
  classCountText: string | number
  studentCountText: string | number
  activeStudentValue: string
  recentEventCountText: string | number
}

export function buildOverviewMetrics(options: BuildOverviewMetricsOptions) {
  const { classCountText, studentCountText, activeStudentValue, recentEventCountText } = options
  return [
    {
      key: 'class-count',
      label: '教学班级',
      value: classCountText,
      hint: '当前纳入教学概览的班级范围',
    },
    {
      key: 'student-count',
      label: '覆盖学生',
      value: studentCountText,
      hint: '当前 scope 内纳入统计的学生样本',
    },
    {
      key: 'active-student',
      label: '活跃学生',
      value: activeStudentValue,
      hint: '近 7 天至少有一次训练动作的学生',
    },
    {
      key: 'recent-event',
      label: '训练事件',
      value: recentEventCountText,
      hint: '提交、实例启动和销毁等行为总量',
    },
  ]
}

export interface TeacherDashboardReviewHighlight {
  key: string
  title: string
  detail: string
  chips: string[]
  tone: 'ready' | 'warning' | 'danger'
}

export function buildReviewHighlights(
  focusClasses: TeacherOverviewClassFocusData[],
  weakDimensions: TeacherOverviewWeakDimensionData[]
): TeacherDashboardReviewHighlight[] {
  const rows: TeacherDashboardReviewHighlight[] = focusClasses.slice(0, 4).map((item) => {
    const riskCopy =
      item.risk_student_count > 0
        ? `${item.class_name} 当前有 ${item.risk_student_count} 名学生需要优先跟进。`
        : `${item.class_name} 当前没有明显掉队学生，可继续维持训练节奏。`
    const weakCopy = item.dominant_weak_dimension
      ? `薄弱维度集中在 ${item.dominant_weak_dimension}。`
      : '薄弱维度还没有形成稳定集中。'

    return {
      key: item.class_name,
      title: `${item.class_name} 复盘摘要`,
      detail: `${riskCopy}${weakCopy}`,
      chips: [
        `活跃率 ${Math.round(item.active_rate)}%`,
        `近 7 天 ${item.recent_event_count} 次事件`,
        item.dominant_weak_dimension || '薄弱维度待观察',
      ],
      tone:
        item.risk_student_count > 0 ? 'danger' : item.active_rate < 65 ? 'warning' : 'ready',
    }
  })

  if (rows.length > 0) {
    return rows
  }

  if (weakDimensions.length > 0) {
    const first = weakDimensions[0]
    return [
      {
        key: first.dimension,
        title: `${first.dimension} 是当前主要薄弱维度`,
        detail: `当前共有 ${first.student_count} 名学生在该方向暴露薄弱信号，适合先安排基础题或统一讲解。`,
        chips: [`${first.student_count} 名学生命中`, '建议统一补强'],
        tone: 'warning',
      },
    ]
  }

  return []
}

export interface TeacherDashboardInterventionTarget {
  id: string
  title: string
  detail: string
  meta: string[]
}

export function buildInterventionTargets(
  students: TeacherStudentItem[]
): TeacherDashboardInterventionTarget[] {
  return students.slice(0, 6).map((student) => {
    const displayName = student.name || student.username
    return {
      id: String(student.id),
      title: displayName,
      detail: `近 7 天 ${student.recent_event_count ?? 0} 次训练动作，累计 ${student.solved_count ?? 0} 题 / ${student.total_score ?? 0} 分。`,
      meta: [
        student.class_name || '班级待识别',
        student.weak_dimension ? `薄弱项 ${student.weak_dimension}` : '薄弱项待观察',
      ],
    }
  })
}
