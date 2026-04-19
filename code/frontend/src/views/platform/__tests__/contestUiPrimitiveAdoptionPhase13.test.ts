import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 13', () => {
  it('awd challenge config dialog should consume shared list heading structure for preview action cards', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading checker-preview-action-card__top"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading__title checker-preview-action-card__title"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<div class="checker-preview-action-card__top">'
    )
  })
})
