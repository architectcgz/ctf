import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 8', () => {
  it('awd challenge config dialog should consume shared list heading structure for config blocks', () => {
    expect(awdChallengeConfigDialogSource).toContain('class="list-heading checker-config-block__head"')
    expect(awdChallengeConfigDialogSource).toContain(
      'class="list-heading__title checker-config-block__title"'
    )
    expect(awdChallengeConfigDialogSource).not.toContain('<header class="checker-config-block__head">')
  })
})
