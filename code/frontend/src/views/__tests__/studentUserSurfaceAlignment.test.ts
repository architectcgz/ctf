import { describe, expect, it } from 'vitest'

import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'

describe('student and user surface alignment', () => {
  it('student dashboard detail pages should use softened control and track tokens instead of bright hardcoded borders', () => {
    expect(studentOverviewSource).toContain('--journal-shell-border:')
    expect(studentOverviewSource).toContain('--journal-soft-border:')
    expect(studentOverviewSource).toContain('--journal-divider:')
    expect(studentOverviewSource).toMatch(/\.journal-metric,[\s\S]*\.journal-rank-summary\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s)
    expect(studentOverviewSource).not.toContain('border-[var(--journal-border)]')
    expect(studentOverviewSource).not.toMatch(/border:\s*1px solid var\(--journal-border\);/)

    expect(studentDifficultySource).toContain('--journal-soft-border:')
    expect(studentDifficultySource).toContain('--journal-track:')
    expect(studentDifficultySource).toMatch(/\.stat-icon\s*\{[\s\S]*border:\s*1px solid var\(--journal-soft-border\);/s)
    expect(studentDifficultySource).not.toContain('rgba(226, 232, 240, 0.72)')
    expect(studentDifficultySource).not.toContain('bg-[rgba(226,232,240,0.65)]')

    expect(studentTimelineSource).toContain('--journal-soft-border:')
    expect(studentTimelineSource).toMatch(/\.stat-icon\s*\{[\s\S]*border:\s*1px solid var\(--journal-soft-border\);/s)
    expect(studentTimelineSource).not.toContain('rgba(226, 232, 240, 0.72)')

    expect(studentRecommendationSource).toContain('--journal-control-border:')
    expect(studentRecommendationSource).toMatch(/\.journal-btn-outline\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(studentRecommendationSource).not.toContain('border-slate-200')
    expect(studentRecommendationSource).not.toContain('bg-slate-50')
    expect(studentRecommendationSource).not.toContain('border-emerald-200')
    expect(studentRecommendationSource).not.toContain('bg-emerald-50')

    expect(studentCategoryProgressSource).toContain('--journal-control-border:')
    expect(studentCategoryProgressSource).toContain('--journal-track:')
    expect(studentCategoryProgressSource).toMatch(/\.category-track\s*\{[\s\S]*background:\s*var\(--journal-track\);/s)
    expect(studentCategoryProgressSource).toMatch(/\.journal-btn-outline\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(studentCategoryProgressSource).not.toContain('rgba(226, 232, 240, 0.65)')
  })

  it('instance and notification pages should soften list shells, controls, and empty-state separators', () => {
    expect(instanceListSource).toContain('--journal-control-border:')
    expect(instanceListSource).toContain('--journal-shell-border:')
    expect(instanceListSource).toContain('--journal-divider:')
    expect(instanceListSource).toMatch(/\.journal-btn\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(instanceListSource).toMatch(/\.instance-list\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s)
    expect(instanceListSource).not.toContain('border-[var(--journal-border)]')
    expect(instanceListSource).not.toContain('border-[var(--journal-border)]/80')
    expect(instanceListSource).not.toMatch(/border:\s*1px solid var\(--journal-border\);/)

    expect(notificationListSource).toContain('--journal-control-border:')
    expect(notificationListSource).toContain('--journal-shell-border:')
    expect(notificationListSource).toContain('--journal-divider:')
    expect(notificationListSource).toMatch(/\.journal-btn\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(notificationListSource).toMatch(/\.notification-list\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s)
    expect(notificationListSource).not.toMatch(/border:\s*1px solid var\(--journal-border\);/)
    expect(notificationListSource).not.toContain('rgba(148, 163, 184, 0.58)')
  })
})
