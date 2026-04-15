import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'
import auditLogSource from '@/views/admin/AuditLog.vue?raw'
import cheatDetectionSource from '@/views/admin/CheatDetection.vue?raw'
import challengeDetailSource from '@/views/admin/ChallengeDetail.vue?raw'
import challengeManageSource from '@/views/admin/ChallengeManage.vue?raw'
import challengePackageFormatSource from '@/views/admin/ChallengePackageFormat.vue?raw'
import imageManageSource from '@/views/admin/ImageManage.vue?raw'

const journalAdminShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-admin-shell.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('admin journal shell shared styles', () => {
  it('应该在共享样式文件中声明 admin shell 的明暗主题壳层', () => {
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin')
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin.journal-hero')
    expect(journalAdminShellSource).toContain('.journal-shell.journal-shell-admin .journal-panel')
    expect(journalAdminShellSource).toContain(
      "[data-theme='dark'] .journal-shell.journal-shell-admin"
    )
  })

  it('admin 管理页应继续通过 journal-shell-admin 接入共享壳层', () => {
    for (const source of [
      contestOrchestrationSource,
      userGovernanceSource,
      auditLogSource,
      challengeManageSource,
      imageManageSource,
      cheatDetectionSource,
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
      imageManageSource,
      cheatDetectionSource,
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
