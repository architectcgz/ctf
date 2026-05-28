import { describe, expect, it } from 'vitest'

import contestAnnouncementsSource from '../ContestAnnouncements.vue?raw'
import contestAnnouncementsTopbarPanelSource from '@/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue?raw'

describe('ContestAnnouncements panel extraction', () => {
  it('应将竞赛公告顶栏收口到 platform contests feature UI', () => {
    expect(contestAnnouncementsSource).toContain(
      'ContestAnnouncementsTopbarPanel,'
    )
    expect(contestAnnouncementsSource).toContain("from '@/features/platform-contests'")
    expect(contestAnnouncementsSource).toContain('<ContestAnnouncementsTopbarPanel')
    expect(contestAnnouncementsTopbarPanelSource).toContain('Contest Announcements')
    expect(contestAnnouncementsTopbarPanelSource).toContain('class="contest-announcement-topbar"')
    expect(contestAnnouncementsTopbarPanelSource).toContain('class="contest-announcement-back"')
  })
})
