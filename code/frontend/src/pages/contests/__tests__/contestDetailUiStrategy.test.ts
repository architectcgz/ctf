import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import contestAnnouncementsWorkspaceSectionSource from '@/components/contests/ContestAnnouncementsWorkspaceSection.vue?raw'
import contestTeamWorkspaceSectionSource from '@/components/contests/ContestTeamWorkspaceSection.vue?raw'
import contestTeamDialogsSource from '@/components/contests/ContestTeamDialogs.vue?raw'
import contestChallengeWorkspaceSource from '@/components/contests/ContestChallengeWorkspacePanel.vue?raw'
import contestTeamPanelSource from '@/components/contests/ContestTeamPanel.vue?raw'

const contestWorkspaceSource = [contestDetailSource, contestChallengeWorkspaceSource].join('\n')

describe('contest detail ui strategy', () => {
  it('contest detail route should compose stable contest workspace sections instead of re-inlining the page surface', () => {
    expect(contestDetailSource).toContain(
      "import ContestOverviewPanel from '@/components/contests/ContestOverviewPanel.vue'"
    )
    expect(contestDetailSource).toContain(
      "import ContestChallengeWorkspacePanel from '@/components/contests/ContestChallengeWorkspacePanel.vue'"
    )
    expect(contestDetailSource).toContain(
      "import ContestAnnouncementsWorkspaceSection from '@/components/contests/ContestAnnouncementsWorkspaceSection.vue'"
    )
    expect(contestDetailSource).toContain(
      "import ContestTeamWorkspaceSection from '@/components/contests/ContestTeamWorkspaceSection.vue'"
    )
    expect(contestDetailSource).toContain(
      "import ContestTeamDialogs from '@/components/contests/ContestTeamDialogs.vue'"
    )
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
