import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'
import contestOperationsHubHeroPanelSource from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue?raw'
import contestOperationsHubWorkspacePanelSource from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue?raw'
import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'
import adminContestFormPanelSource from '@/components/platform/contest/PlatformContestFormPanel.vue?raw'
import adminContestTableSource from '@/components/platform/contest/PlatformContestTable.vue?raw'

const contestOperationsHubCombinedSource = [
  contestOperationsHubSource,
  contestOperationsHubHeroPanelSource,
  contestOperationsHubWorkspacePanelSource,
].join('\n')

describe('contest ui primitive adoption', () => {
  it('contest workspace pages should consume shared ui button and control primitives', () => {
    expect(contestOperationsHubCombinedSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOperationsHubCombinedSource).toContain('class="ui-btn ui-btn--primary ui-btn--sm"')
    expect(contestOperationsHubCombinedSource).toContain('class="contest-ops-hero__actions"')
    expect(contestOperationsHubCombinedSource).toContain('class="contest-ops-actions"')
    expect(contestOperationsHubHeroPanelSource).toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(contestOperationsHubHeroPanelSource).not.toContain('--metric-panel-columns: 4;')
    expect(contestOperationsHubHeroPanelSource).toContain('<Trophy class="h-4 w-4" />')
    expect(contestOperationsHubCombinedSource).toContain(
      'class="workspace-directory-list contest-ops-table"'
    )
    expect(contestOperationsHubCombinedSource).toContain('<WorkspaceDataTable')
    expect(contestOperationsHubCombinedSource).toContain('contestTableColumns')
    expect(contestOperationsHubCombinedSource).not.toContain('class="contest-ops-row"')
    expect(contestOperationsHubCombinedSource).not.toContain('contest-ops-card')
    expect(contestOperationsHubSource).toContain('class="content-pane contest-ops-content"')
    expect(contestOperationsHubCombinedSource).toContain(
      'gap: var(--workspace-directory-page-block-gap, var(--space-5));'
    )
    expect(contestOperationsHubCombinedSource).toContain('padding: 0;')
    expect(contestOperationsHubCombinedSource).not.toContain('padding: 1.5rem;')
    expect(contestOperationsHubCombinedSource).not.toContain('gap: 1rem;')

    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestOrchestrationSource).toContain('class="ui-field contest-filter-field"')
    expect(contestOrchestrationSource).toContain('class="ui-control-wrap"')
    expect(contestOrchestrationSource).toContain('class="ui-control contest-filter-control"')
  })

  it('contest form should consume shared field and button primitives', () => {
    expect(adminContestFormPanelSource).toContain('class="ui-field contest-form-field')
    expect(adminContestFormPanelSource).toContain('class="ui-control-wrap')
    expect(adminContestFormPanelSource).toContain('class="ui-control"')
    expect(adminContestFormPanelSource).toContain('class="ui-btn ui-btn--secondary')
    expect(adminContestFormPanelSource).toContain('class="ui-btn ui-btn--primary')
  })

  it('contest directory rows should consume shared badge and row action primitives', () => {
    expect(adminContestTableSource).toContain('class="ui-badge contest-status-pill')
    expect(adminContestTableSource).toContain('class="ui-row-actions contest-table__actions')
    expect(adminContestTableSource).toContain('class="ui-btn ui-btn--sm ui-btn--primary')
    expect(adminContestTableSource).toContain('<WorkspaceDataTable')
    expect(adminContestTableSource).toContain('contestTableColumns')
    expect(adminContestTableSource).toContain("from '@/components/common/menus/CActionMenu.vue'")
    expect(adminContestTableSource).toContain('class="c-action-menu__trigger c-action-menu__trigger--icon')
    expect(adminContestTableSource).toContain('aria-label="更多竞赛操作"')
    expect(adminContestTableSource).not.toContain('<Teleport to="body">')
    expect(adminContestTableSource).toContain('contest-action--workbench')
    expect(adminContestTableSource).not.toContain('class="contest-action contest-action--primary"')
    expect(adminContestTableSource).not.toContain('class="contest-action contest-action--ghost"')
  })

  it('contest directory buttons should inherit theme tokens instead of page-local color overrides', () => {
    expect(contestOrchestrationSource).toContain('--ui-btn-primary-background: var(--journal-accent);')
    expect(contestOrchestrationSource).toContain(
      '--ui-btn-primary-hover-background: var(--color-primary-hover);'
    )
    expect(contestOrchestrationSource).toContain(
      '--ui-btn-ghost-color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));'
    )
    expect(contestOrchestrationSource).toContain('--action-menu-accent: var(--journal-accent);')

    expect(adminContestTableSource).not.toContain('--ui-btn-primary-bg:')
    expect(adminContestTableSource).not.toMatch(
      /\.contest-action--workbench\s*\{[\s\S]*var\(--color-success\)[\s\S]*\}/
    )
  })
})
