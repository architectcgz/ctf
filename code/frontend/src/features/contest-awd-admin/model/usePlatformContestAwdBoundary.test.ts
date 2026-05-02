import { describe, expect, it } from 'vitest'

import platformContestAwdSource from './usePlatformContestAwd.ts?raw'

describe('usePlatformContestAwd boundary', () => {
  it('应组合 traffic/state 子模块，避免在主组合器内联流量动作与状态标志派生', () => {
    expect(platformContestAwdSource).toContain(
      "import { useAwdTrafficActions } from './useAwdTrafficActions'"
    )
    expect(platformContestAwdSource).toContain(
      "import { useAwdContestStateFlags } from './useAwdContestStateFlags'"
    )
    expect(platformContestAwdSource).not.toContain('async function applyTrafficFilters(')
    expect(platformContestAwdSource).not.toContain('async function setTrafficPage(')
    expect(platformContestAwdSource).not.toContain('async function resetTrafficFilters(')
  })
})
