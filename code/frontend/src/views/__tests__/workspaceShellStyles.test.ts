import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import teacherDashboardPageSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import dashboardViewSource from '@/views/dashboard/DashboardView.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'

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
    expect(workspaceShellStylesSource).toMatch(/--workspace-shell-radius:\s*0;/)
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
      expectNoLocalProperty(source, '.workspace-shell', '--workspace-shell-radius:\\s*0')
      expectNoLocalProperty(source, '.top-tabs', 'gap:\\s*28px')
      expectNoLocalProperty(source, '.content-pane', 'padding:\\s*28px')
      expectNoLocalProperty(source, '.content-pane', 'border-radius:\\s*0')
    }
  })

  it('非 top-tabs 工作区页面应使用共享 content 起始间距', () => {
    expect(workspaceShellStylesSource).toContain('.workspace-shell > .content-pane:first-child')
    expect(workspaceShellStylesSource).toContain('--workspace-content-start-padding-top')
    expect(challengeListSource).toContain('<main class="content-pane">')
    expect(challengeListSource).not.toMatch(/\.content-pane\s*\{[^}]*padding-top:/s)
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

  it('带顶部 tab 的页面不应在 tab 面板内重复渲染 eyebrow', () => {
    expect(scoreboardSource).not.toContain('Contest Scoreboard')
    expect(scoreboardSource).not.toContain('Contest Scoreboard Directory')
    expect(scoreboardSource).not.toContain('Points Scoreboard Directory')
    expect(scoreboardSource).not.toContain(
      'class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"'
    )

    expect(skillProfileSource).not.toContain(
      '<div class="skill-section-kicker">Radar Analysis</div>'
    )
    expect(skillProfileSource).not.toContain('<div class="skill-section-kicker">Weak Points</div>')
    expect(skillProfileSource).not.toContain(
      '<div class="skill-section-kicker">Recommendations</div>'
    )

    expect(classStudentsPageSource).not.toContain(
      '<div class="workspace-overline">Class Workspace</div>'
    )

    for (const [source, label] of [
      [studentOverviewSource, 'Training Journal'],
      [studentRecommendationSource, 'Action Queue'],
      [studentCategoryProgressSource, 'Action Ranking'],
      [studentDifficultySource, 'Intensity Workspace'],
      [studentTimelineSource, 'Timeline Log'],
    ] as const) {
      expect(source).not.toContain(label)
    }

    for (const label of [
      'Progress Signal',
      'Skill Portrait',
      'Student Insight',
      'Trend Review',
      'Review',
      'Intervention',
    ]) {
      expect(teacherDashboardPageSource).not.toContain(`>${label}<`)
    }
  })
})
