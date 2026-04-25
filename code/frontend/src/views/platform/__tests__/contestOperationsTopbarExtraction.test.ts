import { describe, expect, it } from 'vitest'

import contestOperationsSource from '../ContestOperations.vue?raw'

describe('ContestOperations topbar extraction', () => {
  it('应将赛事运维页顶部壳层抽到独立 platform contest 组件', () => {
    expect(contestOperationsSource).toContain(
      "import ContestOperationsTopbarPanel from '@/components/platform/contest/ContestOperationsTopbarPanel.vue'"
    )
    expect(contestOperationsSource).toContain('<ContestOperationsTopbarPanel')
  })
})
