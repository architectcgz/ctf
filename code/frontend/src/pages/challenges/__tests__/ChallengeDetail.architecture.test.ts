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

  it('题目详情 page feature 应通过共享 route transport 承接路由 owner', () => {
    expect(challengeDetailPageSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(challengeDetailPageSource).toContain(
      "import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'"
    )
    expect(challengeDetailPageSource).toContain(
      "import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'"
    )
    expect(challengeDetailPageSource).not.toContain(
      "import { useUrlSyncedTabs } from '@/shared/model/navigation/useUrlSyncedTabs'"
    )
    expect(challengeDetailPageSource).toContain('useRouteQueryTabs<WorkspaceTab>({')
    expect(challengeDetailPageSource).toContain("from './challengeDetailRoutes'")
    expect(challengeDetailPageSource).not.toContain("from 'vue-router'")
    expect(challengeDetailRoutesSource).toContain("name: 'Challenges'")
  })

  it('应仅保留外层主容器卡片并移除内部二级卡片', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.findAll('.challenge-panel')).toHaveLength(0)
  })

  it('工作区应建立满高伸展布局链', () => {
    expect(challengeDetailWorkspaceSource).toContain('min-height: max(100%, calc(100vh - 5rem));')
    expect(challengeWorkspaceShellSource).toContain(
      '.detail-content {\n  display: flex;\n  flex: 1 1 auto;'
    )
    expect(challengeWorkspaceShellSource).toMatch(
      /\.detail-grid,\s*\.workspace-grid\s*{\s*display:\s*grid;\s*flex:\s*1 1 auto;/
    )
  })

  it('题目详情 hero 应使用共享 workspace overline 语义', () => {
    expect(challengeQuestionPanelSource).toContain('workspace-overline')
    expect(challengeQuestionPanelSource).toContain('Question')
    expect(challengeQuestionPanelSource).not.toContain('overline">Question')
  })

  it('题目详情应把 tab 下方面板间距收口到共享 workspace token', () => {
    expect(challengeDetailWorkspaceSource).toContain('--workspace-tabs-panel-gap: var(--space-2);')
    expect(challengeDetailWorkspaceSource).toContain(
      '--workspace-panel-padding-top: var(--workspace-tabs-panel-gap);'
    )
    expect(challengeWorkspaceShellSource).toMatch(
      /\.detail-main,\s*\.content-pane\s*\{[\s\S]*padding:\s*0\s+var\(--space-workspace-content-padding,\s*var\(--space-7\)\)\s+var\(--space-workspace-content-padding,\s*var\(--space-7\)\);/s
    )
    expect(challengeWorkspaceShellSource).toMatch(
      /\.tool-pane\s*\{[\s\S]*padding:\s*var\(--workspace-tabs-panel-gap,\s*var\(--space-2\)\)\s+var\(--space-workspace-content-padding,\s*var\(--space-7\)\)\s+var\(--space-workspace-content-padding,\s*var\(--space-7\)\);/s
    )
  })

  it('题目详情 section heading 应切到共享 workspace overline 语义', () => {
    const combinedSource = [
      challengeDetailWorkspaceSource,
      challengeWorkspaceShellSource,
      challengeQuestionPanelSource,
      challengeSolutionsPanelSource,
      challengeSubmissionRecordsPanelSource,
      challengeWriteupPanelSource,
    ].join('\n')

    for (const label of ['Statement', 'Hints', 'Solutions', 'Submissions', 'My Writeup']) {
      expect(combinedSource).toContain('workspace-overline')
      expect(combinedSource).toContain(label)
      expect(combinedSource).not.toContain(`overline">${label}`)
    }
  })

  it('题目详情剩余局部 kicker 也应统一到 workspace overline 语义', () => {
    expect(challengeActionAsideSource).toContain('workspace-overline')
    expect(challengeActionAsideSource).toContain('Primary Action')
    expect(challengeActionAsideSource).not.toContain('overline">Primary Action')
    expect(challengeDetailWorkspaceSource).not.toMatch(/^\.overline\s*\{/m)
  })

  it('题目详情实例卡片应复用实例实体展示 owner', () => {
    expect(challengeInstanceCardSource).toContain("from '@/entities/instance'")
    expect(challengeInstanceCardSource).toContain('getInstanceStatusTone')
    expect(challengeInstanceCardSource).toContain('getInstanceWaitingQueueLabel')
    expect(challengeInstanceCardSource).toContain('getInstanceWaitingEtaLabel')
    expect(challengeInstanceCardSource).toContain('getInstanceWaitingProgressLabel')
    expect(challengeInstanceCardSource).toContain('formatInstanceAccessDisplay')
    expect(instancePresentationSource).toContain('getInstanceWaitingQueueLabel')
    expect(instancePresentationSource).toContain('getInstanceWaitingEtaLabel')
    expect(instancePresentationSource).toContain('getInstanceWaitingProgressLabel')
  })
})
