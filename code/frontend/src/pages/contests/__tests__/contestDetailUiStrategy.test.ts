import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import contestAnnouncementsWorkspaceSectionSource from '@/features/contest-detail/ui/ContestAnnouncementsWorkspaceSection.vue?raw'
import contestTeamWorkspaceSectionSource from '@/features/contest-detail/ui/ContestTeamWorkspaceSection.vue?raw'
import contestTeamDialogsSource from '@/features/contest-detail/ui/ContestTeamDialogs.vue?raw'
import contestChallengeWorkspaceSource from '@/features/contest-detail/ui/ContestChallengeWorkspacePanel.vue?raw'
import contestTeamPanelSource from '@/features/contest-detail/ui/ContestTeamPanel.vue?raw'

const contestWorkspaceSource = [contestDetailSource, contestChallengeWorkspaceSource].join('\n')

describe('contest detail ui strategy', () => {
  it('contest detail route should compose stable contest workspace sections instead of re-inlining the page surface', () => {
    expect(contestDetailSource).toContain("ContestOverviewPanel,")
    expect(contestDetailSource).toContain("ContestChallengeWorkspacePanel,")
    expect(contestDetailSource).toContain("ContestAnnouncementsWorkspaceSection,")
    expect(contestDetailSource).toContain("ContestTeamWorkspaceSection,")
    expect(contestDetailSource).toContain("ContestTeamDialogs,")
    expect(contestDetailSource).toContain("} from '@/features/contest-detail'")
    expect(contestAnnouncementsWorkspaceSectionSource).toContain('<ContestAnnouncementsPanel')
    expect(contestTeamWorkspaceSectionSource).toContain('<ContestTeamPanel')
    expect(contestTeamDialogsSource).toContain('<CFocusedInputDialog')
  })

  it('contest detail student actions should stay on shared control and button primitives', () => {
    expect(contestWorkspaceSource).toMatch(/class="ui-control-wrap(?:\s+[^\"]+)?"/)
    expect(contestWorkspaceSource).toContain('class="ui-control"')
    expect(contestWorkspaceSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestTeamPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestTeamPanelSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(contestWorkspaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
  })
})
