import { describe, expect, it } from 'vitest'

import contestOperationsSource from '../ContestOperations.vue?raw'

describe('ContestOperations workspace shell', () => {
  it('赛事运维页不应保留额外顶部返回壳层', () => {
    expect(contestOperationsSource).not.toContain('ContestOperationsTopbarPanel')
    expect(contestOperationsSource).not.toContain('class="ops-topbar"')
    expect(contestOperationsSource).toContain('workspace-shell journal-shell journal-shell-admin')
    expect(contestOperationsSource).not.toContain('height: 100vh')
    expect(contestOperationsSource).not.toContain('overflow: hidden')
  })
})
