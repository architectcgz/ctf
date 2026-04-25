import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '@/views/platform/CheatDetection.vue?raw'
import cheatDetectionHeroPanelSource from '@/components/platform/cheat/CheatDetectionHeroPanel.vue?raw'
import cheatDetectionSummaryPanelSource from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'

describe('CheatDetection panel extraction', () => {
  it('应将作弊检测 hero 顶栏抽到独立 platform cheat 组件', () => {
    expect(cheatDetectionSource).toContain(
      "import CheatDetectionWorkspacePanel from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue'"
    )
    expect(cheatDetectionSource).toContain('<CheatDetectionWorkspacePanel')
    expect(cheatDetectionWorkspacePanelSource).toContain(
      "import CheatDetectionHeroPanel from '@/components/platform/cheat/CheatDetectionHeroPanel.vue'"
    )
    expect(cheatDetectionWorkspacePanelSource).toContain('<CheatDetectionHeroPanel')
    expect(cheatDetectionHeroPanelSource).toContain(
      '<div class="workspace-overline">Integrity Workspace</div>'
    )
    expect(cheatDetectionHeroPanelSource).toContain('打开审计日志')
    expect(cheatDetectionHeroPanelSource).toContain('刷新线索')
    expect(cheatDetectionHeroPanelSource).toContain('class="hero-meta-badge__label">最近生成</span>')
  })

  it('应将作弊检测摘要卡抽到独立 platform cheat 组件', () => {
    expect(cheatDetectionWorkspacePanelSource).toContain(
      "import CheatDetectionSummaryPanel from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue'"
    )
    expect(cheatDetectionWorkspacePanelSource).toContain('<CheatDetectionSummaryPanel :summary="riskData.summary" />')
    expect(cheatDetectionSummaryPanelSource).toContain(
      'class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(cheatDetectionSummaryPanelSource).toContain('提交风险账号')
    expect(cheatDetectionSummaryPanelSource).toContain('共享 IP 线索')
    expect(cheatDetectionSummaryPanelSource).toContain('涉及用户总数')
  })
})
