import {
  resolveClassStudentsRouteName,
  resolveStudentAnalysisRouteName,
  resolveStudentManagementRouteName,
} from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function studentReviewArchiveAnalysisRoute(
  role: TeachingWorkspaceRole,
  className: string,
  studentId: string
) {
  return {
    name: resolveStudentAnalysisRouteName(role),
    params: {
      className,
      studentId,
    },
  } as const
}

export function studentReviewArchiveBackRoute(role: TeachingWorkspaceRole, className: string) {
  if (!className) {
    return {
      name: resolveStudentManagementRouteName(role),
    } as const
  }

  return {
    name: resolveClassStudentsRouteName(role),
    params: {
      className,
    },
  } as const
}
