import { describe, expect, it } from 'vitest'

import appEmptySource from '../AppEmpty.vue?raw'

describe('AppEmpty surface styling', () => {
  it('uses softened component-level empty-state borders instead of raw border-border-subtle utility lines', () => {
    expect(appEmptySource).toContain('class="app-empty')
    expect(appEmptySource).not.toContain('border-y border-border-subtle')
    expect(appEmptySource).toMatch(
      /\.app-empty\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--color-border-subtle\) 74%, transparent\);[\s\S]*border-bottom:\s*1px solid color-mix\(in srgb,\s*var\(--color-border-subtle\) 74%, transparent\);/s,
    )
  })
})
