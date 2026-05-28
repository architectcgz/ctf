import { describe, expect, it } from 'vitest'

import adminNotificationPublishDrawerSource from '@/features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue?raw'
import awdChallengeEditorDialogSource from '@/features/platform-awd-challenges/ui/AWDChallengeEditorDialog.vue?raw'
import awdAttackLogDialogSource from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdServiceCheckDialogSource from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdRoundCreateDialogSource from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import platformUserFormDialogSource from '@/features/platform-user-management/ui/PlatformUserFormDialog.vue?raw'
import awdCheckerSaveFlowSource from '@/features/contest-awd-config/model/useAwdCheckerSaveFlow.ts?raw'
import imageManageMutationsSource from '@/features/image-management/model/useImageManageMutations.ts?raw'

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
