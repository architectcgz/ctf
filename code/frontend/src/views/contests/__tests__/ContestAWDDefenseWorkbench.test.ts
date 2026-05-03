import { describe, expect, it } from 'vitest'

import contestAwdDefenseWorkbenchSource from '../ContestAWDDefenseWorkbench.vue?raw'

describe('ContestAWDDefenseWorkbench', () => {
  it('路由壳页应仅做组合，不直接读取 route 参数', () => {
    expect(contestAwdDefenseWorkbenchSource).toContain('useContestAwdDefenseWorkbenchPage')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('useRoute')
  })
})
