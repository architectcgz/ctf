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

  it('未解题时应显示题解锁定态', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('解出题目后可查看推荐题解与社区题解')
    expect(wrapper.text()).not.toContain('精选官方题解')
    expect(challengeApiMocks.getRecommendedChallengeSolutions).not.toHaveBeenCalled()
    expect(challengeApiMocks.getCommunityChallengeSolutions).not.toHaveBeenCalled()
  })
  it('已解题时应通过顶部标签切换到推荐题解、社区题解和编写题解', async () => {
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

    expect(challengeApiMocks.getRecommendedChallengeSolutions).not.toHaveBeenCalled()
    expect(challengeApiMocks.getCommunityChallengeSolutions).not.toHaveBeenCalled()
    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await flushPromises()

    expect(challengeApiMocks.getRecommendedChallengeSolutions).toHaveBeenCalledWith('1')
    expect(challengeApiMocks.getCommunityChallengeSolutions).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('社区题解')
    expect(wrapper.text()).toContain('精选官方题解')

    const communityTab = wrapper.findAll('button').find((node) => node.text().trim() === '社区题解')
    expect(communityTab).toBeTruthy()

    await communityTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('我的 SQLi 复盘')

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '编写题解')
    expect(writeupTab).toBeTruthy()

    await writeupTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('解题过程复盘')
  })
})
