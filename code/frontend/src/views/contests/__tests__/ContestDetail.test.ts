import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia } from 'pinia'
import ContestDetail from '../ContestDetail.vue'
import contestDetailSource from '../ContestDetail.vue?raw'

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
  startContestChallengeInstance: vi.fn(),
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

vi.mock('@/api/contest', () => contestApiMocks)
vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
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
    contestApiMocks.startContestChallengeInstance.mockReset()
    contestApiMocks.submitContestAWDAttack.mockReset()
    contestApiMocks.submitContestFlag.mockReset()
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
    contestApiMocks.startContestChallengeInstance.mockResolvedValue({
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
              challenge_id: '101',
              access_url: 'http://blue.internal',
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
    expect(wrapper.text()).toContain('目标目录')
    expect(wrapper.text()).toContain('Blue')
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

    expect(wrapper.text()).toContain('先在队伍页创建或加入队伍')
    expect(wrapper.text()).toContain('战场')
  })

  it('队伍页创建和加入弹窗应切换到 C 端输入模板', async () => {
    const contestDetailSource = (await import('../ContestDetail.vue?raw')).default

    expect(contestDetailSource).toContain("from '@/components/common/modal-templates/CFocusedInputDialog.vue'")
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

    await wrapper.findAll('button').find((node) => node.text().trim() === '创建队伍')?.trigger('click')
    await flushPromises()
    expect(document.body.textContent).toContain('创建新队伍')
    expect(document.body.textContent).toContain('队伍名称')
    expect(document.body.querySelector('.c-focused-input-shell--plain')).not.toBeNull()

    const closeButtons = Array.from(document.body.querySelectorAll('button'))
    const cancelCreateButton = closeButtons.find((button) => button.textContent?.trim() === '取消')
    cancelCreateButton?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    await wrapper.findAll('button').find((node) => node.text().trim() === '加入队伍')?.trigger('click')
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
              challenge_id: '101',
              access_url: 'http://blue.internal',
            },
          ],
        },
        {
          team_id: '15',
          team_name: 'Green',
          services: [
            {
              challenge_id: '101',
              access_url: '',
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

    expect(wrapper.text()).toContain('防守告警')
    expect(wrapper.text()).toContain('Service A')
    expect(wrapper.text()).toContain('每 15 秒自动刷新')
    expect(wrapper.text()).toContain('Blue')
    expect(wrapper.text()).toContain('Green')

    await wrapper.get('#awd-target-search').setValue('Blue')
    await flushPromises()

    expect(wrapper.text()).toContain('Blue')
    expect(wrapper.text()).not.toContain('Green')

    await wrapper.get('#awd-target-search').setValue('Yellow')
    await flushPromises()

    expect(wrapper.text()).toContain('没有匹配的目标队伍')

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(2)
  })

  it('竞赛详情 hero 应使用共享 workspace overline 语义', () => {
    expect(contestDetailSource).toContain('<div class="workspace-overline">Contest</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Contest</div>')
  })

  it('竞赛详情 section heading 应切到共享 workspace overline 语义', () => {
    expect(contestDetailSource).toContain('<div class="workspace-overline">Rules</div>')
    expect(contestDetailSource).toContain('<div class="workspace-overline">Schedule</div>')
    expect(contestDetailSource).toContain('<div class="workspace-overline">Announcements</div>')
    expect(contestDetailSource).toContain(
      `<div class="workspace-overline">
                  {{ contest.mode === 'awd' ? 'Battle' : 'Challenges' }}
                </div>`
    )
    expect(contestDetailSource).toContain('<div class="workspace-overline">Team</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Rules</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Schedule</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Announcements</div>')
    expect(contestDetailSource).not.toContain('<div class="contest-overline">Team</div>')
  })
})
