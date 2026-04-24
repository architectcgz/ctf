import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'
import contestOperationsHubHeroPanelSource from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue?raw'

describe('ContestOperationsHub panel extraction', () => {
  it('应将赛事运维头部与摘要卡抽到独立 platform contest 组件', () => {
    expect(contestOperationsHubSource).toContain(
      "import ContestOperationsHubHeroPanel from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue'"
    )
    expect(contestOperationsHubSource).toContain('<ContestOperationsHubHeroPanel')
    expect(contestOperationsHubHeroPanelSource).toContain('Event Operations')
    expect(contestOperationsHubHeroPanelSource).toContain('返回竞赛目录')
    expect(contestOperationsHubHeroPanelSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
  })
})
