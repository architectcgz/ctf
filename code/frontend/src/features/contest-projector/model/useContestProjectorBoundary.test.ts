import { describe, expect, it } from 'vitest'

import projectorDataSource from './useContestProjectorData.ts?raw'
import projectorPageSource from './useContestProjectorPage.ts?raw'
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

  it('projector feature model 不应再反向依赖 components/projector 下的 type 与 formatter 文件', () => {
    expect(projectorPageSource).not.toContain("from '@/components/platform/contest/projector/")
    expect(projectorDerivedSource).not.toContain("from '@/components/platform/contest/projector/")
  })
})
