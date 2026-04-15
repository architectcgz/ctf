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
    expect(studentOverviewSource).toMatch(
      /\.journal-inline-item\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s
    )
    expect(studentOverviewSource).toContain('story-metric-grid')
    expect(studentOverviewSource).not.toContain('border-[var(--journal-border)]')
    expect(studentOverviewSource).not.toMatch(/border:\s*1px solid var\(--journal-border\);/)

    expect(studentDifficultySource).toContain('journal-soft-surface')
    expect(studentDifficultySource).toContain('.difficulty-note')
    expect(studentDifficultySource).not.toContain('rgba(226, 232, 240, 0.72)')
    expect(studentDifficultySource).not.toContain('bg-[rgba(226,232,240,0.65)]')

    expect(studentTimelineSource).toContain('journal-soft-surface')
    expect(studentTimelineSource).toMatch(
      /\.stat-icon\s*\{[\s\S]*border:\s*1px solid var\(--journal-soft-border\);/s
    )
    expect(studentTimelineSource).not.toContain('rgba(226, 232, 240, 0.72)')

    expect(journalSoftSurfacesSource).toMatch(
      /\.journal-soft-surface \.journal-btn-outline\s*\{[\s\S]*border:\s*1px solid var\(--journal-control-border\);/s
    )
    expect(studentRecommendationSource).toContain('journal-soft-surface')
    expect(studentRecommendationSource).not.toMatch(/^\.journal-btn-outline\s*\{/m)
    expect(studentRecommendationSource).not.toContain('border-slate-200')
    expect(studentRecommendationSource).not.toContain('bg-slate-50')
    expect(studentRecommendationSource).not.toContain('border-emerald-200')
    expect(studentRecommendationSource).not.toContain('bg-emerald-50')

    expect(studentCategoryProgressSource).toContain('journal-soft-surface')
    expect(studentCategoryProgressSource).toMatch(
      /\.category-track\s*\{[\s\S]*background:\s*var\(--journal-track\);/s
    )
    expect(studentCategoryProgressSource).not.toMatch(/^\.journal-btn-outline\s*\{/m)
    expect(studentCategoryProgressSource).not.toContain('rgba(226, 232, 240, 0.65)')
  })

  it('student timeline 概况卡片应显式采用统一 metric-panel 样式栈', () => {
    expect(studentTimelineSource).toContain(
      'class="timeline-metric-grid mt-5 progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(studentTimelineSource).toContain(
      'class="timeline-metric-card progress-card metric-panel-card"'
    )
    expect(studentTimelineSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(studentTimelineSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(studentTimelineSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(studentTimelineSource).not.toContain('teacher-surface-section')
    expect(studentTimelineSource).toMatch(
      /\.timeline-metric-grid\.metric-panel-default-surface\s*\{[\s\S]*--metric-panel-background:\s*radial-gradient\(/s
    )
  })

  it('student overview 当前排名卡片应切换到 shared metric-panel 卡片栈，而不是继续复用本地 note 边框样式', () => {
    expect(studentOverviewSource).toContain(
      'class="journal-rank-summary mt-5 progress-card metric-panel-card metric-panel-default-surface"'
    )
    expect(studentOverviewSource).toContain(
      'class="journal-rank-summary__label progress-card-label metric-panel-label"'
    )
    expect(studentOverviewSource).toContain(
      'class="journal-rank-summary__value progress-card-value metric-panel-value tech-font"'
    )
    expect(studentOverviewSource).toContain(
      'class="journal-rank-summary__helper progress-card-hint metric-panel-helper"'
    )
    expect(studentOverviewSource).toMatch(
      /\.journal-inline-item\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s
    )
    expect(studentOverviewSource).not.toMatch(
      /\.journal-metric,\s*\.journal-inline-item,\s*\.journal-rank-summary\s*\{/s
    )
  })

  it('student overview 竞技表现统计卡片应复用 shared metric-panel 样式栈，而不是继续使用本地 journal-metric 表面', () => {
    expect(studentOverviewSource).toContain(
      "'story-metric-grid mt-6 progress-strip metric-panel-grid metric-panel-default-surface'"
    )
    expect(studentOverviewSource).toContain(':class="storyMetricGridClass"')
    expect(studentOverviewSource).toContain(
      'class="journal-metric progress-card metric-panel-card"'
    )
    expect(studentOverviewSource).toContain('class="progress-card-label metric-panel-label"')
    expect(studentOverviewSource).toContain('class="progress-card-value metric-panel-value"')
    expect(studentOverviewSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(studentOverviewSource).not.toMatch(/\.journal-metric-accent\s*\{/)
    expect(studentOverviewSource).not.toMatch(
      /\.journal-metric,\s*\.journal-inline-item\s*\{[\s\S]*border:\s*1px solid var\(--journal-shell-border\);/s
    )
  })

  it('student recommendation 应切换到行动优先布局和 shared metric-panel 摘要卡片栈', () => {
    expect(studentRecommendationSource).toContain('现在先练这几道')
    expect(studentRecommendationSource).toContain('当前目标难度')
    expect(studentRecommendationSource).toContain('浏览全部题目')
    expect(studentRecommendationSource).toContain(
      'class="recommendation-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(studentRecommendationSource).toContain(
      'class="recommendation-summary-card progress-card metric-panel-card"'
    )
    expect(studentRecommendationSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(studentRecommendationSource).not.toContain('Top Queue')
    expect(studentRecommendationSource).not.toContain('Full List')
    expect(studentRecommendationSource).not.toContain('推荐摘要')
    expect(studentRecommendationSource).not.toContain('训练动作目录')
    expect(studentRecommendationSource).not.toContain('为什么先做这些')
  })

  it('student category progress 应切换到 shared metric-panel 摘要卡片栈和行动列表，而不是继续保留强弱高亮双卡', () => {
    expect(studentCategoryProgressSource).toContain('优先补这个分类')
    expect(studentCategoryProgressSource).toContain(
      'class="category-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(studentCategoryProgressSource).toContain(
      'class="category-summary-card progress-card metric-panel-card"'
    )
    expect(studentCategoryProgressSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(studentCategoryProgressSource).toContain(
      ':data-test="`category-action-${item.category}`"'
    )
    expect(studentCategoryProgressSource).not.toContain('category-highlight')
    expect(studentCategoryProgressSource).not.toContain('Strongest Direction')
    expect(studentCategoryProgressSource).not.toContain('Weakest Direction')
    expect(studentCategoryProgressSource).not.toContain('Action Directory')
  })

  it('student difficulty 应切换到强度推进工作区和 shared metric-panel 摘要卡片栈，而不是继续保留说明型双栏结构', () => {
    expect(studentDifficultySource).toContain('先推这一档强度')
    expect(studentDifficultySource).toContain(
      'class="difficulty-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(studentDifficultySource).toContain(
      'class="difficulty-summary-card progress-card metric-panel-card"'
    )
    expect(studentDifficultySource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(studentDifficultySource).toContain(':data-test="`difficulty-action-${item.difficulty}`"')
    expect(studentDifficultySource).not.toContain('difficulty-insight-list')
    expect(studentDifficultySource).not.toContain('难度层级总览')
    expect(studentDifficultySource).not.toContain('训练解读')
    expect(studentDifficultySource).not.toContain('Action Directory')
    expect(studentDifficultySource).not.toContain('为什么现在先推这一档')
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
