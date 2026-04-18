import { describe, expect, it } from 'vitest'

import appCardSource from '@/components/common/AppCard.vue?raw'
import appLayoutSource from '@/components/layout/AppLayout.vue?raw'
import pageHeaderSource from '@/components/common/PageHeader.vue?raw'
import skillRadarSource from '@/components/common/SkillRadar.vue?raw'
import radarChartSource from '@/components/charts/RadarChart.vue?raw'
import errorStatusShellSource from '@/components/errors/ErrorStatusShell.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentLegacyOverviewSource from '@/components/dashboard/student/StudentOverviewPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import notificationDropdownSource from '@/components/layout/NotificationDropdown.vue?raw'
import sidebarSource from '@/components/layout/Sidebar.vue?raw'
import topNavSource from '@/components/layout/TopNav.vue?raw'
import cLightActionPopoverSource from '@/components/common/modal-templates/CLightActionPopover.vue?raw'
import challengePackageImportEntrySource from '@/components/admin/challenge/ChallengePackageImportEntry.vue?raw'
import adminDashboardSource from '@/components/admin/dashboard/AdminDashboardPage.vue?raw'
import writeupManageSource from '@/components/admin/writeup/ChallengeWriteupManagePanel.vue?raw'
import awdRoundInspectorSource from '@/components/admin/contest/AWDRoundInspector.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import topologyCanvasBoardSource from '@/components/admin/topology/TopologyCanvasBoard.vue?raw'
import topologyStudioSource from '@/components/admin/topology/ChallengeTopologyStudioPage.vue?raw'
import adminNotificationPublishDrawerSource from '@/components/notifications/AdminNotificationPublishDrawer.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import imageManageSource from '@/views/admin/ImageManage.vue?raw'
import challengeManageSource from '@/views/admin/ChallengeManage.vue?raw'
import adminChallengeDetailSource from '@/views/admin/ChallengeDetail.vue?raw'
import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'

function expectNoHardcodedThemeTokens(
  source: string,
  label: string,
  forbiddenTokens: string[]
): void {
  forbiddenTokens.forEach((token) => {
    expect(source, `${label} 不应包含 ${token}`).not.toContain(token)
  })
}

