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

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/challenges', component: { template: '<div />' } },
      { path: '/challenges/:id', component: { template: '<div />' } },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/challenges')
  await router.isReady()

  const wrapper = mount(ChallengeList, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
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
    expect(wrapper.text()).toContain('开始挑战')
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

    expect(wrapper.text()).toContain('挑战列表加载失败')
    expect(wrapper.text()).toContain('服务暂时不可用，请稍后重试')
    expect(wrapper.text()).not.toContain('请求ID')
  })

  it('应采用平铺目录式题目列表而不是卡片网格', () => {
    expect(challengeListSource).toContain('challenge-directory')
    expect(challengeListSource).toContain('challenge-row')
    expect(challengeListSource).toContain('题目列表')
    expect(challengeListSource).toContain('challenge-search-input')
    expect(challengeListSource).not.toContain('class="challenge-card')
    expect(challengeListSource).not.toContain('Training Range')
    expect(challengeListSource).not.toContain('Challenge Filters')
  })
})
