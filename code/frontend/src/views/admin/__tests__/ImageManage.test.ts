import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageManage from '../ImageManage.vue'

const { getImagesMock, createImageMock, deleteImageMock } = vi.hoisted(() => ({
  getImagesMock: vi.fn(),
  createImageMock: vi.fn(),
  deleteImageMock: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  getImages: getImagesMock,
  createImage: createImageMock,
  deleteImage: deleteImageMock,
}))

function createImageItem(status: 'pending' | 'building' | 'available' | 'failed' = 'available') {
  return {
    id: '1',
    name: 'ubuntu',
    tag: '22.04',
    description: '基础运行环境',
    status,
    created_at: '2024-01-01T00:00:00Z',
  }
}

function createImagePage(status: 'pending' | 'building' | 'available' | 'failed' = 'available') {
  return {
    list: [createImageItem(status)],
    total: 1,
    page: 1,
    page_size: 20,
  }
}

function mountPage() {
  return mount(ImageManage, {
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
}

describe('ImageManage', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    getImagesMock.mockReset()
    createImageMock.mockReset()
    deleteImageMock.mockReset()
    getImagesMock.mockResolvedValue(createImagePage())
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('应该渲染镜像管理页面', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('镜像管理')
  })

  it('应该把镜像名称、标签和描述拆成独立列', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const headers = wrapper.findAll('.image-directory-head span').map((item) => item.text())

    expect(headers).toEqual(['镜像名称', '标签', '描述', '状态', '创建时间', '操作'])
  })

  it('应该支持手动刷新镜像列表', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const refreshButton = wrapper.find('[data-testid="image-refresh-button"]')

    expect(refreshButton.exists()).toBe(true)

    await refreshButton.trigger('click')
    await flushPromises()

    expect(getImagesMock).toHaveBeenCalledTimes(2)
  })

  it('当没有进行中镜像时不应该继续自动轮询', async () => {
    mountPage()

    await flushPromises()

    vi.advanceTimersByTime(10000)
    await flushPromises()

    expect(getImagesMock).toHaveBeenCalledTimes(1)
  })

  it('当存在进行中镜像时应该继续自动轮询', async () => {
    getImagesMock.mockReset()
    getImagesMock.mockResolvedValue(createImagePage('building'))

    mountPage()

    await flushPromises()

    vi.advanceTimersByTime(10000)
    await flushPromises()

    expect(getImagesMock).toHaveBeenCalledTimes(2)
  })
})
