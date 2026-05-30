import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/components/profile/SecuritySettingsWorkspaceShell.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/components/profile/SkillProfileWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/components/profile/UserProfileWorkspaceShell.vue?raw'

const journalUserShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`

describe('profile journal shared utility styles', () => {
  it('应该在共享样式文件中声明 profile 页复用的 tech-font 工具类', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user .tech-font')
  })

  it('profile 页面不应继续在 scoped style 中重复声明 tech-font', () => {
    for (const source of [
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
    ]) {
      expect(extractScopedStyle(source)).not.toMatch(/^\.tech-font\s*\{/m)
    }
  })
})
