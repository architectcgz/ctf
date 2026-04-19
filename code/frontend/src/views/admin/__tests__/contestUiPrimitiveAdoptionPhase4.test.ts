import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/admin/contest/AWDChallengeConfigPanel.vue?raw'
import awdOperationsPanelSource from '@/components/admin/contest/AWDOperationsPanel.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import awdTrafficPanelSource from '@/components/admin/contest/AWDTrafficPanel.vue?raw'
import contestAwdPreflightPanelSource from '@/components/admin/contest/ContestAwdPreflightPanel.vue?raw'

describe('contest ui primitive adoption phase 4', () => {
  it('awd operations panel should consume shared field and button primitives for selector and runtime shell', () => {
    expect(awdOperationsPanelSource).toContain('class="ui-field awd-ops-selector-field"')
    expect(awdOperationsPanelSource).toContain('class="ui-control-wrap"')
    expect(awdOperationsPanelSource).toContain('class="ui-control"')
    expect(awdOperationsPanelSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdOperationsPanelSource).toContain('class="ui-btn ui-btn--primary"')
  })

  it('awd round inspector should consume shared buttons and field primitives for toolbar and filters', () => {
    expect(awdRoundInspectorSource).toContain('class="ui-btn ui-btn--secondary awd-round-toolbar__button"')
    expect(awdRoundInspectorSource).toContain('class="ui-btn ui-btn--primary awd-round-toolbar__button"')
    expect(awdRoundInspectorSource).toContain('class="ui-field awd-round-filter-field"')
    expect(awdRoundInspectorSource).toContain('class="ui-control-wrap awd-round-filter-control"')
    expect(awdRoundInspectorSource).toContain('class="ui-control"')
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
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
  })
})
