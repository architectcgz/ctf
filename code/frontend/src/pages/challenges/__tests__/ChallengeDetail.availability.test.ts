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

  it('草稿题目不可访问时应停留在当前页显示状态提示，并停止额外预取', async () => {
    challengeApiMocks.getChallengeDetail.mockRejectedValueOnce(
      new ApiError('题目为草稿，无法访问', { code: 13005, status: 403 })
    )

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
    expect(wrapper.text()).toContain('草稿题目暂不可访问')
    expect(wrapper.text()).toContain('当前题目还处于草稿状态，尚未开放访问。')
    expect(challengeApiMocks.getMyChallengeWriteupSubmission).not.toHaveBeenCalled()
    expect(challengeApiMocks.getMyChallengeSubmissionRecords).not.toHaveBeenCalled()
  })

  it('已归档题目不可访问时应停留在当前页显示状态提示', async () => {
    challengeApiMocks.getChallengeDetail.mockRejectedValueOnce(
      new ApiError('题目已归档，无法访问', { code: 13005, status: 403 })
    )

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
    expect(wrapper.text()).toContain('已归档题目不可访问')
    expect(wrapper.text()).toContain('当前题目已归档，不再提供访问入口。')
  })

  it('错误态返回题目列表应命中命名 route target', async () => {
    challengeApiMocks.getChallengeDetail.mockRejectedValueOnce(new Error('load failed'))

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()

    const backButton = wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '返回题目列表')
    expect(backButton).toBeTruthy()

    await backButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('Challenges')
    expect(router.currentRoute.value.fullPath).toBe('/challenges')
  })

  it('快速切换题目时不应被旧详情和旧提交记录回写', async () => {
    const staleDetail = createDeferred<{
      id: string
      title: string
      description: string
      category: 'web'
      difficulty: 'easy'
      tags: string[]
      points: number
      need_target: boolean
      is_solved: boolean
      attachment_url?: string
      hints: Array<{ id: string; level: number; title?: string; content?: string }>
    }>()
    const staleRecords =
      createDeferred<
        Array<{ id: string; answer?: string; status: 'correct'; submitted_at: string }>
      >()

    challengeApiMocks.getChallengeDetail.mockImplementation((id: string) => {
      if (id === '1') {
        return staleDetail.promise
      }

      return Promise.resolve({
        id: '2',
        title: 'Fresh Challenge',
        description: '<p>Fresh description</p>',
        category: 'web',
        difficulty: 'easy',
        tags: ['fresh'],
        points: 200,
        need_target: true,
        is_solved: false,
        attachment_url: undefined,
        hints: [],
      })
    })
    challengeApiMocks.getMyChallengeSubmissionRecords.mockImplementation((id: string) => {
      if (id === '1') {
        return staleRecords.promise
      }

      return Promise.resolve([])
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await router.push('/challenges/2')
    await flushPromises()

    staleDetail.resolve({
      id: '1',
      title: 'Stale Challenge',
      description: '<p>Stale description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['stale'],
      points: 100,
      need_target: true,
      is_solved: false,
      attachment_url: undefined,
      hints: [],
    })
    staleRecords.resolve([
      {
        id: 'record-1',
        answer: 'flag{stale}',
        status: 'correct',
        submitted_at: '2026-04-22T08:00:00.000Z',
      },
    ])

    await flushPromises()

    expect(wrapper.text()).toContain('Fresh Challenge')
    expect(wrapper.text()).not.toContain('Stale Challenge')

    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    expect(recordsTab).toBeTruthy()

    await recordsTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('flag{stale}')
    expect(wrapper.text()).toContain('还没有提交记录')
  })
})
