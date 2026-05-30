import { describe, expect, it } from 'vitest'

import contestDialogStateSource from './useContestDialogState.ts?raw'
import contestSaveFlowSource from './useContestSaveFlow.ts?raw'
import platformContestsSource from './usePlatformContests.ts?raw'

describe('platform contests model boundary', () => {
  it('子模块不应反向依赖 usePlatformContests', () => {
    expect(contestDialogStateSource).not.toContain("from './usePlatformContests'")
    expect(contestSaveFlowSource).not.toContain("from './usePlatformContests'")
  })

  it('usePlatformContests 应组合 useContestListState，而不是内联分页请求逻辑', () => {
    expect(platformContestsSource).toContain(
      "import { useContestListState } from './useContestListState'"
    )
    expect(platformContestsSource).not.toContain('usePagination<ContestDetailData>')
    expect(platformContestsSource).not.toContain('getContests({')
  })
})
