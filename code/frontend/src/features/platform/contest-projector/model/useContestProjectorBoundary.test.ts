import { describe, expect, it } from 'vitest'

import projectorDataSource from './useContestProjectorData.ts?raw'
import projectorDerivedSource from './useContestProjectorDerived.ts?raw'

describe('useContestProjector boundary', () => {
  it('data 组合器应下沉轮次选择与快照加载流程，避免回退内联实现', () => {
    expect(projectorDataSource).toContain(
      "import { useProjectorRoundSelection } from './useProjectorRoundSelection'"
    )
    expect(projectorDataSource).toContain(
      "import { useProjectorRoundSnapshotLoader } from './useProjectorRoundSnapshotLoader'"
    )
    expect(projectorDataSource).not.toContain('function chooseLiveRound(')
    expect(projectorDataSource).not.toContain('function chooseDisplayRound(')
    expect(projectorDataSource).not.toContain('async function loadRoundSnapshot(')
    expect(projectorDataSource).not.toContain('function clearRoundSnapshot(')
  })

  it('derived 组合器应复用 builders，避免回退到内联聚合实现', () => {
    expect(projectorDerivedSource).toContain("from './projectorDerivedBuilders'")
    expect(projectorDerivedSource).not.toContain('const serviceLabelMap = new Map<string, string>()')
    expect(projectorDerivedSource).not.toContain('const edgeMap = new Map<string, ContestProjectorAttackEdge>()')
  })
})
