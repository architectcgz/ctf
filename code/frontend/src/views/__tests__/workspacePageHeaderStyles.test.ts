import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import scoreboardViewSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import challengeManageSource from '@/views/admin/ChallengeManage.vue?raw'
import auditLogSource from '@/views/admin/AuditLog.vue?raw'
import imageManageSource from '@/views/admin/ImageManage.vue?raw'
import challengePackageFormatSource from '@/views/admin/ChallengePackageFormat.vue?raw'
import challengeImportManageSource from '@/views/admin/ChallengeImportManage.vue?raw'
import cheatDetectionSource from '@/views/admin/CheatDetection.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import reportExportSource from '@/views/teacher/ReportExport.vue?raw'
import classManagementPageSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import studentManagementPageSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import teacherInstanceManagementPageSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import topologyStudioSource from '@/components/admin/topology/ChallengeTopologyStudioPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import teacherDashboardSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import adminDashboardSource from '@/components/admin/dashboard/AdminDashboardPage.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import writeupManageSource from '@/components/admin/writeup/ChallengeWriteupManagePanel.vue?raw'
import writeupEditorSource from '@/components/admin/writeup/ChallengeWriteupEditorPage.vue?raw'
import writeupViewSource from '@/components/admin/writeup/ChallengeWriteupViewPage.vue?raw'
import pageHeaderSource from '@/components/common/PageHeader.vue?raw'
import adminChallengeDetailSource from '@/views/admin/ChallengeDetail.vue?raw'
import challengeImportPreviewSource from '@/views/admin/ChallengeImportPreview.vue?raw'
import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import notificationDetailSource from '@/views/notifications/NotificationDetail.vue?raw'
import reviewArchiveHeroSource from '@/components/teacher/review-archive/ReviewArchiveHero.vue?raw'
import authEntryShellSource from '@/components/auth/AuthEntryShell.vue?raw'
import errorStatusShellSource from '@/components/errors/ErrorStatusShell.vue?raw'

const sharedStylesSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function expectNoLocalTitleTypography(source: string, selector: string): void {
  const escapedSelector = escapeRegExp(selector)
  expect(source).not.toMatch(new RegExp(`${escapedSelector}\\s*\\{[^}]*font-size:`, 's'))
  expect(source).not.toMatch(new RegExp(`${escapedSelector}\\s*\\{[^}]*letter-spacing:`, 's'))
}

function expectNoLocalCopyTypography(source: string, selector: string): void {
  const escapedSelector = escapeRegExp(selector)
  expect(source).not.toMatch(new RegExp(`${escapedSelector}\\s*\\{[^}]*font-size:`, 's'))
  expect(source).not.toMatch(new RegExp(`${escapedSelector}\\s*\\{[^}]*line-height:`, 's'))
}

function expectSourceToContain(source: string, pattern: string | RegExp): void {
  if (typeof pattern === 'string') {
    expect(source).toContain(pattern)
    return
  }
  expect(source).toMatch(pattern)
}

function expectSourceNotToContain(source: string, pattern: string | RegExp): void {
  if (typeof pattern === 'string') {
    expect(source).not.toContain(pattern)
    return
  }
  expect(source).not.toMatch(pattern)
}

