import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { getContestChallenges, getContestDetail } from '@/api/contest'
import type { ContestChallengeItem, ContestDetailData } from '@/api/contracts'
import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'
import { buildStudentAwdPageModel } from '@/modules/awd/adapters/studentAwdPageAdapter'
import { STUDENT_AWD_PAGES, buildStudentAwdPath } from '@/modules/awd/navigation'
import type { StudentAwdPageKey } from '@/modules/awd/types'

export function useStudentAwdWorkspacePage(pageKey: StudentAwdPageKey) {
  const route = useRoute()
  const contest = ref<ContestDetailData | null>(null)
  const challenges = ref<ContestChallengeItem[]>([])
  const metadataLoading = ref(false)
  const metadataError = ref('')
  const contestId = computed(() => String(route.params.id ?? ''))

  let metadataRequestToken = 0

  const {
    workspace,
    scoreboardRows,
    loading: workspaceLoading,
    error: workspaceError,
    submitResult,
    startingServiceKey,
    submittingKey,
    refreshAll,
    startService,
    submitAttack,
  } = useContestAWDWorkspace({
    contestId,
    contestStatus: computed(() => contest.value?.status),
  })

  async function loadContestMetadata(): Promise<void> {
    if (!contestId.value) {
      contest.value = null
      challenges.value = []
      metadataError.value = ''
      metadataLoading.value = false
      return
    }

    const currentToken = ++metadataRequestToken
    metadataLoading.value = true

    try {
      const [nextContest, nextChallenges] = await Promise.all([
        getContestDetail(contestId.value),
        getContestChallenges(contestId.value),
      ])

      if (currentToken !== metadataRequestToken) {
        return
      }

      contest.value = nextContest
      challenges.value = nextChallenges
      metadataError.value = ''
    } catch (error) {
      if (currentToken !== metadataRequestToken) {
        return
      }

      contest.value = null
      challenges.value = []
      metadataError.value = error instanceof Error ? error.message : '加载 AWD 页面失败，请稍后重试'
    } finally {
      if (currentToken === metadataRequestToken) {
        metadataLoading.value = false
      }
    }
  }

  watch(contestId, () => {
    void loadContestMetadata()
  }, { immediate: true })

  const pageModel = computed(() =>
    buildStudentAwdPageModel({
      contest: contest.value,
      challenges: challenges.value,
      workspace: workspace.value,
      scoreboardRows: scoreboardRows.value,
      selectedPage: pageKey,
      submitResult: submitResult.value,
    })
  )

  const loading = computed(() => metadataLoading.value || workspaceLoading.value)
  const error = computed(() => metadataError.value || workspaceError.value)
  const empty = computed(() => !loading.value && !error.value && !contest.value)

  const layoutProps = computed(() => ({
    contestTitle: pageModel.value.hero.contestTitle,
    pageTitle: pageModel.value.hero.pageTitle,
    pageDescription: pageModel.value.hero.pageDescription,
    pages: STUDENT_AWD_PAGES,
    currentPage: pageKey,
    heroMetrics: pageModel.value.hero.metrics,
    resolvePath: (nextPageKey: string) => buildStudentAwdPath(contestId.value, nextPageKey as StudentAwdPageKey),
    loading: loading.value,
    error: error.value,
    empty: empty.value,
    emptyTitle: '当前没有可展示的 AWD 数据',
    emptyDescription: '请确认赛事状态、队伍加入情况和服务初始化结果。',
  }))

  return {
    contest,
    challenges,
    workspace,
    scoreboardRows,
    submitResult,
    startingServiceKey,
    submittingKey,
    refreshAll,
    startService,
    submitAttack,
    pageModel,
    layoutProps,
  }
}
