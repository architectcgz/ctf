import { resolveTeachingDashboardRouteName } from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function teacherInstanceDashboardRoute(role?: TeachingWorkspaceRole) {
  return {
    name: resolveTeachingDashboardRouteName(role),
  } as const
}
