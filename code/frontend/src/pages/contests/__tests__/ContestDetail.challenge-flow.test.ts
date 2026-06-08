import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import {
  contestApiMocks,
  contestDetailPageSource,
  contestDetailRoutePageSource,
  contestDetailSource,
  contestDetailWorkspaceSource,
  contestPresentationSource,
  contestTeamDialogsSource,
  contestTeamPanelSource,
  contestTeamWorkspaceSectionSource,
  destructiveConfirmMock,
  resetContestDetailTestHarness,
  routeQueryTransportSource,
  router,
  teamPresentationSource,
  webSocketMocks,
} from './ContestDetail.test-harness'
import ContestDetail from '@/pages/contests/ContestDetailRoutePage.vue'

describe('ContestDetail', () => {
  beforeEach(async () => {
    await resetContestDetailTestHarness()
  })

  it('普通竞赛提交反馈应由前端根据结果生成', async () => {
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.submitContestFlag
      .mockResolvedValueOnce({
        is_correct: false,
        points: 0,
        submitted_at: '2024-03-15T09:05:00Z',
      })
      .mockResolvedValueOnce({
        is_correct: true,
        points: 100,
        submitted_at: '2024-03-15T09:06:00Z',
      })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    const challengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(challengeButton).toBeTruthy()
    await challengeButton!.trigger('click')
    await flushPromises()

    const flagInput = wrapper.get('#contest-flag-input')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')
    expect(submitButton).toBeTruthy()

    await flagInput.setValue('flag{wrong}')
    await submitButton!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Flag 错误，请重试')

    await flagInput.setValue('flag{correct}')
    await submitButton!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('正确！+100 分')
  })

  it('普通竞赛题目选中状态应从 URL 恢复并在切换时写回 query', async () => {
    await router.push('/contests/1?panel=challenges&challenge=102')
    await router.isReady()
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
      {
        id: '102',
        challenge_id: '102',
        title: 'Crypto 102',
        category: 'crypto',
        difficulty: 'medium',
        points: 200,
        solved_count: 2,
        is_solved: false,
      },
    ])

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('已选题目')
    expect(wrapper.text()).toContain('主要操作')
    expect(wrapper.text()).toContain('Crypto 102')
    expect(wrapper.text()).toContain('密码')
    expect(wrapper.text()).toContain('200 分')

    const webChallengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(webChallengeButton).toBeTruthy()
    await webChallengeButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.challenge).toBe('101')
    expect(router.currentRoute.value.query.panel).toBe('challenges')
  })

  it('竞赛 Flag 提交进行中遇到回车和点击重叠时只应提交一次', async () => {
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '101',
        challenge_id: '101',
        title: 'Web 101',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.submitContestFlag.mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const challengesTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(challengesTab).toBeTruthy()
    await challengesTab!.trigger('click')
    await flushPromises()

    const challengeButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('Web 101'))
    expect(challengeButton).toBeTruthy()
    await challengeButton!.trigger('click')
    await flushPromises()

    const flagInput = wrapper.get('#contest-flag-input')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')
    expect(submitButton).toBeTruthy()

    await flagInput.setValue('flag{pending}')
    flagInput.element.dispatchEvent(new KeyboardEvent('keyup', { key: 'Enter', bubbles: true }))
    submitButton!.element.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await wrapper.vm.$nextTick()

    expect(contestApiMocks.submitContestFlag).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.submitContestFlag).toHaveBeenCalledWith('1', '101', 'flag{pending}')
  })
})
