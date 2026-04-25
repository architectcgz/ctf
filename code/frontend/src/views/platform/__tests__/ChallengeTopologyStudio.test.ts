import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeTopologyStudioPage from '@/components/platform/topology/ChallengeTopologyStudioPage.vue'
import challengeTopologyStudioPageSource from '@/components/platform/topology/ChallengeTopologyStudioPage.vue?raw'
import topologyTemplateSidePanelSource from '@/components/platform/topology/TopologyTemplateSidePanel.vue?raw'
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
const destructiveConfirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

describe('ChallengeTopologyStudioPage', () => {
  beforeEach(() => {
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    toastMocks.error.mockReset()
    toastMocks.success.mockReset()
    toastMocks.warning.mockReset()
    destructiveConfirmMock.mockReset()
    destructiveConfirmMock.mockResolvedValue(true)

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
    expect(wrapper.classes()).toContain('journal-shell-admin')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).not.toContain('teacher-management-shell')
    expect(wrapper.classes()).not.toContain('teacher-surface')
    expect(wrapper.classes()).not.toContain('teacher-surface-workspace-bg')
  })

  it('管理员挑战拓扑工作台不应继续复用教师端根壳 token', () => {
    expect(challengeTopologyStudioPageSource).not.toContain('teacher-management-shell')
    expect(challengeTopologyStudioPageSource).not.toContain('teacher-surface')
    expect(challengeTopologyStudioPageSource).not.toContain('teacher-surface-workspace-bg')
  })

  it('应使用共享 ui-btn 原语而不是拓扑页私有按钮族', () => {
    const topologySource = `${challengeTopologyStudioPageSource}\n${topologyTemplateSidePanelSource}`

    expect(challengeTopologyStudioPageSource).toContain(
      'class="ui-btn ui-btn--ghost topology-action-btn'
    )
    expect(challengeTopologyStudioPageSource).toContain(
      'class="ui-btn ui-btn--primary topology-action-btn'
    )
    expect(topologySource).toContain('ui-btn ui-btn--secondary topology-action-btn')
    expect(topologySource).toContain('ui-btn ui-btn--danger topology-action-btn')
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

    const deleteButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('删除已保存拓扑'))
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalledWith({
      title: '删除题目拓扑',
      message: '确认删除当前题目已保存的拓扑吗？删除后需要重新保存才能恢复。',
      confirmButtonText: '确认删除',
    })
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

    const deleteButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('删除模板'))
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('模板仍被题目使用，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除模板失败')
  })
})
