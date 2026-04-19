import { describe, expect, it } from 'vitest'

import contestWorkbenchSummaryStripSource from '@/components/platform/contest/ContestWorkbenchSummaryStrip.vue?raw'

describe('contest ui primitive adoption phase 23', () => {
  it('contest workbench summary strip should use the full shared metric panel class stack', () => {
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-workbench-summary-strip"'
    )
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="journal-note progress-card metric-panel-card contest-workbench-summary-strip__item"'
    )
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
