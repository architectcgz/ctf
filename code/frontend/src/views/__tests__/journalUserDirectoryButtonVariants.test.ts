import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'

const journalUserDirectorySource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-directory.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('journal user directory shared button variants', () => {
  it('应该在共享样式文件中声明目录页按钮的 hover/focus 与 primary 变体', () => {
    expect(journalUserDirectorySource).toMatch(
      /:is\(\.challenge-btn, \.contest-btn, \.notification-btn, \.scoreboard-btn, \.instance-btn\):hover,\s*:is\(\.challenge-btn, \.contest-btn, \.notification-btn, \.scoreboard-btn, \.instance-btn\):focus-visible/s
    )
    expect(journalUserDirectorySource).toContain('.challenge-btn-primary')
    expect(journalUserDirectorySource).toContain('.contest-btn-primary')
    expect(journalUserDirectorySource).toContain('.notification-btn-primary')
    expect(journalUserDirectorySource).toContain('.scoreboard-btn-primary')
    expect(journalUserDirectorySource).toContain('.instance-btn-primary')
  })

  it('目标页面不应继续在 scoped style 中重复声明目录按钮通用态和 primary 变体', () => {
    const challengeStyle = extractScopedStyle(challengeListSource)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn:hover,\s*$/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn:focus-visible\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn-primary\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn-primary:hover,\s*$/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn-primary:focus-visible\s*\{/m)

    const notificationStyle = extractScopedStyle(notificationListSource)
    expect(notificationStyle).not.toMatch(/^\.notification-btn-primary\s*\{/m)

    const instanceStyle = extractScopedStyle(instanceListSource)
    expect(instanceStyle).not.toMatch(/^\.instance-btn-primary\s*\{/m)
  })
})
