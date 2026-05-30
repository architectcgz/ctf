import { describe, expect, it } from 'vitest'

import contestEditSource from '@/pages/platform/contests/ContestEditRoutePage.vue?raw'

describe('ContestEdit AWD workspace extraction', () => {
  it('应将 AWD 工作区状态与操作抽到独立 composable', () => {
    expect(contestEditSource).not.toContain('useContestEditAwdWorkspace')
    expect(contestEditSource).toContain('useContestEditPage')
    expect(contestEditSource).toContain('<ContestEditWorkspacePanel')
  })
})
