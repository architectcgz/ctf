import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 12', () => {
  it('awd challenge config dialog should consume shared list heading layout for preview toolbar', () => {
    expect(awdChallengeConfigDialogSource).toContain('class="list-heading checker-preview-toolbar"')
    expect(awdChallengeConfigDialogSource).not.toContain('<div class="checker-preview-toolbar">')
  })
})
