import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import contestEditSource from '../ContestEdit.vue?raw'
import cheatDetectionSource from '../CheatDetection.vue?raw'
import auditLogSource from '../AuditLog.vue?raw'
import imageManageSource from '../ImageManage.vue?raw'
import challengeImportPreviewSource from '../ChallengeImportPreview.vue?raw'
import challengePackageFormatSource from '../ChallengePackageFormat.vue?raw'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'

describe('admin root shell cleanup', () => {
  it.each([
    ['ChallengeDetail.vue', challengeDetailSource],
    ['ContestEdit.vue', contestEditSource],
    ['CheatDetection.vue', cheatDetectionSource],
    ['AuditLog.vue', auditLogSource],
    ['ImageManage.vue', imageManageSource],
    ['ChallengeImportPreview.vue', challengeImportPreviewSource],
    ['ChallengePackageFormat.vue', challengePackageFormatSource],
    ['UserGovernancePage.vue', userGovernanceSource],
  ])('%s 应只保留共享管理员根壳，不再手写外层圆角', (_name, source) => {
    expect(source).toContain('workspace-shell')
    expect(source).toContain('journal-shell-admin')
    expect(source).toContain('journal-hero')
    expect(source).not.toContain('rounded-[30px]')
    expect(source).not.toContain('rounded-[24px]')
  })
})
