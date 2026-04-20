import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'

describe('ContestDetail source', () => {
  it('学生竞赛详情不应再直接内嵌旧 AWD 工作台组件', () => {
    expect(contestDetailSource).not.toContain('ContestAWDWorkspacePanel')
  })
})
