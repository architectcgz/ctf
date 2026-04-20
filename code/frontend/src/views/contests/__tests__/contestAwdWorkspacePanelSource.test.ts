import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台分区 overline 应接入 workspace-overline 共享语义', () => {
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">Defense</div>')
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">My Services</div>')
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">Targets</div>')
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">Status</div>')
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">Scoreboard</div>')
    expect(awdWorkspaceSource).toContain('<div class="workspace-overline">Feedback</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">Defense</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">My Services</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">Targets</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">Status</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">Scoreboard</div>')
    expect(awdWorkspaceSource).not.toContain('<div class="contest-overline">Feedback</div>')
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-overline\s*\{/m)
  })
})
