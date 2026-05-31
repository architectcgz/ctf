import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import contestOperationsHubSource from '@/pages/platform/contests/ContestOperationsHubRoutePage.vue?raw'
import contestOperationsHubHeroPanelSource from '@/features/platform/contests/ui/ContestOperationsHubHeroPanel.vue?raw'
import contestOperationsHubWorkspacePanelSource from '@/features/platform/contests/ui/ContestOperationsHubWorkspacePanel.vue?raw'
import contestOrchestrationSource from '@/features/platform/contests/ui/ContestOrchestrationPage.vue?raw'
import adminContestFormPanelSource from '@/features/platform/contests/ui/PlatformContestFormPanel.vue?raw'
import contestFormActionsSource from '@/features/platform/contests/ui/PlatformContestFormActions.vue?raw'
import contestFormIdentitySectionSource from '@/features/platform/contests/ui/PlatformContestIdentitySection.vue?raw'
import contestFormRulesSectionSource from '@/features/platform/contests/ui/PlatformContestRulesSection.vue?raw'
import contestFormSectionShellSource from '@/features/platform/contests/ui/PlatformContestFormSectionShell.vue?raw'
import contestFormTimelineSectionSource from '@/features/platform/contests/ui/PlatformContestTimelineSection.vue?raw'
import adminContestTableSource from '@/features/platform/contests/ui/PlatformContestTable.vue?raw'
import contestEditSource from '@/pages/platform/contests/ContestEditRoutePage.vue?raw'
import contestEditTopbarPanelSource from '@/features/platform/contests/ui/ContestEditTopbarPanel.vue?raw'
import contestEditWorkspacePanelSource from '@/features/platform/contests/ui/ContestEditWorkspacePanel.vue?raw'
import contestAwdPreflightPanelSource from '@/features/platform/contests/ui/ContestAwdPreflightPanel.vue?raw'
import awdChallengeConfigPanelSource from '@/features/platform/contests/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contests/ui/AWDChallengeConfigHeader.vue?raw'
import awdContestSelectorFieldSource from '@/features/contest-awd-admin/ui/AWDContestSelectorField.vue?raw'
import awdOperationsPanelSource from '@/features/contest-awd-admin/ui/AWDOperationsPanel.vue?raw'
import awdOperationsPreRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue?raw'
import awdRuntimePendingStateSource from '@/features/contest-awd-admin/ui/AWDRuntimePendingState.vue?raw'
import awdRoundCreateDialogSourceBase from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdRoundCreateScoreSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateScoreSection.vue?raw'
import awdRoundCreateSettingsSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateSettingsSection.vue?raw'
import awdAttackLogDialogSourceBase from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdAttackLogDetailsSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue?raw'
import awdAttackLogTargetSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue?raw'
import awdServiceCheckDialogSourceBase from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdServiceCheckResultSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue?raw'
import awdServiceCheckTargetSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue?raw'
import awdOperationsDialogFooterSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue?raw'
import contestChallengeEditorDialogSourceBase from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import contestChallengeSettingsSectionSource from '@/features/contest-workbench/ui/ContestChallengeSettingsSection.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue?raw'
import contestChallengeDirectorySectionSource from '@/features/contest-workbench/ui/ContestChallengeDirectorySection.vue?raw'
import contestChallengeOrchestrationHeaderSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationHeader.vue?raw'
import contestChallengeSummaryStripSource from '@/features/contest-workbench/ui/ContestChallengeSummaryStrip.vue?raw'
import contestWorkbenchSummaryStripSource from '@/features/contest-workbench/ui/ContestWorkbenchSummaryStrip.vue?raw'
import contestAwdChallengeSelectorSectionSource from '@/features/contest-workbench/ui/ContestAwdChallengeSelectorSection.vue?raw'
import awdReadinessChecklistSource from '@/features/awd-readiness/ui/AWDReadinessChecklist.vue?raw'
import awdReadinessOverrideDialogSource from '@/features/awd-readiness/ui/AWDReadinessOverrideDialog.vue?raw'
import awdReadinessSummarySource from '@/features/awd-readiness/ui/AWDReadinessSummary.vue?raw'
import awdReadinessDecisionHUDSource from '@/features/awd-readiness/ui/AWDReadinessDecisionHUD.vue?raw'
import awdRoundHeaderPanelSource from '@/features/awd-inspector/ui/AWDRoundHeaderPanel.vue?raw'
import awdRoundInspectorSourceBase from '@/features/awd-inspector/ui/AWDRoundInspector.vue?raw'
import awdInspectorStatsHudSource from '@/features/awd-inspector/ui/AWDInspectorStatsHud.vue?raw'
import awdInspectorCanvasWorkspaceSource from '@/features/awd-inspector/ui/AWDInspectorCanvasWorkspace.vue?raw'
import awdTrafficPanelSourceBase from '@/features/awd-inspector/ui/AWDTrafficPanel.vue?raw'
import awdTrafficSummaryBandSource from '@/features/awd-inspector/ui/AWDTrafficSummaryBand.vue?raw'
import awdTrafficIntelligenceGridSource from '@/features/awd-inspector/ui/AWDTrafficIntelligenceGrid.vue?raw'
import awdTrafficEventTableSource from '@/features/awd-inspector/ui/AWDTrafficEventTable.vue?raw'

