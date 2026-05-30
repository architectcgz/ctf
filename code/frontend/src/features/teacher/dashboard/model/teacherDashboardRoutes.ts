import { resolveClassManagementRouteName } from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function teacherClassManagementRoute(role?: TeachingWorkspaceRole) {
  return {
    name: resolveClassManagementRouteName(role),
  } as const
}
