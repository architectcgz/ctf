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

  it('manual review 提交后应显示待审核反馈', async () => {
    challengeApiMocks.submitFlag.mockResolvedValue({
      is_correct: false,
      status: 'pending_review',
      submitted_at: '2026-03-12T01:00:00.000Z',
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

    const flagInput = wrapper.find('input[placeholder="flag{...}"]')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')

    await flagInput.setValue('exploit chain answer')
    await submitButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.submitFlag).toHaveBeenCalledWith('1', 'exploit chain answer')
    expect(wrapper.text()).toContain('答案已提交，等待教师审核')
    expect(wrapper.text()).not.toContain('已完成 ✓')
  })

  it('Flag 提交进行中遇到回车和点击重叠时只应提交一次', async () => {
    challengeApiMocks.submitFlag.mockImplementation(() => new Promise(() => {}))

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const flagInput = wrapper.get('input[aria-label="Flag"]')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')

    await flagInput.setValue('flag{pending}')
    flagInput.element.dispatchEvent(new KeyboardEvent('keyup', { key: 'Enter', bubbles: true }))
    submitButton!.element.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await wrapper.vm.$nextTick()

    expect(challengeApiMocks.submitFlag).toHaveBeenCalledTimes(1)
    expect(challengeApiMocks.submitFlag).toHaveBeenCalledWith('1', 'flag{pending}')
  })

  it('正确提交后应提示实例将在 10 分钟后自动关闭', async () => {
    challengeApiMocks.submitFlag.mockResolvedValue({
      is_correct: true,
      status: 'correct',
      points: 100,
      submitted_at: '2026-03-12T01:00:00.000Z',
      instance_shutdown_at: '2026-03-12T01:10:00.000Z',
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

    const flagInput = wrapper.find('input[placeholder="flag{...}"]')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')

    await flagInput.setValue('flag{correct}')
    await submitButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(wrapper.text()).toContain('当前实例将在 10 分钟后自动关闭')
  })

  it('进入题目页后应加载并展示历史提交记录', async () => {
    challengeApiMocks.getMyChallengeSubmissionRecords.mockResolvedValue([
      {
        id: 'record-1',
        status: 'correct',
        submitted_at: '2026-03-12T01:10:00.000Z',
      },
      {
        id: 'record-2',
        status: 'incorrect',
        submitted_at: '2026-03-12T01:00:00.000Z',
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

    expect(challengeApiMocks.getMyChallengeSubmissionRecords).not.toHaveBeenCalled()
    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    expect(recordsTab).toBeTruthy()

    await recordsTab!.trigger('click')
    await flushPromises()

    expect(challengeApiMocks.getMyChallengeSubmissionRecords).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('恭喜你，Flag 正确！')
    expect(wrapper.text()).toContain('Flag 错误，请重试')
  })

  it('提交记录过多时应支持分页切换', async () => {
    challengeApiMocks.getMyChallengeSubmissionRecords.mockResolvedValue(
      Array.from({ length: 11 }, (_, index) => ({
        id: `record-${index + 1}`,
        answer: `flag{${index + 1}}`,
        status: index % 2 === 0 ? 'incorrect' : 'correct',
        submitted_at: `2026-03-${String(20 - index).padStart(2, '0')}T01:00:00.000Z`,
      }))
    )

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    expect(recordsTab).toBeTruthy()

    await recordsTab!.trigger('click')
    await flushPromises()

    expect(wrapper.find('.submission-pagination').exists()).toBe(true)
    expect(wrapper.find('.submission-pagination').text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('flag{1}')
    expect(wrapper.text()).toContain('flag{10}')
    expect(wrapper.text()).not.toContain('flag{11}')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')
    await flushPromises()

    expect(wrapper.find('.submission-pagination').text()).toContain('2 / 2')
    expect(wrapper.text()).toContain('flag{11}')
    expect(wrapper.text()).not.toContain('flag{1}')
  })

  it('Flag 输入应提供可访问标签', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const flagInput = wrapper.find('input[aria-label="Flag"]')
    expect(flagInput.exists()).toBe(true)
  })

  it('题目已解出后仍应允许再次提交 Flag 做校验', async () => {
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
    challengeApiMocks.submitFlag.mockResolvedValueOnce({
      is_correct: true,
      status: 'correct',
      points: 0,
      submitted_at: '2026-03-12T01:15:00.000Z',
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

    const flagInput = wrapper.get('input[aria-label="Flag"]')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')

    expect(flagInput.attributes('disabled')).toBeUndefined()
    expect(submitButton?.attributes('disabled')).toBeUndefined()

    await flagInput.setValue('flag{still-correct}')
    await submitButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.submitFlag).toHaveBeenCalledWith('1', 'flag{still-correct}')
    expect(wrapper.text()).toContain('Flag 校验通过，本题已解出，不重复计分')
  })
})
