import { describe, expect, it } from 'vitest'

import contestChallengeFilterStripSource from '../contest/ContestChallengeFilterStrip.vue?raw'
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

  it('ContestChallengeFilterStrip 应复用 AWD 配置页同一套 metric panel 卡片结构', () => {
    expect(contestChallengeFilterStripSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(contestChallengeFilterStripSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(contestChallengeFilterStripSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestChallengeFilterStripSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestChallengeFilterStripSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(contestChallengeFilterStripSource).not.toContain(
      'contest-challenge-filter-card--active'
    )
  })
})
