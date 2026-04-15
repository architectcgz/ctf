import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import adminContestFormPanelSource from '@/components/admin/contest/AdminContestFormPanel.vue?raw'
import adminContestTableSource from '@/components/admin/contest/AdminContestTable.vue?raw'

describe('contest ui primitive adoption', () => {
  it('contest workspace pages should consume shared ui button and control primitives', () => {
    expect(contestOperationsHubSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOperationsHubSource).toContain('class="ui-btn ui-btn--primary"')

    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--link contest-inline-link"')
    expect(contestOrchestrationSource).toContain('class="ui-field contest-filter-field"')
    expect(contestOrchestrationSource).toContain('class="ui-control-wrap"')
    expect(contestOrchestrationSource).toContain('class="ui-control"')
  })

  it('contest form should consume shared field and button primitives', () => {
    expect(adminContestFormPanelSource).toContain('class="ui-field contest-form-field')
    expect(adminContestFormPanelSource).toContain('class="ui-control-wrap')
    expect(adminContestFormPanelSource).toContain('class="ui-control"')
    expect(adminContestFormPanelSource).toContain('class="ui-btn ui-btn--secondary')
    expect(adminContestFormPanelSource).toContain('class="ui-btn ui-btn--primary')
  })

  it('contest directory rows should consume shared badge and row action primitives', () => {
    expect(adminContestTableSource).toContain('class="ui-badge contest-status-pill')
    expect(adminContestTableSource).toContain('class="ui-row-actions contest-row__actions"')
    expect(adminContestTableSource).toContain('class="ui-btn ui-btn--sm ui-btn--primary')
    expect(adminContestTableSource).toContain('class="ui-btn ui-btn--sm ui-btn--secondary')
    expect(adminContestTableSource).not.toContain('class="contest-action contest-action--primary"')
    expect(adminContestTableSource).not.toContain('class="contest-action contest-action--ghost"')
  })
})
