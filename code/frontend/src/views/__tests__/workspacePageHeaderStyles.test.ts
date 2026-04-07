import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import scoreboardViewSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import challengeManageSource from '@/views/admin/ChallengeManage.vue?raw'
import auditLogSource from '@/views/admin/AuditLog.vue?raw'
import imageManageSource from '@/views/admin/ImageManage.vue?raw'
import challengePackageFormatSource from '@/views/admin/ChallengePackageFormat.vue?raw'
import cheatDetectionSource from '@/views/admin/CheatDetection.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import reportExportSource from '@/views/teacher/ReportExport.vue?raw'
import classManagementPageSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import studentManagementPageSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import teacherInstanceManagementPageSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import topologyStudioSource from '@/components/admin/topology/ChallengeTopologyStudioPage.vue?raw'

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

describe('workspace page header styles', () => {
  it('应该在全局样式中统一声明工作区页面标题与说明文案样式', () => {
    const sharedTitleSelectors = [
      '.workspace-page-title',
      '.journal-page-title',
      '.contest-title',
      '.instance-title',
      '.scoreboard-title',
      '.notification-title',
      '.teacher-title',
      '.manage-title',
      '.hero-title',
      '.admin-page-title',
      '.image-title',
      '.report-title',
      '.topology-hero-title',
    ]

    const sharedCopySelectors = [
      '.workspace-page-copy',
      '.contest-subtitle',
      '.instance-subtitle',
      '.scoreboard-subtitle',
      '.notification-subtitle',
      '.teacher-copy',
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
  })

  it('不应在页面局部重复声明公共标题排版', () => {
    expectNoLocalTitleTypography(contestListSource, '.contest-title')
    expectNoLocalTitleTypography(instanceListSource, '.instance-title')
    expectNoLocalTitleTypography(notificationListSource, '.notification-title')
    expectNoLocalTitleTypography(scoreboardViewSource, '.scoreboard-title')
    expectNoLocalTitleTypography(challengeManageSource, '.manage-title')
    expectNoLocalTitleTypography(auditLogSource, '.admin-page-title')
    expectNoLocalTitleTypography(imageManageSource, '.image-title')
    expectNoLocalTitleTypography(challengePackageFormatSource, '.hero-title')
    expectNoLocalTitleTypography(cheatDetectionSource, '.hero-title')
    expectNoLocalTitleTypography(skillProfileSource, '.journal-page-title')
    expectNoLocalTitleTypography(reportExportSource, '.report-title')
    expectNoLocalTitleTypography(classManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(classStudentsPageSource, '.teacher-title')
    expectNoLocalTitleTypography(studentAnalysisPageSource, '.teacher-title')
    expectNoLocalTitleTypography(studentManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(teacherInstanceManagementPageSource, '.teacher-title')
    expectNoLocalTitleTypography(topologyStudioSource, '.topology-page--template-library .topology-hero-title')
  })

  it('不应在页面局部重复声明公共说明排版', () => {
    expectNoLocalCopyTypography(contestListSource, '.contest-subtitle')
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
    expectNoLocalCopyTypography(reportExportSource, '.report-copy')
    expectNoLocalCopyTypography(topologyStudioSource, '.topology-page--template-library .topology-hero-description')
  })
})
