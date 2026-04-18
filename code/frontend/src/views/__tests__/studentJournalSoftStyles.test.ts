import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentOverviewEditorialSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'
import studentTimelineSource from '@/components/dashboard/student/StudentTimelinePage.vue?raw'

const journalSoftSurfacesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-soft-surfaces.css`,
  'utf-8'
)

describe('student journal soft shared styles', () => {
  it('应该在共享样式文件中声明 student soft journal 的 shell、note 与 eyebrow 规则', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-shell')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-shell-embedded')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-shell.journal-hero')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-note')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-note-label')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-note-value')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-eyebrow')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-eyebrow-soft')
    expect(journalSoftSurfacesSource).toContain(
      "[data-theme='dark'] .journal-soft-surface .journal-note"
    )
  })

  it('student journal 页面应通过根节点 class 接入共享 soft 样式，而不是继续本地重写基础规则', () => {
    for (const source of [
      studentCategoryProgressSource,
      studentDifficultySource,
      studentOverviewEditorialSource,
      studentRecommendationSource,
      studentTimelineSource,
    ]) {
      expect(source).toContain('journal-soft-surface')
      expect(source).not.toMatch(/^\.journal-eyebrow\s*\{/m)
      expect(source).not.toMatch(/^\.journal-shell-embedded,\s*$/m)
      expect(source).not.toMatch(/^\.journal-shell-embedded\s*\{/m)
    }

    for (const source of [
      studentCategoryProgressSource,
      studentDifficultySource,
      studentRecommendationSource,
      studentTimelineSource,
    ]) {
      expect(source).not.toMatch(/^\.journal-note\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-label\s*\{/m)
      expect(source).not.toMatch(/^\.journal-note-value\s*\{/m)
      expect(source).not.toMatch(/^:global\(\[data-theme='dark'\]\) \.journal-note\s*\{/m)
      expect(source).not.toMatch(/^:global\(\[data-theme='dark'\]\) \.journal-shell\s*\{/m)
      expect(source).not.toMatch(/^:global\(\[data-theme='dark'\]\) \.journal-hero\s*\{/m)
    }
  })

  it('student journal 页面应复用共享空态面板样式，而不是重复写裸圆角和主题色类', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-empty-state')

    for (const source of [
      studentCategoryProgressSource,
      studentDifficultySource,
      studentRecommendationSource,
      studentTimelineSource,
    ]) {
      expect(source).toContain('journal-soft-empty-state')
      expect(source).not.toContain(
        'rounded-[22px] border border-dashed border-[var(--journal-shell-border)]'
      )
    }
  })
})
