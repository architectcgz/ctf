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

  it('应该支持保存个人题解草稿', async () => {
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

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '编写题解')
    expect(writeupTab).toBeTruthy()
    await writeupTab!.trigger('click')
    await flushPromises()

    const titleInput = wrapper.find('input[placeholder*="完整链路"]')
    const contentInput = wrapper.find('textarea')
    const draftButton = wrapper.findAll('button').find((node) => node.text().trim() === '保存草稿')

    await titleInput.setValue('我的题解')
    await contentInput.setValue('先找回显，再定位注入。')
    await draftButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.upsertChallengeWriteupSubmission).toHaveBeenCalledWith('1', {
      title: '我的题解',
      content: '先找回显，再定位注入。',
      submission_status: 'draft',
    })
    expect(wrapper.text()).toContain('草稿')
  })

  it('编写题解应通过顶部标签切换进入，默认不显示编辑区', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).not.toContain('解题过程复盘')
    expect(wrapper.find('input[placeholder*="完整链路"]').exists()).toBe(false)
    expect(challengeApiMocks.getMyChallengeWriteupSubmission).not.toHaveBeenCalled()

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '编写题解')
    expect(writeupTab).toBeTruthy()

    await writeupTab!.trigger('click')
    await flushPromises()

    expect(challengeApiMocks.getMyChallengeWriteupSubmission).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('解题过程复盘')
    expect(wrapper.find('input[placeholder*="完整链路"]').exists()).toBe(true)
  })

  it('只有题目标签显示右侧工具区，其他标签应切换为单栏内容', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('Flag 提交')

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '编写题解')

    expect(solutionTab).toBeTruthy()
    expect(recordsTab).toBeTruthy()
    expect(writeupTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('题解区')

    await recordsTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('提交记录')

    await writeupTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('解题过程复盘')
  })

  it('只有切到题目标签时才显示题目基本信息，题解标签下不重复显示题头', async () => {
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

    expect(wrapper.text()).toContain('Solved Challenge')
    expect(wrapper.text()).toContain('分值')

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Solved Challenge')
    expect(wrapper.text()).not.toContain('分值')
    expect(wrapper.text()).toContain('推荐题解')
  })
})
