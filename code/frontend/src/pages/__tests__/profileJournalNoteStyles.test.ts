import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/features/profile/ui/SecuritySettingsWorkspaceShell.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/features/skill-profile/ui/SkillProfileWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/features/profile/ui/UserProfileWorkspaceShell.vue?raw'

const journalNotesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-notes.css`,
  'utf-8'
)
const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`

describe('profile journal note shared styles', () => {
  it('应该在共享样式文件中声明 profile 页复用的 eyebrow soft 与 note 基础规则', () => {
    expect(journalNotesSource).toContain('.journal-shell .journal-eyebrow-soft')
    expect(journalNotesSource).toContain('.journal-shell .journal-note-label')
    expect(journalNotesSource).toContain('.journal-shell .journal-note-helper')
  })

  it('profile 页面不应继续在局部样式里重写共享的基础 note 规则', () => {
    for (const source of [
      userProfileWorkspaceSource,
      skillProfileWorkspaceSource,
      securitySettingsWorkspaceSource,
    ]) {
      expect(source).not.toMatch(/^\.journal-eyebrow-soft\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-label\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-helper\s*\{/m)
    }
  })

  it('security settings 仍可通过变量保留自己的 note 密度', () => {
    expect(securitySettingsWorkspaceSource).toContain('--journal-note-label-size: 0.72rem;')
    expect(securitySettingsWorkspaceSource).toContain('--journal-note-label-weight: 700;')
    expect(securitySettingsWorkspaceSource).toContain('--journal-note-label-spacing: 0.16em;')
    expect(securitySettingsWorkspaceSource).toContain('--journal-note-helper-line-height: 1.45;')
  })
})
