import { describe, expect, it } from 'vitest'

import contestChallengeSummaryStripSource from '@/components/platform/contest/ContestChallengeSummaryStrip.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/components/platform/contest/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest ui primitive adoption phase 24', () => {
  it('contest challenge orchestration summary should use the full shared metric panel class stack', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain('<ContestChallengeSummaryStrip')
    expect(contestChallengeSummaryStripSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"'
    )
    expect(contestChallengeSummaryStripSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(contestChallengeSummaryStripSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestChallengeSummaryStripSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestChallengeSummaryStripSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
