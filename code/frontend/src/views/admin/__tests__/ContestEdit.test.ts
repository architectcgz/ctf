import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestEdit from '../ContestEdit.vue'
import { ApiError } from '@/api/request'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' } as Record<string, string>,
}))

const contestApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  updateContest: vi.fn(),
  getContestAWDReadiness: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
  }
})

describe('ContestEdit', () => {
  beforeEach(() => {
    pushMock.mockReset()
    contestApiMocks.getContest.mockReset()
    contestApiMocks.updateContest.mockReset()
    contestApiMocks.getContestAWDReadiness.mockReset()
    routeState.params = { id: 'contest-1' }

    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 春季校园 CTF',
      description: '校内赛',
      mode: 'jeopardy',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest.mockResolvedValue({
      contest: {
        id: 'contest-1',
        title: '2026 春季校园 CTF（更新）',
        description: '校内赛',
        mode: 'jeopardy',
        status: 'registering',
        starts_at: '2026-03-15T09:00:00.000Z',
        ends_at: '2026-03-15T13:00:00.000Z',
      },
    })
    contestApiMocks.getContestAWDReadiness.mockResolvedValue({
      contest_id: 'contest-1',
      ready: false,
      total_challenges: 1,
      passed_challenges: 0,
      pending_challenges: 0,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 1,
      global_blocking_reasons: [],
      blocking_actions: ['start_contest'],
      items: [
        {
          challenge_id: '101',
          title: 'Challenge 101',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
  })

  it('应该加载竞赛详情并在保存成功后返回赛事目录', async () => {
    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('编辑竞赛')

    await wrapper.get('#contest-title').setValue('2026 春季校园 CTF（更新）')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF（更新）',
      })
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该在 AWD 启动门禁拦截后展示放行弹层并在确认后回到赛事目录', async () => {
    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      description: '攻防赛',
      mode: 'awd',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest
      .mockRejectedValueOnce(new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 }))
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-1',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      })

    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getContestAWDReadiness).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('启动赛事')
    expect(wrapper.text()).toContain('强制继续')

    await wrapper.get('#awd-readiness-override-reason').setValue('teacher drill')
    await wrapper.get('#awd-readiness-override-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenNthCalledWith(
      2,
      'contest-1',
      expect.objectContaining({
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      }),
      { suppressErrorToast: true }
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })
})
