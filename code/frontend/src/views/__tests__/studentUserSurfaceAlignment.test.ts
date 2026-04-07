import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'

const journalSoftSurfacesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-soft-surfaces.css`,
  'utf-8'
)
const journalUserShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)

describe('student and user surface alignment', () => {
  it('student dashboard detail pages should use softened control and track tokens instead of bright hardcoded borders', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-shell')
    expect(journalSoftSurfacesSource).toContain('--journal-shell-border: color-mix')
    expect(journalSoftSurfacesSource).toContain('--journal-soft-border: color-mix')
    expect(journalSoftSurfacesSource).toContain('--journal-divider: color-mix')
    expect(studentOverviewSource).toMatch(/\.journal-metric,[\s\S]*\.journal-rank-summary\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s)
    expect(studentOverviewSource).not.toContain('border-[var(--journal-border)]')
    expect(studentOverviewSource).not.toMatch(/border:\s*1px solid var\(--journal-border\);/)

    expect(studentDifficultySource).toContain('journal-soft-surface')
    expect(studentDifficultySource).toMatch(/\.stat-icon\s*\{[\s\S]*border:\s*1px solid var\(--journal-soft-border\);/s)
    expect(studentDifficultySource).not.toContain('rgba(226, 232, 240, 0.72)')
    expect(studentDifficultySource).not.toContain('bg-[rgba(226,232,240,0.65)]')

    expect(studentTimelineSource).toContain('journal-soft-surface')
    expect(studentTimelineSource).toMatch(/\.stat-icon\s*\{[\s\S]*border:\s*1px solid var\(--journal-soft-border\);/s)
    expect(studentTimelineSource).not.toContain('rgba(226, 232, 240, 0.72)')

    expect(studentRecommendationSource).toContain('journal-soft-surface')
    expect(studentRecommendationSource).toMatch(/\.journal-btn-outline\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(studentRecommendationSource).not.toContain('border-slate-200')
    expect(studentRecommendationSource).not.toContain('bg-slate-50')
    expect(studentRecommendationSource).not.toContain('border-emerald-200')
    expect(studentRecommendationSource).not.toContain('bg-emerald-50')

    expect(studentCategoryProgressSource).toContain('journal-soft-surface')
    expect(studentCategoryProgressSource).toMatch(/\.category-track\s*\{[\s\S]*background:\s*var\(--journal-track\);/s)
    expect(studentCategoryProgressSource).toMatch(/\.journal-btn-outline\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s)
    expect(studentCategoryProgressSource).not.toContain('rgba(226, 232, 240, 0.65)')
  })

  it('instance and notification pages should soften list shells, controls, and empty-state separators', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user')
    expect(journalUserShellSource).toContain('--journal-border:')
    expect(journalUserShellSource).toContain('--journal-surface:')

    expect(instanceListSource).toContain('journal-shell-user')
    expect(instanceListSource).not.toContain('border-[var(--journal-border)]')
    expect(instanceListSource).not.toContain('border-[var(--journal-border)]/80')

    expect(notificationListSource).toContain('journal-shell-user')
    expect(notificationListSource).not.toContain('rgba(148, 163, 184, 0.58)')
  })
})
