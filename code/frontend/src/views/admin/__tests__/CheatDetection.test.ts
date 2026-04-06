import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CheatDetection from '../CheatDetection.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string>,
}))
const adminApiMocks = vi.hoisted(() => ({
  getCheatDetection: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('CheatDetection', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
    adminApiMocks.getCheatDetection.mockReset()
  })

  it('应该默认显示总览 tab，并支持切换到快速入口后跳转到审计日志', async () => {
    adminApiMocks.getCheatDetection.mockResolvedValue({
      generated_at: '2026-03-07T06:00:00.000Z',
      summary: {
        submit_burst_users: 1,
        shared_ip_groups: 1,
        affected_users: 2,
      },
      suspects: [
        {
          user_id: '8',
          username: 'alice',
          submit_count: 9,
          last_seen_at: '2026-03-07T05:58:00.000Z',
          reason: '短时间内提交次数异常偏高',
        },
      ],
      shared_ips: [
        {
          ip: '10.0.0.1',
          user_count: 2,
          usernames: ['alice', 'bob'],
        },
      ],
    })

    const wrapper = mount(CheatDetection)
    await flushPromises()

    expect(wrapper.text()).toContain('作弊检测')
    expect(wrapper.find('#cheat-tab-overview').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#cheat-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#cheat-panel-suspects').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#cheat-panel-shared-ip').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#cheat-panel-actions').attributes('aria-hidden')).toBe('true')

    await wrapper.get('#cheat-tab-actions').trigger('click')
    await flushPromises()

    const quickAction = wrapper
      .findAll('button')
      .find((button) => button.text().includes('查看提交记录'))
    expect(quickAction).toBeTruthy()

    await quickAction!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'AuditLog',
      query: { action: 'submit' },
    })
  })

  it('应该根据 query 预选 tab，并在切换时同步 panel 参数', async () => {
    routeState.query = { panel: 'shared-ip' }

    adminApiMocks.getCheatDetection.mockResolvedValue({
      generated_at: '2026-03-07T06:00:00.000Z',
      summary: {
        submit_burst_users: 1,
        shared_ip_groups: 1,
        affected_users: 2,
      },
      suspects: [
        {
          user_id: '8',
          username: 'alice',
          submit_count: 9,
          last_seen_at: '2026-03-07T05:58:00.000Z',
          reason: '短时间内提交次数异常偏高',
        },
      ],
      shared_ips: [
        {
          ip: '10.0.0.1',
          user_count: 2,
          usernames: ['alice', 'bob'],
        },
      ],
    })

    const wrapper = mount(CheatDetection)
    await flushPromises()

    expect(wrapper.find('#cheat-tab-shared-ip').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#cheat-panel-shared-ip').attributes('aria-hidden')).toBe('false')

    await wrapper.get('#cheat-tab-overview').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      name: 'CheatDetection',
      query: {},
    })
  })
})
