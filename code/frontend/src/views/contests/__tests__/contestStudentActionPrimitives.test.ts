import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'
import awdDefenseServiceListSource from '@/components/contests/awd/AWDDefenseServiceList.vue?raw'
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
    const awdActionSurfaceSource = `${awdWorkspaceSource}\n${awdDefenseServiceListSource}`

    expect(awdActionSurfaceSource).toContain('class="hud-refresh-btn"')
    expect(awdActionSurfaceSource).toContain('class="asset-btn asset-btn--primary"')
    expect(awdActionSurfaceSource).toContain('class="war-room-select"')
    expect(awdActionSurfaceSource).toContain('class="war-room-input"')
    expect(awdActionSurfaceSource).toContain('class="flag-input"')
    expect(awdActionSurfaceSource).toContain('class="submit-btn"')
    expect(awdActionSurfaceSource).toContain('id="awd-target-challenge"')
    expect(awdActionSurfaceSource).toContain('id="awd-target-search"')
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
  })
})
