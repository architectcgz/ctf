import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import contestEditSource from '../ContestEdit.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import auditLogSource from '../AuditLog.vue?raw'
import imageManageSource from '../ImageManage.vue?raw'
import challengeImportPreviewWorkspaceSource from '@/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue?raw'
import challengePackageFormatSource from '../ChallengePackageFormat.vue?raw'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'

describe('admin root shell cleanup', () => {
  it.each([
    ['ChallengeDetail.vue', challengeDetailSource],
    ['ContestEdit.vue', contestEditSource],
    ['CheatDetectionWorkspacePanel.vue', cheatDetectionWorkspacePanelSource],
    ['AuditLog.vue', auditLogSource],
    ['ImageManage.vue', imageManageSource],
    ['ChallengeImportPreviewWorkspacePanel.vue', challengeImportPreviewWorkspaceSource],
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
