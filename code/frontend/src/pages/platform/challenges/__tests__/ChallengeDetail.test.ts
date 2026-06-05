import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeDetail from '@/pages/platform/challenges/ChallengeDetailRoutePage.vue'
import challengeDetailSource from '@/pages/platform/challenges/ChallengeDetailRoutePage.vue?raw'
import adminChallengeTopbarPanelSource from '@/features/platform/challenge-detail/ui/AdminChallengeTopbarPanel.vue?raw'
import platformChallengeDetailPageSource from '@/features/platform/challenge-detail/model/usePlatformChallengeDetailPage.ts?raw'
import platformChallengeDetailRoutesSource from '@/features/platform/challenge-detail/model/platformChallengeDetailRoutes.ts?raw'
import platformChallengeDetailWorkspaceSource from '@/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue?raw'
import { useBackofficeBreadcrumbDetail } from '@/shared/model/layout/useBackofficeBreadcrumbDetail'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: '11' } as Record<string, string>,
  query: {} as Record<string, string>,
}))

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  configureChallengeFlag: vi.fn(),
}))

const challengeApiMocks = vi.hoisted(() => ({
  downloadAttachment: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock, back: vi.fn() }),
  }
})

vi.mock('@/api/admin/authoring', () => adminApiMocks)
vi.mock('@/api/challenge', () => challengeApiMocks)
vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('Admin ChallengeDetail', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    challengeApiMocks.downloadAttachment.mockReset()
    routeState.params = { id: '11' }
    routeState.query = {}
    useBackofficeBreadcrumbDetail().setBreadcrumbDetailTitle()
    adminApiMocks.getChallengeDetail.mockReset()
    adminApiMocks.configureChallengeFlag.mockReset()
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '双节点演练',
      category: 'web',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      attachment_url: 'https://example.com/demo.zip',
      description: 'desc',
      hints: [{ id: 'hint-1', level: 1, title: '入口', content: '观察回显' }],
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })
    challengeApiMocks.downloadAttachment.mockResolvedValue({
      blob: new Blob(['demo']),
      filename: 'demo.zip',
    })
  })

  it('应该默认显示题目详情 tab，并保留独立的拓扑编排入口', async () => {
    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: {
            template: '<div data-testid="challenge-writeup-manage-panel">题解目录</div>',
          },
        },
      },
    })

    expect(wrapper.find('.class-chip').text()).toBe('题目详情')
    await flushPromises()

    expect(wrapper.text()).toContain('题目详情')
    expect(wrapper.text()).toContain('题解管理')
    expect(wrapper.find('#admin-challenge-tab-detail').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-challenge-panel-detail').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#admin-challenge-panel-writeup').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('双节点演练')
    expect(useBackofficeBreadcrumbDetail().breadcrumbDetailTitle.value).toBe('双节点演练')
    expect(
      wrapper
        .find(
          '.challenge-overview-summary.progress-strip.metric-panel-grid.metric-panel-default-surface'
        )
        .exists()
    ).toBe(true)
    expect(wrapper.text()).toContain('基础信息')

    const topologyButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('拓扑编排'))
    expect(topologyButton).toBeTruthy()

    await topologyButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformChallengeTopologyStudio',
      params: { id: '11' },
    })

    wrapper.unmount()
    expect(useBackofficeBreadcrumbDetail().breadcrumbDetailTitle.value).toBeNull()
  })

  it('详情路由页应继续把顶部工作区 owner 留在 challenge detail widget 上', () => {
    expect(challengeDetailSource).toContain('usePlatformChallengeDetailRoutePage')
    expect(challengeDetailSource).not.toContain('useRouteQueryTabs')
    expect(challengeDetailSource).not.toContain('useRoute')
    expect(challengeDetailSource).not.toContain('useRouter')
    expect(challengeDetailSource).toContain(
      "import { PlatformChallengeDetailWorkspace } from '@/widgets/platform-challenge-detail'"
    )
    expect(platformChallengeDetailWorkspaceSource).toContain('<AdminChallengeTopbarPanel')
    expect(challengeDetailSource).not.toContain('admin-btn admin-btn-primary')
    expect(challengeDetailSource).not.toContain('admin-btn admin-btn-ghost')
    expect(adminChallengeTopbarPanelSource).toContain('拓扑编排')
    expect(adminChallengeTopbarPanelSource).toContain('返回 Jeopardy题库')
  })

  it('详情页 page model 应通过 feature route target + transport 处理导航，而不是直接 import vue-router', () => {
    expect(platformChallengeDetailPageSource).toContain(
      "from '@/shared/model/navigation/useRouteNavigationTransport'"
    )
    expect(platformChallengeDetailPageSource).toContain(
      "from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(platformChallengeDetailPageSource).toContain("from './platformChallengeDetailRoutes'")
    expect(platformChallengeDetailPageSource).not.toContain("from 'vue-router'")
    expect(platformChallengeDetailPageSource).toContain('platformChallengeListRoute')
    expect(platformChallengeDetailPageSource).toContain('platformChallengeTopologyStudioRoute')
    expect(platformChallengeDetailPageSource).toContain('platformChallengeWriteupRoute')
    expect(platformChallengeDetailRoutesSource).toContain("name: 'ChallengeManage'")
    expect(platformChallengeDetailRoutesSource).toContain("name: 'PlatformChallengeTopologyStudio'")
    expect(platformChallengeDetailRoutesSource).toContain(
      "name: mode === 'view' ? 'PlatformChallengeWriteupView' : 'PlatformChallengeWriteup'"
    )
  })

  it('应该根据 query 切到题解管理 tab', async () => {
    routeState.query = { panel: 'writeup' }

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: {
            template: '<div data-testid="challenge-writeup-tab">题解目录</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#admin-challenge-tab-writeup').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-challenge-panel-writeup').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('[data-testid="challenge-writeup-tab"]').exists()).toBe(true)
  })

  it('切换题解管理 tab 时应同步更新 panel query', async () => {
    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#admin-challenge-tab-writeup').trigger('click')

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'PlatformChallengeDetail',
      params: { id: '11' },
      query: { panel: 'writeup' },
    })
  })

  it('共享实例题应明确提示答案不做用户隔离', async () => {
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '共享密码题',
      category: 'crypto',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      description: 'desc',
      instance_sharing: 'shared',
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('共享实例只适用于无状态题')
    expect(wrapper.text()).toContain('不提供用户级答案隔离')
    expect(wrapper.text()).toContain('若需隔离答案，请使用 per_user 或 per_team')
  })

  it('共享实例题不应允许保存动态 Flag', async () => {
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '共享密码题',
      category: 'crypto',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      description: 'desc',
      instance_sharing: 'shared',
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    await wrapper.get('select.flag-field-input').setValue('dynamic')
    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存配置'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')

    expect(adminApiMocks.configureChallengeFlag).not.toHaveBeenCalled()
    expect(toastMocks.error).toHaveBeenCalledWith(
      '共享实例只适用于无状态题，不支持动态 Flag；若需隔离答案，请使用 per_user 或 per_team'
    )
  })

  it('管理员下载内部附件时应走带鉴权的下载接口', async () => {
    adminApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '11',
      title: '双节点演练',
      category: 'web',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      attachment_url: '/api/v1/challenges/attachments/imports/demo.zip',
      description: 'desc',
      hints: [],
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })

    const originalCreateElement = document.createElement.bind(document)
    const clickMock = vi.fn()
    let capturedAnchor: HTMLAnchorElement | null = null
    const createElementSpy = vi
      .spyOn(document, 'createElement')
      .mockImplementation((tagName: string) => {
        if (tagName === 'a') {
          const anchor = originalCreateElement(tagName)
          anchor.click = clickMock
          capturedAnchor = anchor
          return anchor
        }
        return originalCreateElement(tagName)
      })

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    const downloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('下载附件'))
    expect(downloadButton).toBeTruthy()
    expect(wrapper.text()).not.toContain('/api/v1/challenges/attachments/imports/demo.zip')

    await downloadButton!.trigger('click')
    await flushPromises()

    expect(challengeApiMocks.downloadAttachment).not.toHaveBeenCalled()
    expect(clickMock).toHaveBeenCalled()
    expect(capturedAnchor).not.toBeNull()
    if (!capturedAnchor) {
      throw new Error('expected download anchor to be created')
    }
    const anchor = capturedAnchor as HTMLAnchorElement
    expect(anchor.href).toContain('/api/v1/challenges/attachments/imports/demo.zip')
    expect(anchor.hasAttribute('download')).toBe(true)

    createElementSpy.mockRestore()
  })

  it('加载失败后卸载页面时应清理延迟跳转定时器', async () => {
    vi.useFakeTimers()
    adminApiMocks.getChallengeDetail.mockRejectedValueOnce(new Error('load failed'))

    const pushSpy = pushMock.mockImplementation(() => Promise.resolve())
    const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout')

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    wrapper.unmount()
    vi.runAllTimers()

    expect(clearTimeoutSpy).toHaveBeenCalled()
    expect(pushSpy).not.toHaveBeenCalledWith({ name: 'ChallengeManage' })

    clearTimeoutSpy.mockRestore()
    vi.useRealTimers()
  })
})
