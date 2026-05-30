import { describe, expect, it } from 'vitest'

import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import challengeDetailSource from '../platform/ChallengeDetail.vue?raw'
import challengeManageSource from '@/features/platform/challenges/ui/ChallengeManagePage.vue?raw'
import challengeImportManageSource from '@/pages/platform/challenges/ChallengeImportManageRoutePage.vue?raw'
import challengePackageFormatSource from '@/pages/platform/challenges/ChallengePackageFormatRoutePage.vue?raw'
import cheatDetectionWorkspaceSource from '../../components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'
import adminDashboardSourceBase from '../../features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import contestOrchestrationSource from '../../features/platform/contests/ui/ContestOrchestrationPage.vue?raw'
import userGovernancePageSource from '../../features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'

const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')

const adminDashboardSource = [
  adminDashboardSourceBase,
  platformOverviewHeroPanelSource,
  platformOverviewAlertsSectionSource,
  platformOverviewHotspotsSectionSource,
].join('\n')

describe('admin full-bleed hero roots', () => {
  it('uses a section root that carries the hero background', () => {
    const sources = [
      auditLogSource,
      challengeDetailSource,
      challengeManageSource,
      challengeImportManageSource,
      challengePackageFormatSource,
      cheatDetectionWorkspaceSource,
      imageManageSource,
      adminDashboardSource,
      contestOrchestrationSource,
      userGovernanceSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      const hasSectionHeroRoot =
        /<section[\s\S]*?class="[^"]*workspace-shell[^"]*journal-shell[^"]*journal-hero[^"]*"/s.test(
          source
        ) ||
        /<section[\s\S]*?class="[^"]*journal-shell[^"]*journal-hero[^"]*workspace-shell[^"]*"/s.test(
          source
        )
      const hasWorkspaceShellRoot = /<div[\s\S]*?class="[^"]*workspace-shell[^"]*"/s.test(source)

      expect(hasSectionHeroRoot || hasWorkspaceShellRoot).toBe(true)
    }
  })
})
