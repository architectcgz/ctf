import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageManage from '../ImageManage.vue'
import imageManageSource from '../ImageManage.vue?raw'
import { ApiError } from '@/api/request'

const { getImagesMock, createImageMock, deleteImageMock } = vi.hoisted(() => ({
  getImagesMock: vi.fn(),
  createImageMock: vi.fn(),
  deleteImageMock: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin', () => ({
  getImages: getImagesMock,
  createImage: createImageMock,
  deleteImage: deleteImageMock,
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
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
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    confirmMock.mockReset()
    getImagesMock.mockResolvedValue(createImagePage())
    confirmMock.mockResolvedValue(true)
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

  it('应在头部展示轻量状态条而不是总量卡片', () => {
    expect(imageManageSource).toContain('class="image-status-strip"')
    expect(imageManageSource).toContain('data-testid="image-status-pill"')
    expect(imageManageSource).toContain(
      '<div class="image-status-strip__note">{{ refreshHint }}</div>'
    )
    expect(imageManageSource).toContain('<div class="workspace-overline">Image Registry</div>')
    expect(imageManageSource).not.toContain('<div class="journal-eyebrow">Image Registry</div>')
    expect(imageManageSource).not.toContain('镜像总量')
    expect(imageManageSource).not.toContain('当前查询结果的镜像总数')
    expect(imageManageSource).not.toContain('这一页已加载的镜像数量')
  })

  it('不应在头部摘要和镜像列表之间重复渲染分割线', () => {
    expect(imageManageSource).toMatch(
      /\.image-header\s*\{[\s\S]*border-bottom:\s*1px solid color-mix\(in srgb, var\(--journal-border\) 88%, transparent\);/s
    )
    expect(imageManageSource).not.toContain('<div class="journal-divider image-divider" />')
    expect(imageManageSource).not.toMatch(/\.image-divider\s*\{/s)
  })

  it('应该把镜像名称、标签和描述拆成独立列', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const headers = wrapper.findAll('.image-directory-head span').map((item) => item.text())

    expect(headers).toEqual(['镜像名称', '标签', '描述', '状态', '创建时间', '操作'])

    const row = wrapper.find('.image-row')
    expect(row.find('.image-row__name').attributes('title')).toBe('ubuntu')
    expect(row.find('.image-row__tag').attributes('title')).toBe('22.04')
    expect(row.find('.image-row__description').attributes('title')).toBe('基础运行环境')
  })

  it('应该为镜像列表长文本保留省略样式和完整悬浮提示', () => {
    expect(imageManageSource).toMatch(/class="image-row__name"[\s\S]*:title="row\.name"/s)
    expect(imageManageSource).toMatch(/class="image-row__tag"[\s\S]*:title="row\.tag"/s)
    expect(imageManageSource).toMatch(
      /class="image-row__description"[\s\S]*:title="row\.description \|\| '未填写镜像说明'"/s
    )
    expect(imageManageSource).toMatch(
      /\.image-row__name\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(imageManageSource).toMatch(
      /\.image-row__tag\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(imageManageSource).toMatch(
      /\.image-row__description\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s
    )
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

  it('当前页无进行中镜像时应展示可用状态条', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const pills = wrapper.findAll('[data-testid="image-status-pill"]')

    expect(pills).toHaveLength(1)
    expect(pills[0].text()).toContain('可用')
    expect(pills[0].text()).toContain('1')
    expect(wrapper.find('.image-status-strip__note').text()).toContain(
      '当前无进行中镜像，可手动刷新'
    )
  })

  it('当前页存在构建中镜像时应展示状态摘要并自动刷新提示', async () => {
    getImagesMock.mockReset()
    getImagesMock.mockResolvedValue(createImagePage('building'))

    const wrapper = mountPage()

    await flushPromises()

    const pills = wrapper.findAll('[data-testid="image-status-pill"]')

    expect(pills).toHaveLength(1)
    expect(pills[0].text()).toContain('构建中')
    expect(pills[0].text()).toContain('1')
    expect(wrapper.find('.image-status-strip__note').text()).toContain(
      '构建中镜像会每 10 秒自动刷新'
    )
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

  it('删除镜像失败时应优先展示接口返回消息', async () => {
    deleteImageMock.mockRejectedValue(
      new ApiError('镜像仍被题目使用，暂时不能删除', { code: 10007, status: 409 })
    )

    const wrapper = mountPage()
    await flushPromises()

    const deleteButton = wrapper.find('.image-row__actions button')
    await deleteButton.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('镜像仍被题目使用，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除失败')
  })
})
