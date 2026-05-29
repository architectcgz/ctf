import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdTrafficEventTableSource from '@/features/awd-inspector/ui/AWDTrafficEventTable.vue?raw'
import awdTrafficIntelligenceGridSource from '@/features/awd-inspector/ui/AWDTrafficIntelligenceGrid.vue?raw'
import awdTrafficPanelSource from '@/features/awd-inspector/ui/AWDTrafficPanel.vue?raw'
import awdTrafficSummaryBandSource from '@/features/awd-inspector/ui/AWDTrafficSummaryBand.vue?raw'

const awdTrafficCombinedSource = [
  awdTrafficPanelSource,
  awdTrafficSummaryBandSource,
  awdTrafficIntelligenceGridSource,
  awdTrafficEventTableSource,
  readFileSync(resolve(process.cwd(), 'src/features/awd-inspector/ui/awdTrafficPanel.css'), 'utf8'),
].join('\n')

describe('AWDTrafficPanel extraction', () => {
  it('AWDTrafficPanel 应把 summary band、intelligence grid 和 event table 下沉到独立子组件', () => {
    expect(awdTrafficPanelSource).toContain('<AWDTrafficSummaryBand')
    expect(awdTrafficPanelSource).toContain('<AWDTrafficIntelligenceGrid')
    expect(awdTrafficPanelSource).toContain('<AWDTrafficEventTable')
    expect(awdTrafficPanelSource).toContain('useAwdTrafficPanel({')
    expect(awdTrafficPanelSource).not.toContain('class="studio-metric-band"')
    expect(awdTrafficPanelSource).not.toContain('class="intelligence-grid"')
    expect(awdTrafficPanelSource).not.toContain('class="drill-down-area"')
    expect(awdTrafficPanelSource).not.toContain('<style scoped>')
  })

  it('提取后的子组件应继续承接 trend、filter 和 pagination primitive', () => {
    expect(awdTrafficSummaryBandSource).toContain('class="metric-pill awd-traffic-summary-card"')
    expect(awdTrafficIntelligenceGridSource).toContain('class="trend-bar-track"')
    expect(awdTrafficEventTableSource).toContain('id="awd-traffic-filter-attacker"')
    expect(awdTrafficEventTableSource).toContain('id="awd-traffic-reset-filters"')
    expect(awdTrafficEventTableSource).toContain('PlatformPaginationControls')
    expect(awdTrafficCombinedSource).toContain('.status-badge')
  })
})
