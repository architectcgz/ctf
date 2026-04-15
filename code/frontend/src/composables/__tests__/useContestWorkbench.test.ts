import { ref } from 'vue'
import { describe, expect, it } from 'vitest'

import { useContestWorkbench } from '../useContestWorkbench'
import type { ContestDetailData } from '@/api/contracts'

function buildContestDetail(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'contest-1',
    title: '2026 春季校园 CTF',
    description: '校内赛',
    mode: 'jeopardy',
    status: 'registering',
    starts_at: '2026-03-15T09:00:00.000Z',
    ends_at: '2026-03-15T13:00:00.000Z',
    ...overrides,
  }
}

describe('useContestWorkbench', () => {
  it('jeopardy 模式仅返回基础阶段', () => {
    const result = useContestWorkbench(ref(buildContestDetail()))

    expect(result.visibleStages.map((item) => item.key)).toEqual(['basics', 'pool'])
    expect(result.defaultStage).toBe('basics')
  })

  it('awd + running 状态默认阶段为 operations', () => {
    const result = useContestWorkbench(
      ref(
        buildContestDetail({
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
        })
      )
    )

    expect(result.visibleStages.map((item) => item.key)).toEqual([
      'basics',
      'pool',
      'awd-config',
      'preflight',
      'operations',
    ])
    expect(result.defaultStage).toBe('operations')
  })

  it('awd + registering 状态默认阶段为 basics 或 pool', () => {
    const result = useContestWorkbench(
      ref(
        buildContestDetail({
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'registering',
        })
      )
    )

    expect(result.visibleStages.map((item) => item.key)).toEqual([
      'basics',
      'pool',
      'awd-config',
      'preflight',
      'operations',
    ])
    expect(['basics', 'pool']).toContain(result.defaultStage)
  })
})
