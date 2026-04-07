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

describe('profile journal shared utility styles', () => {
  it('应该在共享样式文件中声明 profile 页复用的 tech-font 工具类', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user .tech-font')
  })

  it('profile 页面不应继续在 scoped style 中重复声明 tech-font', () => {
    for (const source of [securitySettingsSource, skillProfileSource, userProfileSource]) {
      expect(extractScopedStyle(source)).not.toMatch(/^\.tech-font\s*\{/m)
    }
  })
})
