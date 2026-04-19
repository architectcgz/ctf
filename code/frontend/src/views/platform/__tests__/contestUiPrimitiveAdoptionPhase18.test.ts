import { describe, expect, it } from 'vitest'

import contestAwdPreflightPanelSource from '@/components/platform/contest/ContestAwdPreflightPanel.vue?raw'

describe('contest ui primitive adoption phase 18', () => {
  it('contest awd preflight panel should use shared list heading layout for force start copy block', () => {
    expect(contestAwdPreflightPanelSource).toContain(
      '<header class="list-heading contest-awd-preflight-panel__override-head">'
    )
    expect(contestAwdPreflightPanelSource).not.toContain(`>
        <div>
          <div class="journal-note-label">Override Entry</div>`)
  })
})
