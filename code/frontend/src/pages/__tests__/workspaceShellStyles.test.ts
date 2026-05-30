import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import adminDashboardSourceBase from '@/features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'
import studentCategoryProgressSource from '@/features/student-dashboard/ui/StudentCategoryProgressPage.vue?raw'
import studentDifficultySource from '@/features/student-dashboard/ui/StudentDifficultyPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/features/student-dashboard/ui/StudentRecommendationPage.vue?raw'
import trainingTimelineSource from '@/components/training/TrainingTimelinePanel.vue?raw'
import classManagementPageSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import classStudentsPageSourceBase from '@/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue?raw'
import studentAnalysisPageSourceBase from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import studentAnalysisOverviewHeroPanelSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue?raw'
import teacherDashboardPageSourceBase from '@/features/teacher/dashboard/ui/TeacherDashboardPage.vue?raw'
import teacherDashboardPortraitPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue?raw'
import teacherDashboardStudentInsightPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue?raw'
import teacherDashboardTrendPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue?raw'
import teacherDashboardReviewPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue?raw'
import teacherDashboardInterventionPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue?raw'
import studentManagementPageSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import teacherInstanceManagementPageSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import teacherAwdReviewWorkspaceHeaderSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceHeader.vue?raw'
import awdChallengeImportSectionSource from '@/features/platform/awd-challenges/ui/AwdChallengeImportSection.vue?raw'
import awdChallengeLibraryPageSourceBase from '@/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue?raw'
import awdChallengeLibrarySectionSource from '@/components/platform/awd-service/AwdChallengeLibrarySection.vue?raw'
import awdChallengeWorkspaceHeaderSource from '@/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue?raw'
import awdReviewHeroPanelSource from '@/widgets/awd-review-workspace/AwdReviewHeroPanel.vue?raw'
import auditLogHeroPanelSource from '@/features/audit-log/ui/AuditLogHeroPanel.vue?raw'
import challengeImportHeroPanelSource from '@/components/platform/challenge/ChallengeImportHeroPanel.vue?raw'
import challengeManageHeroPanelSource from '@/features/platform/challenges/ui/ChallengeManageHeroPanel.vue?raw'
import cheatDetectionHeroPanelSource from '@/components/platform/cheat/CheatDetectionHeroPanel.vue?raw'
import classManageHeroPanelSource from '@/features/platform/class-management/ui/ClassManageHeroPanel.vue?raw'
import contestOrchestrationPageSource from '@/features/platform/contests/ui/ContestOrchestrationPage.vue?raw'
import instanceListWorkspaceShellSource from '@/components/instance/InstanceListWorkspaceShell.vue?raw'
import dashboardViewSource from '@/pages/dashboard/DashboardRoutePage.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import notificationDetailSource from '@/pages/notifications/NotificationDetailRoutePage.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import challengeImportManageSource from '@/pages/platform/challenges/ChallengeImportManageRoutePage.vue?raw'
import challengePackageFormatSource from '@/pages/platform/challenges/ChallengePackageFormatRoutePage.vue?raw'
import challengeManageSource from '@/features/platform/challenges/ui/ChallengeManagePage.vue?raw'
import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'
import imageManageHeroPanelSource from '@/features/image-management/ui/ImageManageHeroPanel.vue?raw'
import instanceManageHeroPanelSource from '@/features/platform/instance-management/ui/InstanceManageHeroPanel.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/components/profile/SkillProfileWorkspaceShell.vue?raw'

