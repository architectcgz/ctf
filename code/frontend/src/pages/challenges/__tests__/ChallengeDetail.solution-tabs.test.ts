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

  it('顶部主切换应暴露 tabs 语义，题解页签下仍保留次级 tabs 语义', async () => {
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

    const tablists = wrapper.findAll('[role="tablist"]')
    expect(tablists).toHaveLength(1)

    const topTabs = wrapper.findAll('[role="tab"]')
    expect(topTabs).toHaveLength(4)
    expect(topTabs[0].attributes('aria-selected')).toBe('true')
    expect(topTabs[1].attributes('aria-selected')).toBe('false')

    await topTabs[1].trigger('click')
    await flushPromises()

    const nestedTablists = wrapper.findAll('[role="tablist"]')
    expect(nestedTablists.length).toBeGreaterThanOrEqual(2)
    expect(wrapper.find('[role="tabpanel"]').exists()).toBe(true)
  })

  it('workspace 主标签应支持 ArrowLeft、ArrowRight、Home 和 End 键盘切换', async () => {
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
      attachTo: document.body,
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const questionTab = wrapper.get('#challenge-workspace-tab-question')
    const solutionTab = wrapper.get('#challenge-workspace-tab-solution')
    const writeupTab = wrapper.get('#challenge-workspace-tab-writeup')
    const questionButton = questionTab.element as HTMLButtonElement
    const solutionButton = solutionTab.element as HTMLButtonElement
    const writeupButton = writeupTab.element as HTMLButtonElement

    questionButton.focus()
    expect(questionTab.attributes('aria-selected')).toBe('true')
    expect(solutionTab.attributes('aria-selected')).toBe('false')

    await questionTab.trigger('keydown', { key: 'ArrowLeft' })
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBe('writeup')
    expect(questionTab.attributes('aria-selected')).toBe('false')
    expect(writeupTab.attributes('aria-selected')).toBe('true')
    expect(document.activeElement).toBe(writeupButton)

    await writeupTab.trigger('keydown', { key: 'Home' })
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBeUndefined()
    expect(questionTab.attributes('aria-selected')).toBe('true')
    expect(writeupTab.attributes('aria-selected')).toBe('false')
    expect(document.activeElement).toBe(questionButton)

    await questionTab.trigger('keydown', { key: 'ArrowRight' })
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBe('solution')
    expect(solutionTab.attributes('aria-selected')).toBe('true')
    expect(document.activeElement).toBe(solutionButton)
    expect(wrapper.text()).toContain('推荐题解')

    await solutionTab.trigger('keydown', { key: 'End' })
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBe('writeup')
    expect(writeupTab.attributes('aria-selected')).toBe('true')
    expect(document.activeElement).toBe(writeupButton)
    expect(wrapper.text()).toContain('解题过程复盘')

    wrapper.unmount()
  })

  it('题解子标签应支持 End、Home 和方向键切换', async () => {
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
      attachTo: document.body,
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

    const recommendedTab = wrapper.get('#challenge-solutions-tab-recommended')
    const communityTab = wrapper.get('#challenge-solutions-tab-community')
    const recommendedButton = recommendedTab.element as HTMLButtonElement
    const communityButton = communityTab.element as HTMLButtonElement

    recommendedButton.focus()
    expect(recommendedTab.attributes('aria-selected')).toBe('true')
    expect(communityTab.attributes('aria-selected')).toBe('false')

    await recommendedTab.trigger('keydown', { key: 'End' })
    await wrapper.vm.$nextTick()
    expect(recommendedTab.attributes('aria-selected')).toBe('false')
    expect(communityTab.attributes('aria-selected')).toBe('true')
    expect(document.activeElement).toBe(communityButton)

    await communityTab.trigger('keydown', { key: 'Home' })
    await wrapper.vm.$nextTick()
    expect(recommendedTab.attributes('aria-selected')).toBe('true')
    expect(communityTab.attributes('aria-selected')).toBe('false')
    expect(document.activeElement).toBe(recommendedButton)

    await recommendedTab.trigger('keydown', { key: 'ArrowRight' })
    await wrapper.vm.$nextTick()
    expect(recommendedTab.attributes('aria-selected')).toBe('false')
    expect(communityTab.attributes('aria-selected')).toBe('true')
    expect(document.activeElement).toBe(communityButton)

    wrapper.unmount()
  })
})
