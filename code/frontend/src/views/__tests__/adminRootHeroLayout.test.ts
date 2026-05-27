import { describe, expect, it } from 'vitest'

import auditLogSource from '../platform/AuditLog.vue?raw'
import challengeDetailSource from '../platform/ChallengeDetail.vue?raw'
import challengeManageSource from '../platform/ChallengeManage.vue?raw'
import challengeImportManageSource from '../platform/ChallengeImportManage.vue?raw'
import challengePackageFormatSource from '../platform/ChallengePackageFormat.vue?raw'
import cheatDetectionWorkspaceSource from '../../components/platform/cheat/CheatDetectionWorkspacePanel.vue?raw'
import imageManageSource from '../platform/ImageManage.vue?raw'
import adminDashboardSourceBase from '../../features/platform-overview/ui/PlatformOverviewPage.vue?raw'
import contestOrchestrationSource from '../../components/platform/contest/ContestOrchestrationPage.vue?raw'
import userGovernancePageSource from '../../components/platform/user/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/components/platform/user/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/components/platform/user/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/components/platform/user/UserGovernanceImportPanel.vue?raw'
import platformOverviewAlertsSectionSource from '@/components/platform/dashboard/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/components/platform/dashboard/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/components/platform/dashboard/PlatformOverviewHotspotsSection.vue?raw'

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
