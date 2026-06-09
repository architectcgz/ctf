/// <reference types="node" />
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import auditLogDirectoryPanelSource from '@/features/platform/audit-log/ui/AuditLogDirectoryPanel.vue?raw'
import awdReviewDirectoryPanelSource from '@/widgets/awd-review-workspace/AwdReviewDirectoryPanel.vue?raw'
import challengeManageDirectoryPanelSource from '@/features/platform/challenges/ui/ChallengeManageDirectoryPanel.vue?raw'
import challengeManageHeroPanelSource from '@/features/platform/challenges/ui/ChallengeManageHeroPanel.vue?raw'
import classManageSource from '@/pages/platform/ClassManageRoutePage.vue?raw'
import classManageWorkspacePanelSource from '@/features/platform/class-management/ui/ClassManageWorkspacePanel.vue?raw'
import contestManageCreatePanelSource from '@/features/platform/contest-manage/ui/ContestManageCreatePanel.vue?raw'
import contestManageOverviewPanelSource from '@/features/platform/contest-manage/ui/ContestManageOverviewPanel.vue?raw'
import contestOrchestrationSource from '@/features/platform/contest-manage/ui/ContestOrchestrationPage.vue?raw'
import instanceManageSource from '@/pages/platform/InstanceManageRoutePage.vue?raw'
import instanceManageWorkspacePanelSource from '@/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue?raw'
import adminContestTableSource from '@/features/platform/contest-manage/ui/PlatformContestTable.vue?raw'
import studentManageSource from '@/pages/platform/StudentManageRoutePage.vue?raw'
import studentManageWorkspacePanelSource from '@/features/platform/student-management/ui/StudentManageWorkspacePanel.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import workspaceDirectoryToolbarSource from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue?raw'

const styleSource = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const journalNotesSource = readFileSync(
  resolve(process.cwd(), 'src/assets/styles/journal-notes.css'),
  'utf8'
)

const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceImportPanelSource,
].join('\n')

const contestManagementSource = [
  contestOrchestrationSource,
  contestManageOverviewPanelSource,
  contestManageCreatePanelSource,
  adminContestTableSource,
].join('\n')

describe('admin management surface alignment', () => {
  it('keeps directory spacing, pagination spacing, and summary-grid foundations on shared owners', () => {
    expect(styleSource).toContain('.workspace-directory-section {')
    expect(styleSource).toContain('.workspace-directory-list {')
    expect(styleSource).toContain('.workspace-directory-section > .workspace-directory-pagination')
    expect(styleSource).toContain('--workspace-directory-page-block-gap: var(--space-5);')
    expect(workspaceDirectoryToolbarSource).toContain(
      'margin-bottom: var(--workspace-directory-toolbar-gap-bottom, 1.5rem);'
    )
    expect(journalNotesSource).toContain(
      '.journal-shell-admin :is(.admin-summary-grid, .manage-summary-grid, .image-summary-grid)'
    )
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
  })

  it('keeps user governance on shared summary, directory, and pagination owners', () => {
    expect(userGovernancePageSource).toContain('id="user-panel-overview"')
    expect(userGovernancePageSource).toContain('id="user-panel-import"')
    expect(userGovernanceOverviewPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(userGovernanceOverviewPanelSource).toContain(
      "from '@/shared/ui/common/PagePaginationControls.vue'"
    )
    expect(userGovernanceOverviewPanelSource).toContain('全部用户')
    expect(userGovernanceImportPanelSource).toContain('导入回执')
    expect(userGovernanceSource).not.toContain('<nav class="top-tabs"')
    expect(userGovernanceSource).not.toMatch(/^\.list-heading\s*\{/m)
  })

  it('keeps contest management on shared workspace, directory, and table owners', () => {
    expect(contestOrchestrationSource).toContain('ContestManageOverviewPanel')
    expect(contestOrchestrationSource).toContain('ContestManageCreatePanel')
    expect(contestManageOverviewPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(contestManageOverviewPanelSource).toContain('竞赛列表')
    expect(adminContestTableSource).toContain(
      "from '@/shared/ui/common/PagePaginationControls.vue'"
    )
    expect(adminContestTableSource).toContain("{ key: 'starts_at', label: '开始时间'")
    expect(adminContestTableSource).toContain("{ key: 'ends_at', label: '结束时间'")
    expect(contestManagementSource).not.toContain('<nav class="top-tabs"')
  })

  it('keeps audit log and challenge management on shared directory and pagination owners', () => {
    expect(auditLogSource).toContain('AuditLogHeroPanel')
    expect(auditLogSource).toContain('AuditLogDirectoryPanel')
    expect(auditLogDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(auditLogDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/PagePaginationControls.vue'"
    )
    expect(auditLogDirectoryPanelSource).toContain('操作流水')

    expect(challengeManageHeroPanelSource).toContain('workspace-page-header')
    expect(challengeManageDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(challengeManageDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(challengeManageDirectoryPanelSource).toContain('题目目录')
    expect(challengeManageDirectoryPanelSource).toContain('<WorkspaceDirectoryPagination')
  })

  it('keeps class, student, instance, and awd review pages on thin routes plus shared list owners', () => {
    expect(classManageSource).toContain('ClassManageHeroPanel')
    expect(classManageSource).toContain('ClassManageWorkspacePanel')
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )

    expect(studentManageSource).toContain('StudentManageHeroPanel')
    expect(studentManageSource).toContain('StudentManageWorkspacePanel')
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )

    expect(instanceManageSource).toContain('<PlatformInstanceManagementPage />')
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )

    expect(awdReviewDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(awdReviewDirectoryPanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(awdReviewDirectoryPanelSource).toContain('赛事目录')
  })
})
