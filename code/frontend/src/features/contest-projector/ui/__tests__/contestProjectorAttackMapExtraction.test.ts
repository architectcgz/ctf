import { describe, expect, it } from 'vitest'

import contestProjectorAttackBoardSource from '@/features/contest-projector/ui/ContestProjectorAttackBoard.vue?raw'
import contestProjectorAttackMapSource from '@/features/contest-projector/ui/ContestProjectorAttackMap.vue?raw'
import contestProjectorAttackMapStatsSidebarSource from '@/features/contest-projector/ui/ContestProjectorAttackMapStatsSidebar.vue?raw'
import contestProjectorAttackMapTeamSidebarSource from '@/features/contest-projector/ui/ContestProjectorAttackMapTeamSidebar.vue?raw'

describe('contest projector attack map extraction', () => {
  it('ContestProjectorAttackMap 应把左右侧栏和中央 board 壳体下沉到独立子组件，而不是继续内联整段结构', () => {
    expect(contestProjectorAttackMapSource).toContain('<ContestProjectorAttackMapTeamSidebar')
    expect(contestProjectorAttackMapSource).toContain('<ContestProjectorAttackBoard')
    expect(contestProjectorAttackMapSource).toContain('<ContestProjectorAttackMapStatsSidebar')
    expect(contestProjectorAttackMapSource).toContain('<ContestProjectorAttackDetailOverlay')
    expect(contestProjectorAttackMapSource).not.toContain('class="legend-grid"')
    expect(contestProjectorAttackMapSource).not.toContain('class="map-team-node"')
    expect(contestProjectorAttackMapSource).not.toContain('class="rank-block panel-drilldown"')
    expect(contestProjectorAttackMapSource).not.toContain('const teamDragOffsets = ref<Record<string, TeamDragOffset>>({})')
    expect(contestProjectorAttackMapSource).not.toContain('function updateBeams(): void {')
  })

  it('ContestProjectorAttackBoard 应成为唯一 board DOM mechanics owner，sidebars 只承接稳定展示区块', () => {
    expect(contestProjectorAttackBoardSource).toContain('useProjectorAttackBoard')
    expect(contestProjectorAttackBoardSource).toContain('class="attack-beam-layer"')
    expect(contestProjectorAttackBoardSource).toContain('class="map-team-node"')
    expect(contestProjectorAttackMapTeamSidebarSource).toContain('class="legend-grid"')
    expect(contestProjectorAttackMapTeamSidebarSource).toContain('class="team-list-block panel-drilldown"')
    expect(contestProjectorAttackMapStatsSidebarSource).toContain('class="rank-block panel-drilldown"')
    expect(contestProjectorAttackMapStatsSidebarSource).toContain('class="attack-stat-summary"')
  })
})
