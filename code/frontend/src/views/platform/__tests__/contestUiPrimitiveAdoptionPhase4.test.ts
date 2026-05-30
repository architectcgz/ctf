import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/features/platform/contests/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contests/ui/AWDChallengeConfigHeader.vue?raw'
import awdContestSelectorFieldSource from '@/features/contest-awd-admin/ui/AWDContestSelectorField.vue?raw'
import awdInstanceOrchestrationHeaderSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationHeader.vue?raw'
import awdInstanceOrchestrationMatrixSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationMatrix.vue?raw'
import awdInstanceOrchestrationPanelSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue?raw'
import awdInstanceOrchestrationRowSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationRow.vue?raw'
import awdInspectorCanvasWorkspaceSource from '@/features/awd-inspector/ui/AWDInspectorCanvasWorkspace.vue?raw'
import awdInspectorStatsHudSource from '@/features/awd-inspector/ui/AWDInspectorStatsHud.vue?raw'
import awdOperationsPreRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue?raw'
import awdReadinessDecisionHUDSource from '@/features/awd-readiness/ui/AWDReadinessDecisionHUD.vue?raw'
import awdOperationsPanelSource from '@/features/contest-awd-admin/ui/AWDOperationsPanel.vue?raw'
import awdRoundHeaderPanelSource from '@/features/awd-inspector/ui/AWDRoundHeaderPanel.vue?raw'
import awdRoundInspectorSourceBase from '@/features/awd-inspector/ui/AWDRoundInspector.vue?raw'
import awdRuntimePendingStateSource from '@/features/contest-awd-admin/ui/AWDRuntimePendingState.vue?raw'
import awdTrafficEventTableSource from '@/features/awd-inspector/ui/AWDTrafficEventTable.vue?raw'
import awdTrafficIntelligenceGridSource from '@/features/awd-inspector/ui/AWDTrafficIntelligenceGrid.vue?raw'
import awdTrafficPanelSourceBase from '@/features/awd-inspector/ui/AWDTrafficPanel.vue?raw'
import awdTrafficSummaryBandSource from '@/features/awd-inspector/ui/AWDTrafficSummaryBand.vue?raw'
import contestAwdPreflightPanelSource from '@/features/platform/contests/ui/ContestAwdPreflightPanel.vue?raw'

const awdRoundInspectorSource = [
  awdRoundInspectorSourceBase,
  awdInspectorStatsHudSource,
  awdInspectorCanvasWorkspaceSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/awd-inspector/ui/awdRoundInspector.css'),
    'utf8'
  ),
].join('\n')
const awdChallengeConfigCombinedSource = [
  awdChallengeConfigPanelSource,
  awdChallengeConfigHeaderSource,
  awdChallengeConfigDirectorySectionSource,
  awdChallengeConfigDirectoryRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/platform/contests/ui/awdChallengeConfigPanel.css'),
    'utf8'
  ),
].join('\n')
const awdInstanceOrchestrationCombinedSource = [
  awdInstanceOrchestrationPanelSource,
  awdInstanceOrchestrationHeaderSource,
  awdInstanceOrchestrationMatrixSource,
  awdInstanceOrchestrationRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdInstanceOrchestrationPanel.css'),
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

describe('contest ui primitive adoption phase 4', () => {
  it('awd operations panel should consume shared field and button primitives for selector and runtime shell', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsPreRuntimeStage')
    expect(awdOperationsPreRuntimeStageSource).toContain('name="pending"')
    expect(awdContestSelectorFieldSource).toContain('class="ui-field awd-ops-selector-field"')
    expect(awdContestSelectorFieldSource).toContain('class="ui-control-wrap"')
    expect(awdContestSelectorFieldSource).toContain('class="ui-control"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--primary"')
  })

  it('awd round inspector should keep toolbar and filters in extracted panels', () => {
    expect(awdRoundInspectorSource).toContain('<AWDRoundHeaderPanel')
    expect(awdRoundInspectorSource).toContain('<AWDInspectorCanvasWorkspace')
    expect(awdRoundInspectorSource).not.toContain('<AWDRoundSelectionPanel')
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdRoundInspectorSource).toContain('class="ui-btn ui-btn--secondary awd-inspector-export-button"')
    expect(awdRoundHeaderPanelSource).toContain('class="round-select-native"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--neutral"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--primary"')
    expect(awdTrafficPanelSource).toContain('class="ui-field awd-round-filter-field"')
    expect(awdTrafficPanelSource).toContain('class="ui-control-wrap awd-round-filter-control"')
    expect(awdTrafficPanelSource).toContain('class="ui-control"')
    expect(awdTrafficPanelSource).toContain('class="ui-btn ui-btn--ghost awd-round-filter-search"')
    expect(awdTrafficPanelSource).toContain('class="metric-pill awd-traffic-summary-card"')
  })

  it('awd challenge config panel should consume shared action and row action primitives', () => {
    expect(awdChallengeConfigCombinedSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdChallengeConfigCombinedSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdChallengeConfigCombinedSource).toContain('class="ui-row-actions config-row__actions"')
  })

  it('awd instance orchestration panel should keep shared ops action primitives in extracted surface', () => {
    expect(awdInstanceOrchestrationCombinedSource).toContain('class="ops-btn ops-btn--neutral"')
    expect(awdInstanceOrchestrationCombinedSource).toContain('class="ops-btn ops-btn--primary"')
    expect(awdInstanceOrchestrationCombinedSource).toContain('class="cell-start-btn"')
    expect(awdInstanceOrchestrationCombinedSource).toContain('class="row-start-btn"')
    expect(awdInstanceOrchestrationCombinedSource).toContain('.instance-status--running')
  })

  it('contest awd preflight panel should consume shared primary button primitive', () => {
    expect(contestAwdPreflightPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestAwdPreflightPanelSource).toContain('background: transparent;')
    expect(contestAwdPreflightPanelSource).not.toContain('background: var(--color-bg-base);')
  })

  it('awd readiness decision hud should reuse shared metric panel primitives', () => {
    expect(awdReadinessDecisionHUDSource).toContain(
      'class="decision-hud progress-card metric-panel-card metric-panel-default-surface"'
    )
    expect(awdReadinessDecisionHUDSource).toContain('class="journal-note-label progress-card-label metric-panel-label"')
    expect(awdReadinessDecisionHUDSource).toContain('class="decision-title progress-card-value metric-panel-value"')
    expect(awdReadinessDecisionHUDSource).toContain('class="decision-description progress-card-hint metric-panel-helper"')
    expect(awdReadinessDecisionHUDSource).toContain('--metric-panel-padding: var(--space-2-5) var(--space-3);')
    expect(awdReadinessDecisionHUDSource).toContain('gap: var(--space-2);')
    expect(awdReadinessDecisionHUDSource).not.toContain('无阻塞')
  })
})
