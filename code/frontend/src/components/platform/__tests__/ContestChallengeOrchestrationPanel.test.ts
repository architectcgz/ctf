import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import ContestChallengeOrchestrationPanel from '../contest/ContestChallengeOrchestrationPanel.vue'
import type { AdminContestChallengeViewData } from '@/api/contracts'

const contestApiMocks = vi.hoisted(() => ({
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  getChallenges: vi.fn(),
  listAdminAwdServiceTemplates: vi.fn(),
  createContestAWDService: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateContestAWDService: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteContestAWDService: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    getChallenges: contestApiMocks.getChallenges,
    listAdminAwdServiceTemplates: contestApiMocks.listAdminAwdServiceTemplates,
    createContestAWDService: contestApiMocks.createContestAWDService,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: vi.fn(),
}))

function buildChallenge(
  overrides: Partial<AdminContestChallengeViewData> = {},
): AdminContestChallengeViewData {
  return {
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
    ...overrides,
  }
}

function buildAwdService(overrides: Record<string, unknown> = {}) {
  return {
    id: 'service-1',
    contest_id: 'contest-1',
    challenge_id: '101',
    template_id: '11',
    display_name: 'Web 入门',
    order: 1,
    is_visible: true,
    score_config: {
      points: 120,
      awd_sla_score: 18,
      awd_defense_score: 28,
    },
    runtime_config: {
      checker_type: 'http_standard',
      checker_config: {},
    },
    checker_type: 'http_standard',
    checker_config: {},
    sla_score: 18,
    defense_score: 28,
    validation_state: 'stale',
    last_preview_at: '2026-04-12T08:00:00.000Z',
    last_preview_result: undefined,
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T00:00:00.000Z',
    ...overrides,
  }
}

function mountPanel(props?: Record<string, unknown>) {
  return mount(ContestChallengeOrchestrationPanel, {
    props: {
      contestId: 'contest-1',
      contestMode: 'jeopardy',
      ...props,
    },
    global: {
      stubs: {
        AppEmpty: {
          props: ['title', 'description'],
          template: '<div class="app-empty-stub">{{ title }}|{{ description }}</div>',
        },
        AppLoading: {
          template: '<div class="app-loading-stub"><slot /></div>',
        },
        AdminSurfaceModal: defineComponent({
          name: 'AdminSurfaceModal',
          props: {
            open: { type: Boolean, default: false },
          },
          setup(props, { slots }) {
            return () =>
              props.open
                ? h('div', { class: 'admin-surface-modal-stub' }, [
                    h('div', { class: 'admin-surface-modal-stub__body' }, slots.default?.()),
                    h('div', { class: 'admin-surface-modal-stub__footer' }, slots.footer?.()),
                  ])
                : null
          },
        }),
        CActionMenu: defineComponent({
          name: 'CActionMenu',
          props: {
            open: { type: Boolean, default: false },
          },
          emits: ['update:open'],
          setup(props, { slots, emit }) {
            const toggle = () => emit('update:open', !props.open)
            const close = () => emit('update:open', false)

            return () =>
              h('div', { class: 'c-action-menu-stub' }, [
                slots.trigger?.({
                  open: props.open,
                  toggle,
                  close,
                  setTriggerRef: () => undefined,
                }),
                props.open
                  ? h('div', { class: 'c-action-menu-stub__panel' }, slots.default?.({ close }))
                  : null,
              ])
          },
        }),
      },
    },
  })
}

