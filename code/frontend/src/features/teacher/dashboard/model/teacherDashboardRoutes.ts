import {
  resolveClassManagementRouteName,
  resolveClassReviewRouteName,
  resolveStudentAnalysisRouteName,
} from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function teacherClassManagementRoute(role?: TeachingWorkspaceRole) {
  return {
    name: resolveClassManagementRouteName(role),
  } as const
}

export function teacherStudentAnalysisRoute(
  role: TeachingWorkspaceRole,
  studentId: string,
  className: string
) {
  return {
    name: resolveStudentAnalysisRouteName(role),
    params: {
      className,
      studentId,
    },
  } as const
}

export function teacherClassReviewRoute(role: TeachingWorkspaceRole, className: string) {
  return {
    name: resolveClassReviewRouteName(role),
    params: {
      className,
    },
  } as const
}
