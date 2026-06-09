import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageManageRoutePage from '@/pages/platform/ImageManageRoutePage.vue'
import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'
import imageCreateModalSource from '@/features/platform/image-management/ui/ImageCreateModal.vue?raw'
import imageDetailModalSource from '@/features/platform/image-management/ui/ImageDetailModal.vue?raw'
import imageDirectoryPanelSource from '@/features/platform/image-management/ui/ImageDirectoryPanel.vue?raw'
import imageManageHeroPanelSource from '@/features/platform/image-management/ui/ImageManageHeroPanel.vue?raw'
import imageManagePageSource from '@/features/platform/image-management/model/useImageManagePage.ts?raw'
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

vi.mock('@/api/admin/authoring', () => ({
  getImages: getImagesMock,
  createImage: createImageMock,
  deleteImage: deleteImageMock,
}))
vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
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
  return mount(ImageManageRoutePage, {
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

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const combinedSource = [
  imageManageSource,
  imageDirectoryPanelSource,
  imageManageHeroPanelSource,
].join('\n')

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
    expect(wrapper.text()).toContain('镜像管理')
  })

  it('应在头部展示轻量状态条而不是总量卡片', () => {
    expect(imageManageSource).toContain("from '@/features/platform/image-management'")
    expect(imageManageSource).toContain('ImageManageHeroPanel')
    expect(imageManageSource).toContain('<ImageManageHeroPanel')
    expect(imageManageHeroPanelSource).toContain('data-testid="image-status-pill"')
    expect(imageManageHeroPanelSource).toContain('{{ refreshHint }}')
    expect(imageManageHeroPanelSource).not.toContain('镜像总量')
    expect(imageManageHeroPanelSource).not.toContain('当前查询结果的镜像总数')
    expect(imageManageHeroPanelSource).not.toContain('这一页已加载的镜像数量')
  })

  it('创建镜像弹窗应改用后台表单原语而不是 Element Plus 表单', () => {
    expect(imageManageSource).toContain("from '@/features/platform/image-management'")
    expect(imageManageSource).toContain('ImageCreateModal')
    expect(imageManageSource).toContain('<ImageCreateModal')
    expect(imageCreateModalSource).toContain(
      "from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(imageCreateModalSource).toContain("from '@/entities/image'")
    expect(imageCreateModalSource).not.toContain("from '@/api/admin/authoring'")
    expect(imageCreateModalSource).toContain('<AdminSurfaceModal')
    expect(imageCreateModalSource).not.toContain('<ElForm')
    expect(imageCreateModalSource).not.toContain('<ElFormItem')
    expect(imageCreateModalSource).not.toContain('<ElInput')
    expect(imageCreateModalSource).toContain("type { ImageCreateForm }")
    expect(imageCreateModalSource).toContain("emit('update:name'")
    expect(imageCreateModalSource).toContain("emit('update:tag'")
    expect(imageCreateModalSource).toContain("emit('update:description'")
    expect(imageManagePageSource).toContain(
      "import { useImageManageAutoRefresh } from './useImageManageAutoRefresh'"
    )
    expect(imageManagePageSource).toContain(
      "import { useImageManageMutations } from './useImageManageMutations'"
    )
    expect(imageManagePageSource).toContain("from './imageManagePresentation'")
    expect(imageManagePageSource).toContain('filterAndSortImages(')
    expect(imageManagePageSource).toContain('buildImageStatusSummary(')
    expect(imageManagePageSource).not.toContain('createImage(')
    expect(imageManagePageSource).not.toContain('deleteImage(')
  })

  it('镜像详情弹窗应抽到独立平台组件并保留后台 surface modal', () => {
    expect(imageManageSource).toContain("from '@/features/platform/image-management'")
    expect(imageManageSource).toContain('ImageDetailModal')
    expect(imageManageSource).toContain('<ImageDetailModal')
    expect(imageDetailModalSource).toContain(
      "from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(imageDetailModalSource).toContain('<AdminSurfaceModal')
    expect(imageDetailModalSource).toContain('getStatusLabel')
    expect(imageDetailModalSource).toContain('getStatusStyle')
  })

  it('应该把镜像名称、标签、来源和摘要拆成独立列', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const headers = wrapper.findAll('.workspace-data-table__head-cell').map((item) => item.text())

    expect(headers).toEqual(['镜像名称', '标签', '来源', '摘要', '状态', '验证时间', '操作'])

    const row = wrapper.find('.image-row')
    expect(row.find('.image-row__name').attributes('title')).toBe('ubuntu')
    expect(row.find('.image-row__tag').attributes('title')).toBe('22.04')
    expect(row.find('.image-row__description').attributes('title')).toBe('基础运行环境')
  })

  it('应接入共享目录工具栏和表格，并支持关键词筛选与排序', async () => {
    getImagesMock.mockResolvedValue({
      list: [
        {
          id: '1',
          name: 'zeta',
          tag: '22.04',
          description: 'Zeta image',
          status: 'available',
          created_at: '2024-01-02T00:00:00Z',
        },
        {
          id: '2',
          name: 'alpha',
          tag: '24.04',
          description: 'Alpha image',
          status: 'building',
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    expect(combinedSource).toContain("from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'")
    expect(combinedSource).toContain("from '@/shared/ui/common/WorkspaceDataTable.vue'")
    expect(combinedSource).toContain('<WorkspaceDirectoryToolbar')
    expect(combinedSource).toContain('<WorkspaceDataTable')

    const wrapper = mountPage()
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="检索镜像名称、标签或说明..."]')
    await searchInput.setValue('alp')
    await flushPromises()

    let titles = wrapper.findAll('.image-row__name').map((item) => item.text())
    expect(titles).toEqual(['alpha'])

    await searchInput.setValue('')
    await flushPromises()
    await wrapper.get('.workspace-directory-toolbar__sort-button').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('.workspace-directory-toolbar__menu-item')
      .find((item) => item.text().includes('镜像名称 A-Z'))
      ?.trigger('click')
    await flushPromises()

    titles = wrapper.findAll('.image-row__name').map((item) => item.text())
    expect(titles).toEqual(['alpha', 'zeta'])
  })

  it('镜像目录头部应继续由目录 panel owner 承接标题而不是回退到旧提示壳层', () => {
    expect(combinedSource).toContain('镜像列表')
    expect(combinedSource).not.toContain('image-board__hint')
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

  it('当前页无异常或进行中镜像时不应继续重复展示总量状态条', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const pills = wrapper.findAll('[data-testid="image-status-pill"]')

    expect(pills).toHaveLength(0)
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

    const deleteButton = wrapper
      .findAll('.image-row__actions button')
      .find((button) => button.text().trim() === '删除')
    expect(deleteButton).toBeTruthy()
    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('镜像仍被题目使用，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除失败')
  })

  it('删除镜像时应在本地 owner 上短路重复点击，并禁用对应行按钮', async () => {
    const deleteDeferred = createDeferred<void>()
    deleteImageMock.mockReturnValue(deleteDeferred.promise)

    const wrapper = mountPage()
    await flushPromises()

    const deleteButton = wrapper
      .findAll('.image-row__actions button')
      .find((button) => button.text().trim() === '删除')
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await deleteButton!.trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalledTimes(1)
    expect(deleteImageMock).toHaveBeenCalledTimes(1)
    expect(deleteButton!.attributes('disabled')).toBeDefined()
    expect(deleteButton!.text()).toContain('删除中...')

    deleteDeferred.resolve()
    await flushPromises()

    const reenabledDeleteButton = wrapper
      .findAll('.image-row__actions button')
      .find((button) => button.text().trim() === '删除')

    expect(toastMocks.success).toHaveBeenCalledWith('删除成功')
    expect(reenabledDeleteButton).toBeTruthy()
    expect(reenabledDeleteButton!.attributes('disabled')).toBeUndefined()
    expect(reenabledDeleteButton!.text()).toContain('删除')
  })
})
