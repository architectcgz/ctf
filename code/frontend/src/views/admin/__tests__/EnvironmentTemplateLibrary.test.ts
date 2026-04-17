import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import EnvironmentTemplateLibrary from '../EnvironmentTemplateLibrary.vue'
import ChallengeTopologyStudioPage from '@/components/admin/topology/ChallengeTopologyStudioPage.vue'
import challengeTopologyStudioPageSource from '@/components/admin/topology/ChallengeTopologyStudioPage.vue?raw'
import topologyNodeEditorSource from '@/components/admin/topology/TopologyNodeEditor.vue?raw'

vi.mock('@/api/admin', () => ({
  getChallengeDetail: vi.fn(),
  getImages: vi.fn().mockResolvedValue({
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
  }),
  getChallengeTopology: vi.fn(),
  getEnvironmentTemplates: vi.fn().mockResolvedValue([
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
  ]),
  saveChallengeTopology: vi.fn(),
  deleteChallengeTopology: vi.fn(),
  createEnvironmentTemplate: vi.fn(),
  updateEnvironmentTemplate: vi.fn(),
  deleteEnvironmentTemplate: vi.fn(),
}))

describe('EnvironmentTemplateLibrary', () => {
  it('页面应该直接挂载拓扑工作台，而不是再经过中间包装组件', () => {
    const wrapper = mount(EnvironmentTemplateLibrary, {
      global: {
        stubs: {
          ChallengeTopologyStudioPage: {
            template: '<div data-testid="topology-studio-page" />',
          },
        },
      },
    })

    expect(wrapper.find('[data-testid="topology-studio-page"]').exists()).toBe(true)
  })

  it('应该渲染独立模板库入口和编辑动作', async () => {
    const wrapper = mount(ChallengeTopologyStudioPage, {
      props: {
        mode: 'template-library',
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

    expect(wrapper.text()).toContain('环境模板库')
    expect(wrapper.text()).toContain('双节点模板')
    expect(wrapper.text()).toContain('载入编辑')
    expect(wrapper.text()).toContain('新建空白模板')
    expect(wrapper.text()).not.toContain('应用到挑战')
  })

  it('模板库概览卡片应补齐统一的说明文案', () => {
    expect(challengeTopologyStudioPageSource).toContain(
      'class="topology-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(challengeTopologyStudioPageSource).toContain(
      'class="topology-summary-tile progress-card metric-panel-card"'
    )
    expect(challengeTopologyStudioPageSource).toContain(
      'class="topology-summary-helper progress-card-hint metric-panel-helper"'
    )
    expect(challengeTopologyStudioPageSource).toContain('当前模板草稿中的网络数量')
    expect(challengeTopologyStudioPageSource).toContain('当前模板草稿中的节点数量')
    expect(challengeTopologyStudioPageSource).toContain('当前模板草稿中的连线数量')
    expect(challengeTopologyStudioPageSource).toContain('当前模板草稿中的策略数量')
  })

  it('节点编辑器应改用后台共享字段与按钮原语', () => {
    expect(topologyNodeEditorSource).toContain('class="ui-field')
    expect(topologyNodeEditorSource).toContain('class="ui-control-wrap')
    expect(topologyNodeEditorSource).toContain('class="ui-control')
    expect(topologyNodeEditorSource).toContain('class="ui-btn ui-btn--secondary')
    expect(topologyNodeEditorSource).toContain('class="ui-btn ui-btn--danger')
    expect(topologyNodeEditorSource).not.toContain('rounded-xl border border-border bg-elevated')
  })

  it('模板库页头操作不应保留挑战模式按钮样式分支', () => {
    expect(challengeTopologyStudioPageSource).toContain('class="topology-toolbar-btn topology-toolbar-btn--ghost"')
    expect(challengeTopologyStudioPageSource).toContain('class="topology-toolbar-btn topology-toolbar-btn--primary"')
    expect(challengeTopologyStudioPageSource).not.toContain(
      "isTemplateLibraryMode\n            ? 'topology-toolbar-btn topology-toolbar-btn--ghost'\n            : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'"
    )
    expect(challengeTopologyStudioPageSource).not.toContain(
      "isTemplateLibraryMode\n            ? 'topology-toolbar-btn topology-toolbar-btn--primary'\n            : 'inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90'"
    )
  })

  it('工作台分区底部新增按钮应复用统一工具栏按钮原语', () => {
    expect(challengeTopologyStudioPageSource).toMatch(
      /class="topology-toolbar-btn topology-toolbar-btn--ghost"\s+@click="addNode"[\s\S]*添加节点/
    )
    expect(challengeTopologyStudioPageSource).toMatch(
      /class="topology-toolbar-btn topology-toolbar-btn--ghost"\s+@click="addNetwork"[\s\S]*添加网络/
    )
    expect(challengeTopologyStudioPageSource).toMatch(
      /class="topology-toolbar-btn topology-toolbar-btn--ghost"\s+@click="addLink"[\s\S]*添加连线/
    )
    expect(challengeTopologyStudioPageSource).toMatch(
      /class="topology-toolbar-btn topology-toolbar-btn--ghost"\s+@click="addPolicy"[\s\S]*添加策略/
    )
    expect(challengeTopologyStudioPageSource).not.toMatch(
      /class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"\s+@click="add(Node|Network|Link|Policy)"/
    )
  })
})
