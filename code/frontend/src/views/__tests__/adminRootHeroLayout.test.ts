import { describe, expect, it } from 'vitest'

import auditLogSource from '../admin/AuditLog.vue?raw'
import challengeDetailSource from '../admin/ChallengeDetail.vue?raw'
import cheatDetectionSource from '../admin/CheatDetection.vue?raw'
import adminDashboardSource from '../../components/admin/dashboard/AdminDashboardPage.vue?raw'
import contestOrchestrationSource from '../../components/admin/contest/ContestOrchestrationPage.vue?raw'
import userGovernanceSource from '../../components/admin/user/UserGovernancePage.vue?raw'

describe('admin full-bleed hero roots', () => {
  it('uses a section root that carries the hero background', () => {
    const sources = [
      auditLogSource,
      challengeDetailSource,
      cheatDetectionSource,
      adminDashboardSource,
      contestOrchestrationSource,
      userGovernanceSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      expect(source).toMatch(/<section class="journal-shell[^"]*journal-hero[^"]*min-h-full/s)
    }
  })
})
