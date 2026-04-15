import { computed, proxyRefs, type Ref } from 'vue'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'

export type ContestWorkbenchStageKey =
  | 'basics'
  | 'pool'
  | 'awd-config'
  | 'preflight'
  | 'operations'

export const CONTEST_WORKBENCH_STAGE_ORDER: ContestWorkbenchStageKey[] = [
  'basics',
  'pool',
  'awd-config',
  'preflight',
  'operations',
]

export interface ContestWorkbenchStage {
  key: ContestWorkbenchStageKey
  label: string
  disabled?: boolean
}

export interface ContestWorkbenchSummaryItem {
  key: string
  label: string
  value: string
  hint?: string
}

const BASE_STAGES: ContestWorkbenchStage[] = [
  { key: 'basics', label: '基础信息' },
  { key: 'pool', label: '题目池' },
]

const AWD_STAGES: ContestWorkbenchStage[] = [
  ...BASE_STAGES,
  { key: 'awd-config', label: 'AWD 配置' },
  { key: 'preflight', label: '赛前检查' },
  { key: 'operations', label: '轮次运行' },
]

const CONTEST_MODE_LABELS: Record<'jeopardy' | 'awd', string> = {
  jeopardy: 'Jeopardy',
  awd: 'AWD',
}

const CONTEST_STATUS_LABELS: Partial<Record<ContestStatus, string>> = {
  draft: '草稿',
  registering: '报名中',
  running: '进行中',
  frozen: '已冻结',
  ended: '已结束',
  published: '已发布',
  cancelled: '已取消',
  archived: '已归档',
}

function formatContestStatus(status: ContestStatus): string {
  return CONTEST_STATUS_LABELS[status] ?? status
}

function isAwdContest(contest: ContestDetailData | null): boolean {
  return contest?.mode === 'awd'
}

function hasContestStarted(status: ContestStatus | undefined): boolean {
  return status === 'running' || status === 'frozen' || status === 'ended'
}

function buildAwdReadinessSummary(contest: ContestDetailData | null): string {
  if (!contest || contest.mode !== 'awd') {
    return ''
  }
  if (hasContestStarted(contest.status)) {
    return '轮次运行已开启'
  }
  return '请在开赛前完成 AWD 配置与赛前检查'
}

export function useContestWorkbench(
  contest: Readonly<Ref<ContestDetailData | null>>,
  linkedChallengeCount?: Readonly<Ref<number | null>>
) {
  const visibleStages = computed<ContestWorkbenchStage[]>(() =>
    isAwdContest(contest.value) ? AWD_STAGES : BASE_STAGES
  )

  const defaultStage = computed<ContestWorkbenchStageKey>(() => {
    if (isAwdContest(contest.value)) {
      return hasContestStarted(contest.value?.status) ? 'operations' : 'pool'
    }
    return 'basics'
  })

  const summaryItems = computed<ContestWorkbenchSummaryItem[]>(() => {
    if (!contest.value) {
      return []
    }

    const items: ContestWorkbenchSummaryItem[] = [
      {
        key: 'mode',
        label: '赛事模式',
        value: CONTEST_MODE_LABELS[contest.value.mode === 'awd' ? 'awd' : 'jeopardy'],
      },
      {
        key: 'status',
        label: '当前状态',
        value: formatContestStatus(contest.value.status),
      },
    ]

    if (linkedChallengeCount?.value != null) {
      items.push({
        key: 'challenge-count',
        label: '已关联题目数',
        value: String(linkedChallengeCount.value),
        hint: '题目池调整后，这里会同步显示当前配置情况。',
      })
    }

    if (contest.value.mode === 'awd') {
      items.push({
        key: 'awd-readiness',
        label: 'AWD 准备度',
        value: buildAwdReadinessSummary(contest.value),
      })
    }

    return items
  })

  return proxyRefs({
    visibleStages,
    defaultStage,
    summaryItems,
  })
}
