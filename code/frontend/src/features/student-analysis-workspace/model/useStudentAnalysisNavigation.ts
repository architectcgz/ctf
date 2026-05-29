import type { Ref } from 'vue'

import type { AppRouteTarget } from '@/components/navigation/routeTarget'
import {
  studentAnalysisChallengeDetailRoute,
  studentAnalysisClassStudentsRoute,
  studentAnalysisReviewArchiveRoute,
} from './studentAnalysisRoutes'

interface UseStudentAnalysisNavigationOptions {
  getRole: () => string | undefined
  selectedClassName: Ref<string>
  selectedStudentId: Ref<string>
  openClassStudentsRoute: (target: AppRouteTarget) => void
  openChallengeRoute: (target: AppRouteTarget) => void
  openReviewArchiveRoute: (target: AppRouteTarget) => void
}

export function useStudentAnalysisNavigation(options: UseStudentAnalysisNavigationOptions) {
  const {
    getRole,
    selectedClassName,
    selectedStudentId,
    openClassStudentsRoute,
    openChallengeRoute,
    openReviewArchiveRoute,
  } = options

  function openClassStudents(): void {
    openClassStudentsRoute(studentAnalysisClassStudentsRoute(getRole(), selectedClassName.value))
  }

  function openChallenge(challengeId: string): void {
    openChallengeRoute(studentAnalysisChallengeDetailRoute(challengeId))
  }

  function openReviewArchivePage(): void {
    if (!selectedStudentId.value || !selectedClassName.value) return
    openReviewArchiveRoute(
      studentAnalysisReviewArchiveRoute(
        getRole(),
        selectedClassName.value,
        selectedStudentId.value
      )
    )
  }

  return {
    openClassStudents,
    openChallenge,
    openReviewArchivePage,
  }
}
