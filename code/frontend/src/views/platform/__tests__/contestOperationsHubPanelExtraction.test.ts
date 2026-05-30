import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '@/pages/platform/contests/ContestOperationsHubRoutePage.vue?raw'
import contestOperationsHubHeroPanelSource from '@/features/platform/contests/ui/ContestOperationsHubHeroPanel.vue?raw'

describe('ContestOperationsHub panel extraction', () => {
  it('应将赛事运维头部与摘要卡收口到 platform contests feature UI', () => {
    expect(contestOperationsHubSource).toContain(
      'ContestOperationsHubHeroPanel,'
    )
    expect(contestOperationsHubSource).toContain("from '@/features/platform/contests'")
    expect(contestOperationsHubSource).toContain('<ContestOperationsHubHeroPanel')
    expect(contestOperationsHubHeroPanelSource).toContain('Event Operations')
    expect(contestOperationsHubHeroPanelSource).toContain('返回竞赛目录')
    expect(contestOperationsHubHeroPanelSource).toContain(
      '<header class="workspace-panel-header contest-ops-hero">'
    )
    expect(contestOperationsHubHeroPanelSource).toContain(
      'class="workspace-panel-header__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
  })
})
