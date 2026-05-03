import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestAwdConfig from '../ContestAwdConfig.vue'
import contestAwdConfigSource from '../ContestAwdConfig.vue?raw'
import contestAwdConfigPageSource from '@/features/contest-awd-config/model/useContestAwdConfigPage.ts?raw'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' },
  query: { service: 'service-1' } as Record<string, unknown>,
}))
const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  listContestAWDServices: vi.fn(),
  updateContestAWDService: vi.fn(),
  runContestAWDCheckerPreview: vi.fn(),
}))
const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
    listContestAWDServices: adminApiMocks.listContestAWDServices,
    updateContestAWDService: adminApiMocks.updateContestAWDService,
    runContestAWDCheckerPreview: adminApiMocks.runContestAWDCheckerPreview,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

function buildService(overrides: Record<string, unknown> = {}) {
  return {
    id: 'service-1',
    contest_id: 'contest-1',
    awd_challenge_id: '501',
    title: 'Bank Portal',
    category: 'web',
    difficulty: 'easy',
    display_name: 'Bank Portal',
    order: 1,
    is_visible: true,
    score_config: {
      points: 120,
      awd_sla_score: 1,
      awd_defense_score: 2,
    },
    runtime_config: {
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag', expected_status: 200 },
        get_flag: {
          method: 'GET',
          path: '/api/flag',
          expected_status: 200,
          expected_substring: '{{FLAG}}',
        },
      },
    },
    checker_type: 'http_standard',
    checker_config: {
      put_flag: { method: 'PUT', path: '/api/flag', expected_status: 200 },
      get_flag: {
        method: 'GET',
        path: '/api/flag',
        expected_status: 200,
        expected_substring: '{{FLAG}}',
      },
    },
    sla_score: 1,
    defense_score: 2,
    validation_state: 'pending',
    created_at: '2026-04-01T00:00:00.000Z',
    updated_at: '2026-04-01T00:00:00.000Z',
    ...overrides,
  }
}

function mountPage() {
  return mount(ContestAwdConfig, {
    global: {
      stubs: {
        AppLoading: { template: '<div><slot /></div>' },
        AppEmpty: {
          props: ['title', 'description'],
          template:
            '<div class="app-empty-stub">{{ title }} {{ description }}<slot name="action" /></div>',
        },
      },
    },
  })
}

