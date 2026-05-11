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

  it('student difficulty 行动按钮应选择共享 secondary 变体，禁止页面私有按钮配色', () => {
    expect(studentDifficultySource).toContain("'journal-btn-primary'")
    expect(studentDifficultySource).toContain("'journal-btn-secondary'")
    expect(studentDifficultySource).not.toContain('difficulty-action-item__cta--secondary')
    expect(studentDifficultySource).not.toContain('--journal-soft-button-primary-border')
    expect(studentDifficultySource).not.toContain('--journal-soft-button-primary-background')
    expect(studentDifficultySource).not.toMatch(/\.difficulty-action-item__cta--secondary\s*\{/)
  })

  it('student overview 页面应把雷达图高度和紧凑空态收敛为语义类', () => {
    expect(studentOverviewEditorialSource).toContain('student-overview-radar-height')
    expect(studentOverviewEditorialSource).toContain('journal-soft-empty-state--compact')
    expect(studentOverviewEditorialSource).not.toContain('h-[18rem]')
    expect(studentOverviewEditorialSource).not.toContain('md:h-[21rem]')
    expect(studentOverviewEditorialSource).not.toContain('xl:h-[23rem]')
    expect(studentOverviewEditorialSource).not.toContain('rounded-[18px]')
  })

  it('student journal 页面应复用共享标题、正文与强调图标语义类，而不是继续写主题 utility', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-page-title')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-section-title')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-body-title')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-body-copy')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-meta')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-accent-icon')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-accent-pill')

    for (const source of [
      studentCategoryProgressSource,
      studentDifficultySource,
      studentOverviewEditorialSource,
      studentRecommendationSource,
      studentTimelineSource,
    ]) {
      expect(source).not.toContain('text-[var(--journal-ink)]')
    }

    for (const source of [
      studentCategoryProgressSource,
      studentOverviewEditorialSource,
      studentRecommendationSource,
      studentTimelineSource,
    ]) {
      expect(source).not.toContain('text-[var(--journal-muted)]')
    }

    for (const source of [studentOverviewEditorialSource, studentRecommendationSource]) {
      expect(source).not.toContain('text-[var(--journal-accent-strong)]')
    }

    expect(studentRecommendationSource).not.toContain('border-[var(--journal-accent)]/20')
    expect(studentRecommendationSource).not.toContain('bg-[var(--journal-accent)]/8')
  })
})
