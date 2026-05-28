import { describe, expect, it } from 'vitest'

import contestChallengeFilterStripSource from '@/features/contest-workbench/ui/ContestChallengeFilterStrip.vue?raw'
import contestChallengeDirectorySectionSource from '@/features/contest-workbench/ui/ContestChallengeDirectorySection.vue?raw'
import contestChallengeOrchestrationHeaderSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationHeader.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest challenge orchestration extraction', () => {
  it('ContestChallengeOrchestrationPanel 应将汇总条和 AWD 筛选条下沉到独立子组件，而不是继续在父组件里内联整段结构', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain('<ContestChallengeSummaryStrip')
    expect(contestChallengeOrchestrationPanelSource).toContain('<ContestChallengeFilterStrip')
    expect(contestChallengeOrchestrationPanelSource).toContain(
      '<ContestChallengeOrchestrationHeader'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      '<ContestChallengeDirectorySection'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain("from '../model'")
    expect(contestChallengeOrchestrationPanelSource).toContain('useContestChallengeOrchestration')
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      "from '@/api/admin/contests'"
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      "from '@/api/admin/authoring'"
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain("from '@/api/request'")
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"'
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain(
      'class="contest-challenge-filter"'
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain('class="studio-pane-header"')
    expect(contestChallengeOrchestrationPanelSource).not.toContain('class="studio-table-wrap')
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

  it('header 和 directory section 应承接父面板下沉的稳定展示结构', () => {
    expect(contestChallengeOrchestrationHeaderSource).toContain('class="studio-pane-header"')
    expect(contestChallengeDirectorySectionSource).toContain('class="studio-table-wrap')
    expect(contestChallengeDirectorySectionSource).toContain(
      'class="ui-row-actions contest-challenge-row__actions"'
    )
  })
})
