import type { Ref } from 'vue'
import type { RouteLocationNormalizedLoadedGeneric, Router } from 'vue-router'

import type { AttackSessionQuery } from '@/api/teacher'

type ReviewRouteLike = Pick<RouteLocationNormalizedLoadedGeneric, 'query'>
type ReviewRouterLike = Pick<Router, 'replace'>

interface UseStudentAnalysisReviewQuerySyncOptions {
  route: ReviewRouteLike
  router: ReviewRouterLike
  sessionQuery: Ref<AttackSessionQuery>
  selectedStudentId: Ref<string>
  setSessionQuery: (nextQuery: Partial<AttackSessionQuery>) => void
  loadReviewWorkspace: (studentId: string) => Promise<void>
  reloadAttackSessions: (studentId: string) => Promise<void>
  studentIdFromRoute: () => string
}

export function useStudentAnalysisReviewQuerySync(
  options: UseStudentAnalysisReviewQuerySyncOptions
) {
  const {
    route,
    router,
    sessionQuery,
    selectedStudentId,
    setSessionQuery,
    loadReviewWorkspace,
    reloadAttackSessions,
    studentIdFromRoute,
  } = options

  function reviewWorkspaceQueryFromRoute(): Partial<AttackSessionQuery> {
    return {
      mode:
        route.query.reviewMode === 'practice' ||
        route.query.reviewMode === 'jeopardy' ||
        route.query.reviewMode === 'awd'
          ? route.query.reviewMode
          : undefined,
      result:
        route.query.reviewResult === 'success' ||
        route.query.reviewResult === 'failed' ||
        route.query.reviewResult === 'in_progress' ||
        route.query.reviewResult === 'unknown'
          ? route.query.reviewResult
          : undefined,
      challenge_id:
        typeof route.query.reviewChallengeId === 'string' && route.query.reviewChallengeId.trim()
          ? route.query.reviewChallengeId.trim()
          : undefined,
    }
  }

  function syncReviewWorkspaceQueryFromRoute(): void {
    const nextQuery = reviewWorkspaceQueryFromRoute()

    setSessionQuery({
      with_events: true,
      limit: 20,
      offset: 0,
      ...nextQuery,
    })
  }

  function reviewWorkspaceQueryMatchesState(
    nextQuery: Partial<AttackSessionQuery>
  ): boolean {
    return (
      (nextQuery.mode || undefined) === (sessionQuery.value.mode || undefined) &&
      (nextQuery.result || undefined) === (sessionQuery.value.result || undefined) &&
      (nextQuery.challenge_id || undefined) === (sessionQuery.value.challenge_id || undefined)
    )
  }

  async function updateReviewWorkspaceFilters(
    nextQuery: Partial<{
      challenge_id: string
      mode: 'practice' | 'jeopardy' | 'awd'
      result: 'success' | 'failed' | 'in_progress' | 'unknown'
    }>
  ): Promise<void> {
    const studentId = selectedStudentId.value || studentIdFromRoute()
    const mergedQuery = {
      ...sessionQuery.value,
      ...nextQuery,
      offset: 0,
    }

    setSessionQuery(mergedQuery)

    await router.replace({
      query: {
        ...route.query,
        reviewMode: mergedQuery.mode || undefined,
        reviewResult: mergedQuery.result || undefined,
        reviewChallengeId: mergedQuery.challenge_id || undefined,
      },
    })

    if (Object.prototype.hasOwnProperty.call(nextQuery, 'challenge_id')) {
      await loadReviewWorkspace(studentId)
      return
    }

    await reloadAttackSessions(studentId)
  }

  return {
    reviewWorkspaceQueryFromRoute,
    syncReviewWorkspaceQueryFromRoute,
    reviewWorkspaceQueryMatchesState,
    updateReviewWorkspaceFilters,
  }
}
