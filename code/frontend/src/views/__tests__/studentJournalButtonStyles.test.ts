import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
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
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-secondary')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-outline')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-primary:hover')
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-secondary:hover'
    )
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-btn-outline:hover')
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-primary:focus-visible'
    )
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-secondary:focus-visible'
    )
    expect(journalSoftSurfacesSource).toContain(
      '.journal-soft-surface .journal-btn-outline:focus-visible'
    )
    expect(journalSoftSurfacesSource).toContain(
      'color-mix(in srgb, var(--journal-accent) 50%, var(--journal-control-border))'
    )
    expect(journalSoftSurfacesSource).toContain(
      'color-mix(in srgb, var(--journal-accent) 66%, var(--journal-control-border))'
    )
    expect(journalSoftSurfacesSource).not.toContain(
      'color-mix(in srgb, var(--journal-accent) 50%, transparent)'
    )
    expect(journalSoftSurfacesSource).not.toContain(
      'color-mix(in srgb, var(--journal-accent) 66%, transparent)'
    )
    expect(journalSoftSurfacesSource).toContain('--journal-soft-button-outline-hover-border')
    expect(journalSoftSurfacesSource).toMatch(
      /\[data-theme='dark'\] \.journal-soft-surface \.journal-shell\s*\{[\s\S]*--journal-soft-button-primary-hover-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-accent\) 42%,\s*var\(--journal-control-border\)\s*\);/s
    )
    expect(journalSoftSurfacesSource).toMatch(
      /\[data-theme='dark'\] \.journal-soft-surface \.journal-shell\s*\{[\s\S]*--journal-soft-button-secondary-hover-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-accent\) 28%,\s*var\(--journal-control-border\)\s*\);/s
    )
    expect(journalSoftSurfacesSource).toMatch(
      /\[data-theme='dark'\] \.journal-soft-surface \.journal-shell\s*\{[\s\S]*--journal-soft-button-outline-hover-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-accent\) 28%,\s*var\(--journal-control-border\)\s*\);/s
    )
  })

  it('student dashboard 页面不应继续在 scoped style 中重复声明基础按钮样式', () => {
    for (const source of [
      studentCategoryProgressSource,
      studentDifficultySource,
      studentOverviewEditorialSource,
      studentRecommendationSource,
    ]) {
      const scopedStyle = extractScopedStyle(source)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary,\s*$/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-secondary\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-secondary:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline:hover\s*\{/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-primary:focus-visible,\s*$/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-secondary:focus-visible,\s*$/m)
      expect(scopedStyle).not.toMatch(/^\.journal-btn-outline:focus-visible\s*\{/m)
    }
  })

  it('student dashboard 页面覆写 primary 按钮边框时也必须保留可见边框参照', () => {
    expect(studentRecommendationSource).toContain(
      '--journal-soft-button-primary-border: color-mix('
    )
    expect(studentRecommendationSource).toContain('var(--journal-control-border)')
    expect(studentRecommendationSource).not.toContain(
      '--journal-soft-button-primary-border: color-mix(in srgb, var(--journal-accent) 42%, transparent)'
    )
  })
})
