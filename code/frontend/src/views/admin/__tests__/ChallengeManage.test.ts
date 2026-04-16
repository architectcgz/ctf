import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeManage from '../ChallengeManage.vue'
import challengeManageSource from '../ChallengeManage.vue?raw'
import { ApiError } from '@/api/request'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string>,
}))
const adminApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  createChallenge: vi.fn(),
  createChallengePublishRequest: vi.fn(),
  deleteChallenge: vi.fn(),
  getChallengeDetail: vi.fn(),
  getChallengeFlagConfig: vi.fn(),
  getChallengeImport: vi.fn(),
  getChallenges: vi.fn(),
  getImages: vi.fn(),
  getLatestChallengePublishRequest: vi.fn(),
  listChallengeImports: vi.fn(),
  previewChallengeImport: vi.fn(),
  updateChallenge: vi.fn(),
  configureChallengeFlag: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('ChallengeManage', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
    adminApiMocks.commitChallengeImport.mockReset()
    adminApiMocks.createChallenge.mockReset()
    adminApiMocks.createChallengePublishRequest.mockReset()
    adminApiMocks.deleteChallenge.mockReset()
    adminApiMocks.getChallengeDetail.mockReset()
    adminApiMocks.getChallengeFlagConfig.mockReset()
    adminApiMocks.getChallengeImport.mockReset()
    adminApiMocks.getChallenges.mockReset()
    adminApiMocks.getImages.mockReset()
    adminApiMocks.getLatestChallengePublishRequest.mockReset()
    adminApiMocks.listChallengeImports.mockReset()
    adminApiMocks.previewChallengeImport.mockReset()
    adminApiMocks.updateChallenge.mockReset()
    adminApiMocks.configureChallengeFlag.mockReset()

    adminApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: 'challenge-legacy-id-001',
          title: 'Test Challenge',
          category: 'web',
          difficulty: 'easy',
          status: 'draft',
          points: 100,
          created_at: '2026-04-01T08:00:00.000Z',
          updated_at: '2026-04-01T08:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getLatestChallengePublishRequest.mockResolvedValue({
      id: 'req-1',
      challenge_id: '1',
      status: 'failed',
      active: false,
      failure_summary: 'Flag 未配置',
      created_at: '2026-04-01T08:00:00.000Z',
      updated_at: '2026-04-01T08:01:00.000Z',
    })
    adminApiMocks.listChallengeImports.mockResolvedValue([
      {
        id: 'import-1',
        file_name: 'demo-import.zip',
        slug: 'web-demo',
        title: 'Web Demo',
        description: 'Demo import preview',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        attachments: [{ name: 'demo.zip', path: 'attachments/demo.zip' }],
        hints: [{ level: 1, title: 'Hint 1', content: 'Check login flow' }],
        flag: { type: 'static', prefix: 'flag' },
        runtime: { type: 'container', image_ref: 'ctf/web-demo:latest' },
        extensions: { topology: { source: 'docker/topology.yml', enabled: false } },
        warnings: [],
        created_at: '2026-04-06T09:00:00.000Z',
      },
    ])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应使用统一的管理端 workspace 壳层，而不是页面内重复一套路由级顶栏', () => {
    expect(challengeManageSource).toContain('class="workspace-shell challenge-manage-shell"')
    expect(challengeManageSource).not.toContain('<header class="workspace-topbar">')
    expect(challengeManageSource).toContain('Plus,')
    expect(challengeManageSource).toContain(
      'class="workspace-directory-section challenge-manage-directory"'
    )
    expect(challengeManageSource).not.toContain(
      'class="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-sm font-medium shrink-0"'
    )
  })

  it('应该默认显示题目管理 tab', async () => {
    const wrapper = mount(ChallengeManage)
    await flushPromises()
    expect(wrapper.text()).toContain('题目管理')
    expect(wrapper.text()).toContain('Test Challenge')
  })

  it('题目管理列表应改用共享列表组件，并且不再显示题目 ID 列', async () => {
    expect(challengeManageSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(challengeManageSource).toContain('<WorkspaceDataTable')
    expect(challengeManageSource).not.toContain('>题目 ID<')
    expect(challengeManageSource).not.toContain('检索题目 ID 或名称...')

    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.text()).toContain('题目名称')
    expect(wrapper.text()).not.toContain('题目 ID')
    expect(wrapper.text()).not.toContain('challenge-legacy-id')
  })

  it('筛选排序工具栏和分页应接入共享组件，而不是继续内联实现', () => {
    expect(challengeManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(challengeManageSource).toContain("from '@/components/admin/AdminPaginationControls.vue'")
    expect(challengeManageSource).toContain('<WorkspaceDirectoryToolbar')
    expect(challengeManageSource).toContain('<AdminPaginationControls')
    expect(challengeManageSource).not.toContain('<WorkspaceDirectoryPagination')
    expect(challengeManageSource).not.toContain('<div class="challenge-filter-bar">')
  })

  it('更多操作菜单应浮到表格滚动层之上，而不是渲染在列表容器内部', async () => {
    const wrapper = mount(ChallengeManage, {
      attachTo: document.body,
    })
    await flushPromises()

    await wrapper.get('.challenge-row-menu-button').trigger('click')
    await flushPromises()

    const teleportedMenu = document.body.querySelector('.challenge-row-menu')
    expect(teleportedMenu).not.toBeNull()
    expect(
      wrapper.find('.workspace-directory-list .challenge-row-menu').exists()
    ).toBe(false)

    wrapper.unmount()
  })

  it('支持多选上传并在上传区域下方显示结果', async () => {
    adminApiMocks.previewChallengeImport
      .mockResolvedValueOnce({
        id: 'import-ok',
        file_name: 'ok.zip',
        slug: 'ok-challenge',
        title: 'OK Challenge',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        attachments: [],
        hints: [],
        flag: { type: 'static', prefix: 'flag' },
        runtime: { type: 'container', image_ref: 'ctf/ok:latest' },
        extensions: { topology: { source: '', enabled: false } },
        warnings: [],
        created_at: '2026-04-06T09:10:00.000Z',
      })
      .mockRejectedValueOnce(
        new ApiError('格式错误', {
          code: 10001,
          requestId: 'req_123',
        })
      )

    const wrapper = mount(ChallengeManage)
    await flushPromises()
    await wrapper.get('#challenge-tab-import').trigger('click')
    await flushPromises()

    const fileInput = wrapper.get('input[type="file"]')
    Object.defineProperty(fileInput.element, 'files', {
      value: [new File([''], 'ok.zip'), new File([''], 'bad.zip')]
    })
    await fileInput.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('ok.zip')
    expect(wrapper.text()).toContain('bad.zip')
    expect(wrapper.text()).toContain('错误码 10001')
  })

  it('点击筛选面板内的下拉框时不应立即关闭筛选面板', async () => {
    const wrapper = mount(ChallengeManage, {
      attachTo: document.body,
    })
    await flushPromises()

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    expect(wrapper.find('.workspace-directory-toolbar__filter-panel').exists()).toBe(true)

    await wrapper.get('.challenge-filter-select').trigger('click')
    await flushPromises()

    expect(wrapper.find('.workspace-directory-toolbar__filter-panel').exists()).toBe(true)
    wrapper.unmount()
  })
})
