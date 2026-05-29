import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdAttackLogPanelSource from '@/features/awd-inspector/ui/AWDAttackLogPanel.vue?raw'
import awdInspectorCanvasWorkspaceSource from '@/features/awd-inspector/ui/AWDInspectorCanvasWorkspace.vue?raw'
import awdInspectorStatsHudSource from '@/features/awd-inspector/ui/AWDInspectorStatsHud.vue?raw'
import awdRoundHeaderPanelSource from '@/features/awd-inspector/ui/AWDRoundHeaderPanel.vue?raw'
import awdRoundInspectorSourceBase from '@/features/awd-inspector/ui/AWDRoundInspector.vue?raw'
import awdScoreboardSummaryPanelSource from '@/features/awd-inspector/ui/AWDScoreboardSummaryPanel.vue?raw'
import awdServiceRoundPerformanceTableSource from '@/features/awd-inspector/ui/AWDServiceRoundPerformanceTable.vue?raw'
import awdServiceStatusMatrixSource from '@/features/awd-inspector/ui/AWDServiceStatusMatrix.vue?raw'
import awdServiceStatusPanelSourceBase from '@/features/awd-inspector/ui/AWDServiceStatusPanel.vue?raw'
import awdServiceStatusToolbarSource from '@/features/awd-inspector/ui/AWDServiceStatusToolbar.vue?raw'
import awdTrafficEventTableSource from '@/features/awd-inspector/ui/AWDTrafficEventTable.vue?raw'
import awdTrafficIntelligenceGridSource from '@/features/awd-inspector/ui/AWDTrafficIntelligenceGrid.vue?raw'
import awdTrafficPanelSourceBase from '@/features/awd-inspector/ui/AWDTrafficPanel.vue?raw'
import awdTrafficSummaryBandSource from '@/features/awd-inspector/ui/AWDTrafficSummaryBand.vue?raw'

const awdRoundInspectorSource = [
  awdRoundInspectorSourceBase,
  awdInspectorStatsHudSource,
  awdInspectorCanvasWorkspaceSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/awd-inspector/ui/awdRoundInspector.css'),
    'utf8'
  ),
].join('\n')
const awdServiceStatusPanelSource = [
  awdServiceStatusPanelSourceBase,
  awdServiceStatusToolbarSource,
  awdServiceStatusMatrixSource,
  awdServiceRoundPerformanceTableSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/awd-inspector/ui/awdServiceStatusPanel.css'),
    'utf8'
  ),
].join('\n')
const awdTrafficPanelSource = [
  awdTrafficPanelSourceBase,
  awdTrafficSummaryBandSource,
  awdTrafficIntelligenceGridSource,
  awdTrafficEventTableSource,
  readFileSync(resolve(process.cwd(), 'src/features/awd-inspector/ui/awdTrafficPanel.css'), 'utf8'),
].join('\n')

describe('AWDRoundInspector extraction', () => {
  it('应将态势 HUD 与 tabbed canvas shell 收口到独立子组件，而不是继续堆在 AWDRoundInspector 内', () => {
    expect(awdRoundInspectorSource).toContain('<AWDInspectorStatsHud')
    expect(awdRoundInspectorSource).toContain('<AWDInspectorCanvasWorkspace')
    expect(awdRoundInspectorSourceBase).not.toContain('<div class="awd-stats-hud">')
    expect(awdRoundInspectorSourceBase).not.toContain('<header class="canvas-tabs-header">')
    expect(awdInspectorStatsHudSource).toContain('class="awd-stats-hud"')
    expect(awdInspectorCanvasWorkspaceSource).toContain('class="canvas-tabs-header"')
  })

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
    expect(awdTrafficPanelSource).toContain('PlatformPaginationControls')
    expect(awdTrafficPanelSource).toContain('class="studio-metric-band"')
    expect(awdTrafficPanelSource).toContain('class="intelligence-grid"')
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
    expect(awdServiceStatusPanelSource).toContain('Round Performance Summary')
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
    expect(awdRoundInspectorSource).not.toContain('<AWDRoundSelectionPanel')
    expect(awdRoundInspectorSource).not.toContain('id="awd-round-selector"')
    expect(awdRoundInspectorSource).not.toContain('title="刷新数据"')
    expect(awdRoundInspectorSource).not.toContain('当前正在跟随 live 轮次')

    expect(awdRoundHeaderPanelSource).toContain('class="round-switcher"')
    expect(awdRoundHeaderPanelSource).toContain('class="round-select-native"')
    expect(awdRoundHeaderPanelSource).toContain('title="刷新数据"')
    expect(awdRoundHeaderPanelSource).toContain('当前正在跟随 live 轮次')
    expect(awdRoundHeaderPanelSource).not.toContain('{{ contest.title }}')
  })
})
