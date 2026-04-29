import { describe, expect, it } from 'vitest'

import contestAwdPreflightPanelSource from '@/components/platform/contest/ContestAwdPreflightPanel.vue?raw'

describe('contest ui primitive adoption phase 18', () => {
  it('contest awd preflight panel should not render a force start override block', () => {
    expect(contestAwdPreflightPanelSource).not.toContain('contest-awd-preflight-force-start')
    expect(contestAwdPreflightPanelSource).not.toContain('contest-awd-preflight-panel__override-head')
    expect(contestAwdPreflightPanelSource).not.toContain('Override Entry')
  })
})