describe('workspace page header styles', () => {
  it('应该在全局样式中统一声明工作区页面标题与说明文案样式', () => {
    const sharedTitleSelectors = [
      '.workspace-page-title',
      '.journal-page-title',
      '.challenge-title',
      '.contest-title',
      '.instance-title',
      '.scoreboard-title',
      '.notification-title',
      '.teacher-management-shell .teacher-title',
      '.manage-title',
      '.hero-title',
      '.admin-page-title',
      '.image-title',
      '.report-title',
      '.topology-hero-title',
    ]

    const sharedCopySelectors = [
      '.workspace-page-copy',
      '.challenge-subtitle',
      '.contest-subtitle',
      '.instance-subtitle',
      '.scoreboard-subtitle',
      '.notification-subtitle',
      '.teacher-management-shell .teacher-copy',
      '.admin-page-copy',
      '.image-copy',
      '.report-copy',
      '.skill-overview-copy',
      '.topology-hero-description',
    ]

    for (const selector of sharedTitleSelectors) {
      expect(sharedStylesSource).toContain(selector)
    }

    for (const selector of sharedCopySelectors) {
      expect(sharedStylesSource).toContain(selector)
    }

    expect(sharedStylesSource).toContain(
      '--workspace-page-title-font-size: clamp(24px, 3vw, 34px);'
    )
    expect(sharedStylesSource).toContain('--workspace-page-title-line-height: 1.08;')
    expect(sharedStylesSource).toContain('--workspace-page-title-letter-spacing: -0.03em;')
  })

  it('不应在页面局部重复声明公共标题排版', () => {
    expectNoLocalTitleTypography(contestListSource, '.contest-title')
    expectNoLocalTitleTypography(challengeListSource, '.challenge-title')
    expectNoLocalTitleTypography(instanceListSource, '.instance-title')
    expectNoLocalTitleTypography(notificationListSource, '.notification-title')
    expectNoLocalTitleTypography(scoreboardViewSource, '.scoreboard-title')
    expectNoLocalTitleTypography(challengeManageSource, '.manage-title')
    expectNoLocalTitleTypography(auditLogSource, '.admin-page-title')
    expectNoLocalTitleTypography(imageManageSource, '.image-title')
    expectNoLocalTitleTypography(challengePackageFormatSource, '.hero-title')
    expectNoLocalTitleTypography(cheatDetectionSource, '.hero-title')
    expectNoLocalTitleTypography(skillProfileSource, '.journal-page-title')
    expectNoLocalTitleTypography(classManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(classStudentsPageSource, '.teacher-title')
    expectNoLocalTitleTypography(studentAnalysisPageSource, '.teacher-title')
    expectNoLocalTitleTypography(studentManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(teacherInstanceManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(
      topologyStudioSource,
      '.topology-page--template-library .topology-hero-title'
    )
  })

  it('不应在页面局部重复声明公共说明排版', () => {
    expectNoLocalCopyTypography(contestListSource, '.contest-subtitle')
    expectNoLocalCopyTypography(challengeListSource, '.challenge-subtitle')
    expectNoLocalCopyTypography(instanceListSource, '.instance-subtitle')
    expectNoLocalCopyTypography(notificationListSource, '.notification-subtitle')
    expectNoLocalCopyTypography(scoreboardViewSource, '.scoreboard-subtitle')
    expectNoLocalCopyTypography(auditLogSource, '.admin-page-copy')
    expectNoLocalCopyTypography(imageManageSource, '.image-copy')
    expectNoLocalCopyTypography(classManagementPageSource, '.teacher-copy')
    expectNoLocalCopyTypography(classStudentsPageSource, '.teacher-copy')
    expectNoLocalCopyTypography(studentAnalysisPageSource, '.teacher-copy')
    expectNoLocalCopyTypography(studentManagementPageSource, '.teacher-copy')
    expectNoLocalCopyTypography(teacherInstanceManagementPageSource, '.teacher-copy')
    expectNoLocalCopyTypography(
      topologyStudioSource,
      '.topology-page--template-library .topology-hero-description'
    )
  })

  it('student overview 标题应接入共享页级标题类，而不是继续混用 tab 标题类', () => {
    expect(studentOverviewSource).toContain(
      '<h1 class="journal-page-title workspace-page-title journal-soft-page-title max-w-3xl">'
    )
    expect(studentOverviewSource).not.toContain(
      '<h1 class="journal-page-title workspace-tab-heading__title max-w-3xl text-[var(--journal-ink)]">'
    )
  })

  it('workspace 页级标题应统一接入共享页头类，而不是继续使用 tab 标题类', () => {
    const pageTitleSources = [
      {
        source: studentRecommendationSource,
        include: /<h1 class="journal-page-title workspace-page-title[\s\S]*?>/,
        exclude: /<h1 class="journal-page-title workspace-tab-heading__title[\s\S]*?>/,
      },
      {
        source: studentCategoryProgressSource,
        include: /<h1 class="journal-page-title workspace-page-title[\s\S]*?>/,
        exclude: /<h1 class="journal-page-title workspace-tab-heading__title[\s\S]*?>/,
      },
      {
        source: studentTimelineSource,
        include: /<h1 class="journal-page-title workspace-page-title[\s\S]*?>/,
        exclude: /<h1 class="journal-page-title workspace-tab-heading__title[\s\S]*?>/,
      },
      {
        source: studentDifficultySource,
        include: /<h1 class="journal-page-title workspace-page-title[\s\S]*?>/,
        exclude: /<h1 class="journal-page-title workspace-tab-heading__title[\s\S]*?>/,
      },
      {
        source: teacherDashboardSource,
        include: '<h1 class="hero-title">教学介入台</h1>',
        exclude: '<h1 class="hero-title workspace-tab-heading__title">教学介入台</h1>',
      },
      {
        source: studentAnalysisPageSource,
        include: /<h1 class="teacher-title">[\s\S]*?<\/h1>/,
        exclude: /<h1 class="teacher-title workspace-tab-heading__title">[\s\S]*?<\/h1>/,
      },
      {
        source: adminDashboardSource,
        include: '<h1 class="hero-title">系统值守台</h1>',
        exclude: '<h1 class="hero-title workspace-tab-heading__title">系统值守台</h1>',
      },
      {
        source: userGovernanceSource,
        include: '<h1 class="workspace-page-title">用户治理台</h1>',
        exclude: '<h1 class="workspace-tab-heading__title">用户治理台</h1>',
      },
      {
        source: contestOrchestrationSource,
        include: '<h1 class="workspace-page-title">竞赛目录</h1>',
        exclude: '<h1 class="workspace-page-title workspace-tab-heading__title">竞赛目录</h1>',
      },
      {
        source: challengeManageSource,
        include: /<h1 class="workspace-page-title">(?:题目管理|题目资源管理中心)<\/h1>/,
        exclude:
          /<h1 class="workspace-tab-heading__title">(?:题目管理|题目资源管理中心)<\/h1>/,
      },
      {
        source: challengeImportManageSource,
        include: /<h1 class="workspace-page-title">导入资源包<\/h1>/,
        exclude: /<h1 class="workspace-tab-heading__title">导入资源包<\/h1>/,
      },
      {
        source: challengeImportPreviewSource,
        include: /<PageHeader[\s\S]*title="导入预览"/,
        exclude: '<h1 class="workspace-tab-heading__title">导入预览</h1>',
      },
      {
        source: adminChallengeDetailSource,
        include: '<h1 class="workspace-page-title">题目管理</h1>',
        exclude: '<h1 class="workspace-tab-heading__title">题目管理</h1>',
      },
      {
        source: writeupManageSource,
        include: '<h1 class="workspace-page-title">题解管理</h1>',
        exclude: '<h1 class="workspace-tab-heading__title">题解管理</h1>',
      },
      {
        source: writeupEditorSource,
        include: /<PageHeader[\s\S]*title="题解管理"/,
        exclude: '<h1 class="workspace-tab-heading__title">题解管理</h1>',
      },
      {
        source: writeupViewSource,
        include: /<PageHeader[\s\S]*:title="writeup\.title"/,
        exclude: /<h1 class="workspace-tab-heading__title">\{\{ writeup\.title \}\}<\/h1>/,
      },
      {
        source: topologyStudioSource,
        include: /<h1 class="hero-title">\{\{ heroTitle \}\}<\/h1>/,
        exclude: /<h1 class="hero-title workspace-tab-heading__title">\{\{ heroTitle \}\}<\/h1>/,
      },
      {
        source: pageHeaderSource,
        include: '<h1 class="workspace-page-title">{{ title }}</h1>',
        exclude:
          '<h1 class="text-3xl font-semibold tracking-tight text-text-primary">{{ title }}</h1>',
      },
    ]

    for (const entry of pageTitleSources) {
      expectSourceToContain(entry.source, entry.include)
      expectSourceNotToContain(entry.source, entry.exclude)
    }

    expect(pageHeaderSource).toContain('class="workspace-page-copy"')
    expect(pageHeaderSource).not.toContain(
      'class="max-w-3xl text-sm leading-6 text-text-secondary"'
    )
  })

  it('独立详情和状态页标题也应接入共享页级标题类', () => {
    const specialPageTitleSources = [
      {
        source: challengeDetailSource,
        include: '<h1 class="question-title workspace-page-title">',
        exclude: '<h1 class="question-title">',
      },
      {
        source: contestDetailSource,
        include: '<h1 class="contest-hero__title workspace-page-title">{{ contest.title }}</h1>',
        exclude: '<h1 class="contest-hero__title">{{ contest.title }}</h1>',
      },
      {
        source: notificationDetailSource,
        include: '<h1 class="notification-detail-title workspace-page-title">',
        exclude: '<h1 class="notification-detail-title">',
      },
      {
        source: reviewArchiveHeroSource,
        include: '<h1 class="archive-hero__title workspace-page-title">教学复盘归档</h1>',
        exclude: '<h1 class="archive-hero__title">教学复盘归档</h1>',
      },
      {
        source: authEntryShellSource,
        include: '<h1 class="auth-entry-shell__title workspace-page-title">教学平台入口</h1>',
        exclude: '<h1 class="auth-entry-shell__title">教学平台入口</h1>',
      },
      {
        source: errorStatusShellSource,
        include: '<h1 class="error-status-title workspace-page-title">',
        exclude: '<h1 class="error-status-title">',
      },
    ]

    for (const entry of specialPageTitleSources) {
      expectSourceToContain(entry.source, entry.include)
      expectSourceNotToContain(entry.source, entry.exclude)
    }

    expect(contestDetailSource).toContain('<p class="contest-hero__desc workspace-page-copy">')
    expect(reviewArchiveHeroSource).toContain(
      '<p class="archive-hero__description workspace-page-copy">'
    )
    expect(authEntryShellSource).toContain('<p class="auth-entry-shell__copy workspace-page-copy">')
    expect(errorStatusShellSource).toContain('<p class="error-status-text workspace-page-copy">')
  })

  it('页级说明应统一接入共享页级说明类，而不是继续使用 tab copy', () => {
    const pageCopySources = [
      {
        source: studentOverviewSource,
        include: /<p class="workspace-page-copy max-w-2xl[^"]*">/,
        exclude:
          '<p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">',
      },
      {
        source: studentRecommendationSource,
        include: /<p class="workspace-page-copy max-w-2xl[^"]*">/,
        exclude:
          '<p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">',
      },
      {
        source: studentCategoryProgressSource,
        include: /<p class="workspace-page-copy max-w-2xl[^"]*">/,
        exclude:
          '<p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">',
      },
      {
        source: studentDifficultySource,
        include: /<p class="workspace-page-copy max-w-2xl[^"]*">/,
        exclude:
          '<p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">',
      },
      {
        source: studentTimelineSource,
        include: /<p class="workspace-page-copy max-w-2xl[^"]*">/,
        exclude:
          '<p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">',
      },
      {
        source: userGovernanceSource,
        include: '<p class="workspace-page-copy">',
        exclude: '<p class="workspace-tab-copy">',
      },
      {
        source: challengeManageSource,
        include: '<p class="workspace-page-copy">',
        exclude: '<p class="workspace-tab-copy">',
      },
      {
        source: adminChallengeDetailSource,
        include: '<p class="workspace-page-copy">',
        exclude: '<p class="workspace-tab-copy">',
      },
      {
        source: topologyStudioSource,
        include: '<p class="workspace-page-copy topology-page-copy">',
        exclude: '<p class="workspace-tab-copy topology-page-copy">',
      },
    ]

    for (const entry of pageCopySources) {
      expectSourceToContain(entry.source, entry.include)
      expectSourceNotToContain(entry.source, entry.exclude)
    }
  })

  it('典型工作区页头应复用 PageHeader 组件，而不是继续手写同构头部', () => {
    expect(challengeImportPreviewSource).toContain('<PageHeader')
    expect(userProfileSource).toContain('<PageHeader')
    expect(securitySettingsSource).toContain('<PageHeader')
    expect(writeupEditorSource).toContain('<PageHeader')
    expect(writeupViewSource).toContain('<PageHeader')
  })

  it('高频详情页顶部 tab 触控高度应回到共享默认值', () => {
    expect(challengeDetailSource).toMatch(
      /\.challenge-subtabs\s*\{[\s\S]*--page-top-tab-min-height: 3rem;/s
    )
    expect(challengeDetailSource).not.toMatch(
      /\.challenge-subtabs\s*\{[\s\S]*--page-top-tab-min-height: 2\.5rem;/s
    )
    expect(contestDetailSource).toContain('--page-top-tab-min-height: 3rem;')
    expect(contestDetailSource).not.toContain('--page-top-tab-min-height: 2.5rem;')
  })
})
