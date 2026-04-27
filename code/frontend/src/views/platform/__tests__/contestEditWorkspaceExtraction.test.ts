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

  it('编辑工作台所有 stage 应复用统一切换动画并支持减少动态效果', () => {
    expect(contestEditWorkspacePanelSource).toContain('<Transition')
    expect(contestEditWorkspacePanelSource).toContain('name="studio-stage"')
    expect(contestEditWorkspacePanelSource).toContain('mode="out-in"')
    expect(contestEditWorkspacePanelSource).toContain('class="studio-pane studio-stage-panel"')
    expect(contestEditWorkspacePanelSource).toContain('class="studio-pane studio-pane--operations studio-stage-panel"')
    expect(contestEditWorkspacePanelSource).toContain('@media (prefers-reduced-motion: reduce)')
    expect(contestEditWorkspacePanelSource).not.toContain('class="studio-pane fade-in"')
    expect(contestEditWorkspacePanelSource).not.toContain('@keyframes studioFadeIn')
  })
})
