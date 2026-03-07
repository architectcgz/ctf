import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CheatDetection from '../CheatDetection.vue'

const pushMock = vi.fn()
const adminApiMocks = vi.hoisted(() => ({
  getCheatDetection: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('CheatDetection', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getCheatDetection.mockReset()
  })

  it('应该渲染真实作弊检测结果并支持跳转到审计日志', async () => {
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
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('10.0.0.1')

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
})
