import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import { usePlatformContestAwd } from '@/composables/usePlatformContestAwd'
import { buildAdminAwdPageModel } from '@/modules/awd/adapters/adminAwdPageAdapter'
import { ADMIN_AWD_PAGES, buildAdminAwdPath } from '@/modules/awd/navigation'
import type { AdminAwdPageKey } from '@/modules/awd/types'

export function useAdminAwdWorkspacePage(pageKey: AdminAwdPageKey) {
  const route = useRoute()
  const contest = ref<ContestDetailData | null>(null)
  const metadataLoading = ref(false)
  const metadataError = ref('')
  const contestId = computed(() => String(route.params.id ?? ''))

  const awd = usePlatformContestAwd(contest)
  let metadataRequestToken = 0

  async function loadContestMetadata(): Promise<void> {
    if (!contestId.value) {
      contest.value = null
      metadataError.value = ''
      metadataLoading.value = false
      return
    }

    const currentToken = ++metadataRequestToken
    metadataLoading.value = true

    try {
      const nextContest = await getContest(contestId.value)
      if (currentToken !== metadataRequestToken) {
        return
      }

      contest.value = nextContest
      metadataError.value = nextContest.mode === 'awd' ? '' : '当前赛事不是 AWD 模式'
    } catch (error) {
      if (currentToken !== metadataRequestToken) {
        return
      }

      contest.value = null
      metadataError.value = error instanceof Error ? error.message : '加载赛事数据失败，请稍后重试'
    } finally {
      if (currentToken === metadataRequestToken) {
        metadataLoading.value = false
      }
    }
  }

  watch(
    contestId,
    () => {
      void loadContestMetadata()
    },
    { immediate: true }
  )

  const pageModel = computed(() =>
    buildAdminAwdPageModel({
      contest: contest.value,
      rounds: awd.rounds.value,
      summary: awd.summary.value,
      readiness: awd.readiness.value,
      services: awd.services.value,
      attacks: awd.attacks.value,
      trafficSummary: awd.trafficSummary.value,
      trafficEvents: awd.trafficEvents.value,
      teams: awd.teams.value,
      challengeLinks: awd.challengeLinks.value,
      scoreboardRows: awd.scoreboardRows.value,
      selectedPage: pageKey,
    })
  )

  const loading = computed(
    () =>
      metadataLoading.value ||
      awd.loadingRounds.value ||
      (awd.loadingRoundDetail.value && !awd.summary.value && awd.services.value.length === 0)
  )
  const error = computed(() => metadataError.value)
  const empty = computed(() => !loading.value && !error.value && !contest.value)

  const layoutProps = computed(() => ({
    contestTitle: pageModel.value.hero.contestTitle,
    pageTitle: pageModel.value.hero.pageTitle,
    pageDescription: pageModel.value.hero.pageDescription,
    pages: ADMIN_AWD_PAGES,
    currentPage: pageKey,
    heroMetrics: pageModel.value.hero.metrics,
    resolvePath: (nextPageKey: string) => buildAdminAwdPath(contestId.value, nextPageKey as AdminAwdPageKey),
    loading: loading.value,
    error: error.value,
    empty: empty.value,
    emptyTitle: '当前没有可展示的 AWD 管理数据',
    emptyDescription: '请确认赛事状态、轮次初始化情况与管理接口返回。',
  }))

  function selectRound(roundId: string): void {
    awd.selectedRoundId.value = roundId
  }

  return {
    contest,
    pageModel,
    layoutProps,
    selectedRoundId: awd.selectedRoundId,
    refreshAll: awd.refresh,
    refreshReadiness: awd.refreshReadiness,
    refreshRoundDetail: awd.refreshRoundDetail,
    selectRound,
  }
}
