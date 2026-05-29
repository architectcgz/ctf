import {
  resolveClassManagementRouteName,
  resolveStudentAnalysisRouteName,
  resolveTeachingDashboardRouteName,
} from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function classStudentsStudentAnalysisRoute(
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

export function classStudentsClassManagementRoute(role?: TeachingWorkspaceRole) {
  return {
    name: resolveClassManagementRouteName(role),
  } as const
}

export function classStudentsDashboardRoute(role?: TeachingWorkspaceRole) {
  return {
    name: resolveTeachingDashboardRouteName(role),
  } as const
}
