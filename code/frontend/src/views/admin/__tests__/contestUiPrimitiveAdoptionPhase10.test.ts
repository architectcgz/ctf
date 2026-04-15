import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 10', () => {
  it('awd challenge config dialog should consume shared list heading structure for preview result summary', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading checker-preview-result__head"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading__title checker-preview-result__title"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<header class="checker-preview-result__head">'
    )
  })
})
