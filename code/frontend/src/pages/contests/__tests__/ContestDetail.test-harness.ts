import { vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
export { default as contestDetailSource } from '@/pages/contests/ContestDetailRoutePage.vue?raw'
export { default as contestDetailPageSource } from '@/features/contest-detail/model/useContestDetailPage.ts?raw'
export { default as contestDetailRoutePageSource } from '@/features/contest-detail/model/useContestDetailRoutePage.ts?raw'
export { default as contestDetailWorkspaceSource } from '@/widgets/contest-detail-workspace/ContestDetailWorkspace.vue?raw'
export { default as contestPresentationSource } from '@/entities/contest/model/presentation.ts?raw'
export { default as teamPresentationSource } from '@/entities/team/model/presentation.ts?raw'
export { default as contestTeamDialogsSource } from '@/features/contest-detail/ui/ContestTeamDialogs.vue?raw'
export { default as contestTeamPanelSource } from '@/features/contest-detail/ui/ContestTeamPanel.vue?raw'
export { default as contestTeamWorkspaceSectionSource } from '@/features/contest-detail/ui/ContestTeamWorkspaceSection.vue?raw'
export { default as routeQueryTransportSource } from '@/shared/model/navigation/useRouteQueryTransport.ts?raw'
import { useAuthStore } from '@/stores/auth'

const contestApiMocks = vi.hoisted(() => ({
  getContestDetail: vi.fn(),
  getMyTeam: vi.fn(),
  getContestChallenges: vi.fn(),
  getAnnouncements: vi.fn(),
  getContestAWDWorkspace: vi.fn(),
  getScoreboard: vi.fn(),
  createTeam: vi.fn(),
  joinTeam: vi.fn(),
  kickTeamMember: vi.fn(),
  requestContestAWDTargetAccess: vi.fn(),
  restartContestAWDServiceInstance: vi.fn(),
  startContestAWDServiceInstance: vi.fn(),
  submitContestAWDAttack: vi.fn(),
  submitContestFlag: vi.fn(),
}))

const webSocketMocks = vi.hoisted(() => {
  const connect = vi.fn().mockResolvedValue(undefined)
  const disconnect = vi.fn()
  const handlersByEndpoint = new Map<string, Record<string, (payload: unknown) => void>>()

  return {
    connect,
    disconnect,
    getHandlers: (endpoint: string) => handlersByEndpoint.get(endpoint),
    reset: () => handlersByEndpoint.clear(),
    useWebSocket: vi.fn(
      (endpoint: string, handlers: Record<string, (payload: unknown) => void>) => {
        handlersByEndpoint.set(endpoint, handlers)
        return {
          status: { value: 'idle' as const },
          connect,
          disconnect,
          send: vi.fn(),
        }
      }
    ),
  }
})
const destructiveConfirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/contest', () => contestApiMocks)
vi.mock('@/shared/model/realtime/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))
vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

export let router: any

export async function resetContestDetailTestHarness() {
  vi.useRealTimers()
  contestApiMocks.getContestDetail.mockReset()
  contestApiMocks.getMyTeam.mockReset()
  contestApiMocks.getContestChallenges.mockReset()
  contestApiMocks.getAnnouncements.mockReset()
  contestApiMocks.getContestAWDWorkspace.mockReset()
  contestApiMocks.getScoreboard.mockReset()
  contestApiMocks.createTeam.mockReset()
  contestApiMocks.joinTeam.mockReset()
  contestApiMocks.kickTeamMember.mockReset()
  contestApiMocks.requestContestAWDTargetAccess.mockReset()
  contestApiMocks.restartContestAWDServiceInstance.mockReset()
  contestApiMocks.startContestAWDServiceInstance.mockReset()
  contestApiMocks.submitContestAWDAttack.mockReset()
  contestApiMocks.submitContestFlag.mockReset()
  destructiveConfirmMock.mockReset()
  destructiveConfirmMock.mockResolvedValue(true)
  webSocketMocks.connect.mockClear()
  webSocketMocks.disconnect.mockClear()
  webSocketMocks.useWebSocket.mockClear()
  webSocketMocks.reset()

  contestApiMocks.getContestDetail.mockResolvedValue({
    id: '1',
    title: '2026 春季校园 CTF 挑战赛',
    description: '测试描述',
    status: 'running',
    mode: 'jeopardy',
    starts_at: '2024-03-15T09:00:00Z',
    ends_at: '2024-03-15T21:00:00Z',
  })
  contestApiMocks.getMyTeam.mockResolvedValue(null)
  contestApiMocks.getContestChallenges.mockResolvedValue([])
  contestApiMocks.getAnnouncements.mockResolvedValue([
    {
      id: 'ann-1',
      title: '比赛开始',
      content: '欢迎来到比赛。',
      created_at: '2024-03-15T09:00:00Z',
    },
  ])
  contestApiMocks.getContestAWDWorkspace.mockResolvedValue({
    contest_id: '1',
    current_round: {
      id: '41',
      contest_id: '1',
      round_number: 2,
      status: 'running',
      attack_score: 60,
      defense_score: 40,
      created_at: '2024-03-15T09:00:00Z',
      updated_at: '2024-03-15T09:01:00Z',
    },
    my_team: {
      team_id: '13',
      team_name: 'Red',
    },
    services: [],
    targets: [],
    recent_events: [],
  })
  contestApiMocks.getScoreboard.mockResolvedValue({
    contest: {
      id: '1',
      title: '2026 春季校园 CTF 挑战赛',
      status: 'running',
      started_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    },
    scoreboard: {
      list: [],
      total: 0,
      page: 1,
      page_size: 10,
    },
    frozen: false,
  })
  contestApiMocks.startContestAWDServiceInstance.mockResolvedValue({
    id: '900',
    challenge_id: '101',
    status: 'running',
    share_scope: 'per_team',
    access_url: 'http://red.internal',
    flag_type: 'dynamic',
    expires_at: '2024-03-15T12:00:00Z',
    remaining_extends: 1,
    created_at: '2024-03-15T09:02:00Z',
  })
  contestApiMocks.requestContestAWDTargetAccess.mockResolvedValue({
    access_url: '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
  })
  contestApiMocks.submitContestAWDAttack.mockResolvedValue({
    id: '88',
    round_id: '41',
    attacker_team_id: '13',
    attacker_team: 'Red',
    victim_team_id: '14',
    victim_team: 'Blue',
    awd_challenge_id: 'awd-101',
    attack_type: 'flag_capture',
    source: 'submission',
    submitted_flag: 'flag{demo}',
    is_success: true,
    score_gained: 60,
    created_at: '2024-03-15T09:03:00Z',
  })

  router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/contests', component: { template: '<div>contests</div>' } },
      { path: '/contests/:id', component: { template: '<div />' } },
    ],
  })
  await router.push('/contests/1')
  await router.isReady()
}

export { contestApiMocks, destructiveConfirmMock, webSocketMocks }
