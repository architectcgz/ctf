import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { AdminNotificationPublishDrawer } from '@/features/admin-notification-publisher'
import adminNotificationPublishDrawerSource from '../AdminNotificationPublishDrawer.vue?raw'

const adminApiMocks = vi.hoisted(() => ({
  publishAdminNotification: vi.fn(),
  getUsers: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  dismiss: vi.fn(),
}))

vi.mock('@/api/admin/platform', () => ({
  publishAdminNotification: adminApiMocks.publishAdminNotification,
}))
vi.mock('@/api/admin/users', () => ({
  getUsers: adminApiMocks.getUsers,
}))
vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('AdminNotificationPublishDrawer', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    adminApiMocks.publishAdminNotification.mockReset()
    adminApiMocks.getUsers.mockReset()
    teacherApiMocks.getClasses.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    adminApiMocks.getUsers.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('卸载抽屉后不应继续触发延迟的用户搜索', async () => {
    const wrapper = mount(AdminNotificationPublishDrawer, {
      props: {
        open: true,
      },
      global: {
        stubs: {
          AdminSurfaceDrawer: {
            template: '<div><slot /></div>',
          },
        },
      },
    })

    await wrapper.find('input[name="audience-target"][value="user"]').setValue(true)
    await wrapper.find('input[placeholder="输入用户名/学号搜索"]').setValue('alice')
    await flushPromises()

    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(250)

    expect(adminApiMocks.getUsers).not.toHaveBeenCalled()
  })

  it('指定用户候选项应复用 user entity 的稳定展示标签', async () => {
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        { id: 'u-1', username: 'alice', name: 'Alice Zhang' },
        { id: 'u-2', username: 'bob' },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(AdminNotificationPublishDrawer, {
      props: {
        open: true,
      },
      global: {
        stubs: {
          AdminSurfaceDrawer: {
            template: '<div><slot /></div>',
          },
        },
      },
    })

    await wrapper.find('input[name="audience-target"][value="user"]').setValue(true)
    await wrapper.find('input[placeholder="输入用户名/学号搜索"]').setValue('a')
    await vi.advanceTimersByTimeAsync(250)
    await flushPromises()

    const labels = wrapper.findAll('.publish-check-item span').map((node) => node.text())
    expect(labels).toContain('Alice Zhang (alice)')
    expect(labels).toContain('bob')
    expect(adminNotificationPublishDrawerSource).toContain("from '@/entities/user'")
    expect(adminNotificationPublishDrawerSource).toContain('getUserDisplayLabel')
    expect(adminNotificationPublishDrawerSource).not.toContain('user.name || user.username')
  })
})
