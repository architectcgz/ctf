import { describe, expect, it } from 'vitest'

import contestChallengeOrchestrationPanelSource from '../contest/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest challenge orchestration extraction', () => {
  it('ContestChallengeOrchestrationPanel 应将汇总条和 AWD 筛选条下沉到独立子组件，而不是继续在父组件里内联整段结构', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain('<ContestChallengeSummaryStrip')
    expect(contestChallengeOrchestrationPanelSource).toContain('<ContestChallengeFilterStrip')
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"'
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      'class="contest-challenge-filter"'
    )
  })
})
