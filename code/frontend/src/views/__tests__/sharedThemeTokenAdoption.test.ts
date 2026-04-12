import { describe, expect, it } from 'vitest'

import appCardSource from '@/components/common/AppCard.vue?raw'
import appLayoutSource from '@/components/layout/AppLayout.vue?raw'
import pageHeaderSource from '@/components/common/PageHeader.vue?raw'
import skillRadarSource from '@/components/common/SkillRadar.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import notificationDropdownSource from '@/components/layout/NotificationDropdown.vue?raw'
import sidebarSource from '@/components/layout/Sidebar.vue?raw'
import topNavSource from '@/components/layout/TopNav.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'

describe('shared theme token adoption', () => {
  it('共享层和高频学生页不应继续写死状态色与图表主色', () => {
    expect(appCardSource).not.toContain('rgba(139,148,158,0.08)')
    expect(appLayoutSource).not.toContain('rgba(8,145,178,0.12)')
    expect(appLayoutSource).not.toContain('rgba(255,255,255,0.08)')
    expect(appLayoutSource).not.toContain('rgba(0,0,0,0.16)')

    expect(pageHeaderSource).not.toContain('rgba(15, 23, 42, 0.05)')
    expect(pageHeaderSource).not.toContain('#f8fafc')

    expect(skillRadarSource).not.toContain('rgba(148, 163, 184, 0.18)')
    expect(skillRadarSource).not.toContain('stroke="#0891b2"')
    expect(skillRadarSource).not.toContain('fill="#06b6d4"')
    expect(skillRadarSource).not.toContain('fill="#cbd5e1"')

    expect(studentDifficultySource).not.toContain("beginner: '#10b981'")
    expect(studentDifficultySource).not.toContain("easy: '#22d3ee'")
    expect(studentDifficultySource).not.toContain("medium: '#f59e0b'")
    expect(studentDifficultySource).not.toContain("hard: '#f97316'")
    expect(studentDifficultySource).not.toContain("insane: '#ef4444'")

    expect(studentTimelineSource).not.toContain('color: #10b981;')
    expect(studentTimelineSource).not.toContain('color: #f59e0b;')
    expect(studentTimelineSource).not.toContain('background: #22c55e;')
    expect(studentTimelineSource).not.toContain('background: #10b981;')
    expect(studentTimelineSource).not.toContain('background: #94a3b8;')
    expect(studentTimelineSource).not.toContain('rgba(16, 185, 129, 0.38)')

    expect(studentOverviewSource).not.toContain('background: #10b981;')
    expect(studentOverviewSource).not.toContain('background: #94a3b8;')
    expect(studentOverviewSource).not.toContain('background: #f59e0b;')
    expect(studentOverviewSource).not.toContain('background: #22c55e;')
    expect(studentOverviewSource).not.toContain('rgba(16, 185, 129, 0.4)')
    expect(studentOverviewSource).not.toContain('rgba(16, 185, 129, 0.38)')
    expect(studentOverviewSource).not.toContain('rgba(148, 163, 184, 0.2)')

    expect(studentRecommendationSource).not.toContain('#16a34a')
    expect(studentRecommendationSource).not.toContain('#15803d')

    expect(studentCategoryProgressSource).not.toContain('#0ea5e9')

    expect(notificationDropdownSource).not.toContain('0 8px 18px rgba(15, 23, 42, 0.04)')
    expect(notificationDropdownSource).not.toContain('0 18px 42px rgba(15, 23, 42, 0.14)')

    expect(sidebarSource).not.toContain('0 18px 48px rgba(15, 23, 42, 0.18)')
    expect(sidebarSource).not.toContain('0 18px 48px rgba(15, 23, 42, 0.16)')
    expect(sidebarSource).not.toContain('rgba(99, 102, 241, 0.06)')

    expect(topNavSource).not.toContain('rgba(99, 102, 241, 0.06)')

    expect(skillProfileSource).not.toContain('rgba(148, 163, 184, 0.2)')

    expect(userProfileSource).not.toContain('0 18px 40px rgba(15, 23, 42, 0.05)')
    expect(userProfileSource).not.toContain('border: 1px solid rgba(16, 185, 129, 0.22);')
    expect(userProfileSource).not.toContain('background: rgba(16, 185, 129, 0.08);')
    expect(userProfileSource).not.toContain('background: #10b981;')
    expect(userProfileSource).not.toContain('background: #f59e0b;')
    expect(userProfileSource).not.toContain('background: #94a3b8;')
    expect(userProfileSource).not.toContain('background: #ef4444;')

    expect(securitySettingsSource).not.toContain('rgba(15, 23, 42, 0.95)')
    expect(securitySettingsSource).not.toContain('rgba(2, 6, 23, 0.98)')
    expect(securitySettingsSource).not.toContain('background: #10b981;')
    expect(securitySettingsSource).not.toContain('rgba(16, 185, 129, 0.2)')
  })
})
