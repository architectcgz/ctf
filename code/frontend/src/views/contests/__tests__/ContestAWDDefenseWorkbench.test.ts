import { describe, expect, it } from 'vitest'

import contestAwdDefenseWorkbenchSource from '../ContestAWDDefenseWorkbench.vue?raw'
import studentRoutesSource from '@/router/routes/studentRoutes.ts?raw'

describe('ContestAWDDefenseWorkbench', () => {
  it('页面应通过 feature page owner 组合，不直接在视图层读取 route 或发 API 请求', () => {
    expect(contestAwdDefenseWorkbenchSource).toContain('useContestAwdDefenseWorkbenchPage')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('useRoute')
    expect(contestAwdDefenseWorkbenchSource).not.toContain("from '@/api/contest'")
    expect(contestAwdDefenseWorkbenchSource).not.toContain('防守入口已迁移')
  })

  it('学生路由应暴露独立防守内容页入口', () => {
    expect(studentRoutesSource).toContain("name: 'ContestAWDDefenseWorkbench'")
    expect(studentRoutesSource).toContain("path: 'contests/:id/awd/defense/:serviceId'")
    expect(studentRoutesSource).toContain("title: '防守内容'")
  })
})
