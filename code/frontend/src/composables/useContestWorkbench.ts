import { computed, proxyRefs, type Readonly, type Ref } from 'vue'

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

function isAwdContest(contest: ContestDetailData | null): boolean {
  return contest?.mode === 'awd'
}

function hasContestStarted(status: ContestStatus | undefined): boolean {
  return status === 'running' || status === 'frozen' || status === 'ended'
}

function pickChallengeCount(contest: ContestDetailData | null): string {
  const countCandidates = [
    contest?.meta?.linked_challenge_count,
    contest?.meta?.challenge_count,
    contest?.meta?.total_challenges,
  ]

  for (const candidate of countCandidates) {
    if (typeof candidate === 'number' && Number.isFinite(candidate)) {
      return String(candidate)
    }
  }

  return '待同步'
}

function buildAwdReadinessSummary(contest: ContestDetailData | null): string {
  if (!contest || contest.mode !== 'awd') {
    return ''
  }
  if (hasContestStarted(contest.status)) {
    return '已进入轮次运行占位阶段'
  }
  return '待补齐 Checker、赛前检查与运行配置'
}

export function useContestWorkbench(contest: Readonly<Ref<ContestDetailData | null>>) {
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
        value: CONTEST_STATUS_LABELS[contest.value.status] ?? contest.value.status,
      },
      {
        key: 'challenge-count',
        label: '已关联题目数',
        value: pickChallengeCount(contest.value),
        hint: '工作台骨架阶段先展示汇总信息，题目池详情仍在下方维护。',
      },
    ]

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
