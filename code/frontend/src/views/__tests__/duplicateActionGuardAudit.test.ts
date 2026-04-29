import { describe, expect, it } from 'vitest'

import adminNotificationPublishDrawerSource from '@/components/notifications/AdminNotificationPublishDrawer.vue?raw'
import awdChallengeEditorDialogSource from '@/components/platform/awd-service/AWDChallengeEditorDialog.vue?raw'
import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'
import awdAttackLogDialogSource from '@/components/platform/contest/AWDAttackLogDialog.vue?raw'
import awdServiceCheckDialogSource from '@/components/platform/contest/AWDServiceCheckDialog.vue?raw'
import awdRoundCreateDialogSource from '@/components/platform/contest/AWDRoundCreateDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/components/platform/contest/ContestChallengeEditorDialog.vue?raw'
import platformUserFormDialogSource from '@/components/platform/user/PlatformUserFormDialog.vue?raw'
import imageManagePageSource from '@/composables/useImageManagePage.ts?raw'

describe('duplicate action guard audit', () => {
  it('表单 submit 与按钮 click 共用 handler 时应在本地 owner 上短路 in-flight 状态', () => {
    expect(adminNotificationPublishDrawerSource).toContain('if (publisher.submitting.value) {')
    expect(awdChallengeEditorDialogSource).toContain('if (props.saving) {')
    expect(awdChallengeConfigDialogSource).toContain('if (props.saving) {')
    expect(awdAttackLogDialogSource).toContain('if (props.saving) {')
    expect(awdServiceCheckDialogSource).toContain('if (props.saving) {')
    expect(awdRoundCreateDialogSource).toContain('if (props.saving) {')
    expect(contestChallengeEditorDialogSource).toContain('if (props.saving) {')
    expect(platformUserFormDialogSource).toContain('if (props.saving) {')
    expect(imageManagePageSource).toContain('if (creating.value) {')
  })
})
