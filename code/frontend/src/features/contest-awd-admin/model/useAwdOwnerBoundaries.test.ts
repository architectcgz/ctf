import { describe, expect, it } from 'vitest'

import stateFlagsSource from './useAwdContestStateFlags.ts?raw'
import readinessDecisionSource from './useAwdReadinessDecision.ts?raw'
import roundOperationsSource from './useAwdRoundOperations.ts?raw'
import platformContestAwdSource from './usePlatformContestAwd.ts?raw'

describe('contest AWD owner boundaries', () => {
  it('runtime policy 应继续收口在 useAwdContestStateFlags，而不是散回 mutation 或 readiness owner', () => {
    expect(stateFlagsSource).toContain('export function isAwdRuntimeStageStatus(')
    expect(stateFlagsSource).toContain('const runtimeStageReady = computed(')
    expect(stateFlagsSource).toContain('const canOperateSelectedRound = computed(')
    expect(stateFlagsSource).toContain('const shouldUseCurrentRoundCheck = computed(')
    expect(stateFlagsSource).toContain('const shouldAutoRefresh = computed(')
    expect(stateFlagsSource).not.toContain('getContestAWDReadiness(')
    expect(stateFlagsSource).not.toContain('createContestAWDRound(')
    expect(stateFlagsSource).not.toContain('runContestAWDCurrentRoundCheck(')
  })

  it('readiness override workflow 应继续收口在 useAwdReadinessDecision，而不是重新混入 runtime policy', () => {
    expect(readinessDecisionSource).toContain('async function loadOverrideReadinessSnapshot()')
    expect(readinessDecisionSource).toContain('async function executeOverrideAction(')
    expect(readinessDecisionSource).toContain('toast.error(humanizeRequestError(error, \'读取开赛就绪摘要失败\'))')
    expect(readinessDecisionSource).not.toContain('selectedRoundId')
    expect(readinessDecisionSource).not.toContain('runtimeStageReady')
    expect(readinessDecisionSource).not.toContain('shouldAutoRefresh')
    expect(readinessDecisionSource).not.toContain('canOperateSelectedRound')
  })

  it('round operations 应继续消费外部 policy，而不是重新内联 runtime 或 readiness gate', () => {
    expect(roundOperationsSource).toContain('canOperateSelectedRound: Readonly<ComputedRef<boolean>>')
    expect(roundOperationsSource).toContain('shouldUseCurrentRoundCheck: Readonly<ComputedRef<boolean>>')
    expect(roundOperationsSource).toContain('const shouldRunCurrentRound = shouldUseCurrentRoundCheck.value')
    expect(roundOperationsSource).not.toContain(
      "selectedRound.value?.status === 'running' || !activeRoundId"
    )
    expect(roundOperationsSource).not.toContain(
      'computed(() => Boolean(selectedContest.value && selectedRoundId.value))'
    )
    expect(roundOperationsSource).not.toContain('getContestAWDReadiness(')
  })

  it('usePlatformContestAwd 应继续作为 owner composition root，而不是重新内联 readiness/runtime 逻辑', () => {
    expect(platformContestAwdSource).toContain(
      "import { useAwdContestStateFlags } from './useAwdContestStateFlags'"
    )
    expect(platformContestAwdSource).toContain(
      "import { useAwdReadinessDecision } from './useAwdReadinessDecision'"
    )
    expect(platformContestAwdSource).toContain('} = useAwdReadinessDecision({')
    expect(platformContestAwdSource).toContain('} = useAwdContestStateFlags({')
    expect(platformContestAwdSource).not.toContain('async function refreshReadiness(')
    expect(platformContestAwdSource).not.toContain('async function confirmOverrideAction(')
    expect(platformContestAwdSource).not.toContain('const runtimeStageReady = computed(')
  })
})
