import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'

describe('ContestDetail panel extraction', () => {
  it('应将公告面板和队伍面板抽到独立 contests 组件', () => {
    expect(contestDetailSource).toContain("import ContestAnnouncementsPanel from '@/components/contests/ContestAnnouncementsPanel.vue'")
    expect(contestDetailSource).toContain("import ContestTeamPanel from '@/components/contests/ContestTeamPanel.vue'")
    expect(contestDetailSource).toContain('<ContestAnnouncementsPanel')
    expect(contestDetailSource).toContain('<ContestTeamPanel')
  })
})
