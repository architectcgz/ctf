import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import adminDashboardSourceBase from '@/features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'
import studentOverviewSource from '@/features/student-dashboard/ui/StudentOverviewContent.vue?raw'
import studentRecommendationSource from '@/features/student-dashboard/ui/StudentRecommendationContent.vue?raw'
import studentCategoryProgressSource from '@/features/student-dashboard/ui/StudentCategoryProgressContent.vue?raw'
import studentDifficultySource from '@/features/student-dashboard/ui/StudentDifficultyContent.vue?raw'
import classManagementPageSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import studentAnalysisPageSourceBase from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import studentAnalysisOverviewHeroPanelSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue?raw'
import teacherDashboardPageSourceBase from '@/features/teacher/dashboard/ui/TeacherDashboardPage.vue?raw'
import teacherDashboardPortraitPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue?raw'
import teacherDashboardStudentInsightPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue?raw'
import teacherDashboardTrendPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue?raw'
import teacherDashboardReviewPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue?raw'
import teacherDashboardInterventionPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue?raw'
import contestManageOverviewPanelSource from '@/features/platform/contest-manage/ui/ContestManageOverviewPanel.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import contestListWorkspaceSource from '@/widgets/contest-list-workspace/ContestListWorkspace.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import scoreboardSource from '@/pages/scoreboard/ScoreboardViewRoutePage.vue?raw'

const workspaceShellStylesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)

const adminDashboardSource = [
  adminDashboardSourceBase,
  platformOverviewHeroPanelSource,
  platformOverviewAlertsSectionSource,
  platformOverviewHotspotsSectionSource,
].join('\n')

const teacherDashboardSource = [
  teacherDashboardPageSourceBase,
  teacherDashboardPortraitPanelSource,
  teacherDashboardStudentInsightPanelSource,
  teacherDashboardTrendPanelSource,
  teacherDashboardReviewPanelSource,
  teacherDashboardInterventionPanelSource,
].join('\n')

const studentAnalysisPageSource = [
  studentAnalysisPageSourceBase,
  studentAnalysisOverviewHeroPanelSource,
].join('\n')

const contestListWorkspaceCombinedSource = `${contestListSource}\n${contestListWorkspaceSource}`

describe('workspace shell shared styles', () => {
  it('应该在共享样式文件中声明 workspace shell、顶部 tabs 和 page header 骨架', () => {
    expect(workspaceShellStylesSource).toContain('.workspace-shell')
    expect(workspaceShellStylesSource).toContain('.workspace-shell > .top-tabs')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .content-pane')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .workspace-page-header')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .tab-panel.active')
    expect(workspaceShellStylesSource).toContain('@keyframes workspaceTabPanelIn')
  })

  it('典型页面应继续复用共享 workspace-page-header，而不是退回本地 hero 壳', () => {
    for (const source of [
      adminDashboardSource,
      classManagementPageSource,
      contestListWorkspaceCombinedSource,
    ]) {
      expect(source).toContain('workspace-page-header')
      expect(source).not.toContain('<section class="workspace-hero">')
      expect(source).not.toMatch(/\.workspace-hero\s*\{/)
    }
  })

  it('overview 面板应继续复用 workspace-panel-header，而不是继续占用页级页头', () => {
    expect(contestManageOverviewPanelSource).toContain(
      '<header class="workspace-panel-header contest-overview-head">'
    )
    expect(contestManageOverviewPanelSource).not.toContain(
      '<header class="workspace-page-header contest-overview-head">'
    )
    expect(teacherDashboardSource).toContain(
      '<header class="workspace-panel-header teacher-dashboard-overview-head">'
    )
    expect(studentAnalysisPageSource).toContain(
      '<header class="workspace-panel-header student-analysis-overview-head">'
    )
  })

  it('工作区页面不应继续在局部重复声明共享 shell 基础样式与 tab 动画', () => {
    for (const source of [
      adminDashboardSource,
      teacherDashboardSource,
      studentAnalysisPageSource,
    ]) {
      expect(source).not.toMatch(/\.workspace-shell\s*\{[^}]*border:\s*1px solid/s)
      expect(source).not.toMatch(/\.content-pane\s*\{[^}]*padding:\s*28px/s)
      expect(source).not.toContain('@keyframes tabPanelIn')
      expect(source).not.toContain('animation: tabPanelIn 180ms ease both;')
    }
  })

  it('带顶部 tab 的页面不应在 tab 面板里重复渲染说明型 eyebrow', () => {
    expect(scoreboardSource).not.toContain(
      'class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"'
    )

    for (const label of [
      'Training Journal',
      'Action Queue',
      'Action Ranking',
      'Intensity Workspace',
    ]) {
      expect(studentOverviewSource).not.toContain(label)
      expect(studentRecommendationSource).not.toContain(label)
      expect(studentCategoryProgressSource).not.toContain(label)
      expect(studentDifficultySource).not.toContain(label)
    }
  })
})