describe('ContestChallengeOrchestrationPanel', () => {
  beforeEach(() => {
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.listContestAWDServices.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.listAdminAwdServiceTemplates.mockReset()
    contestApiMocks.createContestAWDService.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateContestAWDService.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteContestAWDService.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
    contestApiMocks.listContestAWDServices.mockResolvedValue([])
  })

  it('应该显示基础编排字段', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        id: 'link-2',
        challenge_id: '102',
        title: 'Crypto 进阶',
        order: 2,
        is_visible: false,
        points: 150,
      }),
      buildChallenge(),
    ])

    const wrapper = mountPanel()

    await flushPromises()

    expect(wrapper.text()).toContain('题目池')
    expect(wrapper.text()).toContain('题目')
    expect(wrapper.text()).toContain('可见性')
    expect(wrapper.text()).toContain('分值')
    expect(wrapper.text()).toContain('顺序')
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).toContain('第 1 位')
  })

  it('应该在 AWD 模式下显示 checker / SLA / 防守分 / 验证状态摘要列', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        awd_checker_type: undefined,
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: undefined,
        awd_checker_last_preview_at: undefined,
      }),
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([buildAwdService()])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Checker')
    expect(wrapper.text()).toContain('验证状态')
    expect(wrapper.text()).toContain('SLA')
    expect(wrapper.text()).toContain('防守分')
    expect(wrapper.text()).toContain('最近试跑')
    expect(wrapper.text()).toContain('HTTP 标准 Checker')
    expect(wrapper.text()).toContain('SLA 18 / 防守 28')
    expect(wrapper.text()).toContain('待重新验证')
  })

  it('应该支持按未配置 AWD 和预检失败筛选', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        id: 'link-1',
        challenge_id: '101',
        title: '未配 Checker 题目',
      }),
      buildChallenge({
        id: 'link-2',
        challenge_id: '102',
        title: '最近失败题目',
        order: 2,
      }),
      buildChallenge({
        id: 'link-3',
        challenge_id: '103',
        title: '最近通过题目',
        order: 3,
      }),
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      buildAwdService({
        id: 'service-2',
        challenge_id: '102',
        display_name: '最近失败题目',
        checker_type: 'legacy_probe',
        validation_state: 'failed',
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {
          checker_type: 'legacy_probe',
          checker_config: {},
        },
      }),
      buildAwdService({
        id: 'service-3',
        challenge_id: '103',
        display_name: '最近通过题目',
        checker_type: 'http_standard',
        validation_state: 'passed',
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {
          checker_type: 'http_standard',
          checker_config: {},
        },
      }),
    ])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    await wrapper.get('#contest-challenge-filter-unconfigured').trigger('click')
    expect(wrapper.text()).toContain('未配 Checker 题目')
    expect(wrapper.text()).not.toContain('最近失败题目')
    expect(wrapper.text()).not.toContain('最近通过题目')

    await wrapper.get('#contest-challenge-filter-validation-failed').trigger('click')
    expect(wrapper.text()).not.toContain('未配 Checker 题目')
    expect(wrapper.text()).toContain('最近失败题目')
    expect(wrapper.text()).not.toContain('最近通过题目')
  })

  it('应该在 AWD 题目池新增时先创建 service，再仅更新关系层分值', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([])
    contestApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '102',
          title: 'Upload Service',
          category: 'web',
          difficulty: 'medium',
          status: 'published',
          points: 150,
          created_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.listAdminAwdServiceTemplates.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 模板',
          slug: 'upload-http',
          category: 'web',
          difficulty: 'medium',
          description: 'http service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.createContestAWDService.mockResolvedValue({
      id: 'service-2',
      contest_id: 'contest-1',
      challenge_id: '102',
      template_id: '11',
      display_name: 'Upload Service',
      order: 3,
      is_visible: false,
      created_at: '2026-03-10T01:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })
    contestApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminAwdServiceTemplates).toHaveBeenCalledWith({
      page: 1,
      page_size: 100,
      status: 'published',
    })

    await wrapper.get('#contest-challenge-select').setValue('102')
    await wrapper.get('#contest-challenge-template').setValue('11')
    await wrapper.get('#contest-challenge-points').setValue('160')
    await wrapper.get('#contest-challenge-order').setValue('3')
    await wrapper.get('#contest-challenge-visibility').setValue('false')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createContestAWDService).toHaveBeenCalledWith('contest-1', {
      challenge_id: 102,
      template_id: 11,
      order: 3,
      is_visible: false,
    })
    expect(contestApiMocks.updateAdminContestChallenge).toHaveBeenCalledWith('contest-1', '102', {
      points: 160,
    })
    expect(contestApiMocks.createAdminContestChallenge).not.toHaveBeenCalled()
  })

  it('应该在 AWD 题目池编辑时同步更新 service 模板，并仅更新关系层分值', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge(),
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      buildAwdService({
        id: 'service-1',
        challenge_id: '101',
        template_id: '11',
      }),
    ])
    contestApiMocks.listAdminAwdServiceTemplates.mockResolvedValue({
      list: [
        {
          id: '11',
          name: '旧模板',
          slug: 'old-template',
          category: 'web',
          difficulty: 'easy',
          description: 'old',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
        {
          id: '12',
          name: '新模板',
          slug: 'new-template',
          category: 'web',
          difficulty: 'medium',
          description: 'new',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.1.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.updateContestAWDService.mockResolvedValue(undefined)
    contestApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-actions-101').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-edit-101').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-challenge-template').setValue('12')
    await wrapper.get('#contest-challenge-points').setValue('140')
    await wrapper.get('#contest-challenge-order').setValue('2')
    await wrapper.get('#contest-challenge-visibility').setValue('false')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContestAWDService).toHaveBeenCalledWith('contest-1', 'service-1', {
      template_id: 12,
      order: 2,
      is_visible: false,
    })
    expect(contestApiMocks.updateAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101', {
      points: 140,
    })
  })

  it('外部题目数据加载失败且尚无成功数据时应展示失败态而不是空态', async () => {
    const wrapper = mountPanel({
      contestMode: 'awd',
      challengeLinks: [],
      loadingExternal: false,
      loadErrorExternal: '同步失败',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('赛事题目暂时不可用')
    expect(wrapper.text()).toContain('同步失败')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })
})
