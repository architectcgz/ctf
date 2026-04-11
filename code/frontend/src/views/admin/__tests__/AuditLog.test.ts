import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AuditLog from '../AuditLog.vue'
import auditLogSource from '../AuditLog.vue?raw'

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

vi.mock('@/api/admin', () => adminApiMocks)

describe('AuditLog', () => {
  beforeEach(() => {
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

  it('应该根据路由 query 加载预置筛选结果', async () => {
    const wrapper = mount(AuditLog)

    await flushPromises()

    expect(adminApiMocks.getAuditLogs).toHaveBeenCalledWith({
      page: 2,
      page_size: 20,
      action: 'submit',
      resource_type: 'challenge',
      actor_user_id: 12,
    })
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('submit')
  })

  it('应该在应用筛选时同步 query', async () => {
    const wrapper = mount(AuditLog)

    await flushPromises()

    const resourceInput = wrapper.find('input[placeholder="资源类型，如 challenge"]')
    await resourceInput.setValue('instance')
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      name: 'AuditLog',
      query: {
        action: 'submit',
        resource_type: 'instance',
        actor_user_id: '12',
      },
    })
  })

  it('应使用统一进度卡片样式展示审计摘要', () => {
    expect(auditLogSource).toContain(
      'class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(auditLogSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(auditLogSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label">激活筛选</div>'
    )
    expect(auditLogSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(auditLogSource).toContain('{{ activeFilterCount }}')
    expect(auditLogSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(auditLogSource).toContain('当前生效的筛选项数量')
  })
})
