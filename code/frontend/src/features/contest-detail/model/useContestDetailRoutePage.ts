import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { useAuthStore } from '@/stores/auth'
import { getContestAccentColor, isStudentVisibleContestStatus } from '@/utils/contest'

import { useContestDetailPage } from './useContestDetailPage'

export type ContestWorkspaceTab = 'overview' | 'announcements' | 'challenges' | 'team'

export function useContestDetailRoutePage() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()

  const contestId = computed(() => String(route.params.id ?? ''))
  const currentUserId = computed(() => authStore.user?.id)
  const selectedChallengeId = computed(() => route.query.challenge)
  const workspaceTabOrder: ContestWorkspaceTab[] = ['overview', 'announcements', 'challenges', 'team']
  const {
    activeTab: activeWorkspaceTab,
    setTabButtonRef,
    selectTab: selectWorkspaceTab,
    handleTabKeydown: handleWorkspaceTabKeydown,
  } = useUrlSyncedTabs<ContestWorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    defaultTab: 'overview',
  })

  function syncSelectedChallengeQuery(challengeId: string | null): void {
    const query = { ...route.query }
    if (challengeId) {
      query.challenge = challengeId
      query.panel = 'challenges'
    } else {
      delete query.challenge
    }

    void router.replace({ query })
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
  const contestAccentStyle = computed<Record<string, string> | undefined>(() => {
    if (!page.contest.value) return undefined
    return {
      '--contest-accent': getContestAccentColor(page.contest.value.status),
    }
  })
  const contestAccessible = computed(() =>
    page.contest.value ? isStudentVisibleContestStatus(page.contest.value.status) : false
  )

  watch(
    () => page.contest.value?.mode,
    (mode) => {
      if (mode === 'awd' && !route.query.panel) {
        selectWorkspaceTab('challenges')
      }
    }
  )

  return {
    router,
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
    contestAccentStyle,
    contestAccessible,
    ...page,
  }
}
