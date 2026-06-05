import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AuditLog from '@/pages/platform/AuditLogRoutePage.vue'
import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import auditActorDetailModalSource from '@/features/audit-log/ui/AuditActorDetailModal.vue?raw'
import auditLogHeroPanelSource from '@/features/audit-log/ui/AuditLogHeroPanel.vue?raw'
import auditLogDirectoryPanelSource from '@/features/audit-log/ui/AuditLogDirectoryPanel.vue?raw'
import auditLogPageSource from '@/features/audit-log/model/useAuditLogPage.ts?raw'
import routeQueryTransportSource from '@/shared/model/navigation/useRouteQueryTransport.ts?raw'

const replaceMock = vi.fn()

const routeState = vi.hoisted(() => ({
  query: {
    action: 'submit',
    resource_type: 'challenge',
    actor_user_id: '12',
    page: '2',
  } as Record<string, string>,
}))

const adminApiMocks = vi.hoisted(() => ({
  getAuditLogs: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ replace: replaceMock }),
  }
})

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('AuditLog', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    replaceMock.mockReset()
    routeState.query = {
      action: 'submit',
      resource_type: 'challenge',
      actor_user_id: '12',
      page: '2',
    }

    adminApiMocks.getAuditLogs.mockReset()
    adminApiMocks.getAuditLogs.mockResolvedValue({
      list: [
        {
          id: 'log-1',
          action: 'submit',
          resource_type: 'challenge',
          resource_id: '5',
          actor_user_id: '12',
          actor_username: 'alice',
          created_at: '2026-03-07T10:00:00Z',
          detail: { status: 'accepted', challenge: 'web-basic' },
        },
      ],
      total: 24,
      page: 2,
      page_size: 20,
    })
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
    document.body.innerHTML = ''
  })

  it('应该根据路由 query 加载预置筛选结果', async () => {
    const wrapper = mount(AuditLog)

    await flushPromises()

    expect(adminApiMocks.getAuditLogs).toHaveBeenLastCalledWith(
      {
        page: 2,
        page_size: 20,
        action: 'submit',
        resource_type: 'challenge',
        actor_user_id: 12,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('submit')
  })

  it('应该在应用筛选时同步 query', async () => {
    const wrapper = mount(AuditLog)

    await flushPromises()

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    await flushPromises()

    const resourceInput = wrapper.find('input[placeholder="资源类型，如 challenge"]')
    await resourceInput.setValue('instance')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      query: {
        action: 'submit',
        resource_type: 'instance',
        actor_user_id: '12',
      },
    })
  })

  it('执行人列应改成点击查看详情，而不是直接常驻显示用户 ID', async () => {
    const wrapper = mount(AuditLog, {
      attachTo: document.body,
    })

    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('ID 12')
    expect(wrapper.text()).not.toContain('查看详情')

    await wrapper.get('.audit-row__actor-link').trigger('click')
    await flushPromises()

    expect(document.body.textContent).toContain('执行人详情')
    expect(document.body.textContent).toContain('用户 ID')
    expect(document.body.textContent).toContain('12')
    expect(document.body.textContent).toContain('challenge #5')
  })

  it('page model 应保留 query owner，但不再直接 import vue-router', () => {
    expect(auditLogPageSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(auditLogPageSource).not.toContain("from 'vue-router'")
    expect(auditLogPageSource).not.toContain('useRoute(')
    expect(auditLogPageSource).not.toContain('useRouter(')
    expect(auditLogPageSource).toContain('const { query, replaceQuery } = useRouteQueryTransport()')
    expect(auditLogPageSource).toContain('await replaceQuery(nextQuery)')
    expect(routeQueryTransportSource).toContain('const router = useRouter()')
  })

  it('路由页应继续作为薄入口，详情抽屉由独立 surface modal owner 承接', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('AuditLogHeroPanel')
    expect(auditLogSource).toContain('AuditLogDirectoryPanel')
    expect(auditLogSource).toContain('AuditActorDetailModal')
    expect(auditLogSource).toContain('<AuditActorDetailModal')
    expect(auditActorDetailModalSource).toContain('<AdminSurfaceModal')
    expect(auditLogDirectoryPanelSource).toContain("from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'")
    expect(auditLogDirectoryPanelSource).toContain("from '@/shared/ui/common/WorkspaceDataTable.vue'")
    expect(auditLogDirectoryPanelSource).toContain("from '@/shared/ui/common/PagePaginationControls.vue'")
  })

  it('审计摘要应继续由 hero panel 单独承接，而不是回退到页面内联壳层', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('AuditLogHeroPanel')
    expect(auditLogSource).toContain('<AuditLogHeroPanel')
    expect(auditLogSource).not.toContain('mt-10 space-y-10')
    expect(auditLogHeroPanelSource).toContain('Audit Log')
    expect(auditLogHeroPanelSource).toContain('本页已加载的日志条数')
  })
})
