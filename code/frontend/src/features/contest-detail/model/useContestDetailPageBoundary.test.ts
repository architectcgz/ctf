import { describe, expect, it } from 'vitest'

import source from './useContestDetailPage.ts?raw'

describe('useContestDetailPage boundary', () => {
  it('应组合 countdown/data-loader/selection-sync 子模块，避免主组合器内联加载与倒计时实现', () => {
    expect(source).toContain("from './useContestDetailCountdown'")
    expect(source).toContain("from './useContestDetailDataLoader'")
    expect(source).toContain("from './useContestDetailSelectionSync'")
    expect(source).not.toContain('function updateCountdown()')
    expect(source).not.toContain('let requestToken = 0')
    expect(source).not.toContain('async function loadPage()')
    expect(source).not.toContain('function normalizeChallengeId(')
  })
})
