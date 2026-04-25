import { describe, expect, it } from 'vitest'

import contestEditSource from '../ContestEdit.vue?raw'
import contestEditTopbarPanelSource from '@/components/platform/contest/ContestEditTopbarPanel.vue?raw'

describe('ContestEdit topbar extraction', () => {
  it('应将竞赛编辑页顶部工作台壳层抽到独立 platform contest 组件', () => {
    expect(contestEditSource).toContain(
      "import ContestEditTopbarPanel from '@/components/platform/contest/ContestEditTopbarPanel.vue'"
    )
    expect(contestEditSource).toContain('<ContestEditTopbarPanel')
    expect(contestEditTopbarPanelSource).toContain('Contest Editor')
    expect(contestEditTopbarPanelSource).toContain('contest-open-announcements')
    expect(contestEditTopbarPanelSource).toContain('保存变更')
  })
})
