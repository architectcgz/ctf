import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import AdminNotificationPublishDrawer from '../AdminNotificationPublishDrawer.vue'

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

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/composables/useToast', () => ({
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
})
