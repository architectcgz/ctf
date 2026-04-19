import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'
import awdReadinessSummarySource from '@/components/platform/contest/AWDReadinessSummary.vue?raw'

describe('contest ui primitive adoption phase 26', () => {
  it('awd challenge config panel should use workspace overline in the active tab panel header', () => {
    expect(awdChallengeConfigPanelSource).toContain(
      '<div class="workspace-overline">AWD Service Config</div>'
    )
    expect(awdChallengeConfigPanelSource).not.toContain(
      '<div class="journal-eyebrow">AWD Service Config</div>'
    )
  })

  it('awd readiness summary should use workspace overline in the active tab panel header', () => {
    expect(awdReadinessSummarySource).toContain(
      '<div class="workspace-overline">AWD Readiness</div>'
    )
    expect(awdReadinessSummarySource).not.toContain(
      '<div class="journal-eyebrow">AWD Readiness</div>'
    )
  })
})
