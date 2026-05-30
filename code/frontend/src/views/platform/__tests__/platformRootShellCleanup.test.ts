import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/pages/platform/challenges/ChallengeDetailRoutePage.vue?raw'
import contestEditSource from '@/pages/platform/contests/ContestEditRoutePage.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'
import challengeImportPreviewWorkspaceSource from '@/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue?raw'
import challengePackageFormatSource from '@/pages/platform/challenges/ChallengePackageFormatRoutePage.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'

const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')

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
