import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
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
      name: 'AuditLog',
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
    expect(wrapper.text()).toContain('查看详情')

    await wrapper.get('.audit-row__actor-trigger').trigger('click')
    await flushPromises()

    expect(document.body.textContent).toContain('执行人详情')
    expect(document.body.textContent).toContain('用户 ID')
    expect(document.body.textContent).toContain('12')
    expect(document.body.textContent).toContain('challenge #5')
  })

  it('筛选区应改成平铺目录筛选，不应继续保留显式应用按钮和说明壳', () => {
    expect(auditLogSource).toContain('class="admin-board workspace-directory-section"')
    expect(auditLogSource).not.toContain('class="admin-section-title">筛选条件</h2>')
    expect(auditLogSource).not.toContain('支持按动作、资源类型与执行人组合筛选。')
    expect(auditLogSource).not.toContain('应用筛选')
    expect(auditLogSource).not.toContain('激活筛选')
    expect(auditLogSource).not.toContain('当前生效的筛选项数量')
    expect(auditLogSource).toContain('placeholder="资源类型，如 challenge"')
    expect(auditLogSource).toContain('placeholder="执行人 ID"')
    expect(auditLogSource).toContain('重置筛选')
    expect(auditLogSource).toContain(':reset-disabled="!hasActiveFilters"')
    expect(auditLogSource).not.toContain('audit-filter-label--ghost')
    expect(auditLogSource).not.toContain('audit-filter-actions')
    expect(auditLogSource).not.toContain('audit-filter-action-row')
  })

  it('应接入共享目录工具栏与列表表格，而不是继续使用原生 table', () => {
    expect(auditLogSource).toContain("from '@/components/common/WorkspaceDirectoryToolbar.vue'")
    expect(auditLogSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(auditLogSource).toContain(
      "from '@/components/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(auditLogSource).toContain('<WorkspaceDirectoryToolbar')
    expect(auditLogSource).toContain('<WorkspaceDataTable')
    expect(auditLogSource).toContain('<AdminSurfaceModal')
    expect(auditLogSource).not.toContain('<section class="audit-filter-strip"')
    expect(auditLogSource).not.toContain('<table class="min-w-full text-sm">')
    expect(auditLogSource).toContain('search-placeholder="检索动作、资源类型、执行人..."')
    expect(auditLogSource).toContain('total-suffix="条日志"')
    expect(auditLogSource).toContain('class="audit-list workspace-directory-list"')
    expect(auditLogSource).toContain('class="audit-row__actor-trigger"')
    expect(auditLogSource).not.toContain('class="audit-row__actor-id"')
    expect(auditLogSource).toMatch(
      /\.admin-board\s*\{[\s\S]*display:\s*grid;[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(auditLogSource).toMatch(
      /\.admin-board :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
  })

  it('应使用统一进度卡片样式展示审计摘要', () => {
    expect(auditLogSource).toContain('<div class="workspace-overline">Audit Log</div>')
    expect(auditLogSource).not.toContain('<div class="journal-eyebrow">Audit Log</div>')
    expect(auditLogSource).toContain(
      'class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(auditLogSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(auditLogSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(auditLogSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(auditLogSource).toContain('本页已加载的日志条数')
  })
})