const classStudentsPageSource = [
  classStudentsPageSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')

const teacherDashboardPageSource = [
  teacherDashboardPageSourceBase,
  teacherDashboardPortraitPanelSource,
  teacherDashboardStudentInsightPanelSource,
  teacherDashboardTrendPanelSource,
  teacherDashboardReviewPanelSource,
  teacherDashboardInterventionPanelSource,
].join('\n')

const adminDashboardSource = [
  adminDashboardSourceBase,
  platformOverviewHeroPanelSource,
  platformOverviewAlertsSectionSource,
  platformOverviewHotspotsSectionSource,
].join('\n')

const teacherInstanceManagementPageSource = [
  teacherInstanceManagementPageSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

const studentAnalysisPageSource = [
  studentAnalysisPageSourceBase,
  studentAnalysisOverviewHeroPanelSource,
].join('\n')

const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')
const awdChallengeLibraryPageSource = [
  awdChallengeLibraryPageSourceBase,
  awdChallengeWorkspaceHeaderSource,
  awdChallengeLibrarySectionSource,
  awdChallengeImportSectionSource,
].join('\n')
import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/components/profile/SecuritySettingsWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/components/profile/UserProfileWorkspaceShell.vue?raw'
import scoreboardDetailSource from '@/pages/scoreboard/ScoreboardDetailRoutePage.vue?raw'
import scoreboardSource from '@/pages/scoreboard/ScoreboardViewRoutePage.vue?raw'
import studentManageHeroPanelSource from '@/features/platform/student-management/ui/StudentManageHeroPanel.vue?raw'
import contestOperationsHubHeroPanelSource from '@/features/platform/contests/ui/ContestOperationsHubHeroPanel.vue?raw'

const workspaceShellStylesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`
const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`

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
      imageManageHeroPanelSource,
      instanceManageHeroPanelSource,
      instanceListWorkspaceSource,
      notificationDetailSource,
      notificationListSource,
      scoreboardDetailSource,
      securitySettingsWorkspaceSource,
      studentManagementPageSource,
      studentManageHeroPanelSource,
      teacherAwdReviewWorkspaceHeaderSource,
      teacherInstanceManagementPageSource,
      userProfileWorkspaceSource,
    ]) {
      expect(source).toContain('workspace-page-header')
      expect(source).not.toContain('<section class="workspace-hero">')
      expect(source).not.toMatch(/\.workspace-hero\s*\{/)
    }
  })

  it('admin 与 workspace 首屏页应把 hero 背景放在 section root 或 workspace shell root 上，而不是退回 div 包裹壳层', () => {
    const sources = [
      adminDashboardSource,
      contestOrchestrationPageSource,
      userGovernanceSource,
      challengeDetailSource,
      challengePackageFormatSource,
      imageManageSource,
      challengeManageSource,
      auditLogSource,
      challengeImportManageSource,
      awdChallengeLibraryPageSource,
      classManagementPageSource,
      teacherInstanceManagementPageSource,
      studentManagementPageSource,
      instanceListWorkspaceSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      const hasSectionHeroRoot =
        /<section[\s\S]*?class="[^"]*workspace-shell[^"]*journal-shell[^"]*journal-hero[^"]*"/s.test(
          source
        ) ||
        /<section[\s\S]*?class="[^"]*journal-shell[^"]*journal-hero[^"]*workspace-shell[^"]*"/s.test(
          source
        )
      const hasWorkspaceShellRoot = /<div[\s\S]*?class="[^"]*workspace-shell[^"]*"/s.test(source)

      expect(hasSectionHeroRoot || hasWorkspaceShellRoot).toBe(true)
    }
  })

  it('overview 工作区面板头部应改用共享 workspace-panel-header，而不是继续占用页级页头', () => {
    expect(userGovernanceSource).toContain('<header class="workspace-panel-header user-overview-head">')
    expect(userGovernanceSource).not.toContain(
      '<header class="workspace-page-header user-overview-head">'
    )
    expect(contestOrchestrationPageSource).toContain(
      '<header class="workspace-panel-header contest-overview-head">'
    )
    expect(contestOrchestrationPageSource).not.toContain(
      '<header class="workspace-page-header contest-overview-head">'
    )
    expect(teacherDashboardPageSource).toContain(
      '<header class="workspace-panel-header teacher-dashboard-overview-head">'
    )
    expect(studentAnalysisPageSource).toContain(
      '<header class="workspace-panel-header student-analysis-overview-head">'
    )
    expect(contestOperationsHubHeroPanelSource).toContain(
      '<header class="workspace-panel-header contest-ops-hero">'
    )
    expect(contestOperationsHubHeroPanelSource).not.toContain(
      '<header class="workspace-page-header contest-ops-hero">'
    )
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

  it('top-tabs 工作区页面也不应在局部重复声明 content-pane 顶部间距', () => {
    expect(studentAnalysisPageSource).toContain('<main class="content-pane">')
    expect(studentAnalysisPageSource).not.toMatch(/\.content-pane\s*\{[^}]*padding-top:/s)
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
    expect(scoreboardSource).not.toContain(
      'class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"'
    )

    expect(skillProfileWorkspaceSource).not.toContain(
      '<div class="skill-section-kicker">Radar Analysis</div>'
    )
    expect(skillProfileWorkspaceSource).not.toContain('<div class="skill-section-kicker">Weak Points</div>')
    expect(skillProfileWorkspaceSource).not.toContain(
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
      [trainingTimelineSource, 'Timeline Log'],
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
