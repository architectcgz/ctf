import { describe, expect, it } from 'vitest'

import awdAttackLogPanelSource from '@/components/admin/contest/AWDAttackLogPanel.vue?raw'
import awdRoundHeaderPanelSource from '@/components/admin/contest/AWDRoundHeaderPanel.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import awdRoundSelectionPanelSource from '@/components/admin/contest/AWDRoundSelectionPanel.vue?raw'
import awdScoreboardSummaryPanelSource from '@/components/admin/contest/AWDScoreboardSummaryPanel.vue?raw'
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

  it('应将流量摘要卡的响应式边框规则收口到语义类，而不是继续把任意选择器类写在模板里', () => {
    expect(awdTrafficPanelSource).toContain('awd-traffic-summary-card')
    expect(awdTrafficPanelSource).not.toContain('md:[&:nth-last-child(-n+2)]:border-b-0')
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

  it('应将排行榜与轮次汇总收口到独立的 AWDScoreboardSummaryPanel，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDScoreboardSummaryPanel')
    expect(awdRoundInspectorSource).not.toContain('排行榜已冻结')
    expect(awdRoundInspectorSource).not.toContain('item.solved_count')
    expect(awdRoundInspectorSource).not.toContain('item.unique_attackers_against')

    expect(awdScoreboardSummaryPanelSource).toContain('实时排行榜')
    expect(awdScoreboardSummaryPanelSource).toContain('排行榜已冻结')
    expect(awdScoreboardSummaryPanelSource).toContain('本轮汇总')
  })

  it('应将顶部操作头部和轮次切换区收口到独立组件，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDRoundHeaderPanel')
    expect(awdRoundInspectorSource).toContain('<AWDRoundSelectionPanel')
    expect(awdRoundInspectorSource).not.toContain('id="awd-round-selector"')
    expect(awdRoundInspectorSource).not.toContain('刷新 AWD 数据')
    expect(awdRoundInspectorSource).not.toContain('当前正在跟随 live 轮次')

    expect(awdRoundHeaderPanelSource).toContain('刷新 AWD 数据')
    expect(awdRoundHeaderPanelSource).toContain('当前正在跟随 live 轮次')
    expect(awdRoundSelectionPanelSource).toContain('id="awd-round-selector"')
    expect(awdRoundSelectionPanelSource).toContain('当前赛事还没有 AWD 轮次')
  })
})
