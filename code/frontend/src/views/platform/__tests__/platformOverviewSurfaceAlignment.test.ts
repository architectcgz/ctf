import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'

describe('admin dashboard surface alignment', () => {
  it('softens the hero primary action border and focus ring to match the dark surface system', () => {
    expect(adminDashboardSource).toMatch(
      /\.quick-actions\s*>\s*\.ui-btn\s*\{[\s\S]*--ui-btn-height:\s*2\.75rem;[\s\S]*--ui-btn-focus-ring:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 16%, transparent\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.quick-actions\s*>\s*\.ui-btn\.ui-btn--primary\s*\{[\s\S]*--ui-btn-primary-hover-shadow:\s*0 12px 24px color-mix\(in srgb,\s*var\(--journal-accent\) 24%, transparent\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.quick-actions\s*>\s*\.ui-btn\.ui-btn--primary\s*\{[\s\S]*--ui-btn-primary-border:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 46%, var\(--journal-border\)\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.quick-actions\s*>\s*\.ui-btn\.ui-btn--ghost\s*\{[\s\S]*--ui-btn-border:\s*var\(--journal-border\);[\s\S]*--ui-btn-background:\s*color-mix\(in srgb,\s*var\(--journal-surface\) 94%, transparent\);/s
    )
  })

  it('softens the alert action rows instead of leaving a bright neutral outline on dark surfaces', () => {
    expect(adminDashboardSource).toMatch(
      /--admin-action-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 72%, transparent\);/
    )
    expect(adminDashboardSource).toMatch(
      /\.admin-action-row\s*\{[\s\S]*border:\s*1px solid var\(--admin-action-border\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.admin-action-row:focus-visible\s*\{[\s\S]*outline:\s*none;[\s\S]*0 0 0 3px color-mix\(in srgb,\s*var\(--journal-accent\) 12%, transparent\)/s
    )
  })

  it('uses list-heading for alerts and hotspot section headers instead of workspace-tab-heading', () => {
    expect(adminDashboardSource).toContain(
      '<h2 class="section-title list-heading__title">当前告警</h2>'
    )
    expect(adminDashboardSource).toContain(
      '<h2 class="section-title list-heading__title">资源热点</h2>'
    )
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">当前告警</h2>')
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">资源热点</h2>')
  })
})
