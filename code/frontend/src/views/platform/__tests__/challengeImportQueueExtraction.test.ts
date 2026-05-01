import { describe, expect, it } from 'vitest'

import challengeImportManageSource from '../ChallengeImportManage.vue?raw'
import challengeImportQueuePanelSource from '@/components/platform/challenge/ChallengeImportQueuePanel.vue?raw'

describe('ChallengeImportManage queue extraction', () => {
  it('应将待确认导入工作区抽到独立 platform challenge 组件', () => {
    expect(challengeImportManageSource).toContain(
      "import ChallengeImportQueuePanel from '@/components/platform/challenge/ChallengeImportQueuePanel.vue'"
    )
    expect(challengeImportManageSource).toContain('<ChallengeImportQueuePanel')
    expect(challengeImportQueuePanelSource).toContain('Import Review')
    expect(challengeImportQueuePanelSource).toContain('待确认导入')
    expect(challengeImportQueuePanelSource).toContain('继续查看预览')
    expect(challengeImportQueuePanelSource).toContain(
      'class="ui-btn ui-btn--primary challenge-queue-action"'
    )
  })

  it('待确认导入面板应把分类与难度标签继续下沉到 challenge entity ui', () => {
    expect(challengeImportQueuePanelSource).toContain(
      "ChallengeCategoryDifficultyPills } from '@/entities/challenge'"
    )
    expect(challengeImportQueuePanelSource).toContain('<ChallengeCategoryDifficultyPills')
    expect(challengeImportQueuePanelSource).not.toContain(':get-category-label=')
    expect(challengeImportQueuePanelSource).not.toContain(':get-difficulty-label=')
  })
})
