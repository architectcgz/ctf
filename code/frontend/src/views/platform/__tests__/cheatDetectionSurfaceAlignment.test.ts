import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '../CheatDetection.vue?raw'
import cheatDetectionHeroPanelSource from '@/components/platform/cheat/CheatDetectionHeroPanel.vue?raw'
import cheatDetectionReviewPanelsSource from '@/components/platform/cheat/CheatDetectionReviewPanels.vue?raw'
import cheatDetectionSummaryPanelSource from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'

const journalNotesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-notes.css`,
  'utf-8'
)
const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')
const cheatDetectionCombinedSource = [
  cheatDetectionSource,
  cheatDetectionWorkspacePanelSource,
  cheatDetectionHeroPanelSource,
  cheatDetectionReviewPanelsSource,
  cheatDetectionSummaryPanelSource,
].join('\n')

describe('cheat detection surface alignment', () => {
  it('softens risk card borders instead of using the full journal border contrast', () => {
    expect(cheatDetectionCombinedSource).toMatch(
      /--cheat-card-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/
    )
    expect(cheatDetectionCombinedSource).toMatch(
      /\.cheat-directory-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--cheat-card-border\);/s
    )
    expect(cheatDetectionCombinedSource).not.toMatch(
      /\.cheat-directory-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--journal-border\);/s
    )
  })

  it('uses a softer section divider so the content bands do not look boxed in', () => {
    expect(cheatDetectionCombinedSource).toMatch(
      /--cheat-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/
    )
    expect(cheatDetectionCombinedSource).toContain(
      '--journal-divider-border: 1px dashed var(--cheat-divider);'
    )
    expect(journalNotesSource).toMatch(
      /\.journal-shell-admin \.journal-divider\s*\{[\s\S]*border-top:\s*var\(/s
    )
    expect(cheatDetectionCombinedSource).not.toMatch(/\.journal-divider\s*\{/s)
  })

  it('uses shared admin shell accent and empty-state border tokens', () => {
    expect(cheatDetectionCombinedSource).toContain(
      '--journal-shell-dark-accent: var(--color-primary-hover);'
    )
    expect(cheatDetectionCombinedSource).toMatch(
      /\.admin-empty\s*\{[\s\S]*border:\s*1px dashed color-mix\(in srgb,\s*var\(--journal-border\) 72%, transparent\);/s
    )
    expect(cheatDetectionCombinedSource).not.toContain('--journal-shell-dark-accent: #60a5fa;')
    expect(cheatDetectionCombinedSource).not.toContain('border: 1px dashed rgba(148, 163, 184, 0.72);')
  })

  it('overrides the empty state top and bottom borders instead of inheriting AppEmpty bright border-y lines', () => {
    expect(cheatDetectionCombinedSource).toMatch(
      /<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有超过阈值的高频提交账号"/s
    )
    expect(cheatDetectionCombinedSource).toMatch(
      /<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有共享 IP 线索"/s
    )
    expect(cheatDetectionCombinedSource).toMatch(
      /\.cheat-empty-state\s*\{[\s\S]*border-top-color:\s*var\(--cheat-divider\);[\s\S]*border-bottom-color:\s*var\(--cheat-divider\);/s
    )
  })

  it('aligns the tab rail with the teacher dashboard underline style instead of pill tabs', () => {
    expect(pageTabsSource).toMatch(
      /\.top-tabs\s*\{[\s\S]*overflow-x:\s*auto;[\s\S]*scrollbar-width:\s*none;/s
    )
    expect(pageTabsSource).toMatch(
      /\.top-tab\s*\{[\s\S]*border-bottom:\s*2px solid transparent;[\s\S]*white-space:\s*nowrap;/s
    )
    expect(pageTabsSource).toContain('--page-top-tab-active-border,')
    expect(pageTabsSource).toMatch(
      /\.top-tab:hover,\s*\.top-tab.active,\s*\.top-tab:focus-visible\s*\{[\s\S]*border-bottom-color:\s*var\(/s
    )
    expect(cheatDetectionCombinedSource).toContain('--page-top-tabs-gap: var(--space-7);')
    expect(cheatDetectionCombinedSource).toContain('--page-top-tab-font-size: var(--font-size-15);')
    expect(cheatDetectionCombinedSource).toContain('--page-top-tab-active-border:')
    expect(cheatDetectionCombinedSource).not.toMatch(/\.top-tabs\s*\{[^}]*display:\s*flex;/s)
    expect(cheatDetectionCombinedSource).not.toMatch(
      /\.top-tab\s*\{[^}]*border-bottom:\s*2px solid transparent;/s
    )
  })

  it('uses shared admin summary-grid styles for cheat overview metric panels', () => {
    expect(journalNotesSource).toContain(
      '.journal-shell-admin :is(.admin-summary-grid, .manage-summary-grid, .image-summary-grid)'
    )
    expect(cheatDetectionCombinedSource).toContain('class="workspace-shell')
    expect(cheatDetectionHeroPanelSource).toContain('<section class="workspace-hero">')
    expect(cheatDetectionHeroPanelSource).toContain('<CheatDetectionSummaryPanel')
    expect(cheatDetectionCombinedSource).toContain(
      'class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(cheatDetectionCombinedSource).not.toContain('class="admin-summary-grid cheat-risk-summary')
    expect(cheatDetectionCombinedSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(cheatDetectionCombinedSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(cheatDetectionCombinedSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(cheatDetectionCombinedSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(cheatDetectionCombinedSource).toContain('--workspace-brand: var(--journal-accent);')
    expect(cheatDetectionCombinedSource).toContain(
      '--workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));'
    )
    expect(cheatDetectionCombinedSource).toContain(
      '--workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));'
    )
    expect(cheatDetectionCombinedSource).toContain(
      '--workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));'
    )
    expect(cheatDetectionCombinedSource).toContain(
      '--workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);'
    )
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
    expect(journalNotesSource).toContain(
      'var(--workspace-brand, var(--journal-accent, var(--color-primary-default))) 18%'
    )
    expect(journalNotesSource).toContain('--workspace-panel-soft,')
    expect(journalNotesSource).toContain('--metric-panel-label-color: color-mix(')
    expect(journalNotesSource).toContain('.progress-card.metric-panel-card .metric-panel-label > :is(svg, .lucide) {')
    expect(journalNotesSource).toContain('padding-inline-end: var(--space-7);')
    expect(journalNotesSource).toContain('top: var(--space-3-5);')
    expect(journalNotesSource).toContain('right: var(--space-4);')
    expect(cheatDetectionCombinedSource).not.toMatch(
      /\.cheat-kpi-summary\.metric-panel-default-surface\.metric-panel-workspace-surface\s*\{/s
    )
    expect(cheatDetectionCombinedSource).not.toContain('class="mt-5 grid gap-3 sm:grid-cols-2"')
    expect(cheatDetectionCombinedSource).not.toContain('class="grid gap-3 md:grid-cols-3"')
  })

  it('uses workspace directory sections and flat rows instead of stacked cards for integrity review flows', () => {
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">高频提交账号</h2>')
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">共享 IP 线索</h2>')
    expect(cheatDetectionCombinedSource).toContain('<h2 class="list-heading__title">审计联动</h2>')
    expect(cheatDetectionCombinedSource).toContain(
      'class="workspace-directory-section cheat-directory-section"'
    )
    expect(cheatDetectionCombinedSource).toContain('class="cheat-directory-list"')
    expect(cheatDetectionCombinedSource).toContain('class="cheat-directory-row"')
    expect(cheatDetectionCombinedSource).toContain('class="cheat-directory-row-meta"')
    expect(cheatDetectionCombinedSource).toContain('class="quick-action-directory"')
    expect(cheatDetectionCombinedSource).toContain('class="quick-action-row"')
    expect(cheatDetectionCombinedSource).not.toContain('class="tab-panel space-y-3"')
    expect(cheatDetectionCombinedSource).not.toContain('class="grid gap-3 lg:grid-cols-2"')
  })
})
