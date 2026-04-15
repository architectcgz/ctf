import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestEdit from '../ContestEdit.vue'
import { ApiError } from '@/api/request'
import type { ContestDetailData } from '@/api/contracts'
import type { VueWrapper } from '@vue/test-utils'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' } as Record<string, string>,
}))

const contestApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  updateContest: vi.fn(),
  getContestAWDReadiness: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  getChallenges: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

const destructiveConfirmMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    getChallenges: contestApiMocks.getChallenges,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

function buildContestDetail(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'contest-1',
    title: '2026 春季校园 CTF',
    description: '校内赛',
    mode: 'jeopardy',
    status: 'registering',
    starts_at: '2026-03-15T09:00:00.000Z',
    ends_at: '2026-03-15T13:00:00.000Z',
    ...overrides,
  }
}

function mountContestEdit() {
  return mount(ContestEdit, {
    global: {
      stubs: {
        ElDialog: {
          props: ['modelValue', 'title'],
          template: '<div><div v-if="title">{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

function getWorkbenchStageRail(wrapper: VueWrapper<any>) {
  return wrapper.get('[role="tablist"][aria-label="竞赛工作台阶段切换"]')
}

describe('ContestEdit', () => {
  beforeEach(() => {
    pushMock.mockReset()
    contestApiMocks.getContest.mockReset()
    contestApiMocks.updateContest.mockReset()
    contestApiMocks.getContestAWDReadiness.mockReset()
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
    destructiveConfirmMock.mockReset()
    routeState.params = { id: 'contest-1' }

    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 春季校园 CTF',
      description: '校内赛',
      mode: 'jeopardy',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest.mockResolvedValue({
      contest: {
        id: 'contest-1',
        title: '2026 春季校园 CTF（更新）',
        description: '校内赛',
        mode: 'jeopardy',
        status: 'registering',
        starts_at: '2026-03-15T09:00:00.000Z',
        ends_at: '2026-03-15T13:00:00.000Z',
      },
    })
    contestApiMocks.getContestAWDReadiness.mockResolvedValue({
      contest_id: 'contest-1',
      ready: false,
      total_challenges: 1,
      passed_challenges: 0,
      pending_challenges: 0,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 1,
      global_blocking_reasons: [],
      blocking_actions: ['start_contest'],
      items: [
        {
          challenge_id: '101',
          title: 'Challenge 101',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    contestApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '101',
          title: 'Web 入门',
          description: '现有题目',
          category: 'web',
          difficulty: 'easy',
          points: 120,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
          flag_config: undefined,
        },
        {
          id: '102',
          title: 'Crypto 进阶',
          description: '新增题目',
          category: 'crypto',
          difficulty: 'medium',
          points: 150,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
          flag_config: undefined,
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createAdminContestChallenge.mockResolvedValue({
      id: 'link-2',
      contest_id: 'contest-1',
      challenge_id: '102',
      title: 'Crypto 进阶',
      category: 'crypto',
      difficulty: 'medium',
      points: 160,
      order: 3,
      is_visible: false,
      awd_checker_type: undefined,
      awd_checker_config: {},
      awd_sla_score: 0,
      awd_defense_score: 0,
      awd_checker_validation_state: 'pending',
      awd_checker_last_preview_at: undefined,
      awd_checker_last_preview_result: undefined,
      created_at: '2026-03-10T01:00:00.000Z',
    })
    contestApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)
    contestApiMocks.deleteAdminContestChallenge.mockResolvedValue(undefined)
    destructiveConfirmMock.mockResolvedValue(true)
  })

  it('应该在普通赛下只展示基础信息与题目池阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目池')
    expect(stageRail.text()).not.toContain('AWD 配置')
    expect(stageRail.text()).not.toContain('赛前检查')
    expect(stageRail.text()).not.toContain('轮次运行')
  })

  it('应该在 AWD 赛事下展示基础信息、题目池、AWD 配置、赛前检查与轮次运行', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目池')
    expect(stageRail.text()).toContain('AWD 配置')
    expect(stageRail.text()).toContain('赛前检查')
    expect(stageRail.text()).toContain('轮次运行')
  })

  it('应该在 AWD 赛事已开赛时默认聚焦轮次运行阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('轮次运行')
  })

  it('应该加载竞赛详情并在保存成功后返回赛事目录', async () => {
    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('编辑竞赛')

    await wrapper.get('#contest-title').setValue('2026 春季校园 CTF（更新）')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF（更新）',
      })
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该在 AWD 启动门禁拦截后展示放行弹层并在确认后回到赛事目录', async () => {
    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      description: '攻防赛',
      mode: 'awd',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest
      .mockRejectedValueOnce(new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 }))
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-1',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      })

    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getContestAWDReadiness).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('启动赛事')
    expect(wrapper.text()).toContain('强制继续')

    await wrapper.get('#awd-readiness-override-reason').setValue('teacher drill')
    await wrapper.get('#awd-readiness-override-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenNthCalledWith(
      2,
      'contest-1',
      expect.objectContaining({
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      }),
      { suppressErrorToast: true }
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该允许管理员在竞赛编辑页编排题目', async () => {
    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-edit-tab-challenges').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminContestChallenges).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('题目编排')
    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 100,
      status: 'published',
    })

    await wrapper.get('#contest-challenge-select').setValue('102')
    await wrapper.get('#contest-challenge-points').setValue('160')
    await wrapper.get('#contest-challenge-order').setValue('3')
    await wrapper.get('#contest-challenge-visibility').setValue('false')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createAdminContestChallenge).toHaveBeenCalledWith('contest-1', {
      challenge_id: 102,
      points: 160,
      order: 3,
      is_visible: false,
    })

    await wrapper.get('#contest-challenge-edit-link-1').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-challenge-points').setValue('140')
    await wrapper.get('#contest-challenge-order').setValue('2')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101', {
      points: 140,
      order: 2,
      is_visible: true,
    })

    await wrapper.get('#contest-challenge-remove-link-1').trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalled()
    expect(contestApiMocks.deleteAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101')
  })
})
