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

  it('新增 AWD 题目应统一从题目编排打开题库选题弹层', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdChallenges.mockResolvedValueOnce({
      list: [
        {
          id: '999',
          name: 'Final Challenge',
          slug: 'final-template',
          category: 'crypto',
          difficulty: 'medium',
          description: 'final service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.find('#awd-challenge-config-create').exists()).toBe(false)
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(wrapper.get('#contest-workbench-stage-tab-pool').attributes('aria-selected')).toBe(
      'true'
    )
    expect(wrapper.find('#contest-challenge-library').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-select').exists()).toBe(false)
    expect(wrapper.html()).not.toContain(['contest', 'template', 'option', '999'].join('-'))
    expect(wrapper.find('#contest-awd-challenge-option-999').exists()).toBe(true)
    expect(wrapper.find('#awd-challenge-config-template').exists()).toBe(false)
    expect(contestApiMocks.getChallenges).not.toHaveBeenCalled()
    expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: 'published',
    })
    expect(wrapper.text()).toContain('Final Challenge')
  })

  it('应该在题目编排 AWD 题库加载失败时给出错误提示而不是留下未处理异常', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdChallenges.mockRejectedValueOnce(new Error('catalog failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('catalog failed')
    expect(wrapper.get('#contest-workbench-stage-tab-pool').attributes('aria-selected')).toBe(
      'true'
    )
    expect(wrapper.find('#contest-awd-challenge-list').exists()).toBe(true)
  })

  it('应该在 AWD 辅助数据仍在加载时显示阶段级加载提示而不是空态', async () => {
    const servicesDeferred = createDeferred<any[]>()

    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices.mockReturnValueOnce(servicesDeferred.promise)

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('正在同步 AWD 配置...')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目')

    servicesDeferred.resolve([
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        awd_challenge_id: '1',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        checker_type: undefined,
        checker_config: {},
        sla_score: 0,
        defense_score: 0,
        validation_state: 'pending',
        last_preview_at: undefined,
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    await flushPromises()
  })
  it('应该在 AWD 赛事的题目池阶段只展示题目编排信息', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {},
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_validation_state: 'stale',
        awd_checker_last_preview_at: '2026-04-12T08:00:00.000Z',
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        awd_challenge_id: '1',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 1,
          awd_defense_score: 2,
        },
        runtime_config: {
          checker_type: 'http_standard',
          checker_config: {},
        },
        checker_type: 'http_standard',
        checker_config: {},
        sla_score: 18,
        defense_score: 28,
        validation_state: 'stale',
        last_preview_at: '2026-04-12T08:00:00.000Z',
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
      },
    ])

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('题目资源')
    expect(wrapper.text()).toContain('可见性')
    expect(wrapper.text()).toContain('分值')
    expect(wrapper.text()).toContain('顺序')
    expect(wrapper.text()).not.toContain('未配置 AWD')
    expect(wrapper.text()).not.toContain('预检失败')
    expect(wrapper.text()).not.toContain('Checker')
    expect(wrapper.text()).not.toContain('SLA 18 / 防守 28')
    expect(wrapper.text()).not.toContain('待重新验证')
  })
})
