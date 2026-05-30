import { describe, expect, it } from 'vitest'

import {
  buildUserGovernancePanelQuery,
  resolveUserGovernancePanel,
} from './useUserGovernancePanelRoute'

describe('useUserGovernancePanelRoute', () => {
  it('应将缺省与旧的 directory query 归一到 overview', () => {
    expect(resolveUserGovernancePanel(undefined)).toBe('overview')
    expect(resolveUserGovernancePanel('directory')).toBe('overview')
  })

  it('应保留 import panel', () => {
    expect(resolveUserGovernancePanel('import')).toBe('import')
    expect(resolveUserGovernancePanel(['import'])).toBe('import')
  })

  it('切到 overview 时应移除 panel query', () => {
    expect(buildUserGovernancePanelQuery({ panel: 'import', keyword: 'alice' }, 'overview')).toEqual(
      { keyword: 'alice' }
    )
  })

  it('切到 import 时应写入 panel query', () => {
    expect(buildUserGovernancePanelQuery({ keyword: 'alice' }, 'import')).toEqual({
      keyword: 'alice',
      panel: 'import',
    })
  })
})
