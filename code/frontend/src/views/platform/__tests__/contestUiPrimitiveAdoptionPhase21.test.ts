import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'

describe('contest ui primitive adoption phase 21', () => {
  it('contest orchestration page should use a dedicated list heading for create panel intro block', () => {
    expect(contestOrchestrationSource).toContain(
      '<header class="list-heading contest-create-head">'
    )
    expect(contestOrchestrationSource).not.toContain(`<section class="workspace-directory-section contest-create-panel">
          <header class="contest-overview-head">`)
  })
})
