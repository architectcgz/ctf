import { describe, expect, it } from 'vitest'

import contestEditWorkspacePanelSource from '@/components/platform/contest/ContestEditWorkspacePanel.vue?raw'
import contestEditSource from '@/views/platform/ContestEdit.vue?raw'

describe('ContestEdit workspace extraction', () => {
  it('应将竞赛编辑 stage 工作区抽到独立组件', () => {
    expect(contestEditSource).toContain(
      "import ContestEditWorkspacePanel from '@/components/platform/contest/ContestEditWorkspacePanel.vue'"
    )
    expect(contestEditSource).toContain('<ContestEditWorkspacePanel')
  })

  it('编辑工作台运行面板应按阶段收窄 AWDOperationsPanel 内容，避免重复显示开赛就绪摘要', () => {
    expect(contestEditWorkspacePanelSource).toContain('operation-panel="inspector"')
    expect(contestEditWorkspacePanelSource).toContain('runtime-content="round-inspector"')
    expect(contestEditWorkspacePanelSource).toContain('operation-panel="instances"')
    expect(contestEditWorkspacePanelSource).toContain('runtime-content="instances"')
  })
})
