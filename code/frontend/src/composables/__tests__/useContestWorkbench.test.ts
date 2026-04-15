import { ref } from 'vue'
import { describe, expect, it } from 'vitest'
import { CONTEST_WORKBENCH_STAGE_ORDER, useContestWorkbench } from '../useContestWorkbench'
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
  it('应该导出统一的工作台阶段顺序常量', () => {
    expect(CONTEST_WORKBENCH_STAGE_ORDER).toEqual([
      'basics',
      'pool',
      'awd-config',
      'preflight',
      'operations',
    ])
  })

  it('jeopardy 模式仅返回基础编排阶段', async () => {
    const result = useContestWorkbench(ref(buildContestDetail()))

    expect(result.visibleStages.map((item) => item.key)).toEqual(['basics', 'pool'])
    expect(result.defaultStage).toBe('basics')
  })

  it('awd + running 状态默认阶段为 operations', async () => {
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

  it('awd + registering 状态默认阶段为 pool', async () => {
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
    expect(result.defaultStage).toBe('pool')
  })

  it('未提供真实题目数时不应展示伪造的已关联题目数摘要', () => {
    const result = useContestWorkbench(
      ref(
        buildContestDetail({
          mode: 'awd',
          status: 'registering',
        })
      )
    )

    expect(result.summaryItems.find((item) => item.key === 'challenge-count')).toBeUndefined()
  })

  it('提供真实题目数时应展示准确的已关联题目数摘要', () => {
    const challengeCount = ref(3)
    const result = useContestWorkbench(
      ref(
        buildContestDetail({
          mode: 'awd',
          status: 'registering',
        })
      ),
      challengeCount
    )

    expect(result.summaryItems.find((item) => item.key === 'challenge-count')).toEqual(
      expect.objectContaining({
        key: 'challenge-count',
        value: '3',
      })
    )
  })
})
