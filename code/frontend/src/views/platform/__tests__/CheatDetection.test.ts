import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'

import cheatDetectionPageSource from '@/features/platform-overview/model/useCheatDetectionPage.ts?raw'
import platformOverviewRoutesSource from '@/features/platform-overview/model/platformOverviewRoutes.ts?raw'
import CheatDetection from '../CheatDetection.vue'
import cheatDetectionSource from '../CheatDetection.vue?raw'

const adminApiMocks = vi.hoisted(() => ({
  getCheatDetection: vi.fn(),
}))

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('CheatDetection', () => {
  beforeEach(() => {
    adminApiMocks.getCheatDetection.mockReset()
  })

  it('应该默认渲染单页风险工作台，并通过 route target 渲染审计联动入口', async () => {
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

    const wrapper = mount(CheatDetection, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('作弊检测')
    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('高频提交账号')
    expect(wrapper.text()).toContain('共享 IP 线索')
    expect(wrapper.text()).toContain('审计联动')

    const quickAction = wrapper
      .findAllComponents(RouterLinkStub)
      .find((link) => link.text().includes('查看提交记录'))
    expect(quickAction).toBeTruthy()
    expect(quickAction!.props('to')).toEqual({
      name: 'AuditLog',
      query: { action: 'submit' },
    })
  })

  it('路由页应仅负责组合，不直接耦合风险检测请求流程', () => {
    expect(cheatDetectionSource).toContain('useCheatDetectionPage')
    expect(cheatDetectionSource).not.toContain("from '@/api/admin/platform'")
    expect(cheatDetectionPageSource).not.toContain("from 'vue-router'")
    expect(cheatDetectionPageSource).toContain('auditLogRoute: buildPlatformAuditLogRoute()')
    expect(cheatDetectionPageSource).toContain('buildAuditRoute: buildPlatformAuditLogRoute')
    expect(platformOverviewRoutesSource).toContain('buildPlatformAuditLogRoute')
  })

  it('应通过 route target contract 渲染账户和共享 IP 的审计入口', async () => {
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

    const wrapper = mount(CheatDetection, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })
    await flushPromises()

    const links = wrapper.findAllComponents(RouterLinkStub)
    expect(
      links.some((link) =>
        JSON.stringify(link.props('to')).includes('"actor_user_id":"8"')
      )
    ).toBe(true)
    expect(
      links.some((link) => JSON.stringify(link.props('to')).includes('"action":"login"'))
    ).toBe(true)
  })

  it('父页应保留风险数据刷新 owner，并把审计跳转下沉为 route target contract', async () => {
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
          RouterLink: RouterLinkStub,
          CheatDetectionWorkspacePanel: {
            props: ['riskData', 'loading', 'auditLogRoute', 'buildAuditRoute'],
            emits: ['refresh'],
            template:
              '<div><div data-testid="risk-state">{{ riskData?.generated_at }}::{{ loading }}</div><div data-testid="audit-route">{{ JSON.stringify(auditLogRoute) }}</div><div data-testid="audit-route-submit">{{ JSON.stringify(buildAuditRoute({ action: \'submit\' })) }}</div><button id="cheat-refresh" type="button" @click="$emit(\'refresh\')">刷新</button></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="risk-state"]').text()).toContain('2026-03-07T06:00:00.000Z')
    expect(wrapper.get('[data-testid="audit-route"]').text()).toContain('"name":"AuditLog"')
    expect(wrapper.get('[data-testid="audit-route-submit"]').text()).toContain('"action":"submit"')
    expect(adminApiMocks.getCheatDetection).toHaveBeenCalledTimes(1)

    await wrapper.get('#cheat-refresh').trigger('click')
    await flushPromises()

    expect(adminApiMocks.getCheatDetection).toHaveBeenCalledTimes(2)
  })
})
