import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const topologyCanvasWorkspaceSectionSource = readFileSync(
  `${process.cwd()}/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue`,
  'utf-8'
)
const topologyNodeSectionSource = readFileSync(
  `${process.cwd()}/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue`,
  'utf-8'
)

describe('async chunk boundaries', () => {
  it('应当将拓扑页的画布与节点编辑器改为异步加载', () => {
    expect(topologyCanvasWorkspaceSectionSource).toContain('defineAsyncComponent')
    expect(topologyCanvasWorkspaceSectionSource).toContain("import('./TopologyCanvasBoard.vue')")
    expect(topologyNodeSectionSource).toContain('defineAsyncComponent')
    expect(topologyNodeSectionSource).toContain("import('./TopologyNodeEditor.vue')")
  })
})
