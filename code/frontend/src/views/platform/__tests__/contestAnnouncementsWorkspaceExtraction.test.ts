import { describe, expect, it } from 'vitest'

import contestAnnouncementsSource from '../ContestAnnouncements.vue?raw'
import contestAnnouncementsWorkspacePanelSource from '@/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue?raw'

describe('ContestAnnouncements workspace extraction', () => {
  it('应将公告发布与历史列表工作区抽到独立 platform contest 组件', () => {
    expect(contestAnnouncementsSource).toContain(
      "import ContestAnnouncementsWorkspacePanel from '@/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue'"
    )
    expect(contestAnnouncementsSource).toContain('<ContestAnnouncementsWorkspacePanel')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('Publish')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('历史公告')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('赛事已结束，公告区仅保留查看能力。')
    expect(contestAnnouncementsWorkspacePanelSource).toContain('contest-announcement-submit')
  })
})
