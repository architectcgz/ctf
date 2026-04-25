import { describe, expect, it } from 'vitest'

import contestEditSource from '@/views/platform/ContestEdit.vue?raw'

describe('ContestEdit workspace extraction', () => {
  it('应将竞赛编辑 stage 工作区抽到独立组件', () => {
    expect(contestEditSource).toContain(
      "import ContestEditWorkspacePanel from '@/components/platform/contest/ContestEditWorkspacePanel.vue'"
    )
    expect(contestEditSource).toContain('<ContestEditWorkspacePanel')
  })
})
