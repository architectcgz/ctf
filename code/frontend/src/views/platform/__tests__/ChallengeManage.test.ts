import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeManage from '../ChallengeManage.vue'
import challengeManageSource from '../ChallengeManage.vue?raw'
import challengeManageHeroPanelSource from '@/components/platform/challenge/ChallengeManageHeroPanel.vue?raw'
import challengeManageDirectoryPanelSource from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue?raw'
import challengeManagePresentationSource from '@/features/platform-challenges/model/useChallengeManagePresentation.ts?raw'

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

vi.mock('@/api/admin/authoring', () => adminApiMocks)

const combinedSource = [
  challengeManageSource,
  challengeManageHeroPanelSource,
  challengeManageDirectoryPanelSource,
].join('\n')

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
    expect(challengeManageSource).toContain(
      'class="workspace-shell challenge-manage-shell journal-shell journal-shell-admin journal-notes-card journal-hero"'
    )
    expect(challengeManageSource).not.toContain('<header class="workspace-topbar">')
    expect(challengeManageSource).toContain('<ChallengeManageHeroPanel')
    expect(challengeManageHeroPanelSource).toMatch(
      /<div class="workspace-overline">\s*Challenge Workspace\s*<\/div>/
    )
    expect(challengeManageHeroPanelSource).toMatch(
      /<h1 class="workspace-page-title">\s*Jeopardy题库\s*<\/h1>/
    )
    expect(challengeManageHeroPanelSource).toMatch(
      /<p class="workspace-page-copy">\s*集中管理 Jeopardy 题目目录、发布状态与题库变更。\s*<\/p>/
    )
    expect(challengeManageSource).not.toContain('Inventory / Challenge Management')
    expect(challengeManageHeroPanelSource).toContain('Plus,')
    expect(challengeManageHeroPanelSource).toContain('Jeopardy题目总计')
    expect(combinedSource).toContain(
      'class="workspace-directory-section challenge-manage-directory"'
    )
    expect(challengeManageSource).not.toContain(
      'class="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-sm font-medium shrink-0"'
    )
  })

  it('应移除顶部 tab，并直接展示统一题目工作台', async () => {
    expect(challengeManageSource).not.toContain('class="top-tabs"')
    expect(challengeManageSource).not.toContain("from '@/composables/useRouteQueryTabs'")

    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.text()).toContain('Jeopardy题库')
    expect(wrapper.text()).toContain('导入题目')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).not.toContain('导入资源包')
    expect(wrapper.text()).not.toContain('导入题目包')
    expect(wrapper.text()).not.toContain('待确认导入')
  })

  it('题目导入入口按钮应使用题目管理页自己的主题变量样式', () => {
    expect(challengeManageHeroPanelSource).toContain('class="challenge-manage-hero-actions"')
    expect(challengeManageHeroPanelSource).toContain(
      'class="header-btn header-btn--primary challenge-manage-import-button"'
    )
    expect(challengeManageHeroPanelSource).toContain('导入题目')
    expect(challengeManageHeroPanelSource).not.toContain('awd-library-hero-actions')
    expect(challengeManageHeroPanelSource).not.toContain('导入资源包')
    expect(challengeManageHeroPanelSource).toMatch(
      /\.challenge-manage-import-button\s*\{[\s\S]*--header-btn-border:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 34%, var\(--journal-border\)\);[\s\S]*--header-btn-background:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 12%, var\(--journal-surface\)\);[\s\S]*--header-btn-color:\s*var\(--journal-accent-strong\);/s
    )
    expect(challengeManageHeroPanelSource).toMatch(
      /\.challenge-manage-import-button\s*\{[\s\S]*--header-btn-focus-ring:\s*color-mix\(in srgb,\s*var\(--journal-accent\) 28%, transparent\);/s
    )
  })

  it('页面壳层应继续复用 journal surface token，而题目更多菜单则应改用共享 action menu primitive', () => {
    expect(challengeManageSource).toContain('--challenge-page-bg')
    expect(challengeManageSource).toContain('--workspace-shell-bg')
    expect(challengeManageSource).toContain('var(--journal-surface)')
    expect(challengeManageSource).toContain('var(--journal-surface-subtle)')
    expect(combinedSource).toContain("from '@/components/common/menus/CActionMenu.vue'")
    expect(combinedSource).not.toContain('--challenge-action-surface')
    expect(combinedSource).not.toContain(":global([data-theme='light']) .challenge-manage-shell")
    expect(combinedSource).not.toContain(":global([data-theme='dark']) .challenge-manage-shell")
    expect(combinedSource).not.toContain('<Teleport to="body">')
    expect(combinedSource).not.toContain('.challenge-row-menu')
  })

  it('最外层 workspace shell 应保留共享边框，而不是自行抹掉 shell 边界', () => {
    expect(challengeManageSource).toContain('.challenge-manage-shell {')
    expect(challengeManageSource).not.toContain('border: none;')
  })

  it('题目管理列表应改用共享列表组件，并且不再显示题目 ID 列', async () => {
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(combinedSource).toContain('<WorkspaceDataTable')
    expect(combinedSource).not.toContain('>题目 ID<')
    expect(combinedSource).not.toContain('检索题目 ID 或名称...')

    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.text()).toContain('题目名称')
    expect(wrapper.text()).not.toContain('题目 ID')
    expect(wrapper.text()).not.toContain('challenge-legacy-id')
  })

  it('筛选排序工具栏和分页应接入共享组件，而不是继续内联实现', () => {
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDirectoryToolbar.vue'")
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDirectoryPagination.vue'")
    expect(combinedSource).toContain('<WorkspaceDirectoryToolbar')
    expect(combinedSource).toContain('<WorkspaceDirectoryPagination')
    expect(combinedSource).not.toContain('<PlatformPaginationControls')
    expect(combinedSource).not.toContain('<div class="challenge-filter-bar">')
  })

  it('题目管理目录区应直接使用目录 section，而不是额外包一层自定义容器', () => {
    expect(combinedSource).toContain(
      '<section class="workspace-directory-section challenge-manage-directory">'
    )
    expect(combinedSource).toContain('<h2 class="list-heading__title">题目目录</h2>')
    expect(combinedSource).not.toContain('<div class="challenge-manage-directory">')
  })

  it('平台题目管理展示层应复用 challenge entity 的分类与难度文案规则', () => {
    expect(challengeManagePresentationSource).toContain("from '@/entities/challenge'")
    expect(challengeManagePresentationSource).not.toContain('function getCategoryLabel(')
    expect(challengeManagePresentationSource).not.toContain('function getDifficultyLabel(')
  })

  it('题目管理页应复用共享 spacing token，而不是给 summary strip 叠加额外上下 margin', () => {
    const summaryGridStyleBlock = challengeManageHeroPanelSource.match(
      /\.manage-summary-grid\s*\{[^}]*\}/s
    )?.[0]

    expect(challengeManageSource).toMatch(
      /\.challenge-manage-content\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap,\s*var\(--space-5\)\);/s
    )
    expect(challengeManageSource).toMatch(
      /\.challenge-manage-panel\s*\{[\s\S]*gap:\s*var\(--workspace-directory-page-block-gap,\s*var\(--space-5\)\);/s
    )
    expect(challengeManageHeroPanelSource).toMatch(
      /\.challenge-manage-hero-panel\s*\{[\s\S]*gap:\s*0;/s
    )
    expect(challengeManageDirectoryPanelSource).not.toContain('.challenge-manage-directory {')
    expect(summaryGridStyleBlock).toBeTruthy()
    expect(summaryGridStyleBlock).not.toContain('margin-top')
    expect(summaryGridStyleBlock).not.toContain('margin-bottom')
  })

  it('更多操作菜单应浮到表格滚动层之上，而不是渲染在列表容器内部', async () => {
    const wrapper = mount(ChallengeManage, {
      attachTo: document.body,
    })
    await flushPromises()

    await wrapper.get('button[aria-haspopup="menu"]').trigger('click')
    await flushPromises()

    const teleportedMenu = document.body.querySelector('[data-action-menu-panel]')
    expect(teleportedMenu).not.toBeNull()
    expect(wrapper.find('.workspace-directory-list [data-action-menu-panel]').exists()).toBe(false)

    wrapper.unmount()
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

  it('切换排序策略后应重排当前页题目顺序', async () => {
    adminApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: 'challenge-z',
          title: 'Zulu Challenge',
          category: 'web',
          difficulty: 'easy',
          status: 'draft',
          points: 100,
          created_at: '2026-04-01T08:00:00.000Z',
          updated_at: '2026-04-01T08:00:00.000Z',
        },
        {
          id: 'challenge-a',
          title: 'Alpha Challenge',
          category: 'pwn',
          difficulty: 'medium',
          status: 'published',
          points: 500,
          created_at: '2026-04-01T08:00:00.000Z',
          updated_at: '2026-04-01T08:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getLatestChallengePublishRequest.mockResolvedValue(null)

    const wrapper = mount(ChallengeManage, {
      attachTo: document.body,
    })
    await flushPromises()

    await wrapper.get('.workspace-directory-toolbar__sort-button').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('.workspace-directory-toolbar__menu-item')
      .find((item) => item.text().includes('标题 A-Z'))
      ?.trigger('click')
    await flushPromises()

    const titles = wrapper.findAll('.challenge-table-title').map((item) => item.text())
    expect(titles).toEqual(['Alpha Challenge', 'Zulu Challenge'])

    wrapper.unmount()
  })

  it('题目目录加载失败时应显示错误态而不是空目录提示', async () => {
    adminApiMocks.getChallenges.mockRejectedValueOnce(new Error('服务暂时不可用'))

    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.text()).toContain('题目目录加载失败')
    expect(wrapper.text()).toContain('服务暂时不可用')
    expect(wrapper.text()).toContain('重新加载')
    expect(wrapper.text()).not.toContain('当前还没有题目，请先前往导入页上传题目包。')
  })
})
