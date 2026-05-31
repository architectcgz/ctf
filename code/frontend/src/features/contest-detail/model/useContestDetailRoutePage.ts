import { computed, watch } from 'vue'

import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'
import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'
import { useAuthStore } from '@/stores/auth'
import { isStudentVisibleContestStatus } from '@/entities/contest'

import { useContestDetailPage } from './useContestDetailPage'

export type ContestWorkspaceTab = 'overview' | 'announcements' | 'challenges' | 'team'

export function useContestDetailRoutePage() {
  const { params, query, replaceQuery } = useRouteQueryTransport()
  const authStore = useAuthStore()

  const contestId = computed(() => String(params.value.id ?? ''))
  const currentUserId = computed(() => authStore.user?.id)
  const selectedChallengeId = computed<string | string[] | undefined>(() => {
    const challenge = query.value.challenge
    if (Array.isArray(challenge)) {
      return challenge.filter((value): value is string => typeof value === 'string')
    }
    return typeof challenge === 'string' ? challenge : undefined
  })
  const workspaceTabOrder: ContestWorkspaceTab[] = ['overview', 'announcements', 'challenges', 'team']
  const {
    activeTab: activeWorkspaceTab,
    setTabButtonRef,
    selectTab: selectWorkspaceTab,
    handleTabKeydown: handleWorkspaceTabKeydown,
  } = useRouteQueryTabs<ContestWorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    defaultTab: 'overview',
  })

  function syncSelectedChallengeQuery(challengeId: string | null): void {
    const nextQuery = { ...query.value }
    if (challengeId) {
      nextQuery.challenge = challengeId
      nextQuery.panel = 'challenges'
    } else {
      delete nextQuery.challenge
    }

    void replaceQuery(nextQuery)
  }

  const page = useContestDetailPage({
    contestId,
    currentUserId,
    selectedChallengeId,
    onSelectedChallengeChange: syncSelectedChallengeQuery,
  })

  const isAWDContest = computed(() => page.contest.value?.mode === 'awd')
  const workspaceTabs = computed<Array<{ id: ContestWorkspaceTab; label: string }>>(() => [
    { id: 'overview', label: '概览' },
    { id: 'announcements', label: '公告' },
    { id: 'challenges', label: isAWDContest.value ? '攻防战场' : '题目' },
    { id: 'team', label: '队伍' },
  ])
  const solvedCount = computed(() => page.challenges.value.filter((item) => item.is_solved).length)
  const totalPoints = computed(() =>
    page.challenges.value.reduce((sum, item) => sum + (item.points || 0), 0)
  )
  const memberCount = computed(() => page.team.value?.members.length ?? 0)
  const contestAccessible = computed(() =>
    page.contest.value ? isStudentVisibleContestStatus(page.contest.value.status) : false
  )

  watch(
    () => page.contest.value?.mode,
    (mode) => {
      if (mode === 'awd' && !query.value.panel) {
        void selectWorkspaceTab('challenges')
      }
    }
  )

  return {
    contestId,
    activeWorkspaceTab,
    setTabButtonRef,
    selectWorkspaceTab,
    handleWorkspaceTabKeydown,
    workspaceTabs,
    isAWDContest,
    solvedCount,
    totalPoints,
    memberCount,
    contestAccessible,
    ...page,
  }
}
