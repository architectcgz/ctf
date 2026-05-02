import { describe, expect, it } from 'vitest'

import contestDialogStateSource from './useContestDialogState.ts?raw'
import contestSaveFlowSource from './useContestSaveFlow.ts?raw'

describe('platform contests model boundary', () => {
  it('子模块不应反向依赖 usePlatformContests', () => {
    expect(contestDialogStateSource).not.toContain("from './usePlatformContests'")
    expect(contestSaveFlowSource).not.toContain("from './usePlatformContests'")
  })
})
