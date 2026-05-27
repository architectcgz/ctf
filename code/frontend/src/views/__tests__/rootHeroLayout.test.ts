import { describe, expect, it } from 'vitest'

import instanceListWorkspaceShellSource from '@/components/instance/InstanceListWorkspaceShell.vue?raw'
import challengeListSource from '../challenges/ChallengeList.vue?raw'
import contestListSource from '../contests/ContestList.vue?raw'
import instanceListSource from '../instances/InstanceList.vue?raw'
import notificationListSource from '../notifications/NotificationList.vue?raw'
import securitySettingsSource from '../profile/SecuritySettings.vue?raw'
import securitySettingsWorkspaceShellSource from '@/components/profile/SecuritySettingsWorkspaceShell.vue?raw'
import skillProfileSource from '../profile/SkillProfile.vue?raw'
import skillProfileWorkspaceShellSource from '@/components/profile/SkillProfileWorkspaceShell.vue?raw'
import userProfileSource from '../profile/UserProfile.vue?raw'
import userProfileWorkspaceShellSource from '@/components/profile/UserProfileWorkspaceShell.vue?raw'
import scoreboardViewSource from '../scoreboard/ScoreboardView.vue?raw'
import categoryProgressSource from '../../features/student-dashboard/ui/StudentCategoryProgressPage.vue?raw'
import difficultyPageSource from '../../features/student-dashboard/ui/StudentDifficultyPage.vue?raw'
import overviewPageSource from '../../components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import recommendationPageSource from '../../features/student-dashboard/ui/StudentRecommendationPage.vue?raw'
import trainingTimelineSource from '../../components/training/TrainingTimelinePanel.vue?raw'

describe('full-bleed hero roots', () => {
  it('uses a section root that carries the hero background', () => {
    const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
    const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
    const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`
    const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`
    const sources = [
      challengeListSource,
      contestListSource,
      instanceListWorkspaceSource,
      notificationListSource,
      scoreboardViewSource,
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
      recommendationPageSource,
      categoryProgressSource,
      trainingTimelineSource,
      difficultyPageSource,
      overviewPageSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      const hasDirectHeroRoot =
        /<section[\s\S]*?class="[^"]*journal-shell[^"]*journal-hero[^"]*flex[^"]*min-h-full[^"]*flex-1[^"]*flex-col[^"]*"/s.test(
          source
        )
      const hasEmbeddableHeroRoot =
        source.includes('embedded?: boolean') &&
        /'(?:(?:workspace-shell )?journal-shell(?: journal-shell-user)? journal-hero)/.test(
          source
        ) &&
        /<section[\s\S]*?class="[^"]*journal-soft-surface[^"]*flex[^"]*min-h-full[^"]*flex-1[^"]*flex-col[^"]*"/s.test(
          source
        )

      expect(hasDirectHeroRoot || hasEmbeddableHeroRoot).toBe(true)
    }
  })
})
