import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/platform/contest/AWDReadinessOverrideDialog.vue?raw'
import awdReadinessSummarySource from '@/components/platform/contest/AWDReadinessSummary.vue?raw'

describe('contest ui primitive adoption phase 25', () => {
  it('awd challenge config panel should use the full shared metric panel class stack', () => {
    expect(awdChallengeConfigPanelSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(awdChallengeConfigPanelSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdChallengeConfigPanelSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdChallengeConfigPanelSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdChallengeConfigPanelSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })

  it('awd readiness summary should use the full shared metric panel class stack', () => {
    expect(awdReadinessSummarySource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-summary-grid"'
    )
    expect(awdReadinessSummarySource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdReadinessSummarySource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdReadinessSummarySource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdReadinessSummarySource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })

  it('awd readiness override dialog should use the full shared metric panel class stack', () => {
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-override-summary"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
