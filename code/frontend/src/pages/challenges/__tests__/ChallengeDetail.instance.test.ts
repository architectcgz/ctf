import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import {
  ApiError,
  challengeActionAsideSource,
  challengeApiMocks,
  challengeDetailPageSource,
  challengeDetailRoutesSource,
  challengeDetailSource,
  challengeDetailWorkspaceSource,
  challengeInstanceCardSource,
  challengeQuestionPanelSource,
  challengeSolutionsPanelSource,
  challengeSubmissionRecordsPanelSource,
  challengeWorkspaceShellSource,
  challengeWriteupPanelSource,
  cleanupChallengeDetailTestHarness,
  createDeferred,
  instanceApiMocks,
  instancePresentationSource,
  resetChallengeDetailTestHarness,
  router,
} from './ChallengeDetail.test-harness'
import ChallengeDetail from '@/pages/challenges/ChallengeDetailRoutePage.vue'

describe('ChallengeDetail', () => {
  beforeEach(() => {
    resetChallengeDetailTestHarness()
  })

  afterEach(() => {
    cleanupChallengeDetailTestHarness()
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

  it('题目已解出时仍应允许重启实例', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Solved Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: true,
      attachment_url: 'https://example.com/file.zip',
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

    expect(wrapper.text()).toContain('重启实例')
    expect(wrapper.text()).not.toContain(
      '当前题目已完成，如仍需验证环境可前往实例列表查看历史实例。'
    )

    const restartButton = wrapper.findAll('button').find((node) => node.text().includes('重启实例'))
    expect(restartButton).toBeTruthy()

    await restartButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.createInstance).toHaveBeenCalledWith('1')
  })

  it('createInstance 返回 pending 时应显示等待文案并触发轮询', async () => {
    vi.useFakeTimers()
    challengeApiMocks.createInstance.mockResolvedValueOnce({
      id: 'inst-1',
      challenge_id: '1',
      status: 'pending',
      access_url: '',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
      queue_position: 3,
      eta_seconds: 120,
      progress: 18,
    })
    instanceApiMocks.getMyInstances.mockReset()
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([]).mockResolvedValueOnce([
      {
        id: 'inst-1',
        challenge_id: 1,
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 2,
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
    await vi.advanceTimersByTimeAsync(100)

    const startButton = wrapper.findAll('button').find((node) => node.text().includes('启动靶机'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('实例正在排队创建')
    expect(wrapper.text()).toContain('等待实例就绪')
    expect(instanceApiMocks.getMyInstances).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(3000)
    await wrapper.vm.$nextTick()

    expect(instanceApiMocks.getMyInstances).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('打开目标')
  })

  it('排队中的实例轮询后若变为 failed 应显示启动失败提示', async () => {
    vi.useFakeTimers()
    challengeApiMocks.createInstance.mockResolvedValueOnce({
      id: 'inst-failed',
      challenge_id: '1',
      status: 'pending',
      access_url: '',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
      queue_position: 2,
      eta_seconds: 90,
      progress: 12,
    })
    instanceApiMocks.getMyInstances.mockReset()
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([]).mockResolvedValueOnce([
      {
        id: 'inst-failed',
        challenge_id: 1,
        status: 'failed',
        access_url: '',
        flag_type: 'static',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 2,
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
    await vi.advanceTimersByTimeAsync(100)

    const startButton = wrapper.findAll('button').find((node) => node.text().includes('启动靶机'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('实例正在排队创建')

    await vi.advanceTimersByTimeAsync(3000)
    await wrapper.vm.$nextTick()

    expect(instanceApiMocks.getMyInstances).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('实例启动失败，当前目标不可访问')
    expect(wrapper.text()).not.toContain('打开目标')
    expect(wrapper.text()).toContain('重启实例')
  })

  it('实例过期后应显示已自动回收并允许重启', async () => {
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([
      {
        id: 'inst-expired',
        challenge_id: 1,
        status: 'expired',
        access_url: '',
        flag_type: 'static',
        expires_at: '2026-03-11T00:00:00.000Z',
        remaining_extends: 0,
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

    expect(wrapper.text()).toContain('已自动回收')
    expect(wrapper.text()).toContain('重启实例')
    expect(wrapper.text()).not.toContain('销毁')
    expect(wrapper.text()).not.toContain('打开目标')

    const restartButton = wrapper.findAll('button').find((node) => node.text().includes('重启实例'))
    expect(restartButton).toBeTruthy()

    await restartButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.createInstance).toHaveBeenCalledWith('1')
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

  it('共享实例应隐藏延时和销毁操作', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Shared Challenge',
      description: '<p>Shared instance</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['shared'],
      points: 100,
      need_target: true,
      flag_type: 'static',
      instance_sharing: 'shared',
      is_solved: false,
      hints: [],
    })
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([
      {
        id: 'inst-shared',
        challenge_id: 1,
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        share_scope: 'shared',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-12T00:00:00.000Z',
        challenge_title: 'Shared Challenge',
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
    expect(wrapper.text()).toContain('共享实例')
    expect(wrapper.text()).toContain('Flag 提交')
    expect(wrapper.text()).toContain('输入当前题目的 Flag 并提交验证。')
    expect(wrapper.text()).not.toContain('延时')
    expect(wrapper.text()).not.toContain('销毁')
  })
})
