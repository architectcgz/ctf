import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'

describe('contest ui primitive adoption phase 22', () => {
  it('contest orchestration page should use shared list heading layout for overview hero header', () => {
    expect(contestOrchestrationSource).toContain(
      '<header class="list-heading contest-overview-head">'
    )
    expect(contestOrchestrationSource).not.toContain('<header class="contest-overview-head">')
  })
})
