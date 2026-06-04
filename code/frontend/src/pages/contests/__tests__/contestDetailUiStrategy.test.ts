import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import contestDetailWorkspaceSource from '@/widgets/contest-detail-workspace/ContestDetailWorkspace.vue?raw'
import contestAnnouncementsWorkspaceSectionSource from '@/features/contest-detail/ui/ContestAnnouncementsWorkspaceSection.vue?raw'
import contestTeamWorkspaceSectionSource from '@/features/contest-detail/ui/ContestTeamWorkspaceSection.vue?raw'
import contestTeamDialogsSource from '@/features/contest-detail/ui/ContestTeamDialogs.vue?raw'

describe('contest detail ui strategy', () => {
  it('contest detail route should delegate page surface to the widget instead of re-inlining workspace sections', () => {
    expect(contestDetailSource).toContain("import { ContestDetailWorkspace } from '@/widgets/contest-detail-workspace'")
    expect(contestDetailSource).toContain('<ContestDetailWorkspace')
    expect(contestDetailSource).not.toContain('ContestOverviewPanel')
    expect(contestDetailSource).not.toContain('ContestAnnouncementsWorkspaceSection')
    expect(contestDetailSource).not.toContain('ContestTeamWorkspaceSection')
    expect(contestDetailSource).not.toContain('ContestTeamDialogs')
    expect(contestDetailWorkspaceSource).toContain('ContestOverviewPanel')
    expect(contestDetailWorkspaceSource).toContain('ContestAnnouncementsWorkspaceSection')
    expect(contestDetailWorkspaceSource).toContain('ContestTeamWorkspaceSection')
    expect(contestDetailWorkspaceSource).toContain('ContestTeamDialogs')
    expect(contestAnnouncementsWorkspaceSectionSource).toContain('<ContestAnnouncementsPanel')
    expect(contestTeamWorkspaceSectionSource).toContain('<ContestTeamPanel')
    expect(contestTeamDialogsSource).toContain('<CFocusedInputDialog')
  })
})
