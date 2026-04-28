import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'

describe('ContestOperationsHub workspace extraction', () => {
  it('应将赛事运维目录工作区抽到独立 platform contest 组件', () => {
    expect(contestOperationsHubSource).toContain(
      "import ContestOperationsHubWorkspacePanel from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue'"
    )
    expect(contestOperationsHubSource).toContain('<ContestOperationsHubWorkspacePanel')
  })
})
