import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'

import ContestChallengeOrchestrationPanel from '../contest/ContestChallengeOrchestrationPanel.vue'
import type { AdminContestChallengeViewData } from '@/api/contracts'

const contestApiMocks = vi.hoisted(() => ({
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  getChallenges: vi.fn(),
  listAdminAwdChallenges: vi.fn(),
  createContestAWDService: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateContestAWDService: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteContestAWDService: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    getChallenges: contestApiMocks.getChallenges,
    listAdminAwdChallenges: contestApiMocks.listAdminAwdChallenges,
    createContestAWDService: contestApiMocks.createContestAWDService,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
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
    awd_challenge_id: '11',
    display_name: 'Web 入门',
    order: 1,
    is_visible: true,
    score_config: {
      points: 120,
      awd_sla_score: 1,
      awd_defense_score: 2,
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
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('ContestChallengeOrchestrationPanel', () => {
  beforeEach(() => {
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.listContestAWDServices.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.listAdminAwdChallenges.mockReset()
    contestApiMocks.createContestAWDService.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateContestAWDService.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteContestAWDService.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
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

  it('应该把题目编排表的题目标题链接到管理员题目详情页', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([buildChallenge()])

    const wrapper = mountPanel()

    await flushPromises()

    const titleLink = wrapper.findAllComponents(RouterLinkStub).find((link) => link.text() === 'Web 入门')

    expect(titleLink?.props('to')).toEqual({
      name: 'PlatformChallengeDetail',
      params: { id: '101' },
    })
  })

  it('AWD 模式下题目编排只展示题目资源和比赛编排字段', async () => {
    contestApiMocks.listContestAWDServices.mockResolvedValue([buildAwdService()])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('题目资源')
    expect(wrapper.text()).toContain('可见性')
    expect(wrapper.text()).toContain('分值')
    expect(wrapper.text()).toContain('顺序')
    expect(wrapper.text()).not.toContain('Checker')
    expect(wrapper.text()).not.toContain('验证状态')
    expect(wrapper.text()).not.toContain('SLA / 防守分')
    expect(wrapper.text()).not.toContain('最近试跑')
    expect(wrapper.text()).not.toContain('HTTP 标准 Checker')
    expect(wrapper.find('.contest-challenge-panel__summary').exists()).toBe(false)
    expect(wrapper.find('.contest-challenge-filters').exists()).toBe(false)
    expect(contestApiMocks.listAdminContestChallenges).not.toHaveBeenCalled()
  })

  it('AWD 题目行操作应直接显示编辑和移除', async () => {
    contestApiMocks.listContestAWDServices.mockResolvedValue([buildAwdService()])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    expect(wrapper.find('#contest-challenge-actions-11').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-open-awd-config-11').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('AWD 编排')
    expect(wrapper.get('#contest-challenge-edit-11').classes()).toContain('ui-row-action--default')
    expect(wrapper.get('#contest-challenge-remove-11').classes()).toContain('ui-btn--danger')
  })

  it('AWD 模式下题目编排只读取 AWD 服务列表，不读取普通 Jeopardy 题目关系', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        id: 'jeopardy-link',
        challenge_id: '909',
        title: '不应出现的 Jeopardy 题',
      }),
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      buildAwdService({
        id: 'service-1',
        display_name: 'AWD Bank Portal',
      }),
    ])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    expect(contestApiMocks.listContestAWDServices).toHaveBeenCalledWith('contest-1')
    expect(contestApiMocks.listAdminContestChallenges).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('AWD Bank Portal')
    expect(wrapper.text()).not.toContain('不应出现的 Jeopardy 题')
  })

  it('AWD 题目编排不应混入 Checker 状态筛选', async () => {
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      buildAwdService({
        id: 'service-1',
        display_name: '未配 Checker 题目',
        checker_type: undefined,
        validation_state: undefined,
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
      }),
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

    expect(wrapper.text()).toContain('未配 Checker 题目')
    expect(wrapper.text()).toContain('最近失败题目')
    expect(wrapper.text()).toContain('最近通过题目')
    expect(wrapper.find('#contest-challenge-filter-unconfigured').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-filter-validation-failed').exists()).toBe(false)
  })

  it('应该在 AWD 题目池新增弹层中从 AWD 题库创建 service', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([])
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 题目',
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
      challenge_id: '11',
      awd_challenge_id: '11',
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

    expect(wrapper.find('.admin-surface-modal-stub').exists()).toBe(true)
    expect(wrapper.find('#contest-challenge-library').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-select').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-template').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-challenge-option-11').exists()).toBe(true)
    expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: 'published',
    })
    expect(contestApiMocks.getChallenges).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('upload-http')
    expect(wrapper.text()).toContain('已就绪')
    expect(wrapper.text()).not.toContain('模板')

    await wrapper.get('#contest-awd-challenge-option-11').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-visibility').exists()).toBe(false)

    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createContestAWDService).toHaveBeenCalledWith('contest-1', {
      awd_challenge_id: 11,
      points: 100,
      order: 0,
      is_visible: true,
    })
    expect(contestApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()
    expect(contestApiMocks.createAdminContestChallenge).not.toHaveBeenCalled()
  })

  it('应该在 AWD 题目池新增弹层中一次关联多个 AWD 题目', async () => {
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 题目',
          slug: 'upload-http',
          category: 'web',
          difficulty: 'medium',
          description: 'http service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          checker_type: 'http_standard',
          checker_config: {
            put_flag: {
              method: 'PUT',
              path: '/api/flag',
              expected_status: 200,
              body_template: '{{FLAG}}',
            },
            get_flag: {
              method: 'GET',
              path: '/api/flag',
              expected_status: 200,
              expected_substring: '{{FLAG}}',
            },
            havoc: {
              method: 'GET',
              path: '/health',
              expected_status: 200,
            },
          },
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
        {
          id: '12',
          name: 'IoT TCP 题目',
          slug: 'iot-tcp',
          category: 'misc',
          difficulty: 'easy',
          description: 'tcp service',
          service_type: 'binary_tcp',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          checker_type: 'legacy_probe',
          checker_config: {
            health_path: '/healthz',
          },
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.createContestAWDService.mockResolvedValue({
      id: 'service-2',
      contest_id: 'contest-1',
      challenge_id: '11',
      awd_challenge_id: '11',
      display_name: 'Upload Service',
      order: 0,
      is_visible: true,
      created_at: '2026-03-10T01:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-awd-challenge-option-12').trigger('click')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createContestAWDService).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.createContestAWDService).toHaveBeenNthCalledWith(1, 'contest-1', {
      awd_challenge_id: 11,
      points: 100,
      order: 0,
      is_visible: true,
      checker_type: 'http_standard',
      checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
          expected_status: 200,
          body_template: '{{FLAG}}',
        },
        get_flag: {
          method: 'GET',
          path: '/api/flag',
          expected_status: 200,
          expected_substring: '{{FLAG}}',
        },
        havoc: {
          method: 'GET',
          path: '/health',
          expected_status: 200,
        },
      },
    })
    expect(contestApiMocks.createContestAWDService).toHaveBeenNthCalledWith(2, 'contest-1', {
      awd_challenge_id: 12,
      points: 100,
      order: 1,
      is_visible: true,
      checker_type: 'legacy_probe',
      checker_config: {
        health_path: '/healthz',
      },
    })
  })

  it('批量关联部分成功时应提示失败项、保留弹层并刷新成功项', async () => {
    const awdServicesState: any[] = []

    contestApiMocks.listContestAWDServices.mockImplementation(async () =>
      awdServicesState.map((item) => ({ ...item }))
    )
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 题目',
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
        {
          id: '12',
          name: 'IoT TCP 题目',
          slug: 'iot-tcp',
          category: 'misc',
          difficulty: 'easy',
          description: 'tcp service',
          service_type: 'binary_tcp',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createContestAWDService
      .mockImplementationOnce(async (_contestId, payload) => {
        awdServicesState.push({
          id: 'service-11',
          contest_id: 'contest-1',
          awd_challenge_id: String(payload.awd_challenge_id),
          display_name: 'Upload HTTP 题目',
          order: payload.order,
          is_visible: payload.is_visible,
          score_config: {
            points: payload.points,
            awd_sla_score: 0,
            awd_defense_score: 0,
          },
          runtime_config: {},
          checker_type: undefined,
          checker_config: {},
          sla_score: 0,
          defense_score: 0,
          validation_state: 'pending',
          last_preview_at: undefined,
          last_preview_result: undefined,
          created_at: '2026-03-10T01:00:00.000Z',
          updated_at: '2026-03-10T01:00:00.000Z',
        })
        return awdServicesState[0]
      })
      .mockRejectedValueOnce(new Error('second failed'))

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-12').trigger('click')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createContestAWDService).toHaveBeenCalledTimes(2)
    expect(toastMocks.warning).toHaveBeenCalledWith('部分 AWD 题目关联失败：IoT TCP 题目')
    expect(toastMocks.success).not.toHaveBeenCalledWith('题目已保存')
    expect(wrapper.find('.admin-surface-modal-stub').exists()).toBe(true)
    expect(wrapper.text()).toContain('Upload HTTP 题目')
  })

  it('批量关联全部失败时应提示错误且保持弹层打开', async () => {
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 题目',
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
        {
          id: '12',
          name: 'IoT TCP 题目',
          slug: 'iot-tcp',
          category: 'misc',
          difficulty: 'easy',
          description: 'tcp service',
          service_type: 'binary_tcp',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createContestAWDService
      .mockRejectedValueOnce(new Error('first failed'))
      .mockRejectedValueOnce(new Error('second failed'))

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-12').trigger('click')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createContestAWDService).toHaveBeenCalledTimes(2)
    expect(toastMocks.error).toHaveBeenCalledWith('部分 AWD 题目关联失败：Upload HTTP 题目、IoT TCP 题目')
    expect(toastMocks.success).not.toHaveBeenCalledWith('题目已保存')
    expect(wrapper.find('.admin-surface-modal-stub').exists()).toBe(true)
  })

  it('应该按关键词和筛选条件加载 AWD 题库候选', async () => {
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-keyword').setValue('bank')
    await wrapper.get('#contest-awd-challenge-service-type').setValue('web_http')
    await wrapper.get('#contest-awd-challenge-deployment-mode').setValue('single_container')
    await wrapper.get('#contest-awd-challenge-readiness').setValue('passed')
    await flushPromises()

    expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      keyword: 'bank',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      readiness_status: 'passed',
      status: 'published',
    })
  })

  it('应该支持 AWD 题库分页翻页', async () => {
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '21',
          name: 'Page 1 AWD',
          slug: 'page-1-awd',
          category: 'web',
          difficulty: 'easy',
          description: 'page 1',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 45,
      page: 1,
      page_size: 20,
    })

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-next-page').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 2,
        page_size: 20,
      }),
    )
  })

  it('应该在 AWD 题目池编辑时同步更新 service 关联题目，并仅更新关系层分值', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge(),
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      buildAwdService({
        id: 'service-1',
        awd_challenge_id: '11',
      }),
    ])
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: '旧题目',
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
          name: '新题目',
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
    await wrapper.get('#contest-challenge-edit-11').trigger('click')
    await flushPromises()

    expect(wrapper.find('#contest-awd-challenge-list').exists()).toBe(false)
    await wrapper.get('#contest-challenge-points').setValue('140')
    await wrapper.get('#contest-challenge-order').setValue('2')
    await wrapper.get('#contest-challenge-visibility').setValue('false')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContestAWDService).toHaveBeenCalledWith('contest-1', 'service-1', {
      awd_challenge_id: 11,
      points: 140,
      order: 2,
      is_visible: false,
    })
    expect(contestApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()
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
