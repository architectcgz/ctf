import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '../CheatDetection.vue?raw'

describe('CheatDetection workspace extraction', () => {
  it('应将作弊检测工作区壳层抽到独立 platform cheat 组件', () => {
    expect(cheatDetectionSource).toContain(
      "import CheatDetectionWorkspacePanel from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue'"
    )
    expect(cheatDetectionSource).toContain('<CheatDetectionWorkspacePanel')
  })
})
