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

  it('独立防守页应把目录状态交给工作台组件统一渲染', () => {
    expect(contestAwdDefenseWorkbenchSource).toContain(':editable-paths="editablePaths"')
    expect(contestAwdDefenseWorkbenchSource).toContain(':current-directory-path="currentDirectoryPath"')
    expect(contestAwdDefenseWorkbenchSource).toContain(':save-error="saveError"')
    expect(contestAwdDefenseWorkbenchSource).toContain(':saving="saveLoading"')
    expect(contestAwdDefenseWorkbenchSource).toContain('@save-file="saveFile"')
    expect(contestAwdDefenseWorkbenchSource).toContain(
      'defense-hero__back ui-btn ui-btn--sm ui-btn--ghost'
    )
    expect(contestAwdDefenseWorkbenchSource).toContain('ui-btn ui-btn--sm ui-btn--secondary')
    expect(contestAwdDefenseWorkbenchSource).toContain('<header class="defense-hero">')
    expect(contestAwdDefenseWorkbenchSource).toContain('border-bottom: 1px solid')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('metric-panel-default-surface')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('metric-panel-workspace-surface')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('defense-back-link')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('@refresh="refreshDirectory"')
    expect(contestAwdDefenseWorkbenchSource).not.toContain('radial-gradient(')
  })

  it('学生路由应暴露独立防守内容页入口', () => {
    expect(studentRoutesSource).toContain("name: 'ContestAWDDefenseWorkbench'")
    expect(studentRoutesSource).toContain("path: 'contests/:id/awd/defense/:serviceId'")
    expect(studentRoutesSource).toContain("title: '防守内容'")
  })
})
