import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '../CheatDetection.vue?raw'
import cheatDetectionReviewPanelsSource from '@/components/platform/cheat/CheatDetectionReviewPanels.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'

describe('CheatDetection review panels extraction', () => {
  it('应将作弊检测风险目录区抽到独立 platform cheat 组件', () => {
    expect(cheatDetectionSource).toContain(
      "import CheatDetectionWorkspacePanel from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue'"
    )
    expect(cheatDetectionSource).toContain('<CheatDetectionWorkspacePanel')
    expect(cheatDetectionWorkspacePanelSource).toContain(
      "import CheatDetectionReviewPanels from '@/components/platform/cheat/CheatDetectionReviewPanels.vue'"
    )
    expect(cheatDetectionWorkspacePanelSource).toContain('<CheatDetectionReviewPanels')
    expect(cheatDetectionReviewPanelsSource).toContain('高频提交账号')
    expect(cheatDetectionReviewPanelsSource).toContain('共享 IP 线索')
    expect(cheatDetectionReviewPanelsSource).toContain('审计联动')
    expect(cheatDetectionReviewPanelsSource).toContain('class="cheat-directory-row"')
    expect(cheatDetectionReviewPanelsSource).toContain('class="quick-action-directory"')
  })
})
