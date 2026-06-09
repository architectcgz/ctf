export interface AwdCheckerPreviewProgressPhase {
  key: string
  label: string
  detail: string
  minElapsedMs: number
  attempt?: number
}

export const AWD_CHECKER_PREVIEW_ATTEMPT_TOTAL = 3

export const AWD_CHECKER_PREVIEW_PROGRESS_PHASES: AwdCheckerPreviewProgressPhase[] = [
  {
    key: 'prepare',
    label: '准备预览环境',
    detail: '正在校验当前 Checker 草稿，并准备目标访问上下文。',
    minElapsedMs: 0,
  },
  {
    key: 'attempt-1',
    label: '第 1 轮试跑',
    detail: '开始执行第一轮请求校验。',
    minElapsedMs: 900,
    attempt: 1,
  },
  {
    key: 'attempt-2',
    label: '第 2 轮试跑',
    detail: '继续执行第二轮稳定性校验。',
    minElapsedMs: 1900,
    attempt: 2,
  },
  {
    key: 'attempt-3',
    label: '第 3 轮试跑',
    detail: '正在完成第三轮试跑。',
    minElapsedMs: 2900,
    attempt: 3,
  },
  {
    key: 'summary',
    label: '汇总结果',
    detail: '正在整理三轮试跑结果并生成最终摘要。',
    minElapsedMs: 3900,
  },
]

export function resolveAwdCheckerPreviewProgressPhaseIndex(elapsedMs: number): number {
  let phaseIndex = 0
  for (let index = 0; index < AWD_CHECKER_PREVIEW_PROGRESS_PHASES.length; index += 1) {
    if (elapsedMs >= AWD_CHECKER_PREVIEW_PROGRESS_PHASES[index].minElapsedMs) {
      phaseIndex = index
    }
  }
  return phaseIndex
}

export function resolveAwdCheckerPreviewProgressPhaseIndexByKey(phaseKey?: string): number {
  if (!phaseKey) {
    return 0
  }
  const index = AWD_CHECKER_PREVIEW_PROGRESS_PHASES.findIndex((item) => item.key === phaseKey)
  return index >= 0 ? index : 0
}

export function formatAwdCheckerPreviewElapsed(elapsedMs: number): string {
  if (!Number.isFinite(elapsedMs) || elapsedMs <= 0) {
    return '0.0s'
  }
  if (elapsedMs < 10_000) {
    return `${(elapsedMs / 1000).toFixed(1)}s`
  }
  return `${Math.round(elapsedMs / 1000)}s`
}
