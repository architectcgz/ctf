import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 9', () => {
  it('awd challenge config dialog should consume shared list heading structure for checker action sections', () => {
    const actionSectionHeads =
      awdChallengeConfigDialogSource.match(/class="list-heading checker-action-section__head"/g) ??
      []
    const actionSectionTitles =
      awdChallengeConfigDialogSource.match(
        /class="list-heading__title checker-action-section__title"/g
      ) ?? []

    expect(actionSectionHeads).toHaveLength(3)
    expect(actionSectionTitles).toHaveLength(3)
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<header class="checker-action-section__head">'
    )
  })
})
