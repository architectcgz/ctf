import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'

describe('admin dashboard surface alignment', () => {
  it('softens the hero primary action border and focus ring to match the dark surface system', () => {
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*>\s*\.ui-btn,\s*\.workspace-alert-actions\s*>\s*\.ui-btn\s*\{[\s\S]*--ui-btn-height:\s*2\.75rem;[\s\S]*--ui-btn-focus-ring:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 16%, transparent\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*>\s*\.ui-btn\.ui-btn--primary\s*\{[\s\S]*--ui-btn-primary-hover-shadow:\s*0 12px 24px color-mix\(in srgb,\s*var\(--journal-accent\) 24%, transparent\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*>\s*\.ui-btn\.ui-btn--primary\s*\{[\s\S]*--ui-btn-primary-border:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 46%, var\(--journal-border\)\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*>\s*\.ui-btn\.ui-btn--ghost,\s*\.workspace-alert-actions\s*>\s*\.ui-btn\.ui-btn--ghost\s*\{[\s\S]*--ui-btn-border:\s*var\(--journal-border\);[\s\S]*--ui-btn-background:\s*color-mix\(in srgb,\s*var\(--journal-surface\) 94%, transparent\);/s
    )
  })

  it('frames the overview hero action rail as a compact operational panel', () => {
    expect(adminDashboardSource).toContain('class="overview-hero-actions"')
    expect(adminDashboardSource).toContain('class="overview-action-grid"')
    expect(adminDashboardSource).toMatch(
      /\.overview-hero-actions\s*\{[\s\S]*border:\s*1px solid var\(--workspace-line-soft\);[\s\S]*border-radius:\s*var\(--workspace-radius-lg\);[\s\S]*background:/s
    )
    expect(adminDashboardSource).toMatch(
      /\.hero-meta-badge\s*\{[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\);/s
    )
  })

  it('uses shared admin summary cards and directory shells instead of page-private premium panels', () => {
    expect(adminDashboardSource).toContain(
      'class="admin-summary-grid overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(adminDashboardSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(adminDashboardSource).toContain('class="workspace-directory-section overview-directory-section"')
    expect(adminDashboardSource).toContain('class="workspace-directory-list overview-list-shell"')
    expect(adminDashboardSource).not.toContain('metric-panel-card--premium')
    expect(adminDashboardSource).not.toContain('metric-panel-grid--premium')
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
