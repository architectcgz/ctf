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

  it('应该加载竞赛详情并在保存成功后返回赛事目录', async () => {
    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('基础信息')

    await wrapper.get('#contest-title').setValue('2026 春季校园 CTF（更新）')
    await submitContestBasicsForm(wrapper)
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF（更新）',
      })
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该在终止进行中的竞赛前弹出二次确认', async () => {
    destructiveConfirmMock.mockResolvedValue(false)
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 春季校园 CTF',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-status').setValue('ended')
    await submitContestBasicsForm(wrapper)

    expect(destructiveConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '确认结束赛事',
      })
    )
    expect(contestApiMocks.updateContest).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('应该在进行中的竞赛切换为已冻结时省略不可修改的时间字段', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 春季校园 CTF',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-status').setValue('frozen')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF',
        status: 'frozen',
      })
    )
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.not.objectContaining({
        starts_at: expect.anything(),
      })
    )
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.not.objectContaining({
        ends_at: expect.anything(),
      })
    )
  })
})
