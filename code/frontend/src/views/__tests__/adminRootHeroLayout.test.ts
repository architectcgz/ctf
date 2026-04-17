import { describe, expect, it } from 'vitest'

import auditLogSource from '../admin/AuditLog.vue?raw'
import challengeDetailSource from '../admin/ChallengeDetail.vue?raw'
import challengeManageSource from '../admin/ChallengeManage.vue?raw'
import challengeImportManageSource from '../admin/ChallengeImportManage.vue?raw'
import challengePackageFormatSource from '../admin/ChallengePackageFormat.vue?raw'
import cheatDetectionSource from '../admin/CheatDetection.vue?raw'
import imageManageSource from '../admin/ImageManage.vue?raw'
import adminDashboardSource from '../../components/admin/dashboard/AdminDashboardPage.vue?raw'
import contestOrchestrationSource from '../../components/admin/contest/ContestOrchestrationPage.vue?raw'
import userGovernanceSource from '../../components/admin/user/UserGovernancePage.vue?raw'

describe('admin full-bleed hero roots', () => {
  it('uses a section root that carries the hero background', () => {
    const sources = [
      auditLogSource,
      challengeDetailSource,
      challengeManageSource,
      challengeImportManageSource,
      challengePackageFormatSource,
      cheatDetectionSource,
      imageManageSource,
      adminDashboardSource,
      contestOrchestrationSource,
      userGovernanceSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      const hasSectionHeroRoot =
        /<section[\s\S]*?class="[^"]*journal-shell[^"]*journal-hero[^"]*min-h-full[^"]*flex-1[^"]*"/s.test(
          source
        )
      const hasWorkspaceShellRoot = /<div[\s\S]*?class="[^"]*workspace-shell[^"]*"/s.test(source)

      expect(hasSectionHeroRoot || hasWorkspaceShellRoot).toBe(true)
    }
  })
})
