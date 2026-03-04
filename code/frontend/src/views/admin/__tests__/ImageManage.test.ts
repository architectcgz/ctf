import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ImageManage from '../ImageManage.vue'

vi.mock('@/api/admin', () => ({
  getImages: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        name: 'ubuntu',
        tag: '22.04',
        source_type: 'registry',
        status: 'ready',
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  }),
  createImage: vi.fn(),
  deleteImage: vi.fn(),
}))

describe('ImageManage', () => {
  it('应该渲染镜像管理页面', async () => {
    const wrapper = mount(ImageManage, {
      global: {
        stubs: {
          ElTable: true,
          ElTableColumn: true,
          ElButton: true,
          ElPagination: true,
          ElDialog: true,
          ElForm: true,
          ElFormItem: true,
          ElInput: true,
          ElSelect: true,
          ElOption: true,
        },
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
  })
})
