import { describe, expect, it } from 'vitest'

import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import awdTrafficPanelSource from '@/components/admin/contest/AWDTrafficPanel.vue?raw'

describe('AWDRoundInspector extraction', () => {
  it('应将攻击流量态势区收口到独立的 AWDTrafficPanel，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdRoundInspectorSource).not.toContain('id="awd-traffic-filter-attacker"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-traffic-filter-victim"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-traffic-filter-status-group"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-traffic-reset-filters"')
    expect(awdRoundInspectorSource).not.toContain('共 ${trafficEventsTotal} 条流量事件')

    expect(awdTrafficPanelSource).toContain('id="awd-traffic-filter-attacker"')
    expect(awdTrafficPanelSource).toContain('id="awd-traffic-filter-victim"')
    expect(awdTrafficPanelSource).toContain('id="awd-traffic-filter-status-group"')
    expect(awdTrafficPanelSource).toContain('id="awd-traffic-reset-filters"')
    expect(awdTrafficPanelSource).toContain('AdminPaginationControls')
  })
})
