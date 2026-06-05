import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '@/pages/platform/CheatDetectionRoutePage.vue?raw'
import cheatDetectionHeroPanelSource from '@/features/platform/cheat-detection/ui/CheatDetectionHeroPanel.vue?raw'
import cheatDetectionReviewPanelsSource from '@/features/platform/cheat-detection/ui/CheatDetectionReviewPanels.vue?raw'
import cheatDetectionSummaryPanelSource from '@/features/platform/cheat-detection/ui/CheatDetectionSummaryPanel.vue?raw'
import cheatDetectionWorkspacePanelSource from '@/features/platform/cheat-detection/ui/CheatDetectionWorkspacePanel.vue?raw'

const journalNotesSource = readFileSync(
  resolve(process.cwd(), 'src/assets/styles/journal-notes.css'),
  'utf8'
)

const cheatDetectionCombinedSource = [
  cheatDetectionSource,
  cheatDetectionWorkspacePanelSource,
  cheatDetectionHeroPanelSource,
  cheatDetectionReviewPanelsSource,
  cheatDetectionSummaryPanelSource,
].join('\n')

describe('cheat detection surface alignment', () => {
  it('keeps the route page as a thin entry and delegates the workspace owner to the feature module', () => {
    expect(cheatDetectionSource).toContain('CheatDetectionWorkspacePanel')
    expect(cheatDetectionSource).toContain('useCheatDetectionPage')
    expect(cheatDetectionSource).not.toContain('workspace-shell')
    expect(cheatDetectionSource).not.toContain('list-heading__title')
  })

  it('uses shared workspace header and summary panel owners for the overview surface', () => {
    expect(cheatDetectionHeroPanelSource).toContain('<header class="workspace-page-header">')
    expect(cheatDetectionHeroPanelSource).toContain('<CheatDetectionSummaryPanel')
    expect(cheatDetectionSummaryPanelSource).toContain(
      'class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(cheatDetectionSummaryPanelSource).not.toMatch(/^\.metric-panel-(?:default|workspace)-surface\s*\{/m)
    expect(journalNotesSource).toContain('.metric-panel-workspace-surface {')
  })

  it('keeps review sections on shared directory and empty-state owners', () => {
    expect(cheatDetectionReviewPanelsSource).toContain(
      'class="workspace-directory-section cheat-directory-section"'
    )
    expect(cheatDetectionReviewPanelsSource).toContain(
      '<h2 class="list-heading__title">高频提交账号</h2>'
    )
    expect(cheatDetectionReviewPanelsSource).toContain(
      '<h2 class="list-heading__title">共享 IP 线索</h2>'
    )
    expect(cheatDetectionReviewPanelsSource).toContain(
      '<h2 class="list-heading__title">审计联动</h2>'
    )
    expect(cheatDetectionReviewPanelsSource).toContain('<AppEmpty')
    expect(cheatDetectionReviewPanelsSource).toContain('class="cheat-empty-state"')
    expect(cheatDetectionReviewPanelsSource).toContain('class="cheat-directory-list"')
    expect(cheatDetectionReviewPanelsSource).toContain('class="quick-action-directory"')
    expect(cheatDetectionReviewPanelsSource).not.toContain('workspace-tab-heading__title')
  })

  it('does not locally redefine shared tab rails or metric-panel primitives', () => {
    expect(cheatDetectionCombinedSource).not.toMatch(/\.top-tabs\s*\{/)
    expect(cheatDetectionCombinedSource).not.toMatch(/\.top-tab\s*\{/)
    expect(cheatDetectionCombinedSource).not.toContain('<nav class="top-tabs"')
    expect(cheatDetectionCombinedSource).not.toMatch(/^\.metric-panel-grid\s*\{/m)
  })
})
