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

  it('应根据 panel query 初始化到对应 workspace tab', async () => {
    await router.push('/challenges/1?panel=writeup')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(router.currentRoute.value.query.panel).toBe('writeup')
    expect(challengeApiMocks.getMyChallengeWriteupSubmission).toHaveBeenCalledWith('1')
    expect(wrapper.find('input[placeholder*="完整链路"]').exists()).toBe(true)
  })

  it('切到非默认 workspace tab 时应写回 panel query', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    expect(recordsTab).toBeTruthy()

    await recordsTab!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBe('records')
    expect(router.currentRoute.value.fullPath).toBe('/challenges/1?panel=records')
  })

  it('切回默认题目 tab 时应清掉 panel query', async () => {
    await router.push('/challenges/1?panel=records')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const questionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题目')
    expect(questionTab).toBeTruthy()

    await questionTab!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBeUndefined()
    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
  })
})
