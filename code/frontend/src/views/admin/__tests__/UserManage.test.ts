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
})
