import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import ContestDetail from '../ContestDetail.vue'
import contestDetailSource from '../ContestDetail.vue?raw'
import contestOverviewPanelSource from '@/components/contests/ContestOverviewPanel.vue?raw'
import contestChallengeWorkspacePanelSource from '@/components/contests/ContestChallengeWorkspacePanel.vue?raw'
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
vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

describe('ContestDetail', () => {
  let router: any

  beforeEach(async () => {
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
      challenge_id: '101',
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
  })

  it('应该渲染竞赛详情页面', async () => {
    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('公告')
    expect(wrapper.text()).toContain('比赛开始')
  })

  it('不应该向学生暴露草稿竞赛详情或报名入口', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '草稿竞赛',
      description: '未开放',
      status: 'draft',
      mode: 'jeopardy',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('当前竞赛暂未开放')
    expect(wrapper.text()).not.toContain('创建队伍')
    expect(wrapper.text()).not.toContain('加入队伍')
    expect(wrapper.text()).not.toContain('草稿')
  })

  it('收到公告实时事件后会刷新公告列表', async () => {
    contestApiMocks.getAnnouncements
      .mockResolvedValueOnce([
        {
          id: 'ann-1',
          title: '比赛开始',
          content: '欢迎来到比赛。',
          created_at: '2024-03-15T09:00:00Z',
        },
      ])
      .mockResolvedValueOnce([
        {
          id: 'ann-1',
          title: '比赛开始',
          content: '欢迎来到比赛。',
          created_at: '2024-03-15T09:00:00Z',
        },
        {
          id: 'ann-2',
          title: '第二条公告',
          content: '新的公告已发布。',
          created_at: '2024-03-15T10:00:00Z',
        },
      ])

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('比赛开始')

    webSocketMocks
      .getHandlers('contests/1/announcements')
      ?.['contest.announcement.created']?.({ contest_id: '1' })

    await flushPromises()

    expect(contestApiMocks.getAnnouncements).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('第二条公告')
  })

  it('AWD 赛事应切换到战场页签并加载学生工作台', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: '101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: '101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValueOnce({
      contest: {
        id: '1',
        title: '2026 春季校园 AWD 联赛',
        status: 'running',
        started_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '13',
            team_name: 'Red',
            score: 158,
            solved_count: 0,
            last_submission_at: '2024-03-15T09:03:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('战场')
    expect(wrapper.text()).toContain('攻击向量')
    expect(wrapper.text()).toContain('目标题目')
    expect(wrapper.text()).toContain('BLUE')
    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledWith('1')
  })

  it('AWD 赛事在未入队时应显示先加入队伍的提示', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      my_team: null,
      services: [],
      targets: [],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('先加入队伍')
    expect(wrapper.text()).toContain('需要先加入队伍后才能进入 AWD 战场。')
    expect(wrapper.text()).toContain('战场')
  })

  it('队伍页创建和加入弹窗应切换到 C 端输入模板', async () => {
    const contestDetailSource = (await import('../ContestDetail.vue?raw')).default

    expect(contestDetailSource).toContain(
      "from '@/components/common/modal-templates/CFocusedInputDialog.vue'"
    )
    expect(contestDetailSource).not.toContain('class="contest-modal"')

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
      attachTo: document.body,
    })

    await router.push('/contests/1?panel=team')
    await router.isReady()
    await flushPromises()

    const teamTab = wrapper.findAll('button').find((node) => node.text().trim() === '队伍')
    expect(teamTab).toBeTruthy()
    await teamTab!.trigger('click')
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '创建队伍')
      ?.trigger('click')
    await flushPromises()
    expect(document.body.textContent).toContain('创建新队伍')
    expect(document.body.textContent).toContain('队伍名称')
    expect(document.body.querySelector('.c-focused-input-shell--plain')).not.toBeNull()

    const closeButtons = Array.from(document.body.querySelectorAll('button'))
    const cancelCreateButton = closeButtons.find((button) => button.textContent?.trim() === '取消')
    cancelCreateButton?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '加入队伍')
      ?.trigger('click')
    await flushPromises()
    expect(document.body.textContent).toContain('加入现有队伍')
    expect(document.body.textContent).toContain('队伍 ID')

    wrapper.unmount()
  })

  it('普通竞赛提交反馈应由前端根据结果生成', async () => {
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.submitContestFlag
      .mockResolvedValueOnce({
        is_correct: false,
        points: 0,
        submitted_at: '2024-03-15T09:05:00Z',
      })
      .mockResolvedValueOnce({
        is_correct: true,
        points: 100,
        submitted_at: '2024-03-15T09:06:00Z',
      })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    const challengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(challengeButton).toBeTruthy()
    await challengeButton!.trigger('click')
    await flushPromises()

    const flagInput = wrapper.get('#contest-flag-input')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')
    expect(submitButton).toBeTruthy()

    await flagInput.setValue('flag{wrong}')
    await submitButton!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Flag 错误，请重试')

    await flagInput.setValue('flag{correct}')
    await submitButton!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('正确！+100 分')
  })

  it('普通竞赛题目选中状态应从 URL 恢复并在切换时写回 query', async () => {
    await router.push('/contests/1?panel=challenges&challenge=102')
    await router.isReady()
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
      {
        id: '102',
        challenge_id: '102',
        title: 'Crypto 102',
        category: 'crypto',
        difficulty: 'medium',
        points: 200,
        solved_count: 2,
        is_solved: false,
      },
    ])

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('已选题目')
    expect(wrapper.text()).toContain('主要操作')
    expect(wrapper.text()).toContain('Crypto 102')
    expect(wrapper.text()).toContain('crypto · 200 分')

    const webChallengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(webChallengeButton).toBeTruthy()
    await webChallengeButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.challenge).toBe('101')
    expect(router.currentRoute.value.query.panel).toBe('challenges')
  })

  it('竞赛 Flag 提交进行中遇到回车和点击重叠时只应提交一次', async () => {
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.submitContestFlag.mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    const challengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(challengeButton).toBeTruthy()
    await challengeButton!.trigger('click')
    await flushPromises()

    const flagInput = wrapper.get('#contest-flag-input')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')
    expect(submitButton).toBeTruthy()

    await flagInput.setValue('flag{pending}')
    flagInput.element.dispatchEvent(new KeyboardEvent('keyup', { key: 'Enter', bubbles: true }))
    submitButton!.element.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await wrapper.vm.$nextTick()

    expect(contestApiMocks.submitContestFlag).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.submitContestFlag).toHaveBeenCalledWith('1', '101', 'flag{pending}')
  })

  it('运行中的 AWD 战场应显示防守告警、自动刷新提示，并支持筛选目标', async () => {
    vi.useFakeTimers()

    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
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
      services: [
        {
          challenge_id: '101',
          access_url: 'http://red.internal',
          service_status: 'compromised',
          checker_type: 'http_standard',
          attack_received: 2,
          sla_score: 0,
          defense_score: 0,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: '101',
              reachable: true,
            },
          ],
        },
        {
          team_id: '15',
          team_name: 'Green',
          services: [
            {
              service_id: '7009',
              challenge_id: '101',
              reachable: false,
            },
          ],
        },
      ],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValue({
      contest: {
        id: '1',
        title: '2026 春季校园 AWD 联赛',
        status: 'running',
        started_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '13',
            team_name: 'Red',
            score: 158,
            solved_count: 0,
            last_submission_at: '2024-03-15T09:03:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('防守监控')
    expect(wrapper.text()).toContain('Service A')
    expect(wrapper.text()).toContain('BLUE')
    expect(wrapper.text()).toContain('GREEN')

    await wrapper.get('#awd-target-search').setValue('Blue')
    await flushPromises()

    expect(wrapper.text()).toContain('BLUE')
    expect(wrapper.text()).not.toContain('GREEN')

    await wrapper.get('#awd-target-search').setValue('Yellow')
    await flushPromises()

    expect(wrapper.text()).toContain('当前题目下没有匹配的目标队伍。')

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(2)
  })

  it('学生 AWD 工作台应展示运行态 service 标识', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: '101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7010',
              challenge_id: '101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [
        {
          id: 'attack-1',
          service_id: '7010',
          challenge_id: '101',
          direction: 'attack_out',
          peer_team_id: '14',
          peer_team_name: 'Blue',
          is_success: true,
          score_gained: 60,
          created_at: '2024-03-15T09:03:00Z',
        },
      ],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7009')
    expect(wrapper.text()).toContain('服务 #7010')
  })

  it('学生 AWD 工作台应优先用 awd service 标识匹配运行态服务', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: 'legacy-9',
          instance_id: '9001',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: 'legacy-9',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('已通过平台代理就绪')
    expect(wrapper.text()).not.toContain('http://blue.runtime.internal')
    expect(wrapper.text()).toContain('代理链路已就绪')
    expect(wrapper.text()).toContain('正常')
  })

  it('学生 AWD 工作台应允许用 awd service 标识切换攻击题目', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
      {
        id: '202',
        challenge_id: '102',
        awd_service_id: '7010',
        title: 'Service B',
        category: 'pwn',
        difficulty: 'hard',
        points: 200,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: 'legacy-101',
          access_url: 'http://red-a.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
        {
          service_id: '7010',
          challenge_id: 'legacy-102',
          access_url: 'http://red-b.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:30Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: 'legacy-101',
              reachable: true,
            },
            {
              service_id: '7010',
              challenge_id: 'legacy-102',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7009')

    await wrapper.get('#awd-target-challenge').setValue('7010')
    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7010')
    expect(wrapper.text()).not.toContain('http://blue-a.internal')
    expect(wrapper.text()).not.toContain('http://blue-b.internal')
  })

  it('学生 AWD 最近反馈应优先按 service 标识回填题目标题', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: 'legacy-101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [],
      recent_events: [
        {
          id: 'attack-1',
          service_id: '7009',
          challenge_id: 'legacy-101',
          direction: 'attack_out',
          peer_team_id: '14',
          peer_team_name: 'Blue',
          is_success: true,
          score_gained: 60,
          created_at: '2024-03-15T09:03:00Z',
        },
      ],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(
      wrapper.findAll('[data-testid="awd-feedback-challenge-title"]').map((node) => node.text())
    ).toContain('Bank Portal')
  })

  it('学生 AWD 服务匹配在存在 awd_service_id 时不应再回退到 challenge_id', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7010',
        title: 'Admin Gateway',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: '101',
          access_url: 'http://wrong.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: '101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('http://wrong.internal')
    expect(wrapper.text()).not.toContain('http://wrong-target.internal')
  })

  it('学生 AWD 面板不应渲染缺少 awd_service_id 的遗留题目', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        title: 'Legacy Gateway',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
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
      services: [
        {
          service_id: '7009',
          challenge_id: '101',
          access_url: 'http://legacy.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('Legacy Gateway')
    expect(wrapper.text()).toContain('当前竞赛暂无可部署服务。')
  })

  it('学生 AWD 提交结果提示应优先按 service 标识回填题目标题', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
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
      services: [
        {
          service_id: '7009',
          challenge_id: 'legacy-101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: 'legacy-101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValue({
      contest: {
        id: '1',
        title: '2026 春季校园 AWD 联赛',
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
    contestApiMocks.submitContestAWDAttack.mockResolvedValueOnce({
      id: '88',
      round_id: '41',
      attacker_team_id: '13',
      attacker_team: 'Red',
      victim_team_id: '14',
      victim_team: 'Blue',
      service_id: '7009',
      challenge_id: 'legacy-101',
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2024-03-15T09:03:00Z',
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    await wrapper.find('input[placeholder="输入获取到的 Flag..."]').setValue('flag{demo}')
    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '提交')
      ?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Bank Portal: 攻击成功，+60 分')
  })

  it('学生 AWD 工作台应通过跨队攻击代理打开目标服务', async () => {
    const openMock = vi.spyOn(window, 'open').mockImplementation(() => null)
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_service_id: '7009',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
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
      services: [
        {
          service_id: '7009',
          challenge_id: 'legacy-101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              challenge_id: 'legacy-101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    await wrapper.get('[data-testid="awd-open-target-7009-14"]').trigger('click')
    await flushPromises()

    expect(contestApiMocks.requestContestAWDTargetAccess).toHaveBeenCalledWith('1', '7009', '14')
    expect(openMock).toHaveBeenCalledWith(
      '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
      '_blank',
      'noopener,noreferrer'
    )

    openMock.mockRestore()
  })

  it('踢出队员前应走统一确认弹窗', async () => {
    contestApiMocks.getMyTeam.mockResolvedValueOnce({
      id: 'team-1',
      name: 'Red',
      captain_user_id: 'user-1',
      invite_code: 'RED-CTF',
      members: [
        { user_id: 'user-1', username: 'alice' },
        { user_id: 'user-2', username: 'bob' },
      ],
    })
    contestApiMocks.kickTeamMember.mockResolvedValue(undefined)

    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().setAuth({ id: 'user-1', username: 'alice', role: 'student' })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [pinia, router],
      },
    })

    await router.push('/contests/1?panel=team')
    await router.isReady()
    await flushPromises()

    const teamTab = wrapper.findAll('button').find((node) => node.text().trim() === '队伍')
    expect(teamTab).toBeTruthy()
    await teamTab!.trigger('click')
    await flushPromises()

    const kickButton = wrapper.findAll('button').find((node) => node.text().trim() === '踢出')
    expect(kickButton).toBeTruthy()
    await kickButton!.trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalledWith({
      title: '踢出成员',
      message: '确定踢出该成员？',
      confirmButtonText: '确认踢出',
    })
    expect(contestApiMocks.kickTeamMember).toHaveBeenCalledWith('1', 'team-1', 'user-2')
  })

  it('竞赛详情 hero 应使用共享 workspace overline 语义', () => {
    expect(contestOverviewPanelSource).toMatch(
      /<div class="workspace-overline">\s*Contest\s*<\/div>/
    )
    expect(contestOverviewPanelSource).not.toContain('<div class="contest-overline">Contest</div>')
  })

  it('竞赛概览数值区域应接入共享 metric panel surface', () => {
    expect(contestOverviewPanelSource).toContain(
      'class="contest-score-rail metric-panel-card metric-panel-workspace-surface"'
    )
    expect(contestOverviewPanelSource).toContain(
      'class="contest-stat-grid metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(contestOverviewPanelSource).toContain(
      'class="contest-stat progress-card metric-panel-card"'
    )
    expect(contestOverviewPanelSource).not.toContain('--metric-panel-grid-gap: 0.85rem;')
    expect(contestOverviewPanelSource).not.toContain('gap: 1.25rem;')
  })

  it('竞赛详情 section heading 应切到共享 workspace overline 语义', () => {
    const combinedSource = [contestDetailSource, contestOverviewPanelSource].join('\n')

    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*Rules\s*<\/div>/)
    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*Schedule\s*<\/div>/)
    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*Announcements\s*<\/div>/)
    expect(combinedSource).toMatch(
      /<div class="workspace-overline">\s*\{\{ contest\.mode === 'awd' \? '战场' : '题目' \}\}\s*<\/div>/
    )
    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*Team\s*<\/div>/)
    expect(combinedSource).not.toContain('<div class="contest-overline">Rules</div>')
    expect(combinedSource).not.toContain('<div class="contest-overline">Schedule</div>')
    expect(combinedSource).not.toContain('<div class="contest-overline">Announcements</div>')
    expect(combinedSource).not.toContain('<div class="contest-overline">Team</div>')
  })

  it('竞赛详情剩余局部 kicker 也应统一到 workspace overline 语义', () => {
    const combinedSource = [contestDetailSource, contestChallengeWorkspacePanelSource].join('\n')

    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*已选题目\s*<\/div>/)
    expect(combinedSource).toMatch(/<div class="workspace-overline">\s*主要操作\s*<\/div>/)
    expect(contestDetailSource).toMatch(/<div class="workspace-overline">\s*Current Team\s*<\/div>/)
    expect(combinedSource).not.toContain('<div class="contest-overline">已选题目</div>')
    expect(combinedSource).not.toContain('<div class="contest-overline">主要操作</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Current Team</div>')
    expect(contestDetailSource).not.toMatch(/^\.contest-overline\s*\{/m)
  })
})
