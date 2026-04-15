import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'

const journalUserShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('profile journal shared button styles', () => {
  it('应该在共享样式文件中声明 profile 页复用的 journal 按钮规则', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user .journal-btn')
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user .journal-btn:hover')
    expect(journalUserShellSource).toContain(
      '.journal-shell.journal-shell-user .journal-btn:focus-visible'
    )
    expect(journalUserShellSource).toContain(
      '.journal-shell.journal-shell-user .journal-btn:disabled'
    )
    expect(journalUserShellSource).toContain(
      '.journal-shell.journal-shell-user .journal-btn--primary'
    )
  })

  it('profile 页面不应继续在 scoped style 中重复声明 journal 按钮基础规则', () => {
    for (const source of [securitySettingsSource, skillProfileSource, userProfileSource]) {
      const scopedStyle = extractScopedStyle(source)
      expect(scopedStyle).not.toMatch(/^\.journal-btn\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn:focus-visible\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn:disabled\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn--primary\s*\{/m)
    }
  })
})
