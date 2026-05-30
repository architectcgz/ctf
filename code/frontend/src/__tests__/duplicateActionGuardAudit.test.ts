import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import adminNotificationPublishDrawerSource from '@/features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue?raw'
import awdChallengeEditorDialogSource from '@/features/platform/awd-challenges/ui/AWDChallengeEditorDialog.vue?raw'
import awdAttackLogDetailsSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue?raw'
import awdAttackLogDialogSourceBase from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdAttackLogTargetSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue?raw'
import awdOperationsDialogFooterSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue?raw'
import awdRoundCreateDialogSource from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdServiceCheckDialogSourceBase from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdServiceCheckResultSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue?raw'
import awdServiceCheckTargetSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue?raw'
import contestChallengeEditorDialogSource from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import platformUserFormDialogSource from '@/features/platform/user-management/ui/PlatformUserFormDialog.vue?raw'
import awdCheckerSaveFlowSource from '@/features/contest-awd-config/model/useAwdCheckerSaveFlow.ts?raw'
import imageManageMutationsSource from '@/features/image-management/model/useImageManageMutations.ts?raw'

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

describe('duplicate action guard audit', () => {
  it('表单 submit 与按钮 click 共用 handler 时应在本地 owner 上短路 in-flight 状态', () => {
    expect(adminNotificationPublishDrawerSource).toContain('if (publisher.submitting.value) {')
    expect(awdChallengeEditorDialogSource).toContain('if (props.saving) {')
    expect(awdCheckerSaveFlowSource).toContain("from '@/api/admin/contests'")
    expect(awdCheckerSaveFlowSource).toContain('saving.value ||')
    expect(awdCheckerSaveFlowSource).toContain('saving.value = true')
    expect(awdCheckerSaveFlowSource).toContain('saving.value = false')
    expect(awdAttackLogDialogSource).toContain('if (props.saving) {')
    expect(awdServiceCheckDialogSource).toContain('if (props.saving) {')
    expect(awdRoundCreateDialogSource).toContain('if (props.saving) {')
    expect(contestChallengeEditorDialogSource).toContain('if (props.saving) {')
    expect(platformUserFormDialogSource).toContain('if (props.saving) {')
    expect(imageManageMutationsSource).toContain('if (creating.value) {')
    expect(imageManageMutationsSource).toContain('if (deletingIds.has(id)) {')
  })
})
