import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdRoundCreateDialogSourceBase from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdRoundCreateScoreSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateScoreSection.vue?raw'
import awdRoundCreateSettingsSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateSettingsSection.vue?raw'
import awdAttackLogDetailsSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue?raw'
import awdAttackLogDialogSourceBase from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdAttackLogTargetSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue?raw'
import awdOperationsDialogFooterSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue?raw'
import awdServiceCheckDialogSourceBase from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdServiceCheckResultSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue?raw'
import awdServiceCheckTargetSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue?raw'
import contestAwdChallengeSelectorSectionSource from '@/features/contest-workbench/ui/ContestAwdChallengeSelectorSection.vue?raw'
import contestChallengeEditorDialogSourceBase from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import contestChallengeSettingsSectionSource from '@/features/contest-workbench/ui/ContestChallengeSettingsSection.vue?raw'

const awdRoundCreateDialogSource = [
  awdRoundCreateDialogSourceBase,
  awdRoundCreateSettingsSectionSource,
  awdRoundCreateScoreSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

const awdAttackLogDialogSource = [
  awdAttackLogDialogSourceBase,
  awdAttackLogTargetSectionSource,
  awdAttackLogDetailsSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

const awdServiceCheckDialogSource = [
  awdServiceCheckDialogSourceBase,
  awdServiceCheckTargetSectionSource,
  awdServiceCheckResultSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

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
