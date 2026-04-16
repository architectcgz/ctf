import { describe, expect, it } from 'vitest'

import contestChallengeOrchestrationPanelSource from '@/components/admin/contest/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest ui primitive adoption phase 24', () => {
  it('contest challenge orchestration summary should use the full shared metric panel class stack', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
