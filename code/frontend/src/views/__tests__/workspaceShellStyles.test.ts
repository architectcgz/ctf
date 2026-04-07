import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/admin/dashboard/AdminDashboardPage.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import teacherDashboardPageSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import dashboardViewSource from '@/views/dashboard/DashboardView.vue?raw'

const workspaceShellStylesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function expectNoLocalProperty(source: string, selector: string, propertyPattern: string): void {
  const escapedSelector = escapeRegExp(selector)
  expect(source).not.toMatch(new RegExp(`${escapedSelector}\\s*\\{[^}]*${propertyPattern}`, 's'))
}

describe('workspace shell shared styles', () => {
  it('应该在共享样式文件中声明 workspace shell 骨架样式', () => {
    expect(workspaceShellStylesSource).toContain('.workspace-shell')
    expect(workspaceShellStylesSource).toContain('.workspace-shell > .workspace-topbar')
    expect(workspaceShellStylesSource).toContain('.workspace-shell > .top-tabs')
    expect(workspaceShellStylesSource).toContain('.workspace-shell > .workspace-grid')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .content-pane')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .tab-panel.active')
    expect(workspaceShellStylesSource).toContain('@keyframes workspaceTabPanelIn')
  })

  it('工作区页面不应继续在局部重复声明骨架壳层样式', () => {
    for (const source of [
      dashboardViewSource,
      adminDashboardSource,
      teacherDashboardPageSource,
      classStudentsPageSource,
      studentAnalysisPageSource,
    ]) {
      expectNoLocalProperty(source, '.workspace-shell', 'border:\\s*1px solid')
      expectNoLocalProperty(
        source,
        '.workspace-shell',
        'box-shadow:\\s*var\\(--workspace-shadow-shell\\)'
      )
      expectNoLocalProperty(source, '.top-tabs', 'gap:\\s*28px')
      expectNoLocalProperty(source, '.content-pane', 'padding:\\s*28px')
    }
  })

  it('工作区页面不应继续在局部重复声明 tab 面板切换动画', () => {
    for (const source of [
      dashboardViewSource,
      adminDashboardSource,
      teacherDashboardPageSource,
      classStudentsPageSource,
      studentAnalysisPageSource,
    ]) {
      expect(source).not.toContain('@keyframes tabPanelIn')
      expect(source).not.toContain('animation: tabPanelIn 180ms ease both;')
    }
  })
})
