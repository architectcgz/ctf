import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 16', () => {
  it('awd challenge config dialog should use a header element for saved validation card top row', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      '<header class="list-heading checker-validation-card__top">'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<div class="list-heading checker-validation-card__top">'
    )
  })
})
