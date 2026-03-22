import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ChallengeDetail from '../ChallengeDetail.vue'

const challengeApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  getChallengeWriteup: vi.fn(),
  submitFlag: vi.fn(),
  unlockHint: vi.fn(),
  createInstance: vi.fn(),
  downloadAttachment: vi.fn(),
}))

const instanceApiMocks = vi.hoisted(() => ({
  getMyInstances: vi.fn(),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

vi.mock('@/api/challenge', () => challengeApiMocks)
vi.mock('@/api/instance', () => instanceApiMocks)

describe('ChallengeDetail', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/challenges/:id', component: { template: '<div />' } }],
    })

    challengeApiMocks.getChallengeDetail.mockResolvedValue({
      id: '1',
      title: 'Test Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: false,
      attachment_url: 'https://example.com/file.zip',
      hints: [
        {
          id: 'hint-1',
          level: 1,
          title: '入口',
          cost_points: 0,
          is_unlocked: false,
        },
      ],
    })
    challengeApiMocks.getChallengeWriteup.mockResolvedValue({
      id: 'writeup-1',
      challenge_id: '1',
      title: '官方题解',
      content: '<p>Exploit path</p>',
      visibility: 'public',
      is_released: true,
      requires_spoiler_warning: true,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })
    challengeApiMocks.submitFlag.mockReset()
    challengeApiMocks.unlockHint.mockReset()
    challengeApiMocks.createInstance.mockResolvedValue({
      id: 'inst-1',
      challenge_id: '1',
      status: 'running',
      access_url: 'http://target.test',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
    })
    challengeApiMocks.downloadAttachment.mockReset()

    instanceApiMocks.getMyInstances.mockResolvedValue([])
    instanceApiMocks.destroyInstance.mockReset()
    instanceApiMocks.extendInstance.mockReset()
    instanceApiMocks.requestInstanceAccess.mockReset()
  })

  it('应该渲染挑战详情', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).toContain('提示系统')
  })

  it('应该支持查看题解并显示 spoiler 警告', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const writeupButton = wrapper.findAll('button').find((node) => node.text().includes('查看题解'))
    expect(writeupButton).toBeTruthy()

    await writeupButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(wrapper.text()).toContain('官方题解')
    expect(wrapper.text()).toContain('请谨慎阅读')
  })

  it('启动靶机后应停留在题目页并显示实例卡片', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const startButton = wrapper.findAll('button').find((node) => node.text().includes('启动靶机'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.createInstance).toHaveBeenCalledWith('1')
    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
    expect(wrapper.text()).toContain('打开目标')
    expect(wrapper.text()).toContain('靶机实例')
  })

  it('已存在实例时应直接显示实例信息', async () => {
    instanceApiMocks.getMyInstances.mockResolvedValue([
      {
        id: 'inst-9',
        challenge_id: 1,
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-12T00:00:00.000Z',
        challenge_title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
      },
    ])

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('打开目标')
    expect(wrapper.text()).toContain('http://127.0.0.1:30000')
    expect(wrapper.text()).toContain('1 次')
    expect(wrapper.text()).not.toContain('挑战信息')
    expect(wrapper.text()).not.toContain('启动靶机')
  })

  it('题目不需要靶机时应展示提示文案', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'No Target Challenge',
      description: '<p>Analyze only</p>',
      category: 'misc',
      difficulty: 'easy',
      tags: ['misc'],
      points: 50,
      need_target: false,
      is_solved: false,
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('该题目不需要靶机')
    expect(wrapper.text()).not.toContain('启动靶机')
  })

  it('应将 markdown 描述渲染为 HTML', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Markdown Challenge',
      description: '# 一级标题\n\n## 二级标题\n\n- item-1',
      category: 'misc',
      difficulty: 'easy',
      tags: ['misc'],
      points: 50,
      need_target: false,
      is_solved: false,
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const content = wrapper.find('.prose')
    expect(content.html()).toContain('<h1')
    expect(content.html()).toContain('<h2')
    expect(content.html()).toContain('<li')
  })
})
