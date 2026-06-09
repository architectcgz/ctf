import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import ChallengeList from '@/pages/challenges/ChallengeListRoutePage.vue'
import challengeListRouteSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import challengeListPageSource from '@/features/challenge-list/ui/ChallengeListPage.vue?raw'
import challengeDirectoryPanelSource from '@/features/challenge-list/ui/ChallengeDirectoryPanel.vue?raw'
import challengeDirectoryRowSource from '@/entities/challenge/ui/ChallengeDirectoryRow.vue?raw'
import challengeListPageModelSource from '@/features/challenge-list/model/useChallengeListPage.ts?raw'
import routeQueryTransportSource from '@/shared/model/navigation/useRouteQueryTransport.ts?raw'
import { getChallenges } from '@/api/challenge'

vi.mock('@/api/challenge', () => ({
  getChallenges: vi.fn(),
}))

const mockedGetChallenges = vi.mocked(getChallenges)

function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/student/dashboard', name: 'Dashboard', component: { template: '<div />' } },
      { path: '/student/skill-profile', name: 'SkillProfile', component: { template: '<div />' } },
      { path: '/challenges', name: 'Challenges', component: { template: '<div />' } },
      { path: '/challenges/:id', name: 'ChallengeDetail', component: { template: '<div />' } },
    ],
  })
}

