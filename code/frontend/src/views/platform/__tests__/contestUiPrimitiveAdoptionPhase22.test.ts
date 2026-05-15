import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'

describe('contest ui primitive adoption phase 22', () => {
  it('contest orchestration overview should adopt the shared panel header contract', () => {
    expect(contestOrchestrationSource).toContain(
      '<header class="workspace-panel-header contest-overview-head">'
    )
    expect(contestOrchestrationSource).toContain('class="workspace-panel-header__intro"')
    expect(contestOrchestrationSource).toContain(
      'class="workspace-panel-header__actions ui-toolbar-actions contest-panel-actions"'
    )
    expect(contestOrchestrationSource).toContain(
      'class="workspace-panel-header__summary admin-summary-grid contest-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
  })
})
