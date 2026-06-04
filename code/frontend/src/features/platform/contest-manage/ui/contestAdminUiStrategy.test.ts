import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '@/pages/platform/contests/ContestOperationsHubRoutePage.vue?raw'
import contestOrchestrationSource from '@/features/platform/contest-manage/ui/ContestOrchestrationPage.vue?raw'
import contestEditSource from '@/pages/platform/contests/ContestEditRoutePage.vue?raw'
import contestEditTopbarPanelSource from '@/features/platform/contest-manage/ui/ContestEditTopbarPanel.vue?raw'
import contestEditWorkspacePanelSource from '@/features/platform/contest-manage/ui/ContestEditWorkspacePanel.vue?raw'
import platformContestEditPageSource from '@/features/platform/contest-manage/ui/PlatformContestEditPage.vue?raw'
import contestAwdPreflightPanelSource from '@/features/platform/contest-manage/ui/ContestAwdPreflightPanel.vue?raw'
import awdChallengeConfigPanelSource from '@/features/platform/contest-manage/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contest-manage/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contest-manage/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contest-manage/ui/AWDChallengeConfigHeader.vue?raw'
import awdOperationsPanelSource from '@/features/contest-awd-admin/ui/AWDOperationsPanel.vue?raw'
import awdOperationsPreRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue?raw'
import contestChallengeDirectorySectionSource from '@/features/contest-workbench/ui/ContestChallengeDirectorySection.vue?raw'
import contestChallengeOrchestrationHeaderSource from '@/features/contest-workbench/ui/ContestChallengeOrchestrationHeader.vue?raw'
import awdReadinessSummarySource from '@/features/awd-readiness/ui/AWDReadinessSummary.vue?raw'

const awdChallengeConfigCombinedSource = [
  awdChallengeConfigPanelSource,
  awdChallengeConfigHeaderSource,
  awdChallengeConfigDirectorySectionSource,
  awdChallengeConfigDirectoryRowSource,
].join('\n')

const contestChallengeOrchestrationCombinedSource = [
  contestChallengeOrchestrationPanelSource,
  contestChallengeOrchestrationHeaderSource,
  contestChallengeDirectorySectionSource,
].join('\n')

describe('contest admin ui strategy', () => {
  it('contest orchestration should keep panel ownership local instead of restoring URL-synced tabs', () => {
    expect(contestOrchestrationSource).toContain('activePanel: ContestManagePanelKey')
    expect(contestOrchestrationSource).toContain('switchPanel: [panel: ContestManagePanelKey]')
    expect(contestOrchestrationSource).toContain('<ContestManageOverviewPanel')
    expect(contestOrchestrationSource).toContain('<ContestManageCreatePanel')
    expect(contestOrchestrationSource).not.toContain(
      "import { useUrlSyncedTabs } from '@/shared/model/navigation/useUrlSyncedTabs'"
    )
    expect(contestOrchestrationSource).not.toContain('useUrlSyncedTabs<ContestPanelKey>(')
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
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })

  it('contest edit route should keep feature-owned announcement and edit workspaces', () => {
    expect(contestEditSource).not.toContain('useContestEditAwdWorkspace')
    expect(contestEditSource).toContain('PlatformContestEditPage')
    expect(platformContestEditPageSource).toContain('useContestEditPage')
    expect(platformContestEditPageSource).toContain('<ContestEditTopbarPanel')
    expect(platformContestEditPageSource).toContain('<ContestEditWorkspacePanel')
    expect(contestEditTopbarPanelSource).toContain('contest-open-announcements')
    expect(contestEditTopbarPanelSource).toContain('保存变更')
    expect(contestEditWorkspacePanelSource).not.toContain('AWDOperationsPanel')
    expect(contestEditWorkspacePanelSource).not.toContain('operation-panel="inspector"')
    expect(contestEditWorkspacePanelSource).not.toContain('runtime-content="round-inspector"')
    expect(contestEditWorkspacePanelSource).not.toContain('operation-panel="instances"')
    expect(contestEditWorkspacePanelSource).not.toContain('runtime-content="instances"')
  })

  it('contest AWD surfaces should keep deprecated focus and force-start blocks out', () => {
    expect(contestAwdPreflightPanelSource).not.toContain('contest-awd-preflight-force-start')
    expect(contestAwdPreflightPanelSource).not.toContain('Override Entry')
    expect(awdChallengeConfigCombinedSource).not.toContain('config-focus-card')
    expect(awdChallengeConfigCombinedSource).not.toContain('当前焦点题目')
    expect(awdReadinessSummarySource).not.toContain('Start Decision')
  })

  it('contest AWD runtime and challenge orchestration should stay split into explicit owners', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsPreRuntimeStage')
    expect(awdOperationsPreRuntimeStageSource).toContain('name="pending"')
    expect(contestChallengeOrchestrationCombinedSource).toContain(
      "from '@/shared/ui/common/menus/CActionMenu.vue'"
    )
    expect(contestChallengeOrchestrationCombinedSource).not.toContain('class="admin-btn')
    expect(contestOperationsHubSource).not.toContain('journal-eyebrow-text')
  })
})
