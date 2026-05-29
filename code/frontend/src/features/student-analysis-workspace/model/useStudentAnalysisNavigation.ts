import type { Ref } from 'vue'

import {
  resolveClassStudentsRouteName,
  resolveStudentReviewArchiveRouteName,
} from '@/utils/teachingWorkspaceRouting'

interface UseStudentAnalysisNavigationOptions {
  getRole: () => string | undefined
  selectedClassName: Ref<string>
  selectedStudentId: Ref<string>
  openClassStudentsRoute: (routeName: string, className: string) => void
  openChallengeRoute: (challengeId: string) => void
  openReviewArchiveRoute: (routeName: string, className: string, studentId: string) => void
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
    openClassStudentsRoute(resolveClassStudentsRouteName(getRole()), selectedClassName.value)
  }

  function openChallenge(challengeId: string): void {
    openChallengeRoute(challengeId)
  }

  function openReviewArchivePage(): void {
    if (!selectedStudentId.value || !selectedClassName.value) return
    openReviewArchiveRoute(
      resolveStudentReviewArchiveRouteName(getRole()),
      selectedClassName.value,
      selectedStudentId.value
    )
  }

  return {
    openClassStudents,
    openChallenge,
    openReviewArchivePage,
  }
}
