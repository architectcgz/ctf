import { describe, expect, it } from 'vitest'

import awdAttackLogDialogSource from '@/components/admin/contest/AWDAttackLogDialog.vue?raw'
import awdRoundCreateDialogSource from '@/components/admin/contest/AWDRoundCreateDialog.vue?raw'
import awdServiceCheckDialogSource from '@/components/admin/contest/AWDServiceCheckDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/components/admin/contest/ContestChallengeEditorDialog.vue?raw'

describe('contest ui primitive adoption phase 3', () => {
  it('awd round create dialog should consume shared field and button primitives', () => {
    expect(awdRoundCreateDialogSource).toContain('class="ui-field')
    expect(awdRoundCreateDialogSource).toContain('class="ui-control-wrap')
    expect(awdRoundCreateDialogSource).toContain('class="ui-control')
    expect(awdRoundCreateDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdRoundCreateDialogSource).toContain('class="ui-btn ui-btn--primary')
  })

  it('awd service check dialog should consume shared field and button primitives', () => {
    expect(awdServiceCheckDialogSource).toContain('class="ui-field')
    expect(awdServiceCheckDialogSource).toContain('class="ui-control-wrap')
    expect(awdServiceCheckDialogSource).toContain('class="ui-control')
    expect(awdServiceCheckDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdServiceCheckDialogSource).toContain('class="ui-btn ui-btn--primary')
  })

  it('awd attack log dialog should consume shared field and button primitives', () => {
    expect(awdAttackLogDialogSource).toContain('class="ui-field')
    expect(awdAttackLogDialogSource).toContain('class="ui-control-wrap')
    expect(awdAttackLogDialogSource).toContain('class="ui-control')
    expect(awdAttackLogDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdAttackLogDialogSource).toContain('class="ui-btn ui-btn--primary')
  })

  it('contest challenge editor dialog should consume shared field and button primitives', () => {
    expect(contestChallengeEditorDialogSource).toContain('class="ui-field contest-challenge-dialog__field')
    expect(contestChallengeEditorDialogSource).toContain('class="ui-control-wrap')
    expect(contestChallengeEditorDialogSource).toContain('class="ui-control')
    expect(contestChallengeEditorDialogSource).toContain(
      'class="ui-btn ui-btn--secondary contest-challenge-dialog__button'
    )
    expect(contestChallengeEditorDialogSource).toContain(
      'class="ui-btn ui-btn--primary contest-challenge-dialog__button'
    )
  })
})
