import type { ComputedRef, Ref } from 'vue'

import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import type { DashboardHighlightItem, DashboardPanelKey } from './studentDashboardTypes'

interface UseStudentDashboardPanelBindingsOptions {
  className: ComputedRef<string | undefined>
  progress: Ref<MyProgressData | null>
  timeline: Ref<TimelineEvent[]>
  recommendations: Ref<RecommendationItem[]>
  skillProfile: Ref<SkillProfileData | null>
  displayName: ComputedRef<string>
  weakDimensions: ComputedRef<string[]>
  categoryStats: ComputedRef<MyProgressData['category_stats']>
  difficultyStats: ComputedRef<MyProgressData['difficulty_stats']>
  completionRate: ComputedRef<number>
  highlightItems: ComputedRef<DashboardHighlightItem[]>
  openChallenge: (challengeId: string) => void
  openChallenges: () => void
  openCategoryChallenges: (category: string) => void
  openDifficultyChallenges: (difficulty: string) => void
  openSkillProfile: () => void
}

export function useStudentDashboardPanelBindings({
  className,
  progress,
  timeline,
  recommendations,
  skillProfile,
  displayName,
  weakDimensions,
  categoryStats,
  difficultyStats,
  completionRate,
  highlightItems,
  openChallenge,
  openChallenges,
  openCategoryChallenges,
  openDifficultyChallenges,
  openSkillProfile,
}: UseStudentDashboardPanelBindingsOptions) {
  function resolveDashboardPanelBindings(panelKey: DashboardPanelKey): Record<string, unknown> {
    switch (panelKey) {
      case 'overview':
        return {
          embedded: true,
          displayName: displayName.value,
          className: className.value,
          progress: progress.value,
          completionRate: completionRate.value,
          highlightItems: highlightItems.value,
          recommendations: recommendations.value,
          timeline: timeline.value,
          weakDimensions: weakDimensions.value,
          skillDimensions: skillProfile.value?.dimensions ?? [],
          onOpenChallenge: openChallenge,
          onOpenChallenges: openChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'recommendation':
        return {
          embedded: true,
          weakDimensions: weakDimensions.value,
          recommendations: recommendations.value,
          onOpenChallenge: openChallenge,
          onOpenChallenges: openChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'category':
        return {
          embedded: true,
          categoryStats: categoryStats.value,
          completionRate: completionRate.value,
          onOpenChallenges: openChallenges,
          onOpenCategoryChallenges: openCategoryChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'timeline':
        return {
          embedded: true,
          timeline: timeline.value,
        }
      case 'difficulty':
        return {
          embedded: true,
          difficultyStats: difficultyStats.value,
          onOpenChallenges: openChallenges,
          onOpenDifficultyChallenges: openDifficultyChallenges,
        }
    }
  }

  return {
    resolveDashboardPanelBindings,
  }
}
