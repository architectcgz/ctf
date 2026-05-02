import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CheatDetection from '../CheatDetection.vue'
import cheatDetectionSource from '../CheatDetection.vue?raw'

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

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('CheatDetection', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
    adminApiMocks.getCheatDetection.mockReset()
  })

  it('应该默认渲染单页风险工作台，并支持从审计联动区跳转到日志', async () => {
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
    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('高频提交账号')
    expect(wrapper.text()).toContain('共享 IP 线索')
    expect(wrapper.text()).toContain('审计联动')

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

  it('路由页应仅负责组合，不直接耦合风险检测请求流程', () => {
    expect(cheatDetectionSource).toContain('useCheatDetectionPage')
    expect(cheatDetectionSource).not.toContain("from '@/api/admin/platform'")
  })

  it('应兼容旧 panel query，并继续渲染同一风险工作台', async () => {
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

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('高频提交账号')
    expect(wrapper.text()).toContain('共享 IP 线索')
    expect(wrapper.text()).toContain('审计联动')
    expect(replaceMock).not.toHaveBeenCalled()
  })

  it('父页应保留风险数据刷新和审计跳转 owner', async () => {
    adminApiMocks.getCheatDetection.mockResolvedValue({
      generated_at: '2026-03-07T06:00:00.000Z',
      summary: {
        submit_burst_users: 1,
        shared_ip_groups: 1,
        affected_users: 2,
      },
      suspects: [],
      shared_ips: [],
    })

    const wrapper = mount(CheatDetection, {
      global: {
        stubs: {
          CheatDetectionWorkspacePanel: {
            props: ['riskData', 'loading'],
            emits: ['open-audit', 'refresh'],
            template:
              '<div><div data-testid="risk-state">{{ riskData?.generated_at }}::{{ loading }}</div><button id="cheat-open-audit" type="button" @click="$emit(\'open-audit\', { action: \'submit\' })">打开审计</button><button id="cheat-refresh" type="button" @click="$emit(\'refresh\')">刷新</button></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="risk-state"]').text()).toContain('2026-03-07T06:00:00.000Z')
    expect(adminApiMocks.getCheatDetection).toHaveBeenCalledTimes(1)

    await wrapper.get('#cheat-open-audit').trigger('click')
    expect(pushMock).toHaveBeenLastCalledWith({
      name: 'AuditLog',
      query: { action: 'submit' },
    })

    await wrapper.get('#cheat-refresh').trigger('click')
    await flushPromises()

    expect(adminApiMocks.getCheatDetection).toHaveBeenCalledTimes(2)
  })
})
