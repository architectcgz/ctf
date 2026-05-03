import { describe, expect, it } from 'vitest'

import topologyStudioPageSource from './useChallengeTopologyStudioPage.ts?raw'

describe('useChallengeTopologyStudioPage boundary', () => {
  it('topology studio feature model 不应反向依赖 components/topology 内部模块', () => {
    expect(topologyStudioPageSource).not.toContain(
      "from '@/components/platform/topology/topologyLayout'"
    )
    expect(topologyStudioPageSource).not.toContain(
      "from '@/components/platform/topology/topologyDraft'"
    )
    expect(topologyStudioPageSource).not.toContain(
      "from '@/components/platform/topology/TopologyCanvasBoard.vue'"
    )
  })
})