describe('shared theme token adoption', () => {
  it('共享组件与基础图表不应回退到硬编码主题色', () => {
    expectNoHardcodedThemeTokens(appCardSource, 'AppCard', ['rgba(139,148,158,0.08)'])
    expectNoHardcodedThemeTokens(appLayoutSource, 'AppLayout', [
      'rgba(8,145,178,0.12)',
      'rgba(255,255,255,0.08)',
      'rgba(0,0,0,0.16)',
    ])
    expectNoHardcodedThemeTokens(pageHeaderSource, 'PageHeader', [
      'rgba(15, 23, 42, 0.05)',
      '#f8fafc',
    ])
    expectNoHardcodedThemeTokens(skillRadarSource, 'SkillRadar', [
      'rgba(148, 163, 184, 0.18)',
      'stroke="#0891b2"',
      'fill="#06b6d4"',
      'fill="#cbd5e1"',
    ])
    expectNoHardcodedThemeTokens(radarChartSource, 'RadarChart', ["|| '#475569'"])
    expectNoHardcodedThemeTokens(errorStatusShellSource, 'ErrorStatusShell', [
      '#0b4f60',
      '#f8feff',
      '#f1f5f9',
    ])
    expectNoHardcodedThemeTokens(cLightActionPopoverSource, 'CLightActionPopover', [
      '#2a7a58',
      '#206346',
      'bg-white',
      'text-slate-700',
      'text-slate-900',
      'border-slate-200',
      'shadow-[0_12px_40px_rgba(0,0,0,0.12)]',
    ])
  })

  it('共享组件不应继续保留低信息度的 Tailwind 任意值魔法数', () => {
    expect(pageHeaderSource).not.toContain('text-[11px]')
    expect(pageHeaderSource).not.toContain('tracking-[0.26em]')

    expect(appCardSource).not.toContain('text-[24px]')
    expect(appCardSource).not.toContain('text-[15px]')
    expect(appCardSource).not.toContain('text-[13px]')
    expect(appCardSource).not.toContain('text-[10px]')
    expect(appCardSource).not.toContain('w-[3px]')
    expect(notificationDropdownSource).not.toContain('w-[1px]')
    expect(sidebarSource).not.toContain('w-[260px]')
    expect(sidebarSource).not.toContain('text-[10px]')
    expect(sidebarSource).not.toContain('text-[13px]')
    expect(sidebarSource).not.toContain('ml-[22px]')
    expect(sidebarSource).not.toContain('-left-[14px]')
    expect(topNavSource).not.toContain('max-w-[1600px]')
    expect(topNavSource).not.toContain('md:text-[15px]')
  })

  it('学生仪表盘与学习工作区不应继续写死状态色', () => {
    expectNoHardcodedThemeTokens(studentDifficultySource, 'StudentDifficultyPage', [
      "beginner: '#10b981'",
      "easy: '#22d3ee'",
      "medium: '#f59e0b'",
      "hard: '#f97316'",
      "insane: '#ef4444'",
    ])
    expectNoHardcodedThemeTokens(studentTimelineSource, 'StudentTimelinePage', [
      'color: #10b981;',
      'color: #f59e0b;',
      'background: #22c55e;',
      'background: #10b981;',
      'background: #94a3b8;',
      'rgba(16, 185, 129, 0.38)',
    ])
    expectNoHardcodedThemeTokens(studentOverviewSource, 'StudentOverviewStyleEditorial', [
      'background: #10b981;',
      'background: #94a3b8;',
      'background: #f59e0b;',
      'background: #22c55e;',
      'rgba(16, 185, 129, 0.4)',
      'rgba(16, 185, 129, 0.38)',
      'rgba(148, 163, 184, 0.2)',
    ])
    expectNoHardcodedThemeTokens(studentLegacyOverviewSource, 'StudentOverviewPage', [
      'rgba(8,47,73,0.35)',
      'rgba(15,23,42,0.55)',
    ])
    expectNoHardcodedThemeTokens(studentRecommendationSource, 'StudentRecommendationPage', [
      '#16a34a',
      '#15803d',
    ])
    expectNoHardcodedThemeTokens(studentCategoryProgressSource, 'StudentCategoryProgressPage', [
      '#0ea5e9',
    ])
    expectNoHardcodedThemeTokens(skillProfileSource, 'SkillProfile', ['rgba(148, 163, 184, 0.2)'])
  })

  it('挑战与榜单页不应回退到旧色值', () => {
    expectNoHardcodedThemeTokens(scoreboardSource, 'ScoreboardView', [
      '#b45309',
      '#475569',
      '#92400e',
    ])
    expectNoHardcodedThemeTokens(challengeListSource, 'ChallengeList', [
      '#0f766e',
      '#7c3aed',
      '#ea580c',
      '#0891b2',
      '#2563eb',
      '#d97706',
    ])
    expectNoHardcodedThemeTokens(challengeDetailSource, 'ChallengeDetail', [
      'rgba(13, 23, 39, 0.06)',
    ])
  })

  it('通知与导航壳层不应带回旧的亮色回退值', () => {
    expectNoHardcodedThemeTokens(
      adminNotificationPublishDrawerSource,
      'AdminNotificationPublishDrawer',
      [
        'var(--color-border, #d8e1ec)',
        'var(--color-text-muted, #64748b)',
        'var(--color-text, #0f172a)',
        'var(--color-primary, #3b82f6)',
        'var(--color-danger, #dc2626)',
        'var(--color-bg-elevated, #fff)',
      ]
    )
    expectNoHardcodedThemeTokens(notificationDropdownSource, 'NotificationDropdown', [
      '0 8px 18px rgba(15, 23, 42, 0.04)',
      '0 18px 42px rgba(15, 23, 42, 0.14)',
    ])
    expectNoHardcodedThemeTokens(sidebarSource, 'Sidebar', [
      '0 18px 48px rgba(15, 23, 42, 0.18)',
      '0 18px 48px rgba(15, 23, 42, 0.16)',
      'rgba(99, 102, 241, 0.06)',
    ])
    expectNoHardcodedThemeTokens(topNavSource, 'TopNav', [
      'rgba(99, 102, 241, 0.06)',
      '#16a34a',
      '#0891b2',
      '#2563eb',
      '#e18a2a',
    ])
  })

  it('个人资料与安全设置页不应回退到浅色状态块', () => {
    expectNoHardcodedThemeTokens(userProfileSource, 'UserProfile', [
      '0 18px 40px rgba(15, 23, 42, 0.05)',
      'border: 1px solid rgba(16, 185, 129, 0.22);',
      'background: rgba(16, 185, 129, 0.08);',
      'background: #10b981;',
      'background: #f59e0b;',
      'background: #94a3b8;',
      'background: #ef4444;',
    ])
    expectNoHardcodedThemeTokens(securitySettingsSource, 'SecuritySettings', [
      'rgba(15, 23, 42, 0.95)',
      'rgba(2, 6, 23, 0.98)',
      'background: #10b981;',
      'rgba(16, 185, 129, 0.2)',
    ])
  })

  it('后台工作区页不应回退到局部硬编码主题令牌', () => {
    expectNoHardcodedThemeTokens(challengePackageImportEntrySource, 'ChallengePackageImportEntry', [
      '#2563eb',
      'rgba(37, 99, 235, 0.12)',
    ])
    expectNoHardcodedThemeTokens(adminDashboardSource, 'AdminDashboardPage', [
      'background: #1d4ed8;',
      'border-color: rgba(37, 99, 235, 0.28);',
    ])
    expectNoHardcodedThemeTokens(writeupManageSource, 'ChallengeWriteupManagePanel', [
      '#38bdf8',
      '#ef4444',
      '#b91c1c',
    ])
    expectNoHardcodedThemeTokens(awdRoundInspectorSource, 'AWDRoundInspector', [
      'rgba(8,145,178,0.15)',
      'rgba(15,23,42,0.94)',
      'border-white/10',
      'bg-white/5',
      'rgba(8,145,178,0.06)',
      'rgba(8,145,178,0)',
    ])
    expectNoHardcodedThemeTokens(contestOrchestrationSource, 'ContestOrchestrationPage', [
      '0 8px 18px rgba(15, 23, 42, 0.035);',
    ])
    expectNoHardcodedThemeTokens(topologyCanvasBoardSource, 'TopologyCanvasBoard', [
      'var(--color-primary) 74%, white',
      'var(--color-warning) 90%, white',
      'var(--color-warning) 90%, #f8fafc',
    ])
    expectNoHardcodedThemeTokens(topologyStudioSource, 'ChallengeTopologyStudioPage', [
      'rgba(15, 23, 42, 0.96)',
      'rgba(15, 23, 42, 0.9)',
    ])
    expectNoHardcodedThemeTokens(imageManageSource, 'ImageManage', [
      "pending: '#8b949e'",
      "building: '#f59e0b'",
      "available: '#10b981'",
      "failed: '#ef4444'",
      'rgba(15, 23, 42, 0.96)',
      'rgba(15, 23, 42, 0.9)',
    ])
    expectNoHardcodedThemeTokens(challengeManageSource, 'ChallengeManage', [
      '#059669',
      '#047857',
      '#dc2626',
      '#b91c1c',
    ])
    expectNoHardcodedThemeTokens(adminChallengeDetailSource, 'AdminChallengeDetail', [
      'rgba(37, 99, 235, 0.18)',
      'rgba(37, 99, 235, 0.08)',
      'rgba(37, 99, 235, 0.42)',
      'rgba(37, 99, 235, 0.12)',
    ])
    expectNoHardcodedThemeTokens(userGovernanceSource, 'UserGovernancePage', [
      "return '#f59e0b'",
      "return '#dc2626'",
      "return '#64748b'",
      'rgba(148, 163, 184, 0.7)',
      'rgba(239, 68, 68, 0.2)',
      'rgba(254, 242, 242, 0.9)',
      'rgba(148, 163, 184, 0.72)',
    ])
  })
})
