import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 11', () => {
  it('awd challenge config dialog should consume shared list heading structure for preview result raw payload section', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading checker-preview-result__json-head"'
    )
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading__title checker-preview-result__json-title"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain(
      '<div class="journal-note-label">原始结果</div>'
    )
  })
})
