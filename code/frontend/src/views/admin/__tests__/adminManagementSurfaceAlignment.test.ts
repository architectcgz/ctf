/// <reference types="node" />
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import auditLogSource from '../AuditLog.vue?raw'
import challengeManageSource from '../ChallengeManage.vue?raw'
import imageManageSource from '../ImageManage.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import adminContestTableSource from '@/components/admin/contest/AdminContestTable.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'

const styleSource = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const journalNotesSource = readFileSync(resolve(process.cwd(), 'src/assets/styles/journal-notes.css'), 'utf8')

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
  })

  it('contest orchestration should soften control and empty-state borders', () => {
    expect(contestOrchestrationSource).toMatch(
      /--admin-control-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 76%, transparent\);/
    )
    expect(contestOrchestrationSource).toContain('class="contest-empty-state"')
    expect(contestOrchestrationSource).toMatch(
      /\.admin-btn-ghost\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s
    )
    expect(contestOrchestrationSource).toMatch(
      /\.contest-empty-state\s*\{[\s\S]*border-top-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);[\s\S]*border-bottom-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/s
    )
  })

  it('contest orchestration should place the tab rail under the workspace topbar before the page title', () => {
    expect(contestOrchestrationSource).toContain('class="workspace-topbar"')
    expect(contestOrchestrationSource).toContain('class="top-tabs"')
    expect(contestOrchestrationSource).toContain('class="content-pane"')
    expect(contestOrchestrationSource.indexOf('class="top-tabs"')).toBeLessThan(
      contestOrchestrationSource.indexOf('赛事编排台')
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
    expect(contestOrchestrationSource).toContain('class="admin-summary-grid contest-overview-summary')
  })

  it('admin summary cards should explicitly adopt metric-panel utility classes', () => {
    expect(auditLogSource).toContain(
      'class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(auditLogSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(auditLogSource).toContain('class="journal-note-label progress-card-label metric-panel-label"')
    expect(auditLogSource).toContain('class="journal-note-value progress-card-value metric-panel-value"')
    expect(auditLogSource).toContain('class="journal-note-helper progress-card-hint metric-panel-helper"')

    expect(journalNotesSource).toContain('.metric-panel-default-surface {')
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
    expect(journalNotesSource).toContain('.progress-card {')
    expect(journalNotesSource).toContain('.progress-card-label {')
    expect(journalNotesSource).toContain('.progress-card-value {')
    expect(journalNotesSource).toContain('.progress-card-hint {')
    expect(journalNotesSource).toContain('--metric-panel-radius: var(--workspace-radius-lg, 18px);')
    expect(journalNotesSource).toContain('--metric-panel-value-size: var(--font-size-26);')
    expect(journalNotesSource).toContain('--metric-panel-helper-line-height: 1.7;')
    expect(challengeManageSource).toContain('class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"')
    expect(challengeManageSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(challengeManageSource).toContain('class="journal-note-label progress-card-label metric-panel-label"')
    expect(challengeManageSource).toContain('class="journal-note-value progress-card-value metric-panel-value"')
    expect(challengeManageSource).toContain('class="journal-note-helper progress-card-hint metric-panel-helper"')
    expect(imageManageSource).toContain(
      'class="image-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(imageManageSource).toContain('class="image-summary-card progress-card metric-panel-card"')
    expect(imageManageSource).toContain('class="progress-card-hint metric-panel-helper"')

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
      'class="admin-summary-grid contest-overview-summary mt-5 progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(contestOrchestrationSource).toContain('class="journal-note progress-card metric-panel-card"')
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
