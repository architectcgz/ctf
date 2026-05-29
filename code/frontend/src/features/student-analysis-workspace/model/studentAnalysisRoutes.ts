import {
  resolveClassStudentsRouteName,
  resolveStudentReviewArchiveRouteName,
} from '@/utils/teachingWorkspaceRouting'

type TeachingWorkspaceRole = string | null | undefined

export function studentAnalysisClassStudentsRoute(
  role: TeachingWorkspaceRole,
  className: string
) {
  return {
    name: resolveClassStudentsRouteName(role),
    params: {
      className,
    },
  } as const
}

export function studentAnalysisChallengeDetailRoute(challengeId: string) {
  return {
    name: 'ChallengeDetail',
    params: {
      id: challengeId,
    },
  } as const
}

export function studentAnalysisReviewArchiveRoute(
  role: TeachingWorkspaceRole,
  className: string,
  studentId: string
) {
  return {
    name: resolveStudentReviewArchiveRouteName(role),
    params: {
      className,
      studentId,
    },
  } as const
}
