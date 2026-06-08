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

  it('路由页应仅负责组合，不直接耦合竞赛编辑加载与保存流程', () => {
    expect(contestEditSource).toContain('PlatformContestEditPage')
    expect(contestEditSource).not.toContain('useContestEditPage')
    expect(contestEditSource).not.toContain("from '@/api/admin/contests'")
    expect(contestEditPageModelSource).not.toContain("from 'vue-router'")
    expect(contestEditPageModelSource).toContain(
      "from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(contestEditPageModelSource).not.toContain('window.location.search')
    expect(platformRoutesSource).toContain("name: 'ContestEdit'")
    expect(platformRoutesSource).toContain(
      "component: () => import('@/pages/platform/contests/ContestEditRoutePage.vue')"
    )
    expect(platformRoutesSource).toContain("contestId: String(route.params.id || '')")
    expect(platformContestEditPageSource).toContain('useContestEditPage')
    expect(platformContestEditPageSource).toContain('ContestEditTopbarPanel')
    expect(platformContestEditPageSource).toContain('ContestEditWorkspacePanel')
    expect(platformContestEditPageSource).toContain('ContestWorkbenchStageTabs')
  })

  it('顶部应提供公告入口 route target', async () => {
    const wrapper = mountContestEdit()

    await flushPromises()

    expect(findRouteLink(wrapper, 'contest-open-announcements')?.props('to')).toEqual({
      name: 'ContestAnnouncements',
      params: { id: 'contest-1' },
    })
  })

  it('应该在普通赛下只展示基础信息与题目池阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())

    const wrapper = mountContestEditWithRealChallengeDialog()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目编排')
    expect(stageRail.text()).not.toContain('AWD 编排')
    expect(stageRail.text()).not.toContain('就绪审计')
    expect(stageRail.text()).not.toContain('轮次运行')
  })

  it('应该在 AWD 赛事下只展示编辑工作台阶段，不混入赛事运维阶段', async () => {
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

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目编排')
    expect(stageRail.text()).toContain('AWD 编排')
    expect(stageRail.text()).toContain('就绪审计')
    expect(stageRail.text()).not.toContain('轮次运行')
    expect(stageRail.text()).not.toContain('实例编排')
  })

  it('应该在赛前检查中列出阻塞项、移除强制开赛入口，并支持返回 AWD 配置', async () => {
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
    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('开赛已锁定')
    expect(wrapper.text()).toContain('开赛门禁')
    expect(wrapper.text()).toContain('轮次创建')
    expect(wrapper.text()).toContain('即时巡检')
    expect(wrapper.text()).toContain('Challenge 101')
    expect(wrapper.text()).toContain('编辑并试跑')
    expect(wrapper.find('#contest-awd-preflight-force-start').exists()).toBe(false)

    expect(findRouteLink(wrapper, 'awd-readiness-edit-1')?.props('to')).toEqual({
      name: 'ContestAWDConfig',
      params: { id: 'contest-1' },
      query: { service: 'service-1' },
    })
    expect(wrapper.text()).not.toContain('当前焦点题目')
  })

  it('题目池更多菜单不应提供 AWD 编排跳转入口', async () => {
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

    expect(wrapper.find('#contest-challenge-actions-1').exists()).toBe(false)
    expect(document.body.querySelector('#contest-challenge-open-awd-config-1')).toBeNull()
    expect(
      getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()
    ).toContain('题目编排')
  })

  it('竞赛编辑页不应渲染轮次运行面板和赛事运维内容', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(wrapper.find('#contest-workbench-stage-tab-instances').exists()).toBe(false)
    expect(wrapper.find('#awd-readiness-edit-101').exists()).toBe(false)
    expect(wrapper.find('.runtime-readiness-strip').exists()).toBe(false)
    expect(wrapper.find('#awd-ops-panel-inspector').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('本轮得分')
    expect(wrapper.text()).not.toContain('攻击流水')
  })

  it('应该在 AWD 赛事已开赛时默认聚焦 AWD 编排阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
  })

  it('旧 operations URL 在编辑页应回落到默认编辑阶段', async () => {
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit?panel=operations')
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

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('题目编排')
    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('尚未进入运行阶段')
  })

  it('AWD 赛事已结束时仍停留在编辑配置阶段，报告导出进入赛事运维页处理', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'ended',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
    expect(wrapper.text()).not.toContain('尚未进入运行阶段')
  })

  it('AWD 题目列表刷新失败时应保留上次成功数据并避免把摘要误报为 0', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices
      .mockResolvedValueOnce([
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
      .mockRejectedValueOnce(new Error('refresh failed'))

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('同步数据'))
      ?.trigger('click')
    await flushPromises()

    expect(toastMocks.error).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })

  it('应该在管理页工作台交接时忽略旧运维子页签并落到编辑阶段', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_ops_panel:contest-1', 'challenges')
    window.history.replaceState(
      {},
      '',
      '/platform/contests/contest-1/edit?panel=operations&opsPanel=inspector'
    )
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.find('#awd-ops-tab-challenges').exists()).toBe(false)
    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(
      getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()
    ).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
  })

  it('旧 operations URL 不再作为编辑页有效阶段保留', async () => {
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit?panel=operations')
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

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('题目编排')
    expect(window.location.search).not.toContain('panel=operations')
  })
})
