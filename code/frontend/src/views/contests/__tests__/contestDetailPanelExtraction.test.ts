import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import contestAnnouncementsWorkspaceSectionSource from '@/components/contests/ContestAnnouncementsWorkspaceSection.vue?raw'
import contestTeamWorkspaceSectionSource from '@/components/contests/ContestTeamWorkspaceSection.vue?raw'
import contestTeamDialogsSource from '@/components/contests/ContestTeamDialogs.vue?raw'

describe('ContestDetail panel extraction', () => {
  it('应将概览、题目工作区、公告/队伍 section 与队伍对话框壳抽到独立 contests 组件', () => {
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
    expect(contestDetailSource).toContain('<ContestOverviewPanel')
    expect(contestDetailSource).toContain('<ContestChallengeWorkspacePanel')
    expect(contestDetailSource).toContain('<ContestAnnouncementsWorkspaceSection')
    expect(contestDetailSource).toContain('<ContestTeamWorkspaceSection')
    expect(contestDetailSource).toContain('<ContestTeamDialogs')
    expect(contestAnnouncementsWorkspaceSectionSource).toContain('<ContestAnnouncementsPanel')
    expect(contestTeamWorkspaceSectionSource).toContain('<ContestTeamPanel')
    expect(contestTeamDialogsSource).toContain('<CFocusedInputDialog')
  })
})
