/// <reference types="node" />
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import auditLogSource from '../AuditLog.vue?raw'
import auditLogDirectoryPanelSource from '@/components/platform/audit/AuditLogDirectoryPanel.vue?raw'
import auditLogHeroPanelSource from '@/components/platform/audit/AuditLogHeroPanel.vue?raw'
import awdReviewIndexSource from '../AWDReviewIndex.vue?raw'
import awdReviewHeroPanelSource from '@/components/platform/awd-review/AwdReviewHeroPanel.vue?raw'
import awdReviewDirectoryPanelSource from '@/components/platform/awd-review/AwdReviewDirectoryPanel.vue?raw'
import challengeManageSource from '../ChallengeManage.vue?raw'
import challengeManageDirectoryPanelSource from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue?raw'
import challengeManageHeroPanelSource from '@/components/platform/challenge/ChallengeManageHeroPanel.vue?raw'
import cheatDetectionHeroPanelSource from '@/components/platform/cheat/CheatDetectionHeroPanel.vue?raw'
import classManageSource from '../ClassManage.vue?raw'
import classManageHeroPanelSource from '@/components/platform/class/ClassManageHeroPanel.vue?raw'
import classManageWorkspacePanelSource from '@/components/platform/class/ClassManageWorkspacePanel.vue?raw'
import instanceManageSource from '../InstanceManage.vue?raw'
import instanceManageHeroPanelSource from '@/components/platform/instance/InstanceManageHeroPanel.vue?raw'
import instanceManageWorkspacePanelSource from '@/components/platform/instance/InstanceManageWorkspacePanel.vue?raw'
import adminChallengeProfilePanelSource from '@/components/platform/challenge/AdminChallengeProfilePanel.vue?raw'
import challengeWriteupManagePanelSource from '@/components/platform/writeup/ChallengeWriteupManagePanel.vue?raw'
import contestEditSource from '../ContestEdit.vue?raw'
import contestEditTopbarPanelSource from '@/components/platform/contest/ContestEditTopbarPanel.vue?raw'
import contestEditWorkspacePanelSource from '@/components/platform/contest/ContestEditWorkspacePanel.vue?raw'
import imageManageSource from '../ImageManage.vue?raw'
import imageManageHeroPanelSource from '@/components/platform/images/ImageManageHeroPanel.vue?raw'
import studentManageSource from '../StudentManage.vue?raw'
import studentManageHeroPanelSource from '@/components/platform/student/StudentManageHeroPanel.vue?raw'
import studentManageWorkspacePanelSource from '@/components/platform/student/StudentManageWorkspacePanel.vue?raw'
import cheatDetectionSource from '../CheatDetection.vue?raw'
import cheatDetectionReviewPanelsSource from '@/components/platform/cheat/CheatDetectionReviewPanels.vue?raw'
import cheatDetectionSummaryPanelSource from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import awdRoundInspectorSource from '@/components/platform/contest/AWDRoundInspector.vue?raw'
import awdTrafficPanelSource from '@/components/platform/contest/AWDTrafficPanel.vue?raw'
import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'
import awdReadinessChecklistSource from '@/components/platform/contest/AWDReadinessChecklist.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/platform/contest/AWDReadinessOverrideDialog.vue?raw'
import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'
import workspaceDataTableSource from '@/components/common/WorkspaceDataTable.vue?raw'
import workspaceDirectoryToolbarSource from '@/components/common/WorkspaceDirectoryToolbar.vue?raw'
import adminContestFormDialogSource from '@/components/platform/contest/PlatformContestFormDialog.vue?raw'
import adminContestFormPanelSource from '@/components/platform/contest/PlatformContestFormPanel.vue?raw'
import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'
import adminContestTableSource from '@/components/platform/contest/PlatformContestTable.vue?raw'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'

const styleSource = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const journalNotesSource = readFileSync(
  resolve(process.cwd(), 'src/assets/styles/journal-notes.css'),
  'utf8'
)
const auditLogCombinedSource = [
  auditLogSource,
  auditLogDirectoryPanelSource,
  auditLogHeroPanelSource,
].join('\n')
const challengeManageCombinedSource = [
  challengeManageSource,
  challengeManageDirectoryPanelSource,
  challengeManageHeroPanelSource,
].join('\n')
const classManageCombinedSource = [
  classManageSource,
  classManageHeroPanelSource,
  classManageWorkspacePanelSource,
].join('\n')
const cheatDetectionCombinedSource = [
  cheatDetectionSource,
  cheatDetectionWorkspacePanelSource,
  cheatDetectionHeroPanelSource,
  cheatDetectionReviewPanelsSource,
  cheatDetectionSummaryPanelSource,
].join('\n')
const imageManageCombinedSource = [imageManageSource, imageManageHeroPanelSource].join('\n')
const instanceManageCombinedSource = [
  instanceManageSource,
  instanceManageHeroPanelSource,
  instanceManageWorkspacePanelSource,
].join('\n')
const studentManageCombinedSource = [
  studentManageSource,
  studentManageHeroPanelSource,
  studentManageWorkspacePanelSource,
].join('\n')
const awdReviewCombinedSource = [
  awdReviewIndexSource,
  awdReviewHeroPanelSource,
  awdReviewDirectoryPanelSource,
].join('\n')
const contestEditCombinedSource = [
  contestEditSource,
  contestEditTopbarPanelSource,
  contestEditWorkspacePanelSource,
].join('\n')