async function mountPageWithRouter(initialPath = '/challenges') {
  const router = createTestRouter()
  await router.push(initialPath)
  await router.isReady()

  const wrapper = mount(ChallengeList, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

async function mountPage(initialPath = '/challenges') {
  const { wrapper } = await mountPageWithRouter(initialPath)
  return wrapper
}

describe('ChallengeList', () => {
  beforeEach(() => {
    mockedGetChallenges.mockReset()
  })

  it('应该渲染挑战列表', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Test Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: ['test'],
          solved_count: 10,
          total_attempts: 20,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('Challenges')
    expect(wrapper.text()).toContain('靶场训练')
    expect(wrapper.text()).toContain('题库概况')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).toContain('题目总数')
    expect(wrapper.text()).toContain('开始做题')
    expect(wrapper.text()).not.toContain('统一查看训练题目')
    expect(wrapper.text()).toContain('10 人解出')
    expect(wrapper.text()).toContain('尝试 20 次')
  })

  it('页面应通过 feature model 获取列表状态，不再直接耦合 challenge api 与分页流程', () => {
    expect(challengeListRouteSource).toContain("ChallengeListPage } from '@/features/challenge-list'")
    expect(challengeListRouteSource).not.toContain("from '@/api/challenge'")
    expect(challengeListRouteSource).not.toContain("from '@/composables/usePagination'")
    expect(challengeListRouteSource).not.toContain('const summaryStats = computed(() => [')
    expect(challengeListRouteSource).not.toContain('async function syncFilterQuery()')
    expect(challengeListRouteSource).not.toContain('watch(')
    expect(challengeListPageSource).toContain("useChallengeListPage } from '../model'")
    expect(challengeListPageModelSource).not.toContain("from 'vue-router'")
    expect(challengeListPageModelSource).toContain(
      "from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(routeQueryTransportSource).toContain('const route = useRoute()')
    expect(routeQueryTransportSource).toContain('const router = useRouter()')
  })

  it('题目目录组件应通过 challenge entity 获取分类与难度展示规则', () => {
    expect(challengeDirectoryPanelSource).toContain("from '@/entities/challenge'")
    expect(challengeDirectoryPanelSource).not.toContain('function getCategoryLabel(')
    expect(challengeDirectoryPanelSource).not.toContain('function getCategoryColor(')
    expect(challengeDirectoryPanelSource).not.toContain('function getDifficultyLabel(')
    expect(challengeDirectoryPanelSource).not.toContain('function getDifficultyColor(')
  })

  it('题目列表不应显示编号前缀', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Hidden Index Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: ['test'],
          solved_count: 10,
          total_attempts: 20,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).not.toContain('CH-1')
  })

  it('搜索时应通过 keyword 参数请求真实筛选', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()
    mockedGetChallenges.mockClear()

    await wrapper.get('#challenge-search-input').setValue('sql')
    await flushPromises()

    expect(mockedGetChallenges).toHaveBeenCalledTimes(1)
    expect(mockedGetChallenges).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        keyword: 'sql',
      })
    )
    expect(mockedGetChallenges.mock.lastCall?.[0]).not.toHaveProperty('search')
  })

  it('旧请求晚返回时不应覆盖新的搜索结果', async () => {
    const initialRequest = createDeferred<{
      list: Array<{
        id: string
        title: string
        category: 'web'
        difficulty: 'easy'
        tags: string[]
        solved_count: number
        total_attempts: number
        is_solved: boolean
        points: number
        created_at: string
      }>
      total: number
      page: number
      page_size: number
    }>()

    mockedGetChallenges.mockImplementation(async (params) => {
      if ((params as { keyword?: string }).keyword === 'sql') {
        return {
          list: [
            {
              id: '2',
              title: 'SQL Search Hit',
              category: 'web',
              difficulty: 'easy',
              tags: ['sql'],
              solved_count: 3,
              total_attempts: 8,
              is_solved: false,
              points: 200,
              created_at: '2024-01-02T00:00:00Z',
            },
          ],
          total: 1,
          page: 1,
          page_size: 20,
        }
      }

      return initialRequest.promise
    })

    const wrapper = await mountPage()

    await wrapper.get('#challenge-search-input').setValue('sql')
    await flushPromises()

    expect(wrapper.text()).toContain('SQL Search Hit')
    expect(wrapper.text()).not.toContain('Initial Full List')

    initialRequest.resolve({
      list: [
        {
          id: '1',
          title: 'Initial Full List',
          category: 'web',
          difficulty: 'easy',
          tags: ['initial'],
          solved_count: 10,
          total_attempts: 20,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('SQL Search Hit')
    expect(wrapper.text()).not.toContain('Initial Full List')
  })

  it('应该显示空列表提示', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('目前还没有题目')
    expect(wrapper.text()).toContain('管理员还没有发布训练题目')
  })

  it('应该显示用户可读的错误信息', async () => {
    mockedGetChallenges.mockRejectedValue(new Error('服务暂时不可用，请稍后重试'))

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('题目列表加载失败')
    expect(wrapper.text()).toContain('服务暂时不可用，请稍后重试')
    expect(wrapper.text()).not.toContain('请求ID')
  })

  it('只有分类和难度标签时不应额外显示暂无标签', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Tagless Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: [],
          solved_count: 10,
          total_attempts: 20,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('Web')
    expect(wrapper.text()).toContain('简单')
    expect(wrapper.text()).not.toContain('暂无标签')
  })

  it('应将积分作为独立列展示而不是放在题目后面', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Point Column Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: ['test'],
          solved_count: 10,
          total_attempts: 20,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('积分')
    expect(wrapper.text()).toContain('100 pts')
  })

  it('应从路由 query 初始化分类和难度筛选', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage('/challenges?category=crypto&difficulty=medium')

    expect((wrapper.get('#challenge-category-filter').element as HTMLSelectElement).value).toBe(
      'crypto'
    )
    expect((wrapper.get('#challenge-difficulty-filter').element as HTMLSelectElement).value).toBe(
      'medium'
    )
    expect(mockedGetChallenges).toHaveBeenCalledTimes(1)
    expect(mockedGetChallenges).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        category: 'crypto',
        difficulty: 'medium',
      })
    )
  })

  it('切换分类和难度时应回写到路由 query', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const { wrapper, router } = await mountPageWithRouter()
    mockedGetChallenges.mockClear()

    await wrapper.get('#challenge-category-filter').setValue('crypto')
    await flushPromises()
    expect(router.currentRoute.value.query).toEqual({ category: 'crypto' })
    expect(mockedGetChallenges).toHaveBeenCalledTimes(1)
    expect(mockedGetChallenges).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        category: 'crypto',
      })
    )
    expect(mockedGetChallenges.mock.lastCall?.[0]).not.toHaveProperty('difficulty')

    await wrapper.get('#challenge-difficulty-filter').setValue('medium')
    await flushPromises()
    expect(router.currentRoute.value.query).toEqual({
      category: 'crypto',
      difficulty: 'medium',
    })
    expect(mockedGetChallenges).toHaveBeenCalledTimes(2)
    expect(mockedGetChallenges).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        category: 'crypto',
        difficulty: 'medium',
      })
    )
  })

  it('页头与目录行应通过 route target，而不是 page model 直接 push', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Route Target Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: ['target'],
          solved_count: 1,
          total_attempts: 2,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const { wrapper, router } = await mountPageWithRouter()

    expect(challengeListPageSource).toContain("import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(challengeDirectoryRowSource).toContain(
      "import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'"
    )
    expect(challengeListPageSource).not.toContain('@click="goToDashboard"')
    expect(challengeListPageSource).not.toContain('@click="openSkillProfile"')
    expect(challengeDirectoryPanelSource).not.toContain("@open-detail")

    const dashboardLink = wrapper.get('a[href="/student/dashboard"]')
    expect(dashboardLink.text()).toContain('返回仪表盘')

    const skillProfileLink = wrapper.get('a[href="/student/skill-profile"]')
    expect(skillProfileLink.text()).toContain('能力画像')

    const detailLink = wrapper.get('a[href="/challenges/1"]')
    expect(detailLink.text()).toContain('Route Target Challenge')

    await detailLink.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
  })

  it('单页结果时也应显式显示分页状态', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Single Page Challenge',
          category: 'web',
          difficulty: 'easy',
          tags: ['test'],
          solved_count: 1,
          total_attempts: 2,
          is_solved: false,
          points: 100,
          created_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage()

    expect(wrapper.text()).toContain('1 / 1')
  })
})
