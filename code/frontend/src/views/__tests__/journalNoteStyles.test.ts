import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'
import auditLogSource from '@/views/platform/AuditLog.vue?raw'
import challengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'
import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import challengePackageFormatSource from '@/views/platform/ChallengePackageFormat.vue?raw'
import cheatDetectionWorkspaceSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import imageManageSource from '@/views/platform/ImageManage.vue?raw'

const journalNotesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-notes.css`,
  'utf-8'
)

describe('journal note shared styles', () => {
  it('应该在共享样式文件中声明 admin journal 的基础样式与变体', () => {
    expect(journalNotesSource).toContain('.journal-shell-admin .journal-eyebrow')
    expect(journalNotesSource).toContain('.journal-shell-admin .journal-note-label')
    expect(journalNotesSource).toContain('.journal-shell-admin .journal-note-value')
    expect(journalNotesSource).toContain('.journal-shell-admin .journal-note-helper')
    expect(journalNotesSource).toContain('.journal-shell-admin .journal-divider')
    expect(journalNotesSource).toContain(
      '.journal-shell-admin.journal-notes-card .journal-note:not(.metric-panel-card) {'
    )
    expect(journalNotesSource).toContain(
      '.journal-shell-admin.journal-notes-rail .journal-note:not(.metric-panel-card) {'
    )
    expect(journalNotesSource).not.toContain(
      '.journal-shell-admin.journal-notes-card .journal-note {'
    )
    expect(journalNotesSource).not.toContain(
      '.journal-shell-admin.journal-notes-rail .journal-note {'
    )
  })

  it('目标 admin 页面应显式声明共享作用域和 note 变体', () => {
    expect(auditLogSource).toContain('journal-shell journal-shell-admin')
    expect(cheatDetectionWorkspaceSource).toContain('journal-shell journal-shell-admin')
    expect(contestOrchestrationSource).toContain(
      'journal-shell journal-shell-admin journal-notes-card'
    )
    expect(userGovernanceSource).toContain('journal-shell journal-shell-admin journal-notes-card')
    expect(challengeManageSource).toContain('journal-shell journal-shell-admin journal-notes-card')
    expect(imageManageSource).toContain('journal-shell journal-shell-admin journal-notes-rail')
    expect(challengePackageFormatSource).toContain('journal-shell journal-shell-admin')
    expect(challengeDetailSource).toContain('journal-shell journal-shell-admin')
  })

  it('admin journal 页面不应继续在局部样式里重写共享的 eyebrow 和 divider', () => {
    for (const source of [
      auditLogSource,
      challengeDetailSource,
      challengeManageSource,
      challengePackageFormatSource,
      cheatDetectionWorkspaceSource,
      contestOrchestrationSource,
      imageManageSource,
      userGovernanceSource,
    ]) {
      expect(source).not.toMatch(/\.journal-eyebrow\s*\{/s)
      expect(source).not.toMatch(/\.journal-divider\s*\{/s)
    }
  })

  it('带 note 的 admin 页面不应继续在局部样式里重写共享的 note 基础样式', () => {
    for (const source of [
      auditLogSource,
      challengeManageSource,
      cheatDetectionWorkspaceSource,
      contestOrchestrationSource,
      imageManageSource,
      userGovernanceSource,
    ]) {
      expect(source).not.toMatch(/^\.journal-note\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-label\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-value\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-helper\s*\{/m)
    }
  })
})
