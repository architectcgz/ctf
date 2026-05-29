import { useRouter } from 'vue-router'

import { useAwdReviewIndex } from './useAwdReviewIndex'

type AwdReviewIndexPageScope = 'teacher' | 'platform'

export function useAwdReviewIndexPage(scope: AwdReviewIndexPageScope) {
  const router = useRouter()
  const index = useAwdReviewIndex()

  function openHome(): void {
    void router.push({
      name: scope === 'teacher' ? 'TeacherDashboard' : 'PlatformOverview',
    })
  }

  function openContest(contestId: string): void {
    void router.push({
      name: scope === 'teacher' ? 'TeacherAWDReviewDetail' : 'PlatformAwdReviewDetail',
      params: { contestId },
    })
  }

  return {
    ...index,
    openHome,
    openContest,
  }
}
