import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeDetailWidgetSource from '@/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import challengeWorkspaceShellSource from '@/features/challenge-detail/ui/ChallengeWorkspaceShell.vue?raw'
import challengeDetailPageSource from '@/features/challenge-detail/model/useChallengeDetailPage.ts?raw'
import challengeDetailPresentationSource from '@/features/challenge-detail/model/useChallengeDetailPresentation.ts?raw'
import challengeDirectoryPanelSource from '@/features/challenge-list/ui/ChallengeDirectoryPanel.vue?raw'
import challengeQuestionPanelSource from '@/features/challenge-detail/ui/ChallengeQuestionPanel.vue?raw'
import challengeSolutionsPanelSource from '@/features/challenge-detail/ui/ChallengeSolutionsPanel.vue?raw'
import challengeSubmissionRecordsPanelSource from '@/features/challenge-detail/ui/ChallengeSubmissionRecordsPanel.vue?raw'
import challengeWriteupPanelSource from '@/features/challenge-detail/ui/ChallengeWriteupPanel.vue?raw'
import challengeActionAsideSource from '@/features/challenge-detail/ui/ChallengeActionAside.vue?raw'

const challengeDetailWorkspaceSource = [
  challengeDetailSource,
  challengeDetailWidgetSource,
  challengeWorkspaceShellSource,
  challengeSolutionsPanelSource,
  challengeSubmissionRecordsPanelSource,
  challengeWriteupPanelSource,
  challengeActionAsideSource,
].join('\n')

describe('challenge page ui strategy', () => {
  it('challenge detail route should stay as a thin page shell that delegates to feature-owned workspace sections', () => {
    expect(challengeDetailSource).toContain(
      "import { useChallengeDetailPage } from '@/features/challenge-detail'"
    )
    expect(challengeDetailSource).toContain(
      "import { ChallengeDetailWorkspace } from '@/widgets/challenge-detail-workspace'"
    )
    expect(challengeDetailSource).toContain('<ChallengeDetailWorkspace')
    expect(challengeDetailSource).not.toContain('<ChallengeWorkspaceShell')
    expect(challengeDetailSource).not.toContain("from '@/api/challenge'")

    expect(challengeDetailWidgetSource).toContain('<ChallengeWorkspaceShell')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeQuestionPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSolutionsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSubmissionRecordsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeWriteupPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeActionAside')
    expect(challengeQuestionPanelSource).toContain("ChallengeMetaStrip } from '@/entities/challenge'")
    expect(challengeQuestionPanelSource).toContain('<ChallengeMetaStrip :challenge="challenge" />')
  })

  it('challenge detail should keep shared shell, tab, button, and semantic status primitives instead of page-local variants', () => {
    expect(challengeDetailWorkspaceSource).toContain(
      'class="journal-shell journal-shell-user journal-hero workspace-shell min-h-full"'
    )
    expect(challengeDetailWorkspaceSource).toContain('class="workspace-tabbar top-tabs"')
    expect(challengeDetailWorkspaceSource).toContain('class="workspace-tab top-tab"')
    expect(challengeDetailWorkspaceSource).toContain('class="solution-tabbar top-tabs challenge-subtabs"')
    expect(challengeDetailWorkspaceSource).toContain('class="solution-tab top-tab challenge-subtab"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-control challenge-input"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-btn ui-btn--primary')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(challengeDetailPresentationSource).toContain('flag-input-wrap--success')
    expect(challengeDetailPresentationSource).not.toContain('border-[var(--color-success)]')
    expect(challengeDetailPresentationSource).not.toContain('bg-[var(--color-success)]')
  })

  it('challenge detail state ownership should keep workspace panel query in the page model while moving workspace tab keyboard logic to the widget', () => {
    expect(challengeDetailPageSource).toContain(
      "import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'"
    )
    expect(challengeDetailPageSource).not.toContain(
      "import { useUrlSyncedTabs } from '@/shared/model/navigation/useUrlSyncedTabs'"
    )
    expect(challengeDetailPageSource).toContain('useRouteQueryTabs<WorkspaceTab>({')
    expect(challengeDetailPageSource).toContain(
      "import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'"
    )
    expect(challengeDetailPageSource).toContain('useTabKeyboardNavigation<ChallengeSolutionTab>({')
    expect(challengeDetailWidgetSource).toContain(
      "import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'"
    )
    expect(challengeDetailWidgetSource).toContain('useTabKeyboardNavigation<WorkspaceTab>({')
    expect(challengeDetailSource).toContain('function setSelectedSolutionId(value: string | null): void {')
    expect(challengeDetailSource).not.toContain('function selectWorkspaceTabFromWidget(')
    expect(challengeDetailSource).not.toContain('function selectSolutionTabFromWidget(')
    expect(challengeDetailSource).not.toContain('function focusTab(id: string): void {')
    expect(challengeDetailSource).not.toContain('function handleSolutionTabKeydown(')
    expect(challengeDetailPresentationSource).not.toContain('function handleSolutionTabKeydown(')
    expect(challengeDetailPresentationSource).not.toContain('focusTab: (tabId: string) => void')
  })

  it('challenge list route should stay as a thin page shell that delegates directory rendering to shared challenge UI', () => {
    expect(challengeListSource).toContain(
      "import { ChallengeDirectoryPanel, useChallengeListPage } from '@/features/challenge-list'"
    )
    expect(challengeListSource).toContain('<ChallengeDirectoryPanel')
    expect(challengeDirectoryPanelSource).toContain(
      "import { ChallengeDirectoryRow } from '@/entities/challenge'"
    )
    expect(challengeDirectoryPanelSource).toContain('<ChallengeDirectoryRow')
    expect(challengeDirectoryPanelSource).not.toContain('class="challenge-row"')
  })
})
