import { describe, expect, it } from 'vitest'

import contestAnnouncementsSource from '../ContestAnnouncements.vue?raw'
import contestAnnouncementsWorkspacePanelSource from '@/features/platform-contests/ui/ContestAnnouncementsWorkspacePanel.vue?raw'

describe('ContestAnnouncements workspace extraction', () => {
  it('应将公告发布与历史列表工作区收口到 platform contests feature UI', () => {
    expect(contestAnnouncementsSource).toContain(
      'ContestAnnouncementsWorkspacePanel,'
    )
    expect(contestAnnouncementsSource).toContain("from '@/features/platform-contests'")
    expect(contestAnnouncementsSource).toContain('<ContestAnnouncementsWorkspacePanel')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('Publish')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('历史公告')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('赛事已结束，公告区仅保留查看能力。')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('contest-announcement-submit')
  })
})
