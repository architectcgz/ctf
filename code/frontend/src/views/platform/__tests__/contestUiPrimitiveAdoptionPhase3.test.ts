import { describe, expect, it } from 'vitest'

import awdAttackLogDialogSource from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdRoundCreateDialogSource from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdServiceCheckDialogSource from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import contestAwdChallengeSelectorSectionSource from '@/features/contest-workbench/ui/ContestAwdChallengeSelectorSection.vue?raw'
import contestChallengeEditorDialogSourceBase from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import contestChallengeSettingsSectionSource from '@/features/contest-workbench/ui/ContestChallengeSettingsSection.vue?raw'

const contestChallengeEditorDialogSource = [
  contestChallengeEditorDialogSourceBase,
  contestAwdChallengeSelectorSectionSource,
  contestChallengeSettingsSectionSource,
].join('\n')

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
