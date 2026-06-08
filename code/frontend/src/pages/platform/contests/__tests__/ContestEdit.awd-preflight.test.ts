import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'

import {
  ApiError,
  buildContestDetail,
  contestApiMocks,
  contestEditPageModelSource,
  contestEditSource,
  createDeferred,
  destructiveConfirmMock,
  findRouteLink,
  getWorkbenchStageRail,
  mountContestEdit,
  mountContestEditWithRealChallengeDialog,
  platformContestEditPageSource,
  platformRoutesSource,
  pushMock,
  resetContestEditTestHarness,
  submitContestBasicsForm,
  toastMocks,
} from './ContestEdit.test-harness'

describe('ContestEdit', () => {
  beforeEach(() => {
    resetContestEditTestHarness()
  })

  it('应该在 AWD 启动门禁拦截后展示错误并停留在当前页面', async () => {
    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      description: '攻防赛',
      mode: 'awd',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest.mockRejectedValueOnce(
      new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    contestApiMocks.getContestAWDReadiness.mockClear()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({ status: 'running' })
    )
    expect(toastMocks.error).toHaveBeenCalledWith('AWD 开赛就绪检查未通过')
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.find('#awd-readiness-override-submit').exists()).toBe(false)
  })

  it('应该在 AWD 启动门禁拦截后不再读取 readiness 放行数据', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.updateContest.mockRejectedValueOnce(
      new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    contestApiMocks.getContestAWDReadiness.mockClear()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(toastMocks.error).toHaveBeenCalledWith('AWD 开赛就绪检查未通过')
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('基础信息')
  })

  it('赛前检查页面不应提供强制开赛入口', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-title').setValue('2026 AWD 联赛（演练版）')
    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()

    expect(wrapper.find('#contest-awd-preflight-force-start').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('强制启动赛事')
    expect(wrapper.text()).not.toContain('强制放行')
    expect(contestApiMocks.updateContest).not.toHaveBeenCalled()
  })

  it('应该在 AWD 辅助请求失败时仍保留工作台而不是进入全局加载错误态', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.getContestAWDReadiness.mockRejectedValue(new Error('readiness failed'))

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('基础信息')
    expect(wrapper.text()).toContain('基础信息')
    expect(wrapper.text()).not.toContain('竞赛详情加载失败')
  })
})
