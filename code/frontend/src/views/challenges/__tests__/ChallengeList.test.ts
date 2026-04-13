import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import ChallengeList from '../ChallengeList.vue'
import challengeListSource from '../ChallengeList.vue?raw'
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
      { path: '/challenges', component: { template: '<div />' } },
      { path: '/challenges/:id', component: { template: '<div />' } },
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
    expect(wrapper.text()).toContain('当前题库概况')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).toContain('题目总数')
    expect(wrapper.text()).toContain('开始做题')
    expect(wrapper.find('.challenge-row-title').attributes('title')).toBe('Test Challenge')
    expect(wrapper.find('.challenge-row-solved').text()).toContain('10 人解出')
    expect(wrapper.find('.challenge-row-attempts').text()).toContain('尝试 20 次')
  })

  it('页头标题与说明应接入共享工作区排版类', () => {
    expect(challengeListSource).toContain('<h1 class="workspace-page-title challenge-title">靶场训练</h1>')
    expect(challengeListSource).not.toContain('按关键词、分类与难度筛选题目，直接进入训练。')
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
    expect(wrapper.find('.challenge-row-index').exists()).toBe(false)
  })

  it('题目页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(challengeListSource).toContain('class="challenge-summary metric-panel-default-surface"')
    expect(challengeListSource).toContain('class="challenge-summary-grid metric-panel-grid"')
    expect(challengeListSource).toContain('class="challenge-summary-item metric-panel-card"')
    expect(challengeListSource).toContain('class="challenge-summary-label metric-panel-label"')
    expect(challengeListSource).toContain('class="challenge-summary-value metric-panel-value"')
    expect(challengeListSource).toContain('class="challenge-summary-helper metric-panel-helper"')
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

    expect(wrapper.find('.challenge-directory-head').text()).toContain('积分')
    expect(wrapper.find('.challenge-row-main .challenge-row-points').exists()).toBe(false)
    expect(wrapper.find('.challenge-row-points').text()).toContain('100 pts')
  })

  it('应从路由 query 初始化分类和难度筛选', async () => {
    mockedGetChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = await mountPage('/challenges?category=crypto&difficulty=medium')

    expect((wrapper.get('#challenge-category-filter').element as HTMLSelectElement).value).toBe('crypto')
    expect((wrapper.get('#challenge-difficulty-filter').element as HTMLSelectElement).value).toBe('medium')
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

  it('应采用平铺目录式题目列表而不是卡片网格', () => {
    expect(challengeListSource).toContain('class="workspace-directory-section challenge-directory-section"')
    expect(challengeListSource).toContain('class="list-heading"')
    expect(challengeListSource).not.toContain('challenge-controls-title')
    expect(challengeListSource).not.toContain('challenge-controls-copy')
    expect(challengeListSource).not.toContain('challenge-filter-pill')
    expect(challengeListSource).not.toContain('激活筛选')
    expect(challengeListSource).toContain('challenge-directory')
    expect(challengeListSource).toContain('challenge-row')
    expect(challengeListSource).not.toContain('</section>\n\n        <div v-if="total > 0" class="challenge-pagination">')
    expect(challengeListSource).toContain('题目列表')
    expect(challengeListSource).toContain('challenge-search-input')
    expect(challengeListSource).toContain('搜索题目标题或描述')
    expect(challengeListSource).not.toContain('challenge-row-index')
    expect(challengeListSource).not.toContain('CH-{{ challengeIndex(index) }}')
    expect(challengeListSource).toContain('<span>分类</span>')
    expect(challengeListSource).toContain('<span>难度</span>')
    expect(challengeListSource).toContain('<span>解出</span>')
    expect(challengeListSource).toContain('<span>尝试</span>')
    expect(challengeListSource).toContain('class="challenge-row-category"')
    expect(challengeListSource).toContain('class="challenge-row-difficulty"')
    expect(challengeListSource).toContain('class="challenge-row-solved"')
    expect(challengeListSource).toContain('class="challenge-row-attempts"')
    expect(challengeListSource).toMatch(/class="challenge-row-title"[\s\S]*:title="challenge\.title"/s)
    expect(challengeListSource).toMatch(/\.challenge-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(challengeListSource).not.toContain('class="challenge-card')
    expect(challengeListSource).not.toContain('Training Range')
    expect(challengeListSource).not.toContain('Challenge Filters')
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

    expect(wrapper.find('.challenge-pagination').exists()).toBe(true)
    expect(wrapper.find('.challenge-pagination').text()).toContain('1 / 1')
  })
})
