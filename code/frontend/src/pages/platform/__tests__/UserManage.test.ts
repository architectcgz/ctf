import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserManage from '@/pages/platform/UserManageRoutePage.vue'
import userManageSource from '@/pages/platform/UserManageRoutePage.vue?raw'
import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import platformUserManagePageSource from '@/features/platform/user-management/model/usePlatformUserManagePage.ts?raw'
import userGovernancePanelRouteSource from '@/features/platform/user-management/model/useUserGovernancePanelRoute.ts?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'

const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')

const adminApiMocks = vi.hoisted(() => ({
  getUsers: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  importUsers: vi.fn(),
  getUserSessions: vi.fn(),
  revokeUserSession: vi.fn(),
  revokeAllUserSessions: vi.fn(),
}))
const pushMock = vi.fn()
const replaceMock = vi.fn()
const destructiveConfirmMock = vi.hoisted(() => vi.fn())
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string>,
}))

vi.mock('@/api/admin/users', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin/users')>('@/api/admin/users')
  return {
    ...actual,
    getUsers: adminApiMocks.getUsers,
    createUser: adminApiMocks.createUser,
    updateUser: adminApiMocks.updateUser,
    deleteUser: adminApiMocks.deleteUser,
    importUsers: adminApiMocks.importUsers,
    getUserSessions: adminApiMocks.getUserSessions,
    revokeUserSession: adminApiMocks.revokeUserSession,
    revokeAllUserSessions: adminApiMocks.revokeAllUserSessions,
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
vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

describe('UserManage', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    adminApiMocks.getUserSessions.mockResolvedValue([])
    destructiveConfirmMock.mockReset()
    destructiveConfirmMock.mockResolvedValue(true)
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
  })

  afterEach(() => {
    vi.useRealTimers()
    document.body.innerHTML = ''
    document.body.style.overflow = ''
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
    adminApiMocks.getUserSessions.mockResolvedValue([
      {
        id: 'session-abcd1234',
        username: 'alice',
        role: 'teacher' as const,
        expires_at: '2026-04-01T12:00:00.000Z',
      },
    ])

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
    expect(wrapper.text()).not.toContain('alice@example.com')
    expect(wrapper.text()).toContain('teacher')

    await wrapper.get('#user-row-detail-1').trigger('click')
    await flushPromises()

    const detailDrawer = document.body.querySelector<HTMLElement>('.user-detail-drawer')
    expect(detailDrawer).not.toBeNull()
    expect(detailDrawer?.textContent).toContain('alice@example.com')
    expect(detailDrawer?.textContent).toContain('未设置姓名')
    expect(adminApiMocks.getUserSessions).toHaveBeenCalledWith('1', expect.any(Object))
    expect(detailDrawer?.textContent).toContain('活跃会话')
    expect(detailDrawer?.textContent).toContain('过期')
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

  it('切换详情用户时应重置撤销全部确认态，避免跨用户残留', async () => {
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
          status: 'locked',
          roles: ['student'],
          created_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getUserSessions.mockResolvedValue([
      {
        id: 'sess-1',
        username: 'alice',
        role: 'teacher',
        expires_at: '2026-04-01T12:00:00.000Z',
      },
    ])

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

    await wrapper.get('#user-row-detail-1').trigger('click')
    await flushPromises()

    const revokeAllButton = document.body.querySelector<HTMLButtonElement>(
      '.user-detail-section__action-btn--danger'
    )
    expect(revokeAllButton).not.toBeNull()
    revokeAllButton?.click()
    await flushPromises()

    expect(document.body.textContent).toContain('确认撤销全部会话')

    await wrapper.get('#user-row-detail-2').trigger('click')
    await flushPromises()

    const drawer = document.body.querySelector<HTMLElement>('.user-detail-drawer')
    expect(drawer?.textContent).toContain('bob@example.com')
    expect(document.body.textContent).not.toContain('确认撤销全部会话')
  })

  it('快速切换详情用户时应中止旧会话请求，并只保留最后一次结果', async () => {
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
          status: 'locked',
          roles: ['student'],
          created_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const pending = new Map<
      string,
      {
        resolve: (value: Array<{ id: string; username: string; role: 'teacher' | 'student'; expires_at: string }>) => void
        signal?: AbortSignal
      }
    >()
    adminApiMocks.getUserSessions.mockImplementation(
      (userId: string, options?: { signal?: AbortSignal }) =>
        new Promise((resolve, reject) => {
          options?.signal?.addEventListener('abort', () => {
            reject(new DOMException('Aborted', 'AbortError'))
          })
          pending.set(userId, { resolve, signal: options?.signal })
        })
    )

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

    await wrapper.get('#user-row-detail-1').trigger('click')
    await Promise.resolve()

    const firstSignal = pending.get('1')?.signal
    expect(firstSignal).toBeInstanceOf(AbortSignal)
    expect(firstSignal?.aborted).toBe(false)

    await wrapper.get('#user-row-detail-2').trigger('click')
    await Promise.resolve()

    expect(firstSignal?.aborted).toBe(true)
    expect(pending.get('2')?.signal).toBeInstanceOf(AbortSignal)

    pending.get('2')?.resolve([
      {
        id: 'sess-2',
        username: 'bob',
        role: 'student',
        expires_at: '2026-04-02T12:00:00.000Z',
      },
    ])
    await flushPromises()

    const drawer = document.body.querySelector<HTMLElement>('.user-detail-drawer')
    expect(drawer?.textContent).toContain('bob@example.com')
    expect(drawer?.textContent).toContain('sess-2')
    expect(drawer?.textContent).not.toContain('alice@example.com')
  })

  it('用户详情里的姓名展示应由 user entity 承接，而不是回退成用户名', () => {
    expect(userGovernanceDetailModalSource).toContain("from '@/entities/user'")
    expect(userGovernanceDetailModalSource).toContain('getUserName')
    expect(userGovernanceDetailModalSource).toContain("{{ getUserName(user, '未设置姓名') }}")
    expect(userGovernanceDetailModalSource).not.toContain('{{ user.name || user.username }}')
  })

  it('用户总览表格里的用户名和姓名展示应由 user entity 承接', () => {
    expect(userGovernanceOverviewPanelSource).toContain("from '@/entities/user'")
    expect(userGovernanceOverviewPanelSource).toContain('getUserDisplayName')
    expect(userGovernanceOverviewPanelSource).toContain('getUserUsernameHandle')
    expect(userGovernanceOverviewPanelSource).not.toContain('@{{ (row as AdminUserListItem).username }}')
    expect(userGovernanceOverviewPanelSource).not.toContain(
      '{{ (row as AdminUserListItem).name || (row as AdminUserListItem).username }}'
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

    expect(userManageSource).toContain("from '@/features/platform/user-management'")
    expect(userManageSource).toContain('UserGovernancePage')
    expect(userManageSource).toContain('usePlatformUserManagePage')
    expect(userManageSource).not.toContain("from '@/features/platform/user-management/ui/'")
    expect(userManageSource).not.toContain('onMounted(')
    expect(userManageSource).not.toContain('confirmDestructiveAction')
    expect(userManageSource).toContain(':active-panel="activePanel"')
    expect(userManageSource).toContain('@switch-panel="switchPanel"')
    expect(userGovernanceSource).not.toContain("from 'vue-router'")
    expect(userGovernanceSource).not.toContain('useRoute(')
    expect(userGovernanceSource).not.toContain('useRouter(')
    expect(userGovernancePanelRouteSource).not.toContain("from 'vue-router'")
    expect(userGovernancePanelRouteSource).not.toContain('useRoute(')
    expect(userGovernancePanelRouteSource).not.toContain('useRouter(')
    expect(userGovernancePanelRouteSource).toContain('resolveUserGovernancePanel')
    expect(userGovernancePanelRouteSource).toContain('buildUserGovernancePanelQuery')
    expect(platformUserManagePageSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(platformUserManagePageSource).not.toContain("from 'vue-router'")
    expect(platformUserManagePageSource).not.toContain('useRoute(')
    expect(platformUserManagePageSource).not.toContain('useRouter(')
    expect(platformUserManagePageSource).toContain(
      'const { query, replaceQuery } = useRouteQueryTransport()'
    )
    expect(platformUserManagePageSource).toContain('resolveUserGovernancePanel(query.value.panel)')
    expect(platformUserManagePageSource).toContain(
      'await replaceQuery(buildUserGovernancePanelQuery(query.value, panel))'
    )
    expect(userGovernanceSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(userGovernanceSource).toContain("from '@/shared/ui/common/WorkspaceDataTable.vue'")
    expect(wrapper.find('.user-list').exists()).toBe(true)
    expect(wrapper.find('.workspace-directory-toolbar').exists()).toBe(true)
    expect(wrapper.findAll('.workspace-data-table__body tr')).toHaveLength(2)
    const headers = wrapper.findAll('.workspace-data-table__head-cell').map((item) => item.text())
    expect(headers).toEqual(['用户', '姓名', '角色', '状态', '操作'])
    expect(wrapper.find('.admin-pagination').exists()).toBe(true)
    const rows = wrapper.findAll('.workspace-data-table__body tr')
    expect(rows[0]?.text()).toContain('alice')
    expect(rows[0]?.text()).not.toContain('alice@example.com')
    expect(rows[0]?.text()).not.toContain('T001')
    expect(rows[0]?.text()).not.toContain('工号：')
    expect(rows[0]?.text()).not.toContain('学号：')
    expect(rows[1]?.text()).toContain('bob')
    expect(rows[1]?.text()).not.toContain('bob@example.com')
    expect(rows[1]?.text()).not.toContain('S002')
    expect(rows[1]?.text()).not.toContain('学号：')
    expect(rows[1]?.text()).not.toContain('工号：')

    expect(document.body.querySelector('.user-detail-drawer')).toBeNull()
    await wrapper.get('#user-row-detail-1').trigger('click')
    await flushPromises()

    const drawer = document.body.querySelector<HTMLElement>('.user-detail-drawer')
    expect(drawer).not.toBeNull()
    expect(drawer?.textContent).toContain('alice')
    expect(drawer?.textContent).toContain('alice@example.com')
    expect(drawer?.textContent).toContain('Class A')
    expect(drawer?.textContent).toContain('T001')
    expect(drawer?.textContent).toContain('2026')

    const closeButton = document.body.querySelector<HTMLButtonElement>('#user-detail-close')
    expect(closeButton).not.toBeNull()
    closeButton?.click()
    await flushPromises()

    expect(document.body.querySelector('.user-detail-drawer')).toBeNull()
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
    const summaryCards = summary.findAll('.progress-card.metric-panel-card')

    expect(summaryCards).toHaveLength(4)
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

    expect(destructiveConfirmMock).toHaveBeenCalledWith({
      title: '删除用户',
      message: '确定删除用户 alice 吗？',
      confirmButtonText: '确认删除',
    })
    expect(adminApiMocks.deleteUser).toHaveBeenCalledWith('1')
  })
})
