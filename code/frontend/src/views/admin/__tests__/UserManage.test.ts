import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserManage from '../UserManage.vue'

const adminApiMocks = vi.hoisted(() => ({
  getUsers: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  importUsers: vi.fn(),
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

describe('UserManage', () => {
  beforeEach(() => {
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
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
    expect(headers).toEqual(['用户', '姓名', '邮箱', '角色', '状态', '班级', '学号 / 工号', '创建时间', '操作'])
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
})
