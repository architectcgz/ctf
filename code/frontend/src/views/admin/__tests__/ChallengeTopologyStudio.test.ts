import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeTopologyStudioPage from '@/components/admin/topology/ChallengeTopologyStudioPage.vue'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  getImages: vi.fn(),
  getChallengeTopology: vi.fn(),
  getEnvironmentTemplates: vi.fn(),
  saveChallengeTopology: vi.fn(),
  deleteChallengeTopology: vi.fn(),
  createEnvironmentTemplate: vi.fn(),
  updateEnvironmentTemplate: vi.fn(),
  deleteEnvironmentTemplate: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  success: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('ChallengeTopologyStudioPage', () => {
  beforeEach(() => {
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    toastMocks.error.mockReset()
    toastMocks.success.mockReset()
    toastMocks.warning.mockReset()

    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '双节点演练',
      category: 'web',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      created_at: '2026-03-10T00:00:00.000Z',
    })
    adminApiMocks.getImages.mockResolvedValue({
      list: [
        {
          id: 'img-1',
          name: 'ctf/web',
          tag: 'latest',
          status: 'available',
          created_at: '2026-03-10T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getChallengeTopology.mockResolvedValue({
      id: '21',
      challenge_id: '11',
      entry_node_key: 'web',
      networks: [{ key: 'default', name: '默认网络' }],
      nodes: [{ key: 'web', name: 'Web', network_keys: ['default'], service_port: 8080 }],
      links: [],
      policies: [],
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })
    adminApiMocks.getEnvironmentTemplates.mockResolvedValue([
      {
        id: '31',
        name: '双节点模板',
        description: 'web + db',
        entry_node_key: 'web',
        networks: [{ key: 'default', name: '默认网络' }],
        nodes: [{ key: 'web', name: 'Web', network_keys: ['default'] }],
        links: [],
        policies: [],
        usage_count: 3,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T02:00:00.000Z',
      },
    ])
    adminApiMocks.saveChallengeTopology.mockResolvedValue(undefined)
    adminApiMocks.deleteChallengeTopology.mockResolvedValue(undefined)
    adminApiMocks.createEnvironmentTemplate.mockResolvedValue(undefined)
    adminApiMocks.updateEnvironmentTemplate.mockResolvedValue(undefined)
    adminApiMocks.deleteEnvironmentTemplate.mockResolvedValue(undefined)

    vi.stubGlobal('confirm', vi.fn(() => true))
  })

  it('应该渲染当前挑战拓扑与模板区块', async () => {
    const wrapper = mount(ChallengeTopologyStudioPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<div><slot name="header" /><slot /><slot name="footer" /></div>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
          SectionCard: { template: '<section><slot /><slot name="footer" /></section>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('双节点演练')
    expect(wrapper.text()).toContain('双节点模板')
    expect(wrapper.text()).toContain('链路策略')
    expect(wrapper.text()).toContain('基础校验已通过')
    expect(wrapper.text()).toContain('当前模板')
  })

  it('应该使用统一的工作区壳层与右侧上下文轨道', async () => {
    const wrapper = mount(ChallengeTopologyStudioPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<div><slot name="header" /><slot /><slot name="footer" /></div>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
          SectionCard: { template: '<section><slot /><slot name="footer" /></section>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.workspace-shell').exists()).toBe(true)
    expect(wrapper.find('.workspace-topbar').exists()).toBe(true)
    expect(wrapper.find('.content-pane').exists()).toBe(true)
    expect(wrapper.find('.context-rail').exists()).toBe(true)
  })

  it('删除拓扑失败时应优先展示接口返回消息', async () => {
    adminApiMocks.deleteChallengeTopology.mockRejectedValue(
      new ApiError('拓扑已被实例引用，暂时不能删除', { status: 409 })
    )

    const wrapper = mount(ChallengeTopologyStudioPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<div><slot name="header" /><slot /><slot name="footer" /></div>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
          SectionCard: { template: '<section><slot /><slot name="footer" /></section>' },
        },
      },
    })

    await flushPromises()

    const deleteButton = wrapper.findAll('button').find((button) => button.text().includes('删除已保存拓扑'))
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('拓扑已被实例引用，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除题目拓扑失败')
  })

  it('删除模板失败时应优先展示接口返回消息', async () => {
    adminApiMocks.deleteEnvironmentTemplate.mockRejectedValue(
      new ApiError('模板仍被题目使用，暂时不能删除', { status: 409 })
    )

    const wrapper = mount(ChallengeTopologyStudioPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<div><slot name="header" /><slot /><slot name="footer" /></div>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
          SectionCard: { template: '<section><slot /><slot name="footer" /></section>' },
        },
      },
    })

    await flushPromises()

    const deleteButton = wrapper.findAll('button').find((button) => button.text().includes('删除模板'))
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('模板仍被题目使用，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除模板失败')
  })
})
