import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserManage from '../UserManage.vue'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'

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
    expect(adminApiMocks.getUsers).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      keyword: undefined,
      role: undefined,
      status: undefined,
    })
  })

  it('应该将用户治理拆成总览、用户列表、导入用户三个标签页', async () => {
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

    expect(wrapper.find('#user-tab-overview').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#user-overview-summary').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#user-directory-filters').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#user-import-start').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('总览')
    expect(wrapper.text()).toContain('用户列表')
    expect(wrapper.text()).toContain('导入用户')
    expect(wrapper.find('#user-overview-summary').text()).toContain('当前用户概况')
    expect(wrapper.find('#user-overview-summary').text()).toContain('用户总量')
    expect(wrapper.text()).toContain('导入回执')

    await wrapper.get('#user-tab-import').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      name: 'UserManage',
      query: { panel: 'import' },
    })
  })

  it('应支持通过 query 直接打开用户列表标签页', async () => {
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

    expect(wrapper.find('#user-tab-directory').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#user-overview-summary').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#user-directory-filters').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('.user-table-shell').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('用户治理台')
    expect(wrapper.find('#user-directory-filters').text()).not.toContain('学生学号')
    expect(wrapper.find('#user-directory-filters').text()).not.toContain('教师工号')
    expect(wrapper.find('#user-directory-filters').text()).toContain('用户目录')
    expect(wrapper.find('#user-directory-filters').text()).not.toContain('筛选条件')
    expect(wrapper.text()).toContain('用户列表')
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

    expect(wrapper.find('.user-table-shell').exists()).toBe(true)
    expect(wrapper.find('.user-table').exists()).toBe(true)
    expect(wrapper.findAll('.user-table tbody tr')).toHaveLength(2)
    expect(wrapper.find('.user-table-accent').exists()).toBe(false)
    const headers = wrapper.findAll('.user-table thead th').map((item) => item.text())
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
    const sectionTitles = wrapper.findAll('section h2').map((item) => item.text())
    expect(sectionTitles.indexOf('用户列表')).toBeLessThan(sectionTitles.indexOf('导入回执'))

    const rows = wrapper.findAll('.user-table tbody tr')
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
    expect(wrapper.find('.user-table .admin-inline-chip').exists()).toBe(false)
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

    const inputs = wrapper.findAll('input.admin-input')
    expect(inputs).toHaveLength(1)
    await inputs[0].setValue('alice')

    expect(adminApiMocks.getUsers).not.toHaveBeenCalled()

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(adminApiMocks.getUsers).toHaveBeenCalledTimes(1)
    expect(adminApiMocks.getUsers).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      keyword: 'alice',
      student_no: undefined,
      teacher_no: undefined,
      role: undefined,
      status: undefined,
    })
  })

  it('用户治理页应使用顶层 tabs 分隔列表和导入区域', () => {
    const userDirectoryPanelStart = userGovernanceSource.indexOf('id="user-directory-filters"')
    const userDirectoryPanelSnippet = userGovernanceSource.slice(
      userDirectoryPanelStart,
      userDirectoryPanelStart + 320
    )

    expect(userGovernanceSource).toContain('user-tab-overview')
    expect(userGovernanceSource).toContain('user-tab-directory')
    expect(userGovernanceSource).toContain('user-tab-import')
    expect(userGovernanceSource).toContain('user-overview-summary')
    expect(userGovernanceSource).toContain('user-directory-filters')
    expect(userGovernanceSource).toContain('user-import-start')
    expect(userGovernanceSource).toContain('<div class="workspace-overline">User Governance</div>')
    expect(userGovernanceSource).not.toContain('<div class="journal-eyebrow">User Governance</div>')
    expect(userGovernanceSource).toMatch(/role="tablist"/s)
    expect(userGovernanceSource.indexOf('User Governance')).toBeLessThan(
      userGovernanceSource.indexOf('role="tablist"')
    )
    expect(userGovernanceSource.indexOf('role="tablist"')).toBeLessThan(
      userGovernanceSource.indexOf('用户治理台')
    )
    expect(userGovernanceSource).not.toContain('user-overview-entry')
    expect(userGovernanceSource).not.toContain('user-panel-directory')
    expect(userGovernanceSource).not.toContain('user-panel-import')
    expect(userGovernanceSource).not.toContain('<main class="content-pane">')
    expect(userDirectoryPanelStart).toBeGreaterThan(-1)
    expect(userDirectoryPanelSnippet).toContain('<div class="list-heading user-directory-head">')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">用户目录</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">用户列表</h2>')
    expect(userDirectoryPanelSnippet).not.toContain('admin-section-head-intro')
    expect(userGovernanceSource).not.toMatch(/筛选与导入[\s\S]*用户列表[\s\S]*导入回执/s)
  })

  it('用户导入流分段页应使用统一目录头样式', () => {
    expect(userGovernanceSource).toContain('class="list-heading admin-section-head-intro"')
    expect(userGovernanceSource).toContain('class="list-heading user-import-receipt-head"')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">导入用户</h2>')
    expect(userGovernanceSource).toContain('<h2 class="list-heading__title">导入回执</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">导入用户</h2>')
    expect(userGovernanceSource).not.toContain('workspace-tab-heading__title">导入回执</h2>')
  })

  it('用户总览头部不应再渲染快捷操作按钮组', () => {
    expect(userGovernanceSource).not.toContain('class="mt-6 flex flex-wrap gap-3"')
    expect(userGovernanceSource).not.toMatch(
      /刷新列表[\s\S]*用户列表[\s\S]*导入用户[\s\S]*创建用户/s
    )
  })

  it('用户总览摘要应内嵌在 overview 区并呈现四个指标卡片', async () => {
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

    const summary = wrapper.get('#user-overview-summary')
    const summaryCards = summary.findAll('.user-overview-stat')

    expect(userGovernanceSource).not.toContain(
      '<article class="journal-brief user-overview-summary'
    )
    expect(summaryCards).toHaveLength(4)
    expect(summary.find('.user-overview-grid').exists()).toBe(true)
    expect(summary.findAll('.user-overview-stat.progress-card.metric-panel-card')).toHaveLength(4)
    expect(summary.findAll('.progress-card-label.metric-panel-label')).toHaveLength(4)
    expect(summary.findAll('.progress-card-value.metric-panel-value')).toHaveLength(4)
    expect(summary.findAll('.progress-card-hint.metric-panel-helper')).toHaveLength(4)
    expect(userGovernanceSource).not.toContain(
      '<div v-if="activePanel === \'overview\'" class="journal-divider mt-6" />'
    )
    expect(summaryCards.map((item) => item.find('.journal-note-label').text())).toEqual([
      '用户总量',
      '活跃账号',
      '教师角色',
      '导入回执',
    ])
  })
})
