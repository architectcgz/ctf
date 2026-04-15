/// <reference types="node" />
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import auditLogSource from '../AuditLog.vue?raw'
import challengeManageSource from '../ChallengeManage.vue?raw'
import challengeDetailSource from '../ChallengeDetail.vue?raw'
import contestEditSource from '../ContestEdit.vue?raw'
import imageManageSource from '../ImageManage.vue?raw'
import cheatDetectionSource from '../CheatDetection.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import awdChallengeConfigPanelSource from '@/components/admin/contest/AWDChallengeConfigPanel.vue?raw'
import awdReadinessSummarySource from '@/components/admin/contest/AWDReadinessSummary.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/admin/contest/AWDReadinessOverrideDialog.vue?raw'
import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'
import adminContestFormDialogSource from '@/components/admin/contest/AdminContestFormDialog.vue?raw'
import adminContestFormPanelSource from '@/components/admin/contest/AdminContestFormPanel.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import adminContestTableSource from '@/components/admin/contest/AdminContestTable.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'

const styleSource = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const journalNotesSource = readFileSync(
  resolve(process.cwd(), 'src/assets/styles/journal-notes.css'),
  'utf8'
)

describe('admin management surface alignment', () => {
  it('audit log should soften table and empty-state borders on dark surfaces', () => {
    expect(auditLogSource).toMatch(
      /--audit-table-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/
    )
    expect(auditLogSource).toMatch(
      /--audit-row-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 62%, transparent\);/
    )
    expect(auditLogSource).toMatch(/class="audit-empty-state[^"]*"/)
    expect(auditLogSource).toContain('border-[var(--audit-table-border)]')
    expect(auditLogSource).toContain('divide-y divide-[var(--audit-row-divider)]')
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
      /\.admin-btn-ghost\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-table-shell\s*\{[\s\S]*border:\s*1px solid var\(--user-table-border\);/s
    )
    expect(userGovernanceSource).toMatch(
      /\.user-table-row\s*\{[\s\S]*border-top:\s*1px solid var\(--user-row-divider\);/s
    )
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">用户目录</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">用户列表</h2>')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">导入用户</h2>')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">导入回执</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">导入用户</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">导入回执</h2>')
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
    expect(contestOrchestrationSource).toMatch(
      /\.admin-btn-ghost\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.contest-empty-state\s*\{[\s\S]*border-top-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);[\s\S]*border-bottom-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/s
    )
    expect(contestOrchestrationSource).toContain(
      '<div class="workspace-overline">Contest Workspace</div>'
    )
    expect(contestOrchestrationSource).toContain('<h1 class="workspace-page-title">赛事管理台</h1>')
    expect(contestOrchestrationSource).toContain(
      '<h2 class="list-heading__title">当前赛事窗口</h2>'
    )
    expect(contestOrchestrationSource).toContain(
      '<div class="workspace-overline">Contest Directory</div>'
    )
    expect(contestOrchestrationSource).toContain('<h2 class="list-heading__title">赛事目录</h2>')
    expect(contestOrchestrationSource).toContain('workspace-directory-empty contest-empty-state')
    expect(contestOrchestrationSource).not.toContain('当前筛选结果')
    expect(contestOrchestrationSource).not.toContain(
      'workspace-tab-heading__title">当前筛选结果</h3>'
    )
    expect(contestOrchestrationSource).not.toContain('class="contest-list-filters"')
    expect(contestOrchestrationSource).toContain('class="contest-filter-grid"')
    expect(contestOrchestrationSource).toContain(
      'class="contest-filter-field contest-filter-field--action"'
    )
    expect(contestOrchestrationSource).not.toContain('class="contest-filter-actions"')
    expect(contestOrchestrationSource).toMatch(
      /\.contest-overview-summary\.metric-panel-default-surface\.metric-panel-workspace-surface\s*\{[\s\S]*--metric-panel-border:\s*color-mix\(in srgb,\s*var\(--workspace-brand\)\s*16%,\s*var\(--workspace-line-soft\)\);[\s\S]*--metric-panel-background:\s*radial-gradient\([\s\S]*linear-gradient\(/s
    )
  })

  it('contest directory rows should expose split schedule columns and dedicated status pills', () => {
    expect(adminContestTableSource).toContain('<span>开始时间</span>')
    expect(adminContestTableSource).toContain('<span>结束时间</span>')
    expect(adminContestTableSource).not.toContain('<span>时间窗口</span>')
    expect(adminContestTableSource).toContain('class="contest-status-pill"')
    expect(adminContestTableSource).toContain('.contest-status-pill--registering')
    expect(adminContestTableSource).toContain('.contest-status-pill--running')
    expect(adminContestTableSource).toContain('.contest-row__starts-at')
    expect(adminContestTableSource).toContain('.contest-row__ends-at')
  })

  it('contest form dialog should adopt the admin workspace dialog shell and section headings', () => {
    expect(adminContestFormDialogSource).toContain('class="contest-form-dialog"')
    expect(adminContestFormDialogSource).toContain(':deep(.contest-form-dialog .el-dialog)')
    expect(adminContestFormDialogSource).toContain('Contest Workspace')
    expect(adminContestFormPanelSource).toContain('<h3 class="list-heading__title">基础信息</h3>')
    expect(adminContestFormPanelSource).toContain('<h3 class="list-heading__title">赛制与时间</h3>')
    expect(adminContestFormPanelSource).toContain(
      'class="contest-form-button contest-form-button--primary"'
    )
  })

  it('contest edit page should use the admin workspace shell and a dedicated back action', () => {
    expect(contestEditSource).toContain('class="workspace-topbar"')
    expect(contestEditSource).toContain('<h1 class="workspace-page-title">编辑竞赛</h1>')
    expect(contestEditSource).toContain('返回竞赛目录')
    expect(contestEditSource).toContain('Contest Editor')
    expect(contestEditSource).toContain('class="workspace-directory-section contest-edit-section"')
  })

  it('awd round inspector traffic filters should stay flattened into the table section instead of using a split intro bar', () => {
    expect(awdRoundInspectorSource).toContain('id="awd-traffic-reset-filters"')
    expect(awdRoundInspectorSource).not.toContain(
      '按攻击方、受害方、题目、状态分桶和路径关键字筛选。'
    )
    expect(awdRoundInspectorSource).not.toContain(
      'class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/60 px-4 py-3"'
    )
  })

  it('awd challenge config and readiness sections should use list-heading for directory blocks', () => {
    expect(awdChallengeConfigPanelSource).toContain('class="workspace-directory-section"')
    expect(awdChallengeConfigPanelSource).toMatch(/class="[^"]*list-heading[^"]*"/)
    expect(awdChallengeConfigPanelSource).toContain('<h3 class="list-heading__title">题目目录</h3>')
    expect(awdChallengeConfigPanelSource).not.toContain(
      'workspace-tab-heading__title">已关联题目</h3>'
    )

    expect(awdReadinessSummarySource).toMatch(/class="[^"]*list-heading[^"]*"/)
    expect(awdReadinessSummarySource).toContain('<h3 class="list-heading__title">系统级阻塞</h3>')
    expect(awdReadinessSummarySource).toContain('<h3 class="list-heading__title">阻塞短名单</h3>')
    expect(awdReadinessSummarySource).not.toContain('workspace-tab-heading__title">系统级阻塞</h3>')
    expect(awdReadinessSummarySource).not.toContain('workspace-tab-heading__title">阻塞短名单</h3>')
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
    expect(challengeDetailSource).toContain(
      '<div class="workspace-overline">Challenge Profile</div>'
    )
    expect(challengeDetailSource).toContain(
      'class="challenge-overview-summary progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(challengeDetailSource).toContain('<h2 class="list-heading__title">基础信息</h2>')
    expect(challengeDetailSource).not.toContain(
      '<div class="journal-note-label">Question Ops</div>'
    )
    expect(challengeDetailSource).toContain('<div class="journal-note-label">Hints</div>')
    expect(challengeDetailSource).toContain('<h2 class="list-heading__title">提示管理</h2>')
    expect(challengeDetailSource).not.toContain('workspace-tab-heading__title">提示管理</h2>')
  })

  it('cheat detection sections should use list-heading for directory headers', () => {
    expect(cheatDetectionSource).toContain('<h2 class="list-heading__title">高频提交账号</h2>')
    expect(cheatDetectionSource).toContain('<h2 class="list-heading__title">共享 IP 线索</h2>')
    expect(cheatDetectionSource).toContain('<h2 class="list-heading__title">快速排查入口</h2>')
    expect(cheatDetectionSource).not.toContain('workspace-tab-heading__title">高频提交账号</h2>')
    expect(cheatDetectionSource).not.toContain('workspace-tab-heading__title">共享 IP 线索</h2>')
    expect(cheatDetectionSource).not.toContain('workspace-tab-heading__title">快速排查入口</h2>')
  })

  it('contest orchestration should place the tab rail at the top of the workspace before the page title', () => {
    expect(contestOrchestrationSource).toContain('class="top-tabs"')
    expect(contestOrchestrationSource).toContain('class="content-pane"')
    expect(contestOrchestrationSource.indexOf('class="top-tabs"')).toBeLessThan(
      contestOrchestrationSource.indexOf('赛事管理台')
    )
  })

  it('admin list pages should use shared directory spacing utilities', () => {
    expect(styleSource).toContain('--workspace-directory-gap-top: 0.75rem;')
    expect(styleSource).toContain('--workspace-directory-gap-pagination: 0.5rem;')
    expect(styleSource).toContain(
      '.workspace-directory-section > :where(.workspace-directory-loading, .workspace-directory-empty, .workspace-directory-list)'
    )
    expect(styleSource).toContain('.workspace-directory-section > .workspace-directory-pagination')

    expect(userGovernanceSource).toContain('class="workspace-directory-section"')
    expect(userGovernanceSource).toContain('class="user-table-shell workspace-directory-list"')
    expect(userGovernanceSource).toContain(
      'class="admin-pagination workspace-directory-pagination"'
    )

    expect(imageManageSource).toContain('class="image-board workspace-directory-section"')
    expect(imageManageSource).toContain('class="image-list workspace-directory-list"')
    expect(imageManageSource).toContain('class="admin-pagination workspace-directory-pagination"')

    expect(auditLogSource).toContain('class="admin-board workspace-directory-section"')
    expect(auditLogSource).toContain('class="audit-table-shell workspace-directory-list')
    expect(auditLogSource).toContain('class="admin-pagination workspace-directory-pagination"')

    expect(challengeManageSource).toContain('class="workspace-directory-section"')
    expect(challengeManageSource).toContain('class="challenge-list workspace-directory-list"')
    expect(challengeManageSource).toContain(
      'class="admin-pagination workspace-directory-pagination"'
    )

    expect(adminContestTableSource).toContain('class="contest-directory workspace-directory-list"')
    expect(adminContestTableSource).toContain(
      'class="admin-pagination workspace-directory-pagination'
    )
  })

  it('admin paginations should expose a shared jump-page control instead of prev-next only', () => {
    expect(userGovernanceSource).toContain('AdminPaginationControls')
    expect(imageManageSource).toContain('AdminPaginationControls')
    expect(auditLogSource).toContain('AdminPaginationControls')
    expect(challengeManageSource).toContain('AdminPaginationControls')
    expect(adminContestTableSource).toContain('AdminPaginationControls')
    expect(awdRoundInspectorSource).toContain('AdminPaginationControls')
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
    expect(userGovernanceSource).toContain('class="admin-summary-grid user-overview-grid')
    expect(contestOrchestrationSource).toContain(
      'class="admin-summary-grid contest-overview-summary'
    )
  })

  it('admin summary cards should explicitly adopt metric-panel utility classes', () => {
    expect(auditLogSource).toContain(
      'class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(auditLogSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(auditLogSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(auditLogSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(auditLogSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )

    expect(journalNotesSource).toContain('.metric-panel-default-surface {')
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
    expect(journalNotesSource).toContain('.progress-card {')
    expect(journalNotesSource).toContain('.progress-card-label {')
    expect(journalNotesSource).toContain('.progress-card-value {')
    expect(journalNotesSource).toContain('.progress-card-hint {')
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
    expect(challengeManageSource).toContain(
      'class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(challengeManageSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(challengeManageSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(challengeManageSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(challengeManageSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(imageManageSource).toContain('class="image-status-strip"')
    expect(imageManageSource).toContain('class="image-status-strip__note">{{ refreshHint }}</div>')
    expect(imageManageSource).not.toContain(
      'class="image-summary-card progress-card metric-panel-card"'
    )

    expect(userGovernanceSource).toContain(
      'class="admin-summary-grid user-overview-grid progress-strip metric-panel-grid metric-panel-default-surface'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note user-overview-stat progress-card metric-panel-card"'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(userGovernanceSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(userGovernanceSource).not.toMatch(
      /\.user-overview-stat \.journal-note-value\s*\{[\s\S]*font-size:\s*clamp\(1\.35rem,\s*2vw,\s*1\.9rem\);/s
    )

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
  })
})
