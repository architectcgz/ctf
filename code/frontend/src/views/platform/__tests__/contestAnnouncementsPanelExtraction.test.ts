import { describe, expect, it } from 'vitest'

import contestAnnouncementsSource from '../ContestAnnouncements.vue?raw'
import contestAnnouncementsTopbarPanelSource from '@/components/platform/contest/ContestAnnouncementsTopbarPanel.vue?raw'

describe('ContestAnnouncements panel extraction', () => {
  it('应将竞赛公告顶栏抽到独立 platform contest 组件', () => {
    expect(contestAnnouncementsSource).toContain(
      "import ContestAnnouncementsTopbarPanel from '@/components/platform/contest/ContestAnnouncementsTopbarPanel.vue'"
    )
    expect(contestAnnouncementsSource).toContain('<ContestAnnouncementsTopbarPanel')
    expect(contestAnnouncementsTopbarPanelSource).toContain('Contest Announcements')
    expect(contestAnnouncementsTopbarPanelSource).toContain('class="contest-announcement-topbar"')
    expect(contestAnnouncementsTopbarPanelSource).toContain('class="contest-announcement-back"')
  })
})
