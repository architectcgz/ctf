interface BuildOverviewDescriptionOptions {
  selectedClassName: string
  hasSummary: boolean
  activeRateText: string
  averageSolvedText: string
  dominantWeakDimension: string
}

export function buildOverviewDescription(options: BuildOverviewDescriptionOptions): string {
  const {
    selectedClassName,
    hasSummary,
    activeRateText,
    averageSolvedText,
    dominantWeakDimension,
  } = options
  if (!selectedClassName) return '当前还没有可展示的班级。'
  if (!hasSummary) return '正在汇总班级近 7 天训练数据。'
  return `近 7 天活跃率 ${activeRateText}，人均解题 ${averageSolvedText}，重点关注 ${dominantWeakDimension} 方向与低活跃学生回流。`
}

interface BuildMetaPillsOptions {
  reviewItemCount: number
  weakDimensionCount: number
  riskStudentCount: number
}

export function buildMetaPills(options: BuildMetaPillsOptions): string[] {
  const { reviewItemCount, weakDimensionCount, riskStudentCount } = options
  return [
    '学习进度主导',
    reviewItemCount ? `${reviewItemCount} 条复盘结论` : '复盘结论待生成',
    weakDimensionCount ? `${weakDimensionCount} 个薄弱方向` : '能力画像待补全',
    riskStudentCount > 0 ? `${riskStudentCount} 名风险学生` : '班级趋势稳定',
  ]
}

interface BuildOverviewMetricsOptions {
  studentCountText: string | number
  activeStudentValue: string
  averageSolvedText: string
  recentEventCountText: string | number
}

export function buildOverviewMetrics(options: BuildOverviewMetricsOptions) {
  const { studentCountText, activeStudentValue, averageSolvedText, recentEventCountText } = options
  return [
    {
      key: 'student-count',
      label: '班级人数',
      value: studentCountText,
      hint: '当前纳入教学分析视图的学生样本',
    },
    {
      key: 'active-student',
      label: '活跃学生',
      value: activeStudentValue,
      hint: '近 7 天至少有一次训练动作的学生',
    },
    {
      key: 'average-solved',
      label: '平均解题',
      value: averageSolvedText,
      hint: '班级当前人均解题数',
    },
    {
      key: 'recent-event',
      label: '训练事件',
      value: recentEventCountText,
      hint: '提交、实例启动和销毁等行为总量',
    },
  ]
}

interface BuildInterventionTipsOptions {
  activeRate?: number
  firstReviewTitle?: string
  dominantWeakDimension: string
}

export function buildInterventionTips(options: BuildInterventionTipsOptions): string[] {
  const { activeRate, firstReviewTitle, dominantWeakDimension } = options
  const tips: string[] = []

  if (activeRate !== undefined) {
    tips.push(
      activeRate < 65
        ? '近 7 天活跃率偏低，优先安排低活跃学生进行补训。'
        : '班级活跃率稳定，建议继续维持节奏并聚焦薄弱维度。'
    )
  }

  if (firstReviewTitle) {
    tips.push(`优先执行「${firstReviewTitle}」对应的课堂动作。`)
  }

  if (dominantWeakDimension !== '待观察') {
    tips.push(`当前薄弱项集中在 ${dominantWeakDimension}，适合先布置该方向基础题。`)
  }

  if (tips.length === 0) {
    tips.push('完成本页排查后可直接导出报告，用于课后复盘与沟通。')
  }

  return tips.slice(0, 3)
}

export function buildTeachingAdvice(
  reviewItems: Array<{ title: string; detail: string }>,
  interventionTips: string[]
) {
  const fromReview = reviewItems.slice(0, 3).map((item) => ({
    title: item.title,
    detail: item.detail,
  }))

  if (fromReview.length > 0) return fromReview

  return interventionTips.map((tip, index) => ({
    title: `教学建议 ${index + 1}`,
    detail: tip,
  }))
}