const contestOperationsHubCombinedSource = [
  contestOperationsHubSource,
  contestOperationsHubHeroPanelSource,
  contestOperationsHubWorkspacePanelSource,
].join('\n')
const contestFormCombinedSource = [
  adminContestFormPanelSource,
  contestFormSectionShellSource,
  contestFormIdentitySectionSource,
  contestFormRulesSectionSource,
  contestFormTimelineSectionSource,
  contestFormActionsSource,
].join('\n')
const contestEditCombinedSource = [contestEditSource, contestEditWorkspacePanelSource].join('\n')
const contestChallengeOrchestrationCombinedSource = [
  contestChallengeOrchestrationPanelSource,
  contestChallengeOrchestrationHeaderSource,
  contestChallengeDirectorySectionSource,
].join('\n')
const awdChallengeConfigCombinedSource = [
  awdChallengeConfigPanelSource,
  awdChallengeConfigHeaderSource,
  awdChallengeConfigDirectorySectionSource,
  awdChallengeConfigDirectoryRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/platform/contests/ui/awdChallengeConfigPanel.css'),
    'utf8'
  ),
].join('\n')
const awdRoundCreateDialogSource = [
  awdRoundCreateDialogSourceBase,
  awdRoundCreateSettingsSectionSource,
  awdRoundCreateScoreSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')
const awdAttackLogDialogSource = [
  awdAttackLogDialogSourceBase,
  awdAttackLogTargetSectionSource,
  awdAttackLogDetailsSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')
const awdServiceCheckDialogSource = [
  awdServiceCheckDialogSourceBase,
  awdServiceCheckTargetSectionSource,
  awdServiceCheckResultSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')
const contestChallengeEditorDialogSource = [
  contestChallengeEditorDialogSourceBase,
  contestAwdChallengeSelectorSectionSource,
  contestChallengeSettingsSectionSource,
].join('\n')
const awdRoundInspectorSource = [
  awdRoundInspectorSourceBase,
  awdInspectorStatsHudSource,
  awdInspectorCanvasWorkspaceSource,
  readFileSync(resolve(process.cwd(), 'src/features/awd-inspector/ui/awdRoundInspector.css'), 'utf8'),
].join('\n')
const awdTrafficPanelSource = [
  awdTrafficPanelSourceBase,
  awdTrafficSummaryBandSource,
  awdTrafficIntelligenceGridSource,
  awdTrafficEventTableSource,
  readFileSync(resolve(process.cwd(), 'src/features/awd-inspector/ui/awdTrafficPanel.css'), 'utf8'),
].join('\n')

describe('contest admin ui strategy', () => {
  it('contest admin workspace pages should consume shared button, field, panel header, and action menu primitives', () => {
    expect(contestOperationsHubHeroPanelSource).toContain('class="header-btn header-btn--ghost"')
    expect(contestOperationsHubCombinedSource).toContain('class="ui-btn ui-btn--primary ui-btn--sm"')
    expect(contestOperationsHubHeroPanelSource).toContain('<header class="workspace-panel-header contest-ops-hero">')
    expect(contestOperationsHubHeroPanelSource).toContain(
      'class="workspace-panel-header__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
    expect(contestOperationsHubCombinedSource).toContain('<WorkspaceDataTable')
    expect(contestOperationsHubCombinedSource).toContain('contestTableColumns')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestOrchestrationSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestOrchestrationSource).toContain('class="ui-field contest-filter-field"')
    expect(contestOrchestrationSource).toContain(
      "import { useUrlSyncedTabs } from '@/shared/model/navigation/useUrlSyncedTabs'"
    )
    expect(contestOrchestrationSource).toContain('useUrlSyncedTabs<ContestPanelKey>(')
    expect(contestOrchestrationSource).toContain('<header class="list-heading contest-create-head">')
    expect(contestOrchestrationSource).toContain(
      '<header class="workspace-panel-header contest-overview-head">'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function resolvePanelFromLocation(): ContestPanelKey {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function syncPanelToLocation(panelKey: ContestPanelKey): void {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function focusTabByIndex(index: number): void {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
    expect(contestFormCombinedSource).toContain('class="ui-field contest-form-field')
    expect(contestFormCombinedSource).toContain('class="ui-control-wrap')
    expect(contestFormCombinedSource).toContain('class="ui-btn ui-btn--secondary')
    expect(contestFormCombinedSource).toContain('class="ui-btn ui-btn--primary')
    expect(adminContestTableSource).toContain('class="ui-badge contest-status-pill')
    expect(adminContestTableSource).toContain('class="ui-row-actions contest-table__actions')
    expect(adminContestTableSource).toContain("from '@/shared/ui/common/menus/CActionMenu.vue'")
    expect(adminContestTableSource).toContain('aria-label="更多竞赛操作"')
    expect(adminContestTableSource).not.toContain('<Teleport to="body">')
    expect(contestEditCombinedSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestEditCombinedSource).toContain('min-height: 0;')
  })

  it('contest admin dialogs and awd runtime surfaces should keep shared field and action primitives with extracted section owners', () => {
    expect(awdRoundCreateDialogSource).toContain('class="ui-field')
    expect(awdRoundCreateDialogSource).toContain('class="ui-control-wrap')
    expect(awdRoundCreateDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdRoundCreateDialogSource).toContain('class="ui-btn ui-btn--primary')
    expect(awdServiceCheckDialogSource).toContain('class="ui-field')
    expect(awdServiceCheckDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdServiceCheckDialogSource).toContain('class="ui-btn ui-btn--primary')
    expect(awdAttackLogDialogSource).toContain('class="ui-field')
    expect(awdAttackLogDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(awdAttackLogDialogSource).toContain('class="ui-btn ui-btn--primary')
    expect(contestChallengeEditorDialogSource).toContain('class="ui-field contest-challenge-dialog__field')
    expect(contestChallengeEditorDialogSource).toContain(
      'class="ui-btn ui-btn--secondary contest-challenge-dialog__button'
    )
    expect(contestChallengeEditorDialogSource).toContain(
      'class="ui-btn ui-btn--primary contest-challenge-dialog__button'
    )
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsPreRuntimeStage')
    expect(awdContestSelectorFieldSource).toContain('class="ui-field awd-ops-selector-field"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdRuntimePendingStateSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdRoundInspectorSource).toContain('<AWDRoundHeaderPanel')
    expect(awdRoundInspectorSource).toContain('<AWDTrafficPanel')
    expect(awdRoundInspectorSource).toContain('class="ui-btn ui-btn--secondary awd-inspector-export-button"')
    expect(awdTrafficPanelSource).toContain('class="ui-field awd-round-filter-field"')
    expect(awdTrafficPanelSource).toContain('class="ui-control-wrap awd-round-filter-control"')
    expect(awdTrafficPanelSource).toContain('class="ui-btn ui-btn--ghost awd-round-filter-search"')
    expect(contestAwdPreflightPanelSource).toContain('class="ui-btn ui-btn--primary"')
  })

  it('contest admin metric panels and workspace overlines should follow the shared summary contract instead of historical local variants', () => {
    expect(contestOrchestrationSource).not.toContain('journal-eyebrow-text')
    expect(contestOperationsHubSource).not.toContain('journal-eyebrow-text')
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdChallengeConfigCombinedSource).toMatch(
      /<div class="workspace-overline">\s*AWD Service Config\s*<\/div>/
    )
    expect(awdChallengeConfigCombinedSource).not.toContain('<div class="journal-eyebrow">AWD Service Config</div>')
    expect(awdReadinessChecklistSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-summary-grid"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-override-summary"'
    )
    expect(awdReadinessDecisionHUDSource).toContain(
      'class="decision-hud progress-card metric-panel-card metric-panel-default-surface"'
    )
    expect(awdReadinessSummarySource).toMatch(
      /<div class="workspace-overline">\s*AWD Readiness\s*<\/div>/
    )
    expect(awdReadinessSummarySource).not.toContain('<div class="journal-eyebrow">AWD Readiness</div>')
    expect(contestWorkbenchSummaryStripSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-workbench-summary-strip"'
    )
    expect(contestChallengeSummaryStripSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary"'
    )
  })

  it('contest admin surfaces should keep current removal constraints instead of reintroducing deprecated blocks', () => {
    expect(contestAwdPreflightPanelSource).not.toContain('contest-awd-preflight-force-start')
    expect(contestAwdPreflightPanelSource).not.toContain('Override Entry')
    expect(awdChallengeConfigCombinedSource).not.toContain('config-focus-card')
    expect(awdChallengeConfigCombinedSource).not.toContain('当前焦点题目')
    expect(awdReadinessSummarySource).toContain('<header class="list-heading readiness-decision__head">')
    expect(awdReadinessSummarySource).not.toContain('Start Decision')
    expect(awdOperationsPreRuntimeStageSource).toContain('name="pending"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--neutral"')
    expect(awdRoundHeaderPanelSource).toContain('class="ops-btn ops-btn--primary"')
    expect(contestChallengeOrchestrationCombinedSource).toContain(
      "from '@/shared/ui/common/menus/CActionMenu.vue'"
    )
    expect(contestChallengeOrchestrationCombinedSource).toContain(
      'class="ui-row-actions contest-challenge-row__actions"'
    )
    expect(contestChallengeOrchestrationCombinedSource).not.toContain('class="admin-btn')
  })

  it('contest route pages should keep feature-owned announcement and edit workspaces instead of rebuilding page-local shells', () => {
    expect(contestEditSource).not.toContain('useContestEditAwdWorkspace')
    expect(contestEditSource).toContain('useContestEditPage')
    expect(contestEditSource).toContain('<ContestEditTopbarPanel')
    expect(contestEditSource).toContain('<ContestEditWorkspacePanel')
    expect(contestEditTopbarPanelSource).toContain('Contest Studio')
    expect(contestEditTopbarPanelSource).toContain('class="studio-edit-label"')
    expect(contestEditTopbarPanelSource).toContain('class="studio-contest-heading"')
    expect(contestEditTopbarPanelSource).toContain(
      'padding: var(--space-4) var(--space-workspace-side-padding) 0;'
    )
    expect(contestEditTopbarPanelSource).toContain('contest-open-announcements')
    expect(contestEditTopbarPanelSource).toContain('保存变更')
    expect(contestEditWorkspacePanelSource).not.toContain('AWDOperationsPanel')
    expect(contestEditWorkspacePanelSource).not.toContain('operation-panel="inspector"')
    expect(contestEditWorkspacePanelSource).not.toContain('runtime-content="round-inspector"')
    expect(contestEditWorkspacePanelSource).not.toContain('operation-panel="instances"')
    expect(contestEditWorkspacePanelSource).not.toContain('runtime-content="instances"')
    expect(contestEditWorkspacePanelSource).toContain('<Transition')
    expect(contestEditWorkspacePanelSource).toContain('name="studio-stage"')
    expect(contestEditWorkspacePanelSource).toContain('mode="out-in"')
    expect(contestEditWorkspacePanelSource).toContain('class="studio-pane studio-stage-panel"')
    expect(contestEditWorkspacePanelSource).toContain('@media (prefers-reduced-motion: reduce)')
    expect(contestEditWorkspacePanelSource).not.toContain('class="studio-pane fade-in"')
    expect(contestEditWorkspacePanelSource).not.toContain('@keyframes studioFadeIn')
  })
})
