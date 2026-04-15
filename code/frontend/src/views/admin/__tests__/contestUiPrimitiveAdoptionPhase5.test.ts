import { describe, expect, it } from 'vitest'

import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'

describe('contest ui primitive adoption phase 5', () => {
  it('awd challenge config dialog should consume shared button and field primitives in the http standard editor', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      'class="ui-btn ui-btn--secondary checker-preset-button"'
    )
    expect(awdChallengeConfigDialogSource).toContain('class="ui-field awd-http-action-field"')
    expect(awdChallengeConfigDialogSource).toContain(
      'class="ui-control-wrap awd-http-action-control"'
    )
    expect(awdChallengeConfigDialogSource).toContain('class="ui-control awd-config-control--mono"')
    expect(awdChallengeConfigDialogSource).not.toContain(
      'class="w-full rounded-xl border border-border bg-surface px-4 py-3'
    )
  })
})
