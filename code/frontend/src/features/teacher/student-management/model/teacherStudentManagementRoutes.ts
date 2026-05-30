import {
  resolveClassManagementRouteName,
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
