import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'
import awdContestSelectorFieldSource from '@/components/platform/contest/AWDContestSelectorField.vue?raw'
import awdReadinessDecisionHUDSource from '@/components/platform/contest/AWDReadinessDecisionHUD.vue?raw'
import awdOperationsPanelSource from '@/components/platform/contest/AWDOperationsPanel.vue?raw'
import awdRoundHeaderPanelSource from '@/components/platform/contest/AWDRoundHeaderPanel.vue?raw'
import awdRoundInspectorSource from '@/components/platform/contest/AWDRoundInspector.vue?raw'
import awdRuntimePendingStateSource from '@/components/platform/contest/AWDRuntimePendingState.vue?raw'
import awdTrafficPanelSource from '@/components/platform/contest/AWDTrafficPanel.vue?raw'
import contestAwdPreflightPanelSource from '@/components/platform/contest/ContestAwdPreflightPanel.vue?raw'

describe('contest ui primitive adoption phase 4', () => {
  it('awd operations panel should consume shared field and button primitives for selector and runtime shell', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDRuntimePendingState')
    expect(awdContestSelectorFieldSource).toContain('class="ui-field awd-ops-selector-field"')
    expect(awdContestSelectorFieldSource).toContain('class="ui-control-wrap"')
    expect(awdContestSelectorFieldSource).toContain('class="ui-control"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--primary"')
  })

  it('awd round inspector should keep toolbar and filters in extracted panels', () => {
    expect(awdRoundInspectorSource).toContain('<AWDRoundHeaderPanel')
    expect(awdRoundInspectorSource).not.toContain('<AWDRoundSelectionPanel')
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdRoundHeaderPanelSource).toContain('class="round-select-native"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--neutral"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--primary"')
    expect(awdTrafficPanelSource).toContain('class="ui-field awd-round-filter-field"')
    expect(awdTrafficPanelSource).toContain('class="ui-control-wrap awd-round-filter-control"')
    expect(awdTrafficPanelSource).toContain('class="ui-control"')
    expect(awdTrafficPanelSource).toContain('class="ui-btn ui-btn--ghost awd-round-filter-search"')
  })

  it('awd challenge config panel should consume shared action and row action primitives', () => {
    expect(awdChallengeConfigPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdChallengeConfigPanelSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdChallengeConfigPanelSource).toContain('class="ui-row-actions config-row__actions"')
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
