import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import {
  contestApiMocks,
  contestDetailPageSource,
  contestDetailRoutePageSource,
  contestDetailSource,
  contestDetailWorkspaceSource,
  contestPresentationSource,
  contestTeamDialogsSource,
  contestTeamPanelSource,
  contestTeamWorkspaceSectionSource,
  destructiveConfirmMock,
  resetContestDetailTestHarness,
  routeQueryTransportSource,
  router,
  teamPresentationSource,
  webSocketMocks,
} from './ContestDetail.test-harness'
import ContestDetail from '@/pages/contests/ContestDetailRoutePage.vue'

describe('ContestDetail', () => {
  beforeEach(async () => {
    await resetContestDetailTestHarness()
  })

  it('队伍页创建和加入弹窗应切换到 C 端输入模板', async () => {
    expect(contestDetailWorkspaceSource).toContain('ContestTeamDialogs,')
    expect(contestDetailWorkspaceSource).toContain("} from '@/features/contest-detail'")
    expect(contestTeamDialogsSource).toContain(
      "from '@/shared/ui/common/modal-templates/CFocusedInputDialog.vue'"
    )

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
      attachTo: document.body,
    })

    await router.push('/contests/1?panel=team')
    await router.isReady()
    await flushPromises()

    const teamTab = wrapper.findAll('button').find((node) => node.text().trim() === '队伍')
    expect(teamTab).toBeTruthy()
    await teamTab!.trigger('click')
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '创建队伍')
      ?.trigger('click')
    await flushPromises()
    expect(document.body.textContent).toContain('创建新队伍')
    expect(document.body.textContent).toContain('队伍名称')

    const closeButtons = Array.from(document.body.querySelectorAll('button'))
    const cancelCreateButton = closeButtons.find((button) => button.textContent?.trim() === '取消')
    cancelCreateButton?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '加入队伍')
      ?.trigger('click')
    await flushPromises()
    expect(document.body.textContent).toContain('加入现有队伍')
    expect(document.body.textContent).toContain('队伍 ID')

    wrapper.unmount()
  })

  it('踢出队员前应走统一确认弹窗', async () => {
    contestApiMocks.getMyTeam.mockResolvedValueOnce({
      id: 'team-1',
      name: 'Red',
      captain_user_id: 'user-1',
      invite_code: 'RED-CTF',
      members: [
        { user_id: 'user-1', username: 'alice' },
        { user_id: 'user-2', username: 'bob' },
      ],
    })
    contestApiMocks.kickTeamMember.mockResolvedValue(undefined)

    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().setAuth({ id: 'user-1', username: 'alice', role: 'student' })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [pinia, router],
      },
    })

    await router.push('/contests/1?panel=team')
    await router.isReady()
    await flushPromises()

    const teamTab = wrapper.findAll('button').find((node) => node.text().trim() === '队伍')
    expect(teamTab).toBeTruthy()
    await teamTab!.trigger('click')
    await flushPromises()

    const kickButton = wrapper.findAll('button').find((node) => node.text().trim() === '踢出')
    expect(kickButton).toBeTruthy()
    await kickButton!.trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalledWith({
      title: '踢出成员',
      message: '确定踢出该成员？',
      confirmButtonText: '确认踢出',
    })
    expect(contestApiMocks.kickTeamMember).toHaveBeenCalledWith('1', 'team-1', 'user-2')
  })
})
