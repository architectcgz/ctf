import { useAwdReviewIndex } from './useAwdReviewIndex'
import {
  buildAwdReviewDetailRoute,
  resolveAwdReviewIndexHomeRoute,
  type AwdReviewIndexPageScope,
} from './awdReviewIndexRoutes'

export function useAwdReviewIndexPage(scope: AwdReviewIndexPageScope) {
  const index = useAwdReviewIndex()

  return {
    ...index,
    homeRoute: resolveAwdReviewIndexHomeRoute(scope),
    buildContestRoute: (contestId: string) => buildAwdReviewDetailRoute(scope, contestId),
  }
}
