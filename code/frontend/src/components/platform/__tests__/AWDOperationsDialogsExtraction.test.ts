import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdAttackLogDetailsSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue?raw'
import awdAttackLogDialogSource from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdAttackLogTargetSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue?raw'
import awdDialogContractsSource from '@/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts?raw'
import awdOperationsDialogFooterSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue?raw'
import awdRoundCreateDialogSource from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdRoundCreateScoreSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateScoreSection.vue?raw'
import awdRoundCreateSettingsSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateSettingsSection.vue?raw'
import awdServiceCheckDialogSource from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdServiceCheckResultSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue?raw'
import awdServiceCheckTargetSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue?raw'

const awdOperationsDialogsCombinedSource = [
  awdRoundCreateDialogSource,
  awdRoundCreateSettingsSectionSource,
  awdRoundCreateScoreSectionSource,
  awdAttackLogDialogSource,
  awdAttackLogTargetSectionSource,
  awdAttackLogDetailsSectionSource,
  awdServiceCheckDialogSource,
  awdServiceCheckTargetSectionSource,
  awdServiceCheckResultSectionSource,
  awdOperationsDialogFooterSource,
  awdDialogContractsSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'),
    'utf8'
  ),
].join('\n')

describe('AWD operations dialogs extraction', () => {
  it('AWD operations dialogs 应把稳定 section 和 footer 下沉到独立子组件', () => {
    expect(awdRoundCreateDialogSource).toContain('<AWDRoundCreateSettingsSection')
    expect(awdRoundCreateDialogSource).toContain('<AWDRoundCreateScoreSection')
    expect(awdRoundCreateDialogSource).toContain('<AWDOperationsDialogFooter')
    expect(awdRoundCreateDialogSource).not.toContain('id="awd-round-number"')
    expect(awdRoundCreateDialogSource).not.toContain('class="awd-round-dialog__footer"')
    expect(awdRoundCreateDialogSource).not.toContain('<style scoped>')
    expect(awdAttackLogDialogSource).toContain('<AWDAttackLogTargetSection')
    expect(awdAttackLogDialogSource).toContain('<AWDAttackLogDetailsSection')
    expect(awdAttackLogDialogSource).toContain('<AWDOperationsDialogFooter')
    expect(awdAttackLogDialogSource).not.toContain('id="awd-attack-team"')
    expect(awdAttackLogDialogSource).not.toContain('class="awd-operations-checkbox"')
    expect(awdAttackLogDialogSource).not.toContain('<style scoped>')
    expect(awdServiceCheckDialogSource).toContain('<AWDServiceCheckTargetSection')
    expect(awdServiceCheckDialogSource).toContain('<AWDServiceCheckResultSection')
    expect(awdServiceCheckDialogSource).toContain('<AWDOperationsDialogFooter')
    expect(awdServiceCheckDialogSource).not.toContain('id="awd-service-team"')
    expect(awdServiceCheckDialogSource).not.toContain('class="awd-operations-field__textarea"')
    expect(awdServiceCheckDialogSource).not.toContain('<style scoped>')
  })

  it('提取后的子组件应继续承接 field primitive、warning、textarea 与 footer shell', () => {
    expect(awdRoundCreateSettingsSectionSource).toContain('id="awd-round-number"')
    expect(awdRoundCreateSettingsSectionSource).toContain('id="awd-round-status"')
    expect(awdRoundCreateScoreSectionSource).toContain('id="awd-attack-score"')
    expect(awdRoundCreateScoreSectionSource).toContain('id="awd-defense-score"')
    expect(awdAttackLogTargetSectionSource).toContain('id="awd-attack-team"')
    expect(awdAttackLogTargetSectionSource).toContain('id="awd-victim-team"')
    expect(awdAttackLogTargetSectionSource).toContain('id="awd-attack-challenge"')
    expect(awdAttackLogDetailsSectionSource).toContain('id="awd-attack-type"')
    expect(awdAttackLogDetailsSectionSource).toContain('id="awd-attack-flag"')
    expect(awdServiceCheckTargetSectionSource).toContain('id="awd-service-team"')
    expect(awdServiceCheckTargetSectionSource).toContain('id="awd-service-challenge"')
    expect(awdServiceCheckTargetSectionSource).toContain('id="awd-service-status"')
    expect(awdServiceCheckResultSectionSource).toContain('id="awd-service-check-result"')
    expect(awdOperationsDialogFooterSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdOperationsDialogFooterSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdDialogContractsSource).toContain('export interface AwdCreateRoundPayload')
    expect(awdDialogContractsSource).toContain('export interface AwdOperationsOverrideDialogState')
    expect(awdOperationsDialogsCombinedSource).toContain('.awd-operations-dialog__footer')
    expect(awdOperationsDialogsCombinedSource).toContain('.awd-operations-checkbox__label')
    expect(awdOperationsDialogsCombinedSource).toContain('.awd-operations-field__textarea')
  })
})
