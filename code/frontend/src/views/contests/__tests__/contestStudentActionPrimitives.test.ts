import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'
import contestChallengeWorkspaceSource from '@/components/contests/ContestChallengeWorkspacePanel.vue?raw'
import contestTeamPanelSource from '@/components/contests/ContestTeamPanel.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'

describe('contest student action primitives', () => {
  it('ContestDetail 应接入共享 ui 按钮与输入控件原语', () => {
    const contestWorkspaceSource = `${contestDetailSource}\n${contestChallengeWorkspaceSource}`

    expect(contestWorkspaceSource).toMatch(/class="ui-control-wrap(?:\s+[^\"]+)?"/)
    expect(contestWorkspaceSource).toContain('class="ui-control"')
    expect(contestWorkspaceSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestTeamPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestTeamPanelSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.flag-submit__input\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.flag-submit__input:focus\s*\{/m)
  })

  it('AWD 工作台应保留当前战场控件原语与稳定选择器', () => {
    expect(awdWorkspaceSource).toContain('class="hud-refresh-btn"')
    expect(awdWorkspaceSource).toContain('class="asset-btn asset-btn--primary"')
    expect(awdWorkspaceSource).toContain('class="war-room-select"')
    expect(awdWorkspaceSource).toContain('class="war-room-input"')
    expect(awdWorkspaceSource).toContain('class="flag-input"')
    expect(awdWorkspaceSource).toContain('class="submit-btn"')
    expect(awdWorkspaceSource).toContain('id="awd-target-challenge"')
    expect(awdWorkspaceSource).toContain('id="awd-target-search"')
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
  })
})
