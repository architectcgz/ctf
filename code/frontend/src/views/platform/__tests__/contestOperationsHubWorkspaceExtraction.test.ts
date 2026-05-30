import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '@/pages/platform/contests/ContestOperationsHubRoutePage.vue?raw'

describe('ContestOperationsHub workspace extraction', () => {
  it('应将赛事运维目录工作区收口到 platform contests feature UI', () => {
    expect(contestOperationsHubSource).toContain(
      'ContestOperationsHubWorkspacePanel,'
    )
    expect(contestOperationsHubSource).toContain("from '@/features/platform/contests'")
    expect(contestOperationsHubSource).toContain('<ContestOperationsHubWorkspacePanel')
  })
})
