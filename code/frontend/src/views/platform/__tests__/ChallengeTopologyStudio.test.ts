import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeTopologyStudioPage from '@/components/platform/topology/ChallengeTopologyStudioPage.vue'
import challengeTopologyStudioPageSource from '@/components/platform/topology/ChallengeTopologyStudioPage.vue?raw'
import topologyChallengeContextRailSource from '@/components/platform/topology/TopologyChallengeContextRail.vue?raw'
import topologyChallengeWorkspaceHeaderSource from '@/components/platform/topology/TopologyChallengeWorkspaceHeader.vue?raw'
import topologyConnectivitySectionsSource from '@/components/platform/topology/TopologyConnectivitySections.vue?raw'
import topologyCanvasQuickEditorSource from '@/components/platform/topology/TopologyCanvasQuickEditor.vue?raw'
import topologyCanvasWorkspaceSectionSource from '@/components/platform/topology/TopologyCanvasWorkspaceSection.vue?raw'
import topologyEntryNodeSectionSource from '@/components/platform/topology/TopologyEntryNodeSection.vue?raw'
import topologyPackageContextPanelSource from '@/components/platform/topology/TopologyPackageContextPanel.vue?raw'
import topologyNetworkSectionSource from '@/components/platform/topology/TopologyNetworkSection.vue?raw'
import topologyNetworkQuickEditorSource from '@/components/platform/topology/TopologyNetworkQuickEditor.vue?raw'
import topologyNodeSectionSource from '@/components/platform/topology/TopologyNodeSection.vue?raw'
import topologyTemplateHeroSectionSource from '@/components/platform/topology/TopologyTemplateHeroSection.vue?raw'
import topologyTemplateSidePanelSource from '@/components/platform/topology/TopologyTemplateSidePanel.vue?raw'
import topologyTemplateWorkbenchSource from '@/components/platform/topology/TopologyTemplateWorkbench.vue?raw'
import challengeTopologyStudioRouteSource from '../ChallengeTopologyStudio.vue?raw'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  getImages: vi.fn(),
  getChallengeTopology: vi.fn(),
  getEnvironmentTemplates: vi.fn(),
  saveChallengeTopology: vi.fn(),
  exportChallengePackage: vi.fn(),
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

