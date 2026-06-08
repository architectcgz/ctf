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

  it('题目池变更后应同步更新 AWD 配置与赛前检查数据', async () => {
    const awdServicesState: any[] = []
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices.mockImplementation(async () =>
      awdServicesState.map((item) => ({ ...item }))
    )
    contestApiMocks.getContestAWDReadiness.mockImplementation(async () => ({
      contest_id: 'contest-1',
      ready: awdServicesState.length > 0,
      total_challenges: awdServicesState.length,
      passed_challenges: awdServicesState.length,
      pending_challenges: 0,
      failed_challenges: 0,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 0,
      global_blocking_reasons: awdServicesState.length > 0 ? [] : ['no_challenges'],
      blocking_actions: awdServicesState.length > 0 ? [] : ['start_contest'],
      items: [],
    }))
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 模板',
          slug: 'upload-http',
          category: 'web',
          difficulty: 'medium',
          description: 'http service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.createContestAWDService.mockImplementation(async (_contestId, payload) => {
      expect(payload).toEqual({
        awd_challenge_id: 11,
        points: 100,
        order: 0,
        is_visible: true,
      })
      awdServicesState.push({
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '11',
        awd_challenge_id: '11',
        title: 'Upload Service',
        category: 'web',
        difficulty: 'medium',
        display_name: 'Upload Service',
        order: 0,
        is_visible: true,
        score_config: {
          points: 100,
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
        created_at: '2026-03-10T01:00:00.000Z',
        updated_at: '2026-03-10T01:00:00.000Z',
      })
      return awdServicesState[0]
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-11').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Upload Service')

    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('可以开赛')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目，无法执行开赛关键动作')
  })

  it('AWD 题目从题目池移除时应删除显式 service 而不是只删 challenge 关联', async () => {
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
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-remove-1').trigger('click')
    await flushPromises()

    expect(contestApiMocks.deleteContestAWDService).toHaveBeenCalledWith('contest-1', 'service-1')
    expect(contestApiMocks.deleteAdminContestChallenge).not.toHaveBeenCalled()
  })

  it('AWD 配置变更后题目池应同步更新，并允许继续打开新增对话框', async () => {
    const awdServicesState: any[] = [
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
    ]
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices.mockImplementation(async () =>
      awdServicesState.map((item) => ({ ...item }))
    )
    contestApiMocks.createContestAWDService.mockImplementation(async (_contestId, payload) => {
      expect(payload).toEqual({
        awd_challenge_id: 1,
        points: 100,
        order: 0,
        is_visible: true,
      })
      const created = {
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '1',
        awd_challenge_id: '1',
        title: 'Bank Portal AWD',
        category: 'web',
        difficulty: 'medium',
        display_name: 'Bank Portal AWD',
        order: 0,
        is_visible: true,
        score_config: {
          points: 100,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        created_at: '2026-03-10T01:00:00.000Z',
        updated_at: '2026-03-10T01:00:00.000Z',
      }
      awdServicesState.push(created)
      return created
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-1').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Bank Portal AWD')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    expect(wrapper.find('#contest-challenge-library').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-challenge-option-1').exists()).toBe(true)
  })

  it('题目编排新增 AWD 题目保存失败时应提示错误并保持弹层打开', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.createContestAWDService.mockRejectedValueOnce(new Error('save failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-1').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('部分 AWD 题目关联失败：Bank Portal AWD')
    expect(wrapper.find('#contest-awd-challenge-option-1').exists()).toBe(true)
  })
})
