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
        description: '基础运行环境',
        status: 'available',
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

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('镜像管理')
  })

  it('应该把镜像名称、标签和描述拆成独立列', async () => {
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

    const headers = wrapper.findAll('.image-directory-head span').map((item) => item.text())

    expect(headers).toEqual(['镜像名称', '标签', '描述', '状态', '创建时间', '操作'])
  })
})
