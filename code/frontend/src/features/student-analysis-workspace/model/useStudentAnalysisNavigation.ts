import type { Ref } from 'vue'
import type { Router } from 'vue-router'

import {
  resolveClassStudentsRouteName,
  resolveStudentReviewArchiveRouteName,
} from '@/utils/teachingWorkspaceRouting'

interface UseStudentAnalysisNavigationOptions {
  router: Router
  getRole: () => string | undefined
  selectedClassName: Ref<string>
  selectedStudentId: Ref<string>
}

export function useStudentAnalysisNavigation(options: UseStudentAnalysisNavigationOptions) {
  const { router, getRole, selectedClassName, selectedStudentId } = options

  function openClassStudents(): void {
    router.push({
      name: resolveClassStudentsRouteName(getRole()),
      params: { className: selectedClassName.value },
    })
  }

  function openChallenge(challengeId: string): void {
    router.push(`/challenges/${challengeId}`)
  }

  function openReviewArchivePage(): void {
    if (!selectedStudentId.value || !selectedClassName.value) return
    router.push({
      name: resolveStudentReviewArchiveRouteName(getRole()),
      params: {
        className: selectedClassName.value,
        studentId: selectedStudentId.value,
      },
    })
  }

  return {
    openClassStudents,
    openChallenge,
    openReviewArchivePage,
  }
}
