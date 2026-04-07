import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentOverviewEditorialSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'

const journalSoftSurfacesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-soft-surfaces.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('student journal shared button styles', () => {
  it('应该在共享样式文件中声明 soft journal 的 primary 与 outline 按钮规则', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-primary')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-outline')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-primary:hover')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-outline:hover')
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-primary:focus-visible'
    )
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-outline:focus-visible'
    )
  })

  it('student dashboard 页面不应继续在 scoped style 中重复声明基础按钮样式', () => {
    for (const source of [
      studentCategoryProgressSource,
      studentOverviewEditorialSource,
      studentRecommendationSource,
    ]) {
      const scopedStyle = extractScopedStyle(source)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary,\s*$/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary:focus-visible,\s*$/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline:focus-visible\s*\{/m)
    }
  })
})
