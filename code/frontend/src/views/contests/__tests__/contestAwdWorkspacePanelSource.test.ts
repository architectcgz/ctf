import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台应保留当前战情面板结构与运行态 service 标识', () => {
    expect(awdWorkspaceSource).toContain('DEFENSE MONITOR')
    expect(awdWorkspaceSource).toContain('ATTACK VECTOR')
    expect(awdWorkspaceSource).toContain('FIELD INTEL')
    expect(awdWorkspaceSource).toContain('RECENT FEEDBACK')
    expect(awdWorkspaceSource).toContain('formatServiceRef')
    expect(awdWorkspaceSource).toContain('id="awd-target-challenge"')
    expect(awdWorkspaceSource).toContain('id="awd-target-search"')
    expect(awdWorkspaceSource).toContain('data-testid="awd-feedback-challenge-title"')
  })
})
