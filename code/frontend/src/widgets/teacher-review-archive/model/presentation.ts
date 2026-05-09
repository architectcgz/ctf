import type { Component } from 'vue'
import { CheckCircle, Clock3, Trophy } from 'lucide-vue-next'

export const REVIEW_ARCHIVE_STATE_COPY = {
  errorTitle: '复盘归档加载失败',
  reload: '重新加载',
  emptyTitle: '暂无复盘归档',
  emptyDescription: '当前学生还没有可展示的复盘归档数据。',
} as const

export const REVIEW_ARCHIVE_SUMMARY_COPY = {
  summaryTitle: '训练摘要',
  summarySubtitle: '将当前归档的关键指标收束为一页课堂摘要。',
  solvedRateLabel: '完成率',
  solvedRateHintPrefix: '已完成',
  correctSubmissionLabel: '有效提交',
  correctSubmissionHint: '归档内命中 Flag 的提交次数',
  latestActivityLabel: '最近活跃',
  latestActivityHint: '归档内最后一条训练活动',
  skillTitle: '能力画像',
  skillSubtitle: '优先识别当前最强与最弱的训练维度。',
} as const

type ReviewArchiveSummaryCardTone = 'primary' | 'warning' | 'neutral'

export interface ReviewArchiveSummaryCardSchema {
  key: 'solved_rate' | 'correct_submission' | 'latest_activity'
  tone: ReviewArchiveSummaryCardTone
  label: string
  value: string | number
  hint: string
  valueClass?: string
  icon: Component
}

interface ReviewArchiveSkillDimensionItem {
  key: string
  name: string
  value: number
}

interface ReviewArchiveSummaryCardInput {
  solvedRate: number
  totalSolved: number
  totalChallenges: number
  correctSubmissionCount: number
  formattedLastActivity: string
}

export function buildReviewArchiveSummaryCards(
  input: ReviewArchiveSummaryCardInput
): ReviewArchiveSummaryCardSchema[] {
  return [
    {
      key: 'solved_rate',
      tone: 'primary',
      label: REVIEW_ARCHIVE_SUMMARY_COPY.solvedRateLabel,
      value: `${input.solvedRate}%`,
      hint: `${REVIEW_ARCHIVE_SUMMARY_COPY.solvedRateHintPrefix} ${input.totalSolved} / ${input.totalChallenges}`,
      icon: Trophy,
    },
    {
      key: 'correct_submission',
      tone: 'warning',
      label: REVIEW_ARCHIVE_SUMMARY_COPY.correctSubmissionLabel,
      value: input.correctSubmissionCount,
      hint: REVIEW_ARCHIVE_SUMMARY_COPY.correctSubmissionHint,
      icon: CheckCircle,
    },
    {
      key: 'latest_activity',
      tone: 'neutral',
      label: REVIEW_ARCHIVE_SUMMARY_COPY.latestActivityLabel,
      value: input.formattedLastActivity,
      hint: REVIEW_ARCHIVE_SUMMARY_COPY.latestActivityHint,
      valueClass: 'summary-card__value--time',
      icon: Clock3,
    },
  ]
}

export function rankReviewArchiveSkillDimensions(
  dimensions: ReviewArchiveSkillDimensionItem[]
): ReviewArchiveSkillDimensionItem[] {
  return [...dimensions].sort((left, right) => right.value - left.value)
}