vi.mock('@/api/admin/authoring', () => adminApiMocks)
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
      source_type: 'package_import',
      source_path: 'docker/topology.yml',
      sync_status: 'clean',
      package_revision_id: '501',
      last_export_revision_id: '502',
      package_baseline: {
        entry_node_key: 'web',
        networks: [{ key: 'default', name: '默认网络' }],
        nodes: [{ key: 'web', name: 'Web', network_keys: ['default'], service_port: 8080 }],
        links: [],
        policies: [],
      },
      package_files: [
        { path: 'docker/Dockerfile', size: 32 },
        { path: 'docker/topology.yml', size: 256 },
      ],
      package_revisions: [
        {
          id: '502',
          revision_no: 2,
          source_type: 'exported',
          package_slug: 'dual-node-demo',
          topology_source_path: 'docker/topology.yml',
          created_at: '2026-03-10T03:00:00.000Z',
          updated_at: '2026-03-10T03:00:00.000Z',
        },
      ],
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
    adminApiMocks.exportChallengePackage.mockResolvedValue({
      challenge_id: '11',
      revision_id: '502',
      archive_path: '/tmp/dual-node-demo.zip',
      source_dir: '/tmp/source',
      file_name: 'dual-node-demo.zip',
      download_url: '/api/v1/authoring/challenges/11/package-export/download?revision_id=502',
      created_at: '2026-03-10T03:00:00.000Z',
    })
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
    expect(wrapper.text()).toContain('题包基线已接入')
    expect(wrapper.text()).toContain('docker/topology.yml')
    expect(wrapper.text()).toContain('dual-node-demo')
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

  it('路由壳页应仅做组合，不直接绑定路由实例', () => {
    expect(challengeTopologyStudioRouteSource).toContain('useChallengeTopologyStudioRoutePage')
    expect(challengeTopologyStudioRouteSource).not.toContain('useRoute')
    expect(challengeTopologyStudioRouteSource).not.toContain('useRouter')
  })

  it('应使用共享 ui-btn 原语而不是拓扑页私有按钮族', () => {
    const topologySource = `${challengeTopologyStudioPageSource}\n${topologyTemplateSidePanelSource}\n${topologyNetworkSectionSource}\n${topologyConnectivitySectionsSource}\n${topologyNodeSectionSource}\n${topologyCanvasQuickEditorSource}\n${topologyNetworkQuickEditorSource}\n${topologyEntryNodeSectionSource}\n${topologyPackageContextPanelSource}`

    expect(topologyChallengeWorkspaceHeaderSource).toContain(
      'class="ui-btn ui-btn--ghost topology-action-btn'
    )
    expect(topologyChallengeWorkspaceHeaderSource).toContain(
      'class="ui-btn ui-btn--primary topology-action-btn'
    )
    expect(topologySource).toContain('ui-btn ui-btn--secondary topology-action-btn')
    expect(topologySource).toContain('ui-btn ui-btn--danger topology-action-btn')
  })

  it('画布工作区应从父页下沉到独立组件，同时保留父页 selection owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyCanvasWorkspaceSection')
    expect(challengeTopologyStudioPageSource).not.toContain('title="图形画布"')
    expect(challengeTopologyStudioPageSource).not.toContain('<div v-else-if="selectedNodeDraft"')
    expect(challengeTopologyStudioPageSource).not.toContain('<div v-else-if="selectedEdgeMeta"')
    expect(topologyCanvasWorkspaceSectionSource).toContain('title="图形画布"')
    expect(topologyCanvasWorkspaceSectionSource).toContain('<TopologyCanvasQuickEditor')
    expect(topologyCanvasWorkspaceSectionSource).toContain("emit('setInteractionMode'")
    expect(topologyCanvasQuickEditorSource).toContain('画布快速编辑')
    expect(topologyCanvasQuickEditorSource).toContain("emit('updateSelectedNodeField'")
    expect(topologyCanvasQuickEditorSource).toContain("emit('updateSelectedEdgeKind'")
  })

  it('网络快速编辑应从父页下沉到独立组件，同时保留 draft.networks owner', () => {
    expect(challengeTopologyStudioPageSource).not.toContain('<TopologyNetworkQuickEditor')
    expect(topologyCanvasWorkspaceSectionSource).toContain('<TopologyNetworkQuickEditor')
    expect(challengeTopologyStudioPageSource).not.toContain('v-model="network.key"')
    expect(challengeTopologyStudioPageSource).not.toContain('v-model="network.name"')
    expect(challengeTopologyStudioPageSource).not.toContain('v-model="network.internal"')
    expect(topologyNetworkQuickEditorSource).toContain('网络快速编辑')
    expect(topologyNetworkQuickEditorSource).toContain("emit('updateNetwork'")
  })

  it('入口节点卡片应从父页下沉到独立组件，同时保留 entry node owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyEntryNodeSection')
    expect(challengeTopologyStudioPageSource).not.toContain('SectionCard title="入口节点"')
    expect(challengeTopologyStudioPageSource).not.toContain('v-model="draft.entry_node_key"')
    expect(topologyEntryNodeSectionSource).toContain('title="入口节点"')
    expect(topologyEntryNodeSectionSource).toContain("emit('updateEntryNodeKey'")
  })

  it('challenge context rail 应从父页下沉到独立组件，同时保留导出与模板动作 owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyChallengeContextRail')
    expect(challengeTopologyStudioPageSource).not.toContain('SectionCard title="题包来源"')
    expect(challengeTopologyStudioPageSource).not.toContain('SectionCard title="题包文件"')
    expect(challengeTopologyStudioPageSource).not.toContain('SectionCard title="修订历史"')
    expect(topologyChallengeContextRailSource).toContain('<TopologyStatusNotes')
    expect(topologyChallengeContextRailSource).toContain('<TopologyPackageContextPanel')
    expect(topologyChallengeContextRailSource).toContain('<TopologyTemplateSidePanel')
    expect(topologyChallengeContextRailSource).toContain("emit('exportPackage')")
    expect(topologyChallengeContextRailSource).toContain("emit('update:templateKeyword'")
    expect(topologyPackageContextPanelSource).toContain('title="题包来源"')
    expect(topologyPackageContextPanelSource).toContain('title="题包文件"')
    expect(topologyPackageContextPanelSource).toContain('title="修订历史"')
    expect(topologyPackageContextPanelSource).toContain("emit('exportPackage')")
  })

  it('template workbench 应从父页下沉到独立组件，同时保留 draft 与模板动作 owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyTemplateWorkbench')
    expect(challengeTopologyStudioPageSource).not.toContain('class="template-toolbar-tab"')
    expect(challengeTopologyStudioPageSource).not.toContain("activeWorkbenchTab === 'visual'")
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyCanvasWorkspaceSection')
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyEntryNodeSection')
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyNodeSection')
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyNetworkSection')
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyConnectivitySections')
    expect(topologyTemplateWorkbenchSource).toContain('<TopologyTemplateSidePanel')
    expect(topologyTemplateWorkbenchSource).toContain("emit('update:activeWorkbenchTab'")
    expect(topologyTemplateWorkbenchSource).toContain("emit('loadTemplate'")
  })

  it('template hero 应从父页下沉到独立组件，同时保留 summary 与 status 数据 owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyTemplateHeroSection')
    expect(challengeTopologyStudioPageSource).not.toContain('class="topology-hero-grid')
    expect(challengeTopologyStudioPageSource).not.toContain('class="topology-hero-kicker"')
    expect(topologyTemplateHeroSectionSource).toContain('class="topology-hero-grid')
    expect(topologyTemplateHeroSectionSource).toContain('真实接口')
    expect(topologyTemplateHeroSectionSource).toContain('<TopologySummaryGrid')
    expect(topologyTemplateHeroSectionSource).toContain('<TopologyStatusNotes')
  })

  it('challenge header 应从父页下沉到独立组件，同时保留返回与保存动作 owner', () => {
    expect(challengeTopologyStudioPageSource).toContain('<TopologyChallengeWorkspaceHeader')
    expect(challengeTopologyStudioPageSource).not.toContain('class="workspace-topbar topology-workspace-topbar"')
    expect(challengeTopologyStudioPageSource).not.toContain('class="workspace-tab-heading topology-page-heading"')
    expect(topologyChallengeWorkspaceHeaderSource).toContain('class="workspace-topbar topology-workspace-topbar"')
    expect(topologyChallengeWorkspaceHeaderSource).toContain('class="workspace-tab-heading topology-page-heading"')
    expect(topologyChallengeWorkspaceHeaderSource).toContain("emit('back')")
    expect(topologyChallengeWorkspaceHeaderSource).toContain("emit('save')")
    expect(topologyChallengeWorkspaceHeaderSource).toContain('<TopologySummaryGrid')
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
