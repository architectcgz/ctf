import type { Ref } from 'vue'
import type { Router } from 'vue-router'

import {
  resolveClassManagementRouteName,
  resolveClassStudentsRouteName,
  resolveStudentAnalysisRouteName,
  resolveStudentReviewArchiveRouteName,
} from '@/utils/teachingWorkspaceRouting'

interface UseTeacherStudentAnalysisNavigationOptions {
  router: Router
  getRole: () => string | undefined
  selectedClassName: Ref<string>
  selectedStudentId: Ref<string>
}

export function useTeacherStudentAnalysisNavigation(
  options: UseTeacherStudentAnalysisNavigationOptions
) {
  const { router, getRole, selectedClassName, selectedStudentId } = options

  function selectClass(className: string): void {
    router.push({
      name: resolveClassStudentsRouteName(getRole()),
      params: { className },
    })
  }

  function openClassManagement(): void {
    router.push({ name: resolveClassManagementRouteName(getRole()) })
  }

  function openClassStudents(): void {
    router.push({
      name: resolveClassStudentsRouteName(getRole()),
      params: { className: selectedClassName.value },
    })
  }

  function selectStudent(studentId: string): void {
    router.push({
      name: resolveStudentAnalysisRouteName(getRole()),
      params: {
        className: selectedClassName.value,
        studentId,
      },
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
    selectClass,
    openClassManagement,
    openClassStudents,
    selectStudent,
    openChallenge,
    openReviewArchivePage,
  }
}
