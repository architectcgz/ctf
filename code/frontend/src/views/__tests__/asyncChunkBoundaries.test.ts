import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const topologyStudioPageSource = readFileSync(
  `${process.cwd()}/src/components/platform/topology/ChallengeTopologyStudioPage.vue`,
  'utf-8'
)

describe('async chunk boundaries', () => {
  it('应当将拓扑页的画布与节点编辑器改为异步加载', () => {
    expect(topologyStudioPageSource).toContain('defineAsyncComponent')
    expect(topologyStudioPageSource).toContain("import('./TopologyCanvasBoard.vue')")
    expect(topologyStudioPageSource).toContain("import('./TopologyNodeEditor.vue')")
  })
})
