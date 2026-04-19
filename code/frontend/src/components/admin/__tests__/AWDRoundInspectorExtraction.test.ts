import { describe, expect, it } from 'vitest'

import awdAttackLogPanelSource from '@/components/admin/contest/AWDAttackLogPanel.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import awdServiceStatusPanelSource from '@/components/admin/contest/AWDServiceStatusPanel.vue?raw'
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

  it('应将服务状态表收口到独立的 AWDServiceStatusPanel，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDServiceStatusPanel')
    expect(awdRoundInspectorSource).not.toContain('id="awd-service-filter-team"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-service-filter-status"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-service-filter-source"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-service-filter-alert"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-export-services"')

    expect(awdServiceStatusPanelSource).toContain('id="awd-service-filter-team"')
    expect(awdServiceStatusPanelSource).toContain('id="awd-service-filter-status"')
    expect(awdServiceStatusPanelSource).toContain('id="awd-service-filter-source"')
    expect(awdServiceStatusPanelSource).toContain('id="awd-service-filter-alert"')
    expect(awdServiceStatusPanelSource).toContain('id="awd-export-services"')
  })

  it('应将攻击日志表收口到独立的 AWDAttackLogPanel，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDAttackLogPanel')
    expect(awdRoundInspectorSource).not.toContain('id="awd-attack-filter-team"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-attack-filter-result"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-attack-filter-source"')
    expect(awdRoundInspectorSource).not.toContain('id="awd-export-attacks"')

    expect(awdAttackLogPanelSource).toContain('id="awd-attack-filter-team"')
    expect(awdAttackLogPanelSource).toContain('id="awd-attack-filter-result"')
    expect(awdAttackLogPanelSource).toContain('id="awd-attack-filter-source"')
    expect(awdAttackLogPanelSource).toContain('id="awd-export-attacks"')
  })
})
