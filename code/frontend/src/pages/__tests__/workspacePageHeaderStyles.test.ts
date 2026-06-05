import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import contestListWorkspaceSource from '@/widgets/contest-list-workspace/ContestListWorkspace.vue?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import instanceListWorkspaceShellSource from '@/features/instance-list/ui/InstanceListWorkspaceShell.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import studentOverviewSource from '@/features/student-dashboard/ui/StudentOverviewContent.vue?raw'
import studentRecommendationSource from '@/features/student-dashboard/ui/StudentRecommendationContent.vue?raw'
import studentCategoryProgressSource from '@/features/student-dashboard/ui/StudentCategoryProgressContent.vue?raw'
import studentDifficultySource from '@/features/student-dashboard/ui/StudentDifficultyContent.vue?raw'
import trainingTimelineContentSource from '@/entities/training-timeline/ui/TrainingTimelineContent.vue?raw'
import classManagementPageSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import studentManagementPageSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import teacherInstanceManagementPageSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import teacherDashboardSourceBase from '@/features/teacher/dashboard/ui/TeacherDashboardPage.vue?raw'
import teacherDashboardPortraitPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue?raw'
import teacherDashboardStudentInsightPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue?raw'
import teacherDashboardTrendPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue?raw'
import teacherDashboardReviewPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue?raw'
import teacherDashboardInterventionPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue?raw'
import adminDashboardSourceBase from '@/features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import contestManageOverviewPanelSource from '@/features/platform/contest-manage/ui/ContestManageOverviewPanel.vue?raw'
import pageHeaderSource from '@/shared/ui/common/PageHeader.vue?raw'
import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeDetailWidgetSource from '@/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue?raw'
import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import contestDetailWidgetSource from '@/widgets/contest-detail-workspace/ContestDetailWorkspace.vue?raw'
import contestOverviewPanelSource from '@/features/contest-detail/ui/ContestOverviewPanel.vue?raw'
import errorStatusShellSource from '@/shared/ui/errors/ErrorStatusShell.vue?raw'

const sharedStylesSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

const contestListWorkspaceCombinedSource = `${contestListSource}\n${contestListWorkspaceSource}`
const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`
const teacherInstanceManagementPageSource = [
  teacherInstanceManagementPageSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')
const teacherDashboardSource = [
  teacherDashboardSourceBase,
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
const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')
const challengeDetailWorkspaceSource = [
  challengeDetailSource,
  challengeDetailWidgetSource,
].join('\n')
const contestDetailWorkspaceSource = [
  contestDetailSource,
  contestDetailWidgetSource,
  contestOverviewPanelSource,
].join('\n')

describe('workspace page header styles', () => {
  it('应该在全局样式中声明共享页级标题与说明文案入口', () => {
    expect(sharedStylesSource).toContain('.workspace-page-title')
    expect(sharedStylesSource).toContain('.workspace-page-copy')
    expect(sharedStylesSource).toContain('--workspace-page-title-font-size:')
    expect(sharedStylesSource).toContain('--workspace-page-title-line-height:')
    expect(sharedStylesSource).toContain('--workspace-page-title-letter-spacing:')
  })

  it('典型页面标题应接入 workspace-page-title，而不是继续混用 workspace-tab-heading 标题', () => {
    expect(studentOverviewSource).toContain('workspace-page-title')
    expect(studentRecommendationSource).toContain('workspace-page-title')
    expect(studentCategoryProgressSource).toContain('workspace-page-title')
    expect(studentDifficultySource).toContain('workspace-page-title')
    expect(trainingTimelineContentSource).toContain('workspace-page-title')
    expect(classManagementPageSource).toContain('workspace-page-title')
    expect(studentManagementPageSource).toContain('workspace-page-title')
    expect(teacherInstanceManagementPageSource).toContain('workspace-page-title')
    expect(teacherDashboardSource).toContain('workspace-page-title')
    expect(adminDashboardSource).toContain('workspace-page-title')
    expect(userGovernanceSource).toContain('workspace-page-title')

    for (const source of [
      studentOverviewSource,
      studentRecommendationSource,
      studentCategoryProgressSource,
      studentDifficultySource,
      classManagementPageSource,
      studentManagementPageSource,
      teacherInstanceManagementPageSource,
    ]) {
      expect(source).not.toContain('workspace-tab-heading__title')
    }
  })

  it('页级说明应接入 workspace-page-copy，而不是继续使用 workspace-tab-copy', () => {
    for (const source of [
      studentOverviewSource,
      studentRecommendationSource,
      studentCategoryProgressSource,
      studentDifficultySource,
      trainingTimelineContentSource,
      classManagementPageSource,
      studentManagementPageSource,
      teacherInstanceManagementPageSource,
      userGovernanceSource,
    ]) {
      expect(source).toContain('workspace-page-copy')
      expect(source).not.toContain('workspace-tab-copy')
    }
  })

  it('典型页头 owner 应继续收口到 workspace-page-header 或 PageHeader', () => {
    expect(contestListWorkspaceCombinedSource).toContain('workspace-page-header')
    expect(instanceListWorkspaceSource).toContain('workspace-page-header')
    expect(challengeListSource).toContain('workspace-page-header')
    expect(pageHeaderSource).toContain('workspace-page-title')
    expect(pageHeaderSource).toContain('workspace-page-copy')
    expect(contestDetailWorkspaceSource).toContain('workspace-page-copy')
    expect(errorStatusShellSource).toContain('workspace-page-title')
  })

  it('overview 面板应继续复用 workspace-panel-header 结构', () => {
    expect(userGovernanceSource).toContain('<header class="workspace-panel-header user-overview-head">')
    expect(userGovernanceSource).not.toContain(
      '<header class="workspace-page-header user-overview-head">'
    )
    expect(contestManageOverviewPanelSource).toContain(
      '<header class="workspace-panel-header contest-overview-head">'
    )
    expect(contestManageOverviewPanelSource).not.toContain(
      '<header class="workspace-page-header contest-overview-head">'
    )
  })
})