describe('admin management surface alignment', () => {
  it('audit log should soften table and empty-state borders on dark surfaces', () => {
    expect(auditLogCombinedSource).toMatch(
      /--audit-table-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/
    )
    expect(auditLogCombinedSource).toMatch(
      /--audit-row-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 62%, transparent\);/
    )
    expect(auditLogCombinedSource).toMatch(/class="audit-empty-state[^"]*"/)
    expect(auditLogCombinedSource).toContain('class="audit-list workspace-directory-list"')
    expect(auditLogCombinedSource).toMatch(
      /\.audit-list\s*\{[\s\S]*border:\s*1px solid var\(--audit-table-border\);/s
    )
    expect(auditLogCombinedSource).toMatch(
      /\.audit-list :deep\(\.workspace-data-table__row\)\s*\{[\s\S]*border-bottom-color:\s*var\(--audit-row-divider\);/s
    )
  })

  it('user governance should soften control and table shell borders', () => {
    expect(userGovernanceSource).toMatch(
      /--admin-control-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 76%, transparent\);/
    )
    expect(userGovernanceSource).toMatch(
      /--user-table-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 72%, transparent\);/
    )
    expect(userGovernanceSource).toMatch(
      /--user-row-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 58%, transparent\);/
    )
    expect(userGovernanceSource).toMatch(
      /\.user-panel-actions\s*>\s*\.ui-btn,[\s\S]*\.workspace-directory-empty\s+\.ui-btn\s*\{[\s\S]*--ui-btn-height:\s*2\.75rem;[\s\S]*--ui-btn-radius:\s*1rem;/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-panel-actions\s*>\s*\.ui-btn\.ui-btn--ghost\s*\{[\s\S]*--ui-btn-border:\s*var\(--admin-control-border\);[\s\S]*--ui-btn-background:\s*color-mix\(in srgb,\s*var\(--journal-surface\) 94%, transparent\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-row__actions\s*>\s*\.ui-btn\.ui-btn--secondary\s*\{[\s\S]*--ui-btn-border:\s*var\(--admin-control-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-table-shell\s*\{[\s\S]*--workspace-directory-shell-border:\s*var\(--user-table-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-table-shell\s*\{[\s\S]*--workspace-directory-row-divider:\s*var\(--user-row-divider\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-table-row\s*\{[\s\S]*border-top:\s*1px solid var\(--user-row-divider\);/s
    )
    expect(userGovernanceSource).toMatch(
      /<h2 class="list-heading__title">\s*全部用户\s*<\/h2>/
    )
    expect(userGovernanceSource).toMatch(
      /<h2 class="workspace-page-title">\s*导入用户\s*<\/h2>/
    )
    expect(userGovernanceSource).toMatch(
      /<h2 class="list-heading__title">\s*导入回执\s*<\/h2>/
    )
    expect(userGovernanceSource).toContain('<header class="workspace-tab-heading user-overview-head">')
    expect(userGovernanceSource).toContain('<header class="workspace-tab-heading user-import-head">')
    expect(userGovernanceSource).not.toContain('<header class="list-heading user-overview-head">')
    expect(userGovernanceSource).not.toContain('<header class="list-heading user-import-head">')
    expect(userGovernanceSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(userGovernanceSource).not.toMatch(
      /\.user-directory-section,\s*\.user-import-panel,\s*\.user-import-receipt-section\s*\{/s
    )
    expect(userGovernanceSource).not.toMatch(
      /\.user-directory-head\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(userGovernanceSource).not.toMatch(
      /\.user-directory-section :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(userGovernanceSource).not.toContain('<nav class="top-tabs"')
    expect(userGovernanceSource).toContain('id="user-panel-overview"')
    expect(userGovernanceSource).toContain('id="user-panel-import"')
  })

  it('contest orchestration should soften control and empty-state borders', () => {
    expect(contestOrchestrationSource).toMatch(
      /--admin-control-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 76%, transparent\);/
    )
    expect(contestOrchestrationSource).toMatch(
      /--workspace-panel:\s*color-mix\(in srgb,\s*var\(--color-bg-surface\) 90%, var\(--color-bg-base\)\);/
    )
    expect(contestOrchestrationSource).toMatch(
      /--workspace-line-soft:\s*color-mix\(in srgb,\s*var\(--color-text-primary\) 10%, transparent\);/
    )
    expect(contestOrchestrationSource).toContain('workspace-directory-empty contest-empty-state')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOrchestrationSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(contestOrchestrationSource).toContain('<WorkspaceDirectoryToolbar')
    expect(contestOrchestrationSource).toContain('class="ui-field contest-filter-field"')
    expect(contestOrchestrationSource).toContain('class="ui-control contest-filter-control"')
    expect(contestOrchestrationSource).toMatch(
      /\.contest-empty-state\s*\{[\s\S]*border-top-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);[\s\S]*border-bottom-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/s
    )
    expect(contestOrchestrationSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(contestOrchestrationSource).toMatch(
      /\.contest-directory-section,\s*\.contest-create-panel\s*\{[\s\S]*--workspace-directory-section-padding:\s*var\(--space-5\)\s*var\(--space-5-5\);/s
    )
    expect(contestOrchestrationSource).not.toMatch(
      /\.contest-directory-section :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(contestOrchestrationSource).toMatch(
      /<div class="workspace-overline">\s*Contest Workspace\s*<\/div>/
    )
    expect(contestOrchestrationSource).toMatch(
      /<h1 class="workspace-page-title">\s*竞赛目录\s*<\/h1>/
    )
    expect(contestOrchestrationSource).toMatch(
      /<h2 class="list-heading__title">\s*竞赛列表\s*<\/h2>/
    )
    expect(contestOrchestrationSource).toContain('workspace-directory-empty contest-empty-state')
    expect(contestOrchestrationSource).not.toContain('当前筛选结果')
    expect(contestOrchestrationSource).not.toContain(
      'workspace-tab-heading__title">当前筛选结果</h3>'
    )
    expect(contestOrchestrationSource).not.toContain('<nav class="top-tabs"')
    expect(contestOrchestrationSource).not.toContain('class="contest-list-filters"')
    expect(contestOrchestrationSource).toContain('class="contest-filter-stack"')
    expect(contestOrchestrationSource).not.toContain('class="contest-filter-grid"')
    expect(contestOrchestrationSource).not.toContain('contest-filter-field--action')
    expect(contestOrchestrationSource).not.toContain('class="contest-filter-actions"')
    expect(contestOrchestrationSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(contestOrchestrationSource).toMatch(
      /\.contest-directory-section,\s*\.contest-create-panel\s*\{[\s\S]*--workspace-directory-section-padding:\s*var\(--space-5\)\s*var\(--space-5-5\);/s
    )
    expect(contestOrchestrationSource).not.toMatch(
      /\.contest-directory-section :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.contest-overview-summary\.metric-panel-default-surface\.metric-panel-workspace-surface\s*\{[\s\S]*--metric-panel-border:\s*color-mix\(in srgb,\s*var\(--workspace-brand\)\s*16%,\s*var\(--workspace-line-soft\)\);/s
    )
    expect(contestOrchestrationSource).not.toMatch(
      /\.contest-overview-summary\.metric-panel-default-surface\.metric-panel-workspace-surface\s*\{[\s\S]*--metric-panel-background:/s
    )
  })

  it('contest directory rows should expose split schedule columns and dedicated status pills', () => {
    expect(adminContestTableSource).toContain('<span>开始时间</span>')
    expect(adminContestTableSource).toContain('<span>结束时间</span>')
    expect(adminContestTableSource).not.toContain('<span>时间窗口</span>')
    expect(adminContestTableSource).toContain('class="ui-badge contest-status-pill"')
    expect(adminContestTableSource).toContain('.contest-status-pill--registering')
    expect(adminContestTableSource).toContain('.contest-status-pill--running')
    expect(adminContestTableSource).toContain('.contest-row__starts-at')
    expect(adminContestTableSource).toContain('.contest-row__ends-at')
  })

  it('contest form dialog should adopt the admin workspace dialog shell and section headings', () => {
    expect(adminContestFormDialogSource).toContain('class="contest-form-dialog"')
    expect(adminContestFormDialogSource).toContain(
      "from '@/components/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(adminContestFormDialogSource).toContain(
      ':deep(.contest-form-dialog .modal-template-panel--classic)'
    )
    expect(adminContestFormDialogSource).toContain('Contest Workspace')
    expect(adminContestFormPanelSource).toMatch(
      /<h3 class="list-heading__title">\s*基础信息\s*<\/h3>/
    )
    expect(adminContestFormPanelSource).toMatch(
      /<h3 class="list-heading__title">\s*赛制与时间\s*<\/h3>/
    )
    expect(adminContestFormPanelSource).toContain(
      'class="ui-btn ui-btn--primary contest-form-button contest-form-button--primary"'
    )
  })

  it('contest edit page should use the admin workspace shell and a dedicated back action', () => {
    expect(contestEditCombinedSource).toMatch(/class="[^"]*\bworkspace-topbar\b[^"]*"/)
    expect(contestEditCombinedSource).toMatch(
      /<h1[\s\S]*class="studio-contest-heading"[\s\S]*>\s*\{\{ pageTitle \}\}\s*<\/h1>/
    )
    expect(contestEditCombinedSource).toContain('class="studio-edit-label"')
    expect(contestEditCombinedSource).toContain('返回竞赛目录')
    expect(contestEditCombinedSource).toContain('Contest Studio')
    expect(contestEditCombinedSource).toContain('class="workspace-directory-section contest-edit-section"')
  })

  it('awd round inspector traffic filters should stay flattened into the table section instead of using a split intro bar', () => {
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdTrafficPanelSource).toContain('id="awd-traffic-reset-filters"')
    expect(awdRoundInspectorSource).not.toContain(
      '按攻击方、受害方、题目、状态分桶和路径关键字筛选。'
    )
    expect(awdRoundInspectorSource).not.toContain(
      'class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/60 px-4 py-3"'
    )
  })

  it('awd challenge config and readiness sections should use list-heading for directory blocks', () => {
    expect(awdChallengeConfigPanelSource).toMatch(/class="[^"]*\bworkspace-directory-section\b[^"]*"/)
    expect(awdChallengeConfigPanelSource).toMatch(/class="[^"]*list-heading[^"]*"/)
    expect(awdChallengeConfigPanelSource).toMatch(
      /<h3 class="list-heading__title">\s*题目目录\s*<\/h3>/
    )
    expect(awdChallengeConfigPanelSource).not.toContain(
      'workspace-tab-heading__title">已关联题目</h3>'
    )

    expect(awdReadinessChecklistSource).toMatch(/class="[^"]*list-heading[^"]*"/)
    expect(awdReadinessChecklistSource).toMatch(
      /<h3 class="list-heading__title">\s*系统级阻塞\s*<\/h3>/
    )
    expect(awdReadinessChecklistSource).toMatch(
      /<h3 class="list-heading__title">\s*阻塞短名单\s*<\/h3>/
    )
    expect(awdReadinessChecklistSource).not.toContain('workspace-tab-heading__title">系统级阻塞</h3>')
    expect(awdReadinessChecklistSource).not.toContain('workspace-tab-heading__title">阻塞短名单</h3>')
  })

  it('awd readiness override dialog should use list-heading for override sections', () => {
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="workspace-directory-section readiness-override-section"'
    )
    expect(awdReadinessOverrideDialogSource).toMatch(/class="[^"]*list-heading[^"]*"/)
    expect(awdReadinessOverrideDialogSource).toContain(
      '<h3 class="list-heading__title">系统级阻塞</h3>'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      '<h3 class="list-heading__title">阻塞题目</h3>'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      '<h3 class="list-heading__title">填写本次放行原因</h3>'
    )
    expect(awdReadinessOverrideDialogSource).not.toContain(
      'workspace-tab-heading__title">系统级阻塞</h3>'
    )
    expect(awdReadinessOverrideDialogSource).not.toContain(
      'workspace-tab-heading__title">阻塞题目</h3>'
    )
    expect(awdReadinessOverrideDialogSource).not.toContain(
      'workspace-tab-heading__title">填写本次放行原因</h3>'
    )
  })

  it('awd challenge config dialog should use plain block titles instead of workspace-tab-heading titles', () => {
    expect(awdChallengeConfigDialogSource).toContain('checker-config-block__title')
    expect(awdChallengeConfigDialogSource).not.toContain(
      'workspace-tab-heading__title checker-config-block__title'
    )
    expect(awdChallengeConfigDialogSource).toContain('>最终 JSON 预览</h3>')
    expect(awdChallengeConfigDialogSource).toContain('>最近一次已保存校验</h3>')
    expect(awdChallengeConfigDialogSource).toContain('>试跑 Checker</h3>')
  })

  it('challenge detail hint section should use list-heading for the hint directory header', () => {
    expect(adminChallengeProfilePanelSource).toMatch(
      /<div class="workspace-overline">\s*Challenge Profile\s*<\/div>/
    )
    expect(adminChallengeProfilePanelSource).toContain(
      'class="challenge-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(adminChallengeProfilePanelSource).toContain('<Tags class="h-4 w-4" />')
    expect(adminChallengeProfilePanelSource).toContain('<Gauge class="h-4 w-4" />')
    expect(adminChallengeProfilePanelSource).toContain('<Trophy class="h-4 w-4" />')
    expect(adminChallengeProfilePanelSource).toContain('<CircleDot class="h-4 w-4" />')
    expect(adminChallengeProfilePanelSource).toMatch(
      /<h2 class="list-heading__title">\s*基础信息\s*<\/h2>/
    )
    expect(adminChallengeProfilePanelSource).not.toContain(
      '<div class="journal-note-label">Question Ops</div>'
    )
    expect(adminChallengeProfilePanelSource).toMatch(
      /<div class="journal-note-label">\s*Hints\s*<\/div>/
    )
    expect(adminChallengeProfilePanelSource).toMatch(
      /<h2 class="list-heading__title">\s*提示管理\s*<\/h2>/
    )
    expect(adminChallengeProfilePanelSource).not.toContain('workspace-tab-heading__title">提示管理</h2>')
  })

  it('cheat detection sections should use list-heading for directory headers', () => {
    expect(cheatDetectionHeroPanelSource).toContain('<section class="workspace-hero">')
    expect(cheatDetectionHeroPanelSource).toContain('<CheatDetectionSummaryPanel')
    expect(cheatDetectionCombinedSource).toMatch(
      /\.cheat-workbench\s*\{[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(cheatDetectionCombinedSource).toMatch(
      /\.cheat-directory-section\s*\{[\s\S]*gap:\s*var\(--space-4\);[\s\S]*padding:\s*0;/s
    )
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">高频提交账号</h2>')
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">共享 IP 线索</h2>')
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">审计联动</h2>')
    expect(cheatDetectionCombinedSource).not.toContain(
      'workspace-tab-heading__title">高频提交账号</h2>'
    )
    expect(cheatDetectionCombinedSource).not.toContain(
      'workspace-tab-heading__title">共享 IP 线索</h2>'
    )
    expect(cheatDetectionCombinedSource).not.toContain(
      'workspace-tab-heading__title">审计联动</h2>'
    )
  })

  it('contest orchestration should merge metrics and directory into one workspace instead of keeping a top tab rail', () => {
    expect(contestOrchestrationSource).not.toContain('class="top-tabs"')
    expect(contestOrchestrationSource).toContain('id="contest-panel-overview"')
    expect(contestOrchestrationSource).not.toContain('id="contest-panel-list"')
    expect(contestOrchestrationSource).toContain(
      'class="workspace-directory-section contest-directory-section"'
    )
  })

  it('admin list pages should use shared directory spacing utilities', () => {
    expect(styleSource).toContain('--workspace-directory-gap-top: 0.75rem;')
    expect(styleSource).toContain('--workspace-directory-gap-pagination: 0.5rem;')
    expect(styleSource).toContain('--workspace-directory-page-block-gap: var(--space-5);')
    expect(styleSource).toContain('--workspace-directory-toolbar-gap-bottom: 0;')
    expect(styleSource).toContain('.workspace-directory-section {')
    expect(styleSource).toContain('.workspace-directory-section > .list-heading {')
    expect(styleSource).toContain('.list-heading {')
    expect(styleSource).toContain('.list-heading__title {')
    expect(styleSource).toContain('.workspace-directory-loading,')
    expect(styleSource).toContain('.workspace-directory-list {')
    expect(styleSource).toContain('--workspace-directory-shell-border')
    expect(styleSource).toContain('.workspace-directory-list .workspace-data-table__head-cell')
    expect(styleSource).toContain('.workspace-directory-list .workspace-data-table__row')
    expect(styleSource).toContain('@media (max-width: 768px) {')
    expect(styleSource).toContain('.list-heading {')
    expect(styleSource).toContain(
      '.workspace-directory-section > :where(.workspace-directory-loading, .workspace-directory-empty, .workspace-directory-list)'
    )
    expect(styleSource).toContain('.workspace-directory-section > .workspace-directory-pagination')
    expect(workspaceDataTableSource).not.toMatch(
      /\.workspace-data-table-shell\s*\{[^}]*\bborder:\s*none;/s
    )
    expect(workspaceDataTableSource).not.toMatch(
      /\.workspace-data-table-shell\s*\{[^}]*\bbackground:\s*transparent;/s
    )

    expect(userGovernanceSource).toContain('workspace-directory-section')
    expect(userGovernanceSource).toContain(
      'class="user-table-shell workspace-directory-list user-list"'
    )
    expect(userGovernanceSource).toContain(
      'class="admin-pagination workspace-directory-pagination"'
    )

    expect(imageManageSource).toContain('class="image-board workspace-directory-section"')
    expect(imageManageSource).toContain('class="image-list workspace-directory-list"')
    expect(imageManageSource).toContain('class="admin-pagination workspace-directory-pagination"')

    expect(auditLogCombinedSource).toContain('class="admin-board workspace-directory-section"')
    expect(auditLogCombinedSource).toContain('class="audit-list workspace-directory-list"')
    expect(auditLogCombinedSource).toContain('class="admin-pagination workspace-directory-pagination"')

    expect(challengeManageCombinedSource).toContain(
      'class="workspace-directory-section challenge-manage-directory"'
    )
    expect(challengeManageCombinedSource).toContain('class="challenge-list workspace-directory-list"')
    expect(challengeManageCombinedSource).toContain('class="workspace-directory-loading"')
    expect(challengeManageCombinedSource).toContain('class="workspace-directory-empty"')
    expect(challengeManageCombinedSource).toContain('<WorkspaceDirectoryPagination')

    expect(adminContestTableSource).toContain('class="contest-directory workspace-directory-list"')
    expect(adminContestTableSource).toContain(
      'class="admin-pagination workspace-directory-pagination'
    )

    expect(classManageCombinedSource).toContain('class="workspace-directory-section admin-class-manage-directory"')
    expect(classManageCombinedSource).toContain('class="workspace-directory-list admin-class-manage-table"')
    expect(classManageCombinedSource).toContain('<WorkspaceDirectoryPagination')
    expect(classManageCombinedSource).toMatch(
      /\.admin-class-manage-shell__content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap\);/s
    )

    expect(studentManageCombinedSource).toContain(
      'class="workspace-directory-section admin-student-manage-directory"'
    )
    expect(studentManageCombinedSource).toContain(
      'class="workspace-directory-list admin-student-manage-table"'
    )
    expect(studentManageCombinedSource).toContain('<WorkspaceDirectoryPagination')
    expect(studentManageCombinedSource).toMatch(
      /\.admin-student-manage-shell__content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap\);/s
    )

    expect(instanceManageCombinedSource).toContain(
      'class="workspace-directory-section admin-instance-manage-directory"'
    )
    expect(instanceManageCombinedSource).toContain(
      'class="workspace-directory-list admin-instance-manage-table"'
    )
    expect(instanceManageCombinedSource).toContain('<WorkspaceDirectoryPagination')
    expect(instanceManageCombinedSource).toMatch(
      /\.admin-instance-manage-shell__content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap\);/s
    )

    expect(awdReviewCombinedSource).toContain(
      'class="workspace-directory-section admin-awd-review-directory"'
    )
    expect(awdReviewCombinedSource).toContain(
      'class="workspace-directory-list admin-awd-review-table"'
    )
    expect(awdReviewCombinedSource).toContain('class="workspace-directory-loading"')
    expect(awdReviewCombinedSource).toContain('class="workspace-directory-empty"')
    expect(awdReviewIndexSource).toMatch(
      /\.admin-awd-review-shell__content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap\);/s
    )

    expect(challengeManageSource).toMatch(
      /\.challenge-manage-content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap,\s*var\(--space-5\)\);/s
    )
    expect(challengeManageSource).toMatch(
      /\.challenge-manage-panel\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap,\s*var\(--space-5\)\);/s
    )
    expect(challengeManageCombinedSource).toMatch(
      /\.challenge-manage-hero-panel\s*\{[\s\S]*gap:\s*0;/s
    )
    expect(challengeManageCombinedSource).toMatch(
      /\.challenge-metric-head\s*\{[\s\S]*margin-bottom:\s*var\(--space-2\);/s
    )
    expect(challengeManageCombinedSource).toMatch(
      /\.challenge-metric-value-wrap\s+\.metric-panel-value,\s*[\s\S]*\.challenge-metric-value-wrap\s+\.metric-panel-helper\s*\{[\s\S]*margin-top:\s*0;/s
    )
    expect(workspaceDirectoryToolbarSource).toContain(
      'margin-bottom: var(--workspace-directory-toolbar-gap-bottom, 1.5rem);'
    )
    expect(auditLogSource).not.toMatch(
      /\.admin-board :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(imageManageSource).not.toMatch(
      /\.image-board :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(challengeManageSource).not.toMatch(
      /\.challenge-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:/s
    )
    expect(classManageCombinedSource).not.toMatch(
      /\.admin-class-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(studentManageCombinedSource).not.toMatch(
      /\.admin-student-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(instanceManageCombinedSource).not.toMatch(
      /\.admin-instance-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
  })

  it('admin paginations should expose a shared jump-page control instead of prev-next only', () => {
    expect(userGovernanceSource).toContain('PlatformPaginationControls')
    expect(imageManageSource).toContain('PlatformPaginationControls')
    expect(auditLogCombinedSource).toContain('PlatformPaginationControls')
    expect(challengeManageCombinedSource).toContain('WorkspaceDirectoryPagination')
    expect(adminContestTableSource).toContain('PlatformPaginationControls')
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdTrafficPanelSource).toContain('PlatformPaginationControls')
  })

  it('admin summary grids should use shared summary-grid base styles', () => {
    expect(journalNotesSource).toContain(
      '.journal-shell-admin :is(.admin-summary-grid, .manage-summary-grid, .image-summary-grid)'
    )

    expect(challengeManageSource).not.toMatch(
      /^\.manage-summary-grid\s*\{[\s\S]*display:\s*grid;[\s\S]*grid-template-columns:/m
    )
    expect(auditLogSource).not.toMatch(
      /^\.admin-summary-grid\s*\{[\s\S]*display:\s*grid;[\s\S]*grid-template-columns:/m
    )
    expect(imageManageSource).not.toMatch(
      /^\.image-summary-grid\s*\{[\s\S]*display:\s*grid;[\s\S]*grid-template-columns:/m
    )
    expect(auditLogSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(imageManageSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(classManageSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(studentManageSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(instanceManageSource).not.toMatch(/^\.list-heading\s*\{/m)
    expect(userGovernanceSource).toContain('class="admin-summary-grid user-overview-grid')
    expect(contestOrchestrationSource).toContain(
      'class="admin-summary-grid contest-overview-summary'
    )
    expect(classManageCombinedSource).toContain(
      'class="admin-summary-grid admin-class-manage-shell__summary'
    )
    expect(classManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-class-manage-shell"'
    )
    expect(classManageSource).toMatch(
      /\.admin-class-manage-shell\s*\{[\s\S]*--workspace-line-soft:\s*color-mix\(in srgb,\s*var\(--color-text-primary\) 10%, transparent\);/s
    )
    expect(studentManageCombinedSource).toContain(
      'class="admin-summary-grid admin-student-manage-shell__summary'
    )
    expect(studentManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-student-manage-shell"'
    )
    expect(studentManageSource).toMatch(
      /\.admin-student-manage-shell\s*\{[\s\S]*--workspace-line-soft:\s*color-mix\(in srgb,\s*var\(--color-text-primary\) 10%, transparent\);/s
    )
    expect(instanceManageCombinedSource).toContain(
      'class="admin-summary-grid admin-instance-manage-shell__summary'
    )
    expect(instanceManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-instance-manage-shell"'
    )
    expect(instanceManageSource).toMatch(
      /\.admin-instance-manage-shell\s*\{[\s\S]*--workspace-line-soft:\s*color-mix\(in srgb,\s*var\(--color-text-primary\) 10%, transparent\);/s
    )
    expect(awdReviewCombinedSource).toContain('class="admin-summary-grid admin-awd-review-shell__summary')
  })

  it('admin summary cards should explicitly adopt metric-panel utility classes', () => {
    expect(auditLogCombinedSource).toContain(
      'class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(auditLogCombinedSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(auditLogCombinedSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(auditLogCombinedSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(auditLogCombinedSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )

    expect(journalNotesSource).toContain('.metric-panel-default-surface {')
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
    expect(styleSource).toContain('--workspace-hero-summary-gap: var(--space-5);')
    expect(journalNotesSource).toContain(
      '.workspace-hero + :where(.progress-strip, .admin-summary-grid, .manage-summary-grid),'
    )
    expect(journalNotesSource).toContain(
      'margin-top: var(--workspace-hero-summary-gap, var(--space-5));'
    )
    expect(journalNotesSource).toContain('.progress-card {')
    expect(journalNotesSource).toContain(
      '--metric-panel-padding: var(--space-3-5) var(--space-4) var(--space-3-5);'
    )
    expect(journalNotesSource).toContain('.progress-card-label {')
    expect(journalNotesSource).toContain('.progress-card-value {')
    expect(journalNotesSource).toContain('.progress-card-hint {')
    expect(journalNotesSource).toContain('position: relative;')
    expect(journalNotesSource).toContain('display: block;')
    expect(journalNotesSource).toContain('min-height: 1rem;')
    expect(journalNotesSource).toContain('padding-inline-end: var(--space-7);')
    expect(journalNotesSource).toContain(
      '.progress-card.metric-panel-card .metric-panel-label > :is(svg, .lucide) {'
    )
    expect(journalNotesSource).toContain('top: var(--space-3-5);')
    expect(journalNotesSource).toContain('right: var(--space-4);')
    expect(journalNotesSource).toContain(
      'font-size: var(--metric-panel-label-size, var(--font-size-11));'
    )
    expect(journalNotesSource).toContain(
      'font-size: var(--metric-panel-value-size, var(--font-size-26));'
    )
    expect(journalNotesSource).toContain(
      'font-size: var(--metric-panel-helper-size, var(--font-size-13));'
    )
    expect(journalNotesSource).toContain('--metric-panel-label-size: var(--font-size-11);')
    expect(journalNotesSource).toContain('--metric-panel-radius: var(--workspace-radius-lg, 18px);')
    expect(journalNotesSource).toContain('--metric-panel-value-size: var(--font-size-26);')
    expect(journalNotesSource).toContain('--metric-panel-helper-size: var(--font-size-13);')
    expect(journalNotesSource).toContain('--metric-panel-helper-line-height: 1.7;')
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
    expect(journalNotesSource).toContain(
      'var(--workspace-brand, var(--journal-accent, var(--color-primary-default))) 18%'
    )
    expect(journalNotesSource).toContain('--workspace-panel-soft,')
    expect(journalNotesSource).toMatch(
      /\.journal-shell-admin :is\(\.admin-summary-grid, \.manage-summary-grid, \.image-summary-grid\) > \.journal-note \.journal-note-label\s*\{[\s\S]*color:\s*var\(--metric-panel-label-color,\s*var\(--journal-muted\)\);/s
    )
    expect(journalNotesSource).toMatch(
      /\.journal-shell-admin :is\(\.admin-summary-grid, \.manage-summary-grid, \.image-summary-grid\) > \.journal-note \.journal-note-value\s*\{[\s\S]*color:\s*var\(--metric-panel-value-color,\s*var\(--journal-ink\)\);/s
    )
    expect(journalNotesSource).toMatch(
      /\.journal-shell-admin :is\(\.admin-summary-grid, \.manage-summary-grid, \.image-summary-grid\) > \.journal-note \.journal-note-helper\s*\{[\s\S]*color:\s*var\(--metric-panel-helper-color,\s*var\(--journal-muted\)\);/s
    )
    expect(journalNotesSource).toContain(
      '.journal-shell-admin.journal-notes-card .journal-note:not(.metric-panel-card) {'
    )
    expect(journalNotesSource).not.toContain(
      '.journal-shell-admin.journal-notes-card .journal-note {'
    )
    expect(challengeManageCombinedSource).toContain(
      'class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(challengeManageCombinedSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(challengeManageCombinedSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(challengeManageCombinedSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(challengeManageCombinedSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(imageManageCombinedSource).toContain('class="image-status-strip"')
    expect(imageManageCombinedSource).toContain(
      'class="image-status-strip__note">{{ refreshHint }}</div>'
    )
    expect(imageManageCombinedSource).not.toContain(
      'class="image-summary-card progress-card metric-panel-card"'
    )

    expect(userGovernanceSource).toContain(
      'class="admin-summary-grid user-overview-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface'
    )
    expect(userGovernanceSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(userGovernanceSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(userGovernanceSource).not.toContain('user-overview-stat')

    expect(contestOrchestrationSource).toContain(
      'class="admin-summary-grid contest-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(contestOrchestrationSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(contestOrchestrationSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestOrchestrationSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestOrchestrationSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )

    expect(classManageCombinedSource).toContain(
      'class="admin-summary-grid admin-class-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(studentManageCombinedSource).toContain(
      'class="admin-summary-grid admin-student-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(instanceManageCombinedSource).toContain(
      'class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(awdReviewCombinedSource).toContain(
      'class="admin-summary-grid admin-awd-review-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
  })

  it('workspace summary headers should divide actions from metric cards', () => {
    expect(userGovernanceSource).toMatch(
      /\.user-overview-head\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.contest-overview-head\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(awdReviewHeroPanelSource).toMatch(
      /\.admin-awd-review-shell__hero\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(classManageHeroPanelSource).toMatch(
      /\.workspace-hero\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(studentManageHeroPanelSource).toMatch(
      /\.workspace-hero\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(instanceManageHeroPanelSource).toMatch(
      /\.workspace-hero\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(adminChallengeProfilePanelSource).toMatch(
      /\.challenge-detail-header\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft,/s
    )
    expect(challengeWriteupManagePanelSource).toMatch(
      /\.writeup-manage-header\s*\{[\s\S]*padding-bottom:\s*var\(--space-6\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft,/s
    )
  })
})
