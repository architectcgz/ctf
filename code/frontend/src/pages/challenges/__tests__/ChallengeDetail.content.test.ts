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

  it('提示内容应支持前端展开查看且不再调用解锁接口', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).not.toContain('先观察登录表单的参数。')
    expect(wrapper.text()).not.toContain('解锁提示')

    const toggleButton = wrapper.find('button.hint-toggle')
    expect(toggleButton.exists()).toBe(true)
    expect(toggleButton.text()).toContain('展开提示')

    await toggleButton.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('先观察登录表单的参数。')
    expect(challengeApiMocks.unlockHint).not.toHaveBeenCalled()
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

  it('点击分值侧栏达到阈值后应显示试探提示', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).not.toContain('这块区域的情报价值，低于你现在的期待。')

    const scoreRail = wrapper.get('.score-rail')
    await scoreRail.trigger('click')
    await scoreRail.trigger('click')
    await scoreRail.trigger('click')
    await scoreRail.trigger('click')

    expect(wrapper.text()).toContain('这块区域的情报价值，低于你现在的期待。')
    expect(wrapper.text()).toContain('下载附件')
    expect(wrapper.text()).toContain('题目描述')
  })

  it('学生下载内部附件时应直接走浏览器原生下载链路', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Test Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: false,
      attachment_url: '/api/v1/challenges/attachments/imports/demo.zip',
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const originalCreateElement = document.createElement.bind(document)
    const clickMock = vi.fn()
    let capturedAnchor: HTMLAnchorElement | null = null
    const createElementSpy = vi
      .spyOn(document, 'createElement')
      .mockImplementation((tagName: string) => {
        if (tagName === 'a') {
          const anchor = originalCreateElement(tagName)
          anchor.click = clickMock
          capturedAnchor = anchor
          return anchor
        }
        return originalCreateElement(tagName)
      })

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()

    const downloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('下载附件'))
    expect(downloadButton).toBeTruthy()

    await downloadButton!.trigger('click')
    await flushPromises()

    expect(challengeApiMocks.downloadAttachment).not.toHaveBeenCalled()
    expect(clickMock).toHaveBeenCalled()
    expect(capturedAnchor).not.toBeNull()
    if (!capturedAnchor) {
      throw new Error('expected download anchor to be created')
    }
    const anchor = capturedAnchor as HTMLAnchorElement
    expect(anchor.href).toContain('/api/v1/challenges/attachments/imports/demo.zip')
    expect(anchor.hasAttribute('download')).toBe(true)

    createElementSpy.mockRestore()
  })
})
