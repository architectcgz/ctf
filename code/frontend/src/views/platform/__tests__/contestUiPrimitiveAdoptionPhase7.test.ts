import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 7', () => {
  it('awd challenge config dialog should consume shared journal-note surfaces in preview feedback areas', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-notice checker-preview-notice--error"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-notice checker-preview-notice--warning"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-notice checker-preview-notice--success"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-validation-card"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-result__json"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain('.checker-preview-notice {')
  })
})
