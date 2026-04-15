import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 6', () => {
  it('awd challenge config dialog should consume shared badge and note primitives in preview panels', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      "up: 'ui-badge ui-badge--pill ui-badge--soft checker-preview-status checker-preview-status--up'"
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'return `ui-badge ui-badge--pill ui-badge--soft checker-validation-chip checker-validation-chip--${state}`'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-action-card"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="journal-note checker-preview-target-card"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain('.checker-preview-status {')
    expect(awdChallengeConfigDialogSource).not.toContain('.checker-validation-chip {')
  })
})
