import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/admin/dashboard/AdminDashboardPage.vue?raw'

describe('admin dashboard surface alignment', () => {
  it('softens the hero primary action border and focus ring to match the dark surface system', () => {
    expect(adminDashboardSource).toMatch(
      /\.admin-btn\s*\{[\s\S]*border:\s*1px solid transparent;[\s\S]*box-shadow:\s*var\(--admin-btn-shadow,\s*none\);/s,
    )
    expect(adminDashboardSource).toMatch(
      /\.admin-btn:focus-visible\s*\{[\s\S]*outline:\s*none;[\s\S]*0 0 0 3px color-mix\(in srgb,\s*var\(--journal-accent\) 16%, transparent\)/s,
    )
    expect(adminDashboardSource).toMatch(
      /\.admin-btn-primary\s*\{[\s\S]*--admin-btn-shadow:\s*0 12px 24px color-mix\(in srgb,\s*var\(--journal-accent\) 24%, transparent\);/s,
    )
    expect(adminDashboardSource).toMatch(
      /\.admin-btn-primary\s*\{[\s\S]*border-color:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 46%, var\(--journal-border\)\);/s,
    )
  })
})
