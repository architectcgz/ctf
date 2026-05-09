import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageManage from '../ImageManage.vue'
import imageManageSource from '../ImageManage.vue?raw'
import imageCreateModalSource from '@/components/platform/images/ImageCreateModal.vue?raw'
import imageDetailModalSource from '@/components/platform/images/ImageDetailModal.vue?raw'
import imageDirectoryPanelSource from '@/components/platform/images/ImageDirectoryPanel.vue?raw'
import imageManageHeroPanelSource from '@/components/platform/images/ImageManageHeroPanel.vue?raw'
import imageManagePageSource from '@/features/image-management/model/useImageManagePage.ts?raw'
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
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('镜像管理')
  })

  it('应在头部展示轻量状态条而不是总量卡片', () => {
    expect(imageManageSource).toContain(
      "import ImageManageHeroPanel from '@/components/platform/images/ImageManageHeroPanel.vue'"
    )
    expect(imageManageSource).toContain('<ImageManageHeroPanel')
    expect(imageManageHeroPanelSource).toContain('class="image-status-strip"')
    expect(imageManageHeroPanelSource).toContain('data-testid="image-status-pill"')
    expect(imageManageHeroPanelSource).toMatch(
      /<div class="image-status-strip__note">\s*\{\{ refreshHint \}\}\s*<\/div>/
    )
    expect(imageManageHeroPanelSource).toMatch(
      /<div class="workspace-overline">\s*Image Registry\s*<\/div>/
    )
    expect(imageManageHeroPanelSource).not.toContain(
      '<div class="journal-eyebrow">Image Registry</div>'
    )
    expect(imageManageHeroPanelSource).not.toContain('镜像总量')
    expect(imageManageHeroPanelSource).not.toContain('当前查询结果的镜像总数')
    expect(imageManageHeroPanelSource).not.toContain('这一页已加载的镜像数量')
  })

  it('创建镜像弹窗应改用后台表单原语而不是 Element Plus 表单', () => {
    expect(imageManageSource).toContain(
      "import ImageCreateModal from '@/components/platform/images/ImageCreateModal.vue'"
    )
    expect(imageManageSource).toContain('<ImageCreateModal')
    expect(imageCreateModalSource).toContain(
      "from '@/components/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(imageCreateModalSource).toContain("from '@/entities/image'")
    expect(imageCreateModalSource).not.toContain("from '@/api/admin/authoring'")
    expect(imageCreateModalSource).toContain('<AdminSurfaceModal')
    expect(imageCreateModalSource).not.toContain('<ElForm')
    expect(imageCreateModalSource).not.toContain('<ElFormItem')
    expect(imageCreateModalSource).not.toContain('<ElInput')
    expect(imageCreateModalSource).toContain('class="ui-field')
    expect(imageCreateModalSource).toContain('class="ui-control-wrap')
    expect(imageCreateModalSource).toContain('class="ui-control')
    expect(imageCreateModalSource).toContain('class="ui-btn ui-btn--secondary')
    expect(imageCreateModalSource).toContain('class="ui-btn ui-btn--primary')
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
    expect(imageManageSource).toContain(
      "import ImageDetailModal from '@/components/platform/images/ImageDetailModal.vue'"
    )
    expect(imageManageSource).toContain('<ImageDetailModal')
    expect(imageDetailModalSource).toContain(
      "from '@/components/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(imageDetailModalSource).toContain('<AdminSurfaceModal')
    expect(imageDetailModalSource).toContain('class="image-detail__grid"')
  })

  it('头部操作应改用共享 header-btn 原语而不是页面私有 admin-btn 按钮族', () => {
    expect(imageManageHeroPanelSource).toContain('class="header-actions image-header__actions"')
    expect(imageManageHeroPanelSource).toContain('class="header-btn header-btn--ghost"')
    expect(imageManageHeroPanelSource).toContain('class="header-btn header-btn--primary"')
    expect(imageDirectoryPanelSource).toContain('class="ui-btn ui-btn--sm ui-btn--primary"')
    expect(combinedSource).toContain('class="ui-btn ui-btn--sm ui-btn--danger"')
    expect(imageManageSource).not.toContain('admin-btn admin-btn-ghost')
    expect(imageManageSource).not.toContain('admin-btn admin-btn-primary')
    expect(imageManageSource).not.toContain('admin-btn admin-btn-danger')
  })

  it('不应在头部摘要和镜像列表之间重复渲染分割线', () => {
    expect(imageManageHeroPanelSource).toMatch(
      /\.image-header\s*\{[\s\S]*border-bottom:\s*1px solid color-mix\(in srgb, var\(--journal-border\) 88%, transparent\);/s
    )
    expect(imageManageSource).not.toContain('<div class="journal-divider image-divider" />')
    expect(imageManageSource).not.toMatch(/\.image-divider\s*\{/s)
  })

  it('立即刷新按钮应使用与目录筛选控件一致的外边框语义', () => {
    expect(imageManageHeroPanelSource).toMatch(
      /\.image-header__actions\s*>\s*\[data-testid='image-refresh-button'\]\s*\{[\s\S]*--header-btn-border:\s*var\(--image-toolbar-control-border\);[\s\S]*--header-btn-background:\s*var\(--image-toolbar-control-background\);[\s\S]*box-shadow:\s*var\(--image-toolbar-control-shadow\);/s
    )
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

    expect(combinedSource).toContain("from '@/components/common/WorkspaceDirectoryToolbar.vue'")
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(combinedSource).toContain('<WorkspaceDirectoryToolbar')
    expect(combinedSource).toContain('<WorkspaceDataTable')
    expect(combinedSource).toMatch(
      /\.image-board\s*\{[\s\S]*display:\s*grid;[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(combinedSource).toMatch(
      /\.image-board :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )

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

  it('镜像目录头部应只保留左侧标题组并恢复目录标题样式', () => {
    expect(combinedSource).toContain('<header class="list-heading image-board__head">')
    expect(combinedSource).toMatch(
      /<h2 class="list-heading__title image-section-title">\s*镜像列表\s*<\/h2>/
    )
    expect(combinedSource).not.toContain('image-board__hint')
    expect(combinedSource).toMatch(/\.image-board__head\s*\{[\s\S]*margin-bottom:\s*0;/s)
    expect(combinedSource).toMatch(
      /\.list-heading__title\s*\{[\s\S]*font-size:\s*clamp\(1\.2rem,\s*1rem\s*\+\s*0\.5vw,\s*1\.45rem\);/s
    )
  })

  it('镜像状态摘要应位于头部标题区域右侧', () => {
    expect(imageManageHeroPanelSource).toMatch(
      /<div class="image-header__intro">[\s\S]*<div class="image-header__copy">[\s\S]*<div\s+class="image-status-strip"[\s\S]*<div class="image-header__side">/s
    )
    expect(imageManageHeroPanelSource).toMatch(
      /\.image-header__intro\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+minmax\(18rem,\s*auto\);/s
    )
  })

  it('应该为镜像列表长文本保留省略样式和完整悬浮提示', () => {
    expect(combinedSource).toMatch(
      /class="image-row__name"[\s\S]*:title="\(row as AdminImageListItem\)\.name"/s
    )
    expect(combinedSource).toMatch(
      /class="image-row__tag"[\s\S]*:title="\(row as AdminImageListItem\)\.tag"/s
    )
    expect(combinedSource).toMatch(
      /class="image-row__description"[\s\S]*:title="[\s\S]*\(row as AdminImageListItem\)\.last_error \|\|[\s\S]*\(row as AdminImageListItem\)\.digest \|\|[\s\S]*\(row as AdminImageListItem\)\.description \|\|[\s\S]*'未生成摘要'/s
    )
    expect(combinedSource).toMatch(
      /\.image-row__name\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(combinedSource).toMatch(
      /\.image-row__tag\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(combinedSource).toMatch(
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
})
