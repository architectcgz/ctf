import { resolveAwdReviewIndexRouteName } from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function awdReviewIndexRoute(role: TeachingWorkspaceRole) {
  return {
    name: resolveAwdReviewIndexRouteName(role),
  } as const
}
