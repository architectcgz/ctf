import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'

describe('contest ui primitive adoption phase 20', () => {
  it('awd challenge config panel should use a header element for active challenge focus card intro block', () => {
    expect(awdChallengeConfigPanelSource).toContain(
      '<header class="list-heading config-focus-card__head">'
    )
    expect(awdChallengeConfigPanelSource).not.toContain(`>
        <div>
          <div class="journal-note-label">Current Focus</div>`)
  })
})
