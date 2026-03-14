import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ChallengeTopologyStudioPage from '@/components/admin/topology/ChallengeTopologyStudioPage.vue'

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

describe('EnvironmentTemplateLibraryPage', () => {
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
})