describe('ContestAwdConfig', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    adminApiMocks.getContest.mockReset()
    adminApiMocks.listContestAWDServices.mockReset()
    adminApiMocks.updateContestAWDService.mockReset()
    adminApiMocks.runContestAWDCheckerPreview.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    routeState.params.id = 'contest-1'
    routeState.query = { service: 'service-1' }
    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      mode: 'awd',
      status: 'registering',
    })
    adminApiMocks.listContestAWDServices.mockResolvedValue([buildService()])
  })

  it('路由页拆分为稳定的 AWD 配置子组件', () => {
    expect(contestAwdConfigSource).toContain(
      "import { useContestAwdConfigPage } from '@/features/contest-awd-config'"
    )
    expect(contestAwdConfigSource).not.toContain("from '@/api/admin/contests'")
    expect(contestAwdConfigSource).toContain(
      "import ContestAwdConfigTopbar from '@/components/platform/contest/ContestAwdConfigTopbar.vue'"
    )
    expect(contestAwdConfigSource).toContain(
      "import ContestAwdServiceDirectory from '@/components/platform/contest/ContestAwdServiceDirectory.vue'"
    )
    expect(contestAwdConfigSource).toContain(
      "import ContestAwdDebugStation from '@/components/platform/contest/ContestAwdDebugStation.vue'"
    )
    expect(contestAwdConfigSource).toContain(
      "import ContestAwdConfigFooter from '@/components/platform/contest/ContestAwdConfigFooter.vue'"
    )
    expect(contestAwdConfigSource).toContain('<ContestAwdServiceDirectory')
    expect(contestAwdConfigSource).toContain('<ContestAwdDebugStation')
    expect(contestAwdConfigPageSource).not.toContain(
      "from '@/components/platform/contest/awdCheckerConfigSupport'"
    )
    expect(contestAwdConfigPageSource).toContain(
      "import { useAwdCheckerPreviewFlow } from './useAwdCheckerPreview'"
    )
    expect(contestAwdConfigPageSource).toContain(
      "import { useAwdCheckerSaveFlow } from './useAwdCheckerSaveFlow'"
    )
    expect(contestAwdConfigPageSource).toContain(
      "import { useContestAwdConfigDataLoader } from './useContestAwdConfigDataLoader'"
    )
    expect(contestAwdConfigPageSource).toContain("from './awdCheckerLabels'")
    expect(contestAwdConfigPageSource).not.toContain('runContestAWDCheckerPreview')
    expect(contestAwdConfigPageSource).not.toContain('updateContestAWDService')
    expect(contestAwdConfigPageSource).not.toContain('getContest')
    expect(contestAwdConfigPageSource).not.toContain('listContestAWDServices')
  })

  it('使用独立页面编辑服务配置，并锁定 checker 类型', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(adminApiMocks.listContestAWDServices).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('AWD 服务配置')
    expect(wrapper.text()).toContain('Bank Portal')
    expect(wrapper.text()).toContain('HTTP 标准 Checker')
    expect(wrapper.find('#awd-challenge-config-checker-type').exists()).toBe(false)
    expect(wrapper.get('#awd-config-json-preview').text()).toContain('"put_flag"')

    await wrapper.find('input[type="number"]').setValue(3)
    await wrapper.findAll('input[type="number"]')[1].setValue(4)
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存配置'))
      ?.trigger('click')
    await flushPromises()

    expect(adminApiMocks.updateContestAWDService).toHaveBeenCalledWith('contest-1', 'service-1', {
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
      },
      awd_sla_score: 3,
      awd_defense_score: 4,
    })
  })

  it('query 中的服务不存在时回退到第一个 AWD 服务', async () => {
    routeState.query = { service: 'missing-service' }
    adminApiMocks.listContestAWDServices.mockResolvedValue([
      buildService({ id: 'service-2', display_name: 'Blog Service', order: 2 }),
    ])

    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.text()).toContain('Blog Service')
    expect(replaceMock).toHaveBeenCalledWith({
      name: 'ContestAWDConfig',
      params: { id: 'contest-1' },
      query: { service: 'service-2' },
    })
  })

  it('TCP checker 步骤默认全部收起，并支持切换展开步骤', async () => {
    adminApiMocks.listContestAWDServices.mockResolvedValue([
      buildService({
        checker_type: 'tcp_standard',
        checker_config: {
          timeout_ms: 5000,
          steps: [
            { send: 'PING\\n', expect_contains: 'PONG', timeout_ms: 1000 },
            { send_template: 'SET {{FLAG}}\\n', expect_contains: 'OK', timeout_ms: 2000 },
          ],
        },
      }),
    ])

    const wrapper = mountPage()

    await flushPromises()

    const stepPanels = wrapper.findAll('.checker-action-section--tcp')
    expect(stepPanels).toHaveLength(2)
    expect(stepPanels[0].classes()).toContain('is-collapsed')
    expect(stepPanels[1].classes()).toContain('is-collapsed')
    expect(wrapper.text()).toContain('发送 SET {{FLAG}}\\n · 期望 OK · 2000ms')

    await stepPanels[0].find('.checker-step-toggle').trigger('click')

    expect(stepPanels[0].classes()).not.toContain('is-collapsed')
    expect(stepPanels[1].classes()).toContain('is-collapsed')

    await stepPanels[1].find('.checker-step-toggle').trigger('click')

    expect(stepPanels[0].classes()).toContain('is-collapsed')
    expect(stepPanels[1].classes()).not.toContain('is-collapsed')
  })

  it('试跑完成后保存会带上预览令牌', async () => {
    adminApiMocks.runContestAWDCheckerPreview.mockResolvedValue({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: { status_reason: 'healthy' },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        awd_challenge_id: '501',
      },
      preview_token: 'token-1',
    })

    const wrapper = mountPage()

    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('试跑 Checker'))
      ?.trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存并写入试跑结果'))
      ?.trigger('click')
    await flushPromises()

    expect(adminApiMocks.runContestAWDCheckerPreview).toHaveBeenCalledWith('contest-1', {
      awd_challenge_id: 501,
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
      },
      preview_flag: 'flag{preview}',
    })
    expect(adminApiMocks.updateContestAWDService).toHaveBeenCalledWith(
      'contest-1',
      'service-1',
      expect.objectContaining({ awd_checker_preview_token: 'token-1' })
    )
  })
})
