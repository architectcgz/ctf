import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/features/platform/contests/ui/ContestOrchestrationPage.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import cheatDetectionWorkspaceSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import challengeDetailSource from '@/pages/platform/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeManageSource from '@/features/platform/challenges/ui/ChallengeManagePage.vue?raw'
import challengeImportManageSource from '@/views/platform/ChallengeImportManage.vue?raw'
import challengePackageFormatSource from '@/views/platform/ChallengePackageFormat.vue?raw'
import imageManageSource from '@/views/platform/ImageManage.vue?raw'

const journalAdminShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-admin-shell.css`,
  'utf-8'
)
const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('admin journal shell shared styles', () => {
  it('应该在共享样式文件中声明 admin shell 的明暗主题壳层', () => {
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin')
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin.journal-hero')
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin .journal-panel')
    expect(journalAdminShellSource).toMatch(
      /\.journal-shell\.journal-shell-admin\.journal-hero\s*\{[\s\S]*border-radius:\s*0\s*!important;/s
    )
    expect(journalAdminShellSource).toContain(
      "[data-theme='dark'] .journal-shell.journal-shell-admin"
    )
  })

  it('admin 壳层在暗色模式下应为 secondary 与 ghost 按钮提供低对比深色 token，而不是白底按钮', () => {
    expect(journalAdminShellSource).toContain("[data-theme='dark'] .journal-shell.journal-shell-admin")
    expect(journalAdminShellSource).toMatch(
      /\[data-theme='dark'\] \.journal-shell\.journal-shell-admin\s*\{[\s\S]*--ui-btn-secondary-background:\s*color-mix\(/s
    )
    expect(journalAdminShellSource).toMatch(
      /\[data-theme='dark'\] \.journal-shell\.journal-shell-admin\s*\{[\s\S]*--ui-btn-secondary-border:\s*color-mix\(/s
    )
    expect(journalAdminShellSource).toMatch(
      /\[data-theme='dark'\] \.journal-shell\.journal-shell-admin\s*\{[\s\S]*--ui-btn-ghost-hover-background:\s*color-mix\(/s
    )
  })

  it('admin 管理页应继续通过 journal-shell-admin 接入共享壳层', () => {
    for (const source of [
      contestOrchestrationSource,
      userGovernanceSource,
      auditLogSource,
      challengeManageSource,
      challengeImportManageSource,
      imageManageSource,
      cheatDetectionWorkspaceSource,
      challengeDetailSource,
      challengePackageFormatSource,
    ]) {
      expect(source).toContain('journal-shell-admin')
    }
  })

  it('这些页面不应继续本地重写整套 admin shell 与 dark hero', () => {
    for (const source of [
      contestOrchestrationSource,
      userGovernanceSource,
      auditLogSource,
      challengeManageSource,
      challengeImportManageSource,
      imageManageSource,
      cheatDetectionWorkspaceSource,
      challengeDetailSource,
      challengePackageFormatSource,
    ]) {
      const style = extractScopedStyle(source)

      expect(style).not.toContain('--journal-ink: var(--color-text-primary);')
      expect(style).not.toMatch(/^\.journal-hero\b/m)
      expect(style).not.toMatch(/^:global\(\[data-theme='dark'\]\) \.journal-shell\b/m)
      expect(style).not.toMatch(/^:global\(\[data-theme='dark'\]\) \.journal-hero\b/m)
    }
  })
})
