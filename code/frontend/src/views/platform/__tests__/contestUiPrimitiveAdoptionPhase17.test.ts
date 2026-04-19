import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 17', () => {
  it('awd challenge config dialog should use a header element for preview toolbar', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      '<header class="list-heading checker-preview-toolbar">'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<div class="list-heading checker-preview-toolbar">'
    )
  })
})
