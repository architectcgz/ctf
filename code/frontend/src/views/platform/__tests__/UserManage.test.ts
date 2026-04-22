import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserManage from '../UserManage.vue'
import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'

const adminApiMocks = vi.hoisted(() => ({
  getUsers: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  importUsers: vi.fn(),
}))
const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string>,
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getUsers: adminApiMocks.getUsers,
    createUser: adminApiMocks.createUser,
    updateUser: adminApiMocks.updateUser,
    deleteUser: adminApiMocks.deleteUser,
    importUsers: adminApiMocks.importUsers,
  }
})
vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

describe('UserManage', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应该渲染真实用户列表', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('用户治理台')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('alice@example.com')
    expect(wrapper.text()).toContain('teacher')
    expect(adminApiMocks.getUsers).toHaveBeenCalledWith(
      {
        page: 1,
        page_size: 20,
        keyword: undefined,
        role: undefined,
        status: undefined,
        student_no: undefined,
        teacher_no: undefined,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
  })

  it('应该将用户总览与目录合并为一个工作台，并保留导入用户独立面板', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.find('#user-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#user-panel-import').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('用户治理台')
    expect(wrapper.text()).toContain('全部用户')
    expect(wrapper.text()).toContain('创建用户')
    expect(wrapper.text()).toContain('导入用户')
    expect(wrapper.find('#user-panel-overview').text()).toContain('用户总量')
    expect(wrapper.find('#user-panel-overview').text()).toContain('导入回执')

    await wrapper.get('#user-open-import').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      name: 'UserManage',
      query: { panel: 'import' },
    })
  })

  it('应将旧的 directory query 兼容到默认工作台视图', async () => {
    routeState.query = { panel: 'directory' }
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          student_no: 'S001',
          teacher_no: 'T001',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#user-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#user-panel-import').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('.user-list').exists()).toBe(true)
    expect(wrapper.text()).toContain('用户治理台')
    expect(wrapper.find('#user-panel-overview').text()).not.toContain('学生学号')
    expect(wrapper.find('#user-panel-overview').text()).not.toContain('教师工号')
    expect(wrapper.find('#user-panel-overview').text()).toContain('全部用户')
  })

  it('应该使用统一容器渲染用户分段列表', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          student_no: 'S001',
          teacher_no: 'T001',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
        {
          id: '2',
          username: 'bob',
          email: 'bob@example.com',
          class_name: 'Class B',
          student_no: 'S002',
          teacher_no: '',
          status: 'locked',
          roles: ['student'],
          created_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(userGovernanceSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(userGovernanceSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(userGovernanceSource).toContain('<WorkspaceDirectoryToolbar')
    expect(userGovernanceSource).toContain('<WorkspaceDataTable')
    expect(userGovernanceSource).not.toContain('<table class="user-table min-w-full text-sm">')
    expect(wrapper.find('.user-list').exists()).toBe(true)
    expect(wrapper.find('.workspace-directory-toolbar').exists()).toBe(true)
    expect(wrapper.findAll('.workspace-data-table__body tr')).toHaveLength(2)
    expect(wrapper.find('.user-table-accent').exists()).toBe(false)
    const headers = wrapper.findAll('.workspace-data-table__head-cell').map((item) => item.text())
    expect(headers).toEqual([
      '用户',
      '姓名',
      '邮箱',
      '角色',
      '状态',
      '班级',
      '学号 / 工号',
      '创建时间',
      '操作',
    ])
    expect(wrapper.find('.admin-pagination').exists()).toBe(true)
    const rows = wrapper.findAll('.workspace-data-table__body tr')
    expect(rows[0]?.text()).toContain('alice')
    expect(rows[0]?.text()).toContain('alice@example.com')
    expect(rows[0]?.text()).toContain('T001')
    expect(rows[0]?.text()).not.toContain('工号：')
    expect(rows[0]?.text()).not.toContain('学号：')
    expect(rows[1]?.text()).toContain('bob')
    expect(rows[1]?.text()).toContain('bob@example.com')
    expect(rows[1]?.text()).toContain('S002')
    expect(rows[1]?.text()).not.toContain('学号：')
    expect(rows[1]?.text()).not.toContain('工号：')
    expect(wrapper.findAll('.user-action-btn')).toHaveLength(4)
    expect(wrapper.find('.user-list .admin-inline-chip').exists()).toBe(false)
  })

  it('文本筛选应在节流后再请求用户列表', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    adminApiMocks.getUsers.mockClear()

    const inputs = wrapper.findAll('.workspace-directory-toolbar__search-input')
    expect(inputs).toHaveLength(1)
    await inputs[0].setValue('alice')

    expect(adminApiMocks.getUsers).not.toHaveBeenCalled()

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(adminApiMocks.getUsers).toHaveBeenCalledTimes(1)
    expect(adminApiMocks.getUsers).toHaveBeenLastCalledWith(
      {
        page: 1,
        page_size: 20,
        keyword: 'alice',
        student_no: undefined,
        teacher_no: undefined,
        role: undefined,
        status: undefined,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
  })

  it('用户目录筛选与列表应切到共享目录原语', () => {
    expect(userGovernanceSource).toContain('workspace-directory-section')
    expect(userGovernanceSource).toContain(
      'class="user-table-shell workspace-directory-list user-list"'
    )
    expect(userGovernanceSource).toContain(
      'search-placeholder="用户名 / 邮箱 / 班级 / 学号 / 工号"'
    )
    expect(userGovernanceSource).toContain('filter-panel-title="用户筛选"')
    expect(userGovernanceSource).toContain('total-suffix="个用户"')
    expect(userGovernanceSource).not.toContain('class="mt-5 grid gap-4"')
    expect(userGovernanceSource).not.toContain('<table class="user-table min-w-full text-sm">')
  })

  it('用户治理页应改用共享 ui-btn 原语而不是页面私有 admin-btn 按钮族', () => {
    expect(userGovernanceSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(userGovernanceSource).toContain('class="ui-btn ui-btn--primary"')
    expect(userGovernanceSource).toContain('class="ui-btn ui-btn--secondary user-action-btn"')
    expect(userGovernanceSource).toContain('class="ui-btn ui-btn--danger user-action-btn"')
    expect(userGovernanceSource).not.toContain('admin-btn admin-btn-ghost')
    expect(userGovernanceSource).not.toContain('admin-btn admin-btn-primary')
    expect(userGovernanceSource).not.toContain('admin-btn admin-btn-danger')
    expect(userGovernanceSource).not.toContain('admin-btn-compact')
  })

  it('用户治理页应改成 SaaS 全景工作台，并仅保留导入独立面板', () => {
    const overviewPanelStart = userGovernanceSource.indexOf('id="user-panel-overview"')
    const overviewPanelSnippet = userGovernanceSource.slice(
      overviewPanelStart,
      overviewPanelStart + 640
    )
    const importPanelStart = userGovernanceSource.indexOf('id="user-panel-import"')
    const importPanelSnippet = userGovernanceSource.slice(importPanelStart, importPanelStart + 420)

    expect(userGovernanceSource).toContain('id="user-panel-overview"')
    expect(userGovernanceSource).toContain('id="user-panel-import"')
    expect(userGovernanceSource).not.toContain('user-tab-overview')
    expect(userGovernanceSource).not.toContain('user-tab-directory')
    expect(userGovernanceSource).not.toContain('user-tab-import')
    expect(userGovernanceSource).not.toMatch(/role="tablist"/s)
    expect(userGovernanceSource).toContain('<main class="content-pane">')
    expect(overviewPanelStart).toBeGreaterThan(-1)
    expect(overviewPanelSnippet).toContain('<div class="workspace-overline">User Workspace</div>')
    expect(overviewPanelSnippet).toContain('<h1 class="workspace-page-title">用户治理台</h1>')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">全部用户</h2>')
    expect(userGovernanceSource).toContain('<WorkspaceDirectoryToolbar')
    expect(overviewPanelSnippet).not.toContain('<nav class="top-tabs"')
    expect(importPanelStart).toBeGreaterThan(-1)
    expect(userGovernanceSource).toContain('<div class="workspace-overline">User Import</div>')
    expect(userGovernanceSource).toContain('<h2 class="workspace-page-title">导入用户</h2>')
    expect(userGovernanceSource).toMatch(
      /\.user-directory-section :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
  })

  it('用户导入流应保留独立导入面板和回执区', () => {
    expect(userGovernanceSource).toContain('class="workspace-directory-section user-import-panel"')
    expect(userGovernanceSource).toContain('class="workspace-tab-heading user-import-head"')
    expect(userGovernanceSource).toContain('<h2 class="workspace-page-title">导入用户</h2>')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">导入回执</h2>')
    expect(userGovernanceSource).toContain('id="user-return-overview"')
  })

  it('用户工作台头部应暴露全局操作按钮，而不是顶层 tabs', () => {
    expect(userGovernanceSource).toContain('id="user-open-import"')
    expect(userGovernanceSource).toContain('id="user-open-create"')
    expect(userGovernanceSource).toContain('刷新列表')
    expect(userGovernanceSource).not.toContain('<nav class="top-tabs"')
  })

  it('用户工作台摘要应内嵌在 overview 区并呈现四个指标卡片', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
        {
          id: '2',
          username: 'bob',
          email: 'bob@example.com',
          class_name: 'Class B',
          status: 'inactive',
          roles: ['student'],
          created_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    const summary = wrapper.get('#user-panel-overview')
    const summaryCards = summary.findAll('.user-overview-stat')

    expect(summaryCards).toHaveLength(4)
    expect(summary.find('.user-overview-grid').exists()).toBe(true)
    expect(summary.findAll('.user-overview-stat.progress-card.metric-panel-card')).toHaveLength(4)
    expect(summary.findAll('.progress-card-label.metric-panel-label')).toHaveLength(4)
    expect(summary.findAll('.progress-card-value.metric-panel-value')).toHaveLength(4)
    expect(summary.findAll('.progress-card-hint.metric-panel-helper')).toHaveLength(4)
    expect(summaryCards.map((item) => item.find('.journal-note-label').text())).toEqual([
      '用户总量',
      '活跃账号',
      '教师角色',
      '导入回执',
    ])
  })

  it('删除用户失败时不应抛到全局错误页', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: '1',
          username: 'alice',
          email: 'alice@example.com',
          class_name: 'Class A',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.deleteUser.mockRejectedValue(new Error('删除失败'))
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    const wrapper = mount(UserManage, {
      global: {
        stubs: {
          UserGovernancePage: {
            props: ['list'],
            template:
              '<button id="delete-user" type="button" @click="$emit(\'deleteUser\', list[0].id)">删除用户</button>',
          },
          PlatformUserFormDialog: true,
        },
      },
    })

    await flushPromises()

    await expect(wrapper.get('#delete-user').trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(adminApiMocks.deleteUser).toHaveBeenCalledWith('1')
  })
})
