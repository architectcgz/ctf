import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import classManagementPageSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import teacherDashboardPageSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import studentManagementPageSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import teacherInstanceManagementPageSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import teacherAwdReviewWorkspaceHeaderSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspaceHeader.vue?raw'
import awdChallengeLibraryPageSource from '@/components/platform/awd-service/AWDChallengeLibraryPage.vue?raw'
import awdReviewHeroPanelSource from '@/components/platform/awd-review/AwdReviewHeroPanel.vue?raw'
import auditLogHeroPanelSource from '@/components/platform/audit/AuditLogHeroPanel.vue?raw'
import challengeImportHeroPanelSource from '@/components/platform/challenge/ChallengeImportHeroPanel.vue?raw'
import challengeManageHeroPanelSource from '@/components/platform/challenge/ChallengeManageHeroPanel.vue?raw'
import cheatDetectionHeroPanelSource from '@/components/platform/cheat/CheatDetectionHeroPanel.vue?raw'
import classManageHeroPanelSource from '@/components/platform/class/ClassManageHeroPanel.vue?raw'
import contestOrchestrationPageSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'
import dashboardViewSource from '@/views/dashboard/DashboardView.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationDetailSource from '@/views/notifications/NotificationDetail.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import challengeImportManageSource from '@/views/platform/ChallengeImportManage.vue?raw'
import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import imageManageSource from '@/views/platform/ImageManage.vue?raw'
import imageManageHeroPanelSource from '@/components/platform/images/ImageManageHeroPanel.vue?raw'
import instanceManageHeroPanelSource from '@/components/platform/instance/InstanceManageHeroPanel.vue?raw'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import scoreboardDetailSource from '@/views/scoreboard/ScoreboardDetail.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import studentManageHeroPanelSource from '@/components/platform/student/StudentManageHeroPanel.vue?raw'

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
    expect(workspaceShellStylesSource).toContain('.workspace-shell .content-pane')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .workspace-page-header')
    expect(workspaceShellStylesSource).not.toContain('.workspace-shell > .workspace-grid')
    expect(workspaceShellStylesSource).toContain('.workspace-shell .tab-panel.active')
    expect(workspaceShellStylesSource).toContain('@keyframes workspaceTabPanelIn')
  })

  it('首屏页面头部应使用共享 workspace-page-header 分隔线结构', () => {
    expect(workspaceShellStylesSource).toMatch(
      /\.workspace-shell \.workspace-page-header\s*\{[\s\S]*grid-template-columns:\s*var\(--workspace-page-header-columns,\s*minmax\(0,\s*1fr\)\s+auto\);[\s\S]*padding-bottom:\s*var\(--workspace-page-header-padding-bottom,\s*var\(--space-6\)\);[\s\S]*border-bottom:\s*1px solid/s
    )
    expect(workspaceShellStylesSource).toMatch(
      /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.workspace-shell \.workspace-page-header\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\);/s
    )
    expect(challengeListSource).toContain('class="workspace-page-header challenge-topbar"')

    for (const source of [
      adminDashboardSource,
      awdChallengeLibraryPageSource,
      awdReviewHeroPanelSource,
      auditLogHeroPanelSource,
      challengeImportHeroPanelSource,
      challengeManageHeroPanelSource,
      cheatDetectionHeroPanelSource,
      classManageHeroPanelSource,
      classManagementPageSource,
      contestListSource,
      contestOrchestrationPageSource,
      imageManageHeroPanelSource,
      instanceManageHeroPanelSource,
      instanceListSource,
      notificationDetailSource,
      notificationListSource,
      scoreboardDetailSource,
      securitySettingsSource,
      studentManagementPageSource,
      studentManageHeroPanelSource,
      teacherAwdReviewWorkspaceHeaderSource,
      teacherInstanceManagementPageSource,
      userGovernanceSource,
      userProfileSource,
    ]) {
      expect(source).toContain('workspace-page-header')
      expect(source).not.toContain('<section class="workspace-hero">')
      expect(source).not.toMatch(/\.workspace-hero\s*\{/)
    }
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

    for (const source of [
      challengeManageSource,
      challengeImportManageSource,
      awdChallengeLibraryPageSource,
    ]) {
      expect(source).toContain('content-pane')
      expect(source).not.toContain('<div class="workspace-grid">')
      expect(source).not.toMatch(/\.content-pane\s*\{[^}]*padding-top:/s)
    }

    expect(imageManageSource).toContain('<main class="content-pane">')
    expect(imageManageSource).not.toMatch(/\.content-pane\s*\{[^}]*padding-top:/s)
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
