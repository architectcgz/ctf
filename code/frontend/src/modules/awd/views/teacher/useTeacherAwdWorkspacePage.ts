import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useTeacherAwdReviewDetail } from '@/composables/useTeacherAwdReviewDetail'
import { buildTeacherAwdPageModel } from '@/modules/awd/adapters/teacherAwdPageAdapter'
import { TEACHER_AWD_PAGES, buildTeacherAwdPath } from '@/modules/awd/navigation'
import type { TeacherAwdPageKey } from '@/modules/awd/types'

export function useTeacherAwdWorkspacePage(pageKey: TeacherAwdPageKey) {
  const route = useRoute()
  const detail = useTeacherAwdReviewDetail()

  const pageModel = computed(() =>
    buildTeacherAwdPageModel({
      review: detail.review.value,
      selectedPage: pageKey,
    })
  )

  const empty = computed(() => !detail.loading.value && !detail.error.value && !detail.review.value)

  const layoutProps = computed(() => ({
    contestTitle: pageModel.value.hero.contestTitle,
    pageTitle: pageModel.value.hero.pageTitle,
    pageDescription: pageModel.value.hero.pageDescription,
    pages: TEACHER_AWD_PAGES,
    currentPage: pageKey,
    heroMetrics: pageModel.value.hero.metrics,
    resolvePath: (nextPageKey: string) =>
      buildTeacherAwdPath(detail.contestId.value, nextPageKey as TeacherAwdPageKey),
    loading: detail.loading.value,
    error: detail.error.value || '',
    empty: empty.value,
    emptyTitle: '当前没有可展示的 AWD 复盘数据',
    emptyDescription: '请确认赛事轮次、教学复盘归档和导出状态。',
  }))

  function setRound(roundNumber?: number): void {
    const query = {
      ...route.query,
    } as Record<string, string>

    if (roundNumber) {
      query.round = String(roundNumber)
    } else {
      delete query.round
    }

    void detail.router.replace({
      name: String(route.name || 'TeacherAwdOverview'),
      params: { contestId: detail.contestId.value },
      query,
    })
  }

  return {
    ...detail,
    pageModel,
    layoutProps,
    setRound,
  }
}
