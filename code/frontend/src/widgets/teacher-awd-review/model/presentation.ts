export interface TeacherAwdReviewSummaryStats {
  roundCount: number
  teamCount: number
  serviceCount: number
  attackCount: number
  trafficCount: number
}

export interface TeacherAwdReviewSummaryItem {
  label: string
  value: string | number
  hint: string
  valueClass?: string
}

export const TEACHER_AWD_REVIEW_WORKSPACE_COPY = {
  overline: 'AWD Review',
  title: 'AWD复盘',
  descriptionSuffix:
    '多维复盘攻防实战过程。通过轮次下钻与流量回溯，协助教师评估学生的防御加固能力与漏洞挖掘表现。',
} as const

export function buildTeacherAwdReviewSummaryItems(
  summaryStats: TeacherAwdReviewSummaryStats,
  polling: boolean
): TeacherAwdReviewSummaryItem[] {
  return [
    {
      label: '轮次范围',
      value: summaryStats.roundCount,
      hint: '当前视图覆盖的轮次数量',
    },
    {
      label: '参与队伍',
      value: summaryStats.teamCount,
      hint: '当前视图包含的队伍数量',
    },
    {
      label: '服务 / 攻击 / 流量',
      value: `${summaryStats.serviceCount} / ${summaryStats.attackCount} / ${summaryStats.trafficCount}`,
      hint: '证据总量与服务运行信号',
    },
    {
      label: '导出状态',
      value: polling ? '后台处理中...' : '链路就绪',
      hint: '归档与教师报告导出链路状态',
      valueClass: 'awd-review-status-text',
    },
  ]
}
