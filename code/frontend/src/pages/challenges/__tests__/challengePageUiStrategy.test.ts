import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import challengeWorkspaceShellSource from '@/features/challenge-detail/ui/ChallengeWorkspaceShell.vue?raw'
import challengeDetailPageSource from '@/features/challenge-detail/model/useChallengeDetailPage.ts?raw'
import challengeDetailPresentationSource from '@/features/challenge-detail/model/useChallengeDetailPresentation.ts?raw'
import challengeDirectoryPanelSource from '@/components/challenge/ChallengeDirectoryPanel.vue?raw'
import challengeQuestionPanelSource from '@/components/challenge/ChallengeQuestionPanel.vue?raw'
import challengeSolutionsPanelSource from '@/features/challenge-detail/ui/ChallengeSolutionsPanel.vue?raw'
import challengeSubmissionRecordsPanelSource from '@/features/challenge-detail/ui/ChallengeSubmissionRecordsPanel.vue?raw'
import challengeWriteupPanelSource from '@/components/challenge/ChallengeWriteupPanel.vue?raw'
import challengeActionAsideSource from '@/components/challenge/ChallengeActionAside.vue?raw'

const challengeDetailWorkspaceSource = [
  challengeDetailSource,
  challengeWorkspaceShellSource,
  challengeSolutionsPanelSource,
  challengeSubmissionRecordsPanelSource,
  challengeWriteupPanelSource,
  challengeActionAsideSource,
].join('\n')

describe('challenge page ui strategy', () => {
  it('challenge detail route should stay as a thin page shell that delegates to feature-owned workspace sections', () => {
    expect(challengeDetailSource).toContain(
      "import { ChallengeWorkspaceShell, useChallengeDetailPage } from '@/features/challenge-detail'"
    )
    expect(challengeDetailSource).toContain('<ChallengeWorkspaceShell')
    expect(challengeDetailSource).not.toContain("from '@/api/challenge'")

    expect(challengeWorkspaceShellSource).toContain('<ChallengeQuestionPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSolutionsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSubmissionRecordsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeWriteupPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeActionAside')
    expect(challengeQuestionPanelSource).toContain("ChallengeMetaStrip } from '@/entities/challenge'")
    expect(challengeQuestionPanelSource).toContain('<ChallengeMetaStrip :challenge="challenge" />')
  })

  it('challenge detail should keep shared shell, tab, button, and semantic status primitives instead of page-local variants', () => {
    expect(challengeDetailSource).toContain(
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

  it('challenge detail state ownership should keep tab keyboard logic in the page model and keep presentation pure', () => {
    expect(challengeDetailPageSource).toContain(
      "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
    )
    expect(challengeDetailPageSource).toContain('useTabKeyboardNavigation<ChallengeSolutionTab>({')
    expect(challengeDetailSource).not.toContain('function focusTab(id: string): void {')
    expect(challengeDetailSource).not.toContain('function handleSolutionTabKeydown(')
    expect(challengeDetailPresentationSource).not.toContain('function handleSolutionTabKeydown(')
    expect(challengeDetailPresentationSource).not.toContain('focusTab: (tabId: string) => void')
  })

  it('challenge list route should stay as a thin page shell that delegates directory rendering to shared challenge UI', () => {
    expect(challengeListSource).toContain(
      "import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'"
    )
    expect(challengeListSource).toContain('<ChallengeDirectoryPanel')
    expect(challengeDirectoryPanelSource).toContain(
      "import { ChallengeDirectoryRow } from '@/entities/challenge'"
    )
    expect(challengeDirectoryPanelSource).toContain('<ChallengeDirectoryRow')
    expect(challengeDirectoryPanelSource).not.toContain('class="challenge-row"')
  })
})
