import { describe, expect, it } from 'vitest'

import projectorPageSource from './useContestProjectorPage.ts?raw'
import projectorDerivedSource from './useContestProjectorDerived.ts?raw'

describe('useContestProjector boundary', () => {
  it('projector feature model 不应再反向依赖 components/projector 下的 type 与 formatter 文件', () => {
    expect(projectorPageSource).not.toContain("from '@/components/platform/contest/projector/")
    expect(projectorDerivedSource).not.toContain("from '@/components/platform/contest/projector/")
  })
})
