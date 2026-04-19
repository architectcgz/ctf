import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 15', () => {
  it('awd challenge config dialog should consume shared list heading layout for saved validation card top row', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading checker-validation-card__top"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain('<div class="checker-validation-card__top">')
  })
})
