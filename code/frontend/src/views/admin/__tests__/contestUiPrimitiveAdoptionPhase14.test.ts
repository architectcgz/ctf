import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 14', () => {
  it('awd challenge config dialog should consume shared list heading structure for preview target cards', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading checker-preview-target-card__top"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading__title checker-preview-target-card__title checker-preview-target-card__url"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<div class="checker-preview-target-card__top">'
    )
  })
})
