import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAdminNotificationPublisher } from '@/composables/useAdminNotificationPublisher'

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

describe('useAdminNotificationPublisher', () => {
  beforeEach(() => {
    adminApiMocks.publishAdminNotification.mockReset()
    adminApiMocks.getUsers.mockReset()
    teacherApiMocks.getClasses.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()

    adminApiMocks.publishAdminNotification.mockResolvedValue({
      batch_id: 'batch-1',
      recipient_count: 42,
    })
    adminApiMocks.getUsers.mockResolvedValue({
      list: [
        {
          id: 'u-1',
          username: 'alice',
          name: 'Alice',
          status: 'active',
          roles: ['student'],
          created_at: '2026-03-31T08:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A' }, { name: 'Class B' }])
  })

  it('assembles audience_rules payload for role/class/user/all modes', async () => {
    const publisher = useAdminNotificationPublisher()
    publisher.form.type = 'system'
    publisher.form.title = '停机维护'
    publisher.form.content = '今晚 23:00 - 23:30 维护。'

    publisher.audienceTarget.value = 'role'
    publisher.selectedRoles.value = ['teacher', 'admin']
    expect(publisher.buildPayload().audience_rules).toEqual({
      mode: 'union',
      rules: [{ type: 'role', values: ['teacher', 'admin'] }],
    })

    publisher.audienceTarget.value = 'class'
    publisher.selectedClasses.value = ['Class A']
    expect(publisher.buildPayload().audience_rules).toEqual({
      mode: 'union',
      rules: [{ type: 'class', values: ['Class A'] }],
    })

    publisher.audienceTarget.value = 'user'
    publisher.selectedUserIds.value = ['u-1', 'u-2']
    expect(publisher.buildPayload().audience_rules).toEqual({
      mode: 'union',
      rules: [{ type: 'user', values: ['u-1', 'u-2'] }],
    })

    publisher.audienceTarget.value = 'all'
    expect(publisher.buildPayload().audience_rules).toEqual({
      mode: 'union',
      rules: [{ type: 'all' }],
    })

    await publisher.searchUsers('alice')
    await publisher.loadClasses()

    expect(adminApiMocks.getUsers).toHaveBeenCalledWith({ page: 1, page_size: 20, keyword: 'alice' })
    expect(teacherApiMocks.getClasses).toHaveBeenCalledTimes(1)
  })

  it('submits publish payload successfully and returns publish receipt', async () => {
    const publisher = useAdminNotificationPublisher()
    publisher.form.type = 'contest'
    publisher.form.title = '春季赛提醒'
    publisher.form.content = '报名将在今晚截止。'
    publisher.form.link = 'https://ctf.example.test/contests/1'
    publisher.audienceTarget.value = 'role'
    publisher.selectedRoles.value = ['student']

    const result = await publisher.submit()

    expect(adminApiMocks.publishAdminNotification).toHaveBeenCalledWith({
      type: 'contest',
      title: '春季赛提醒',
      content: '报名将在今晚截止。',
      link: 'https://ctf.example.test/contests/1',
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'role', values: ['student'] }],
      },
    })
    expect(result).toEqual({ batch_id: 'batch-1', recipient_count: 42 })
    expect(toastMocks.success).toHaveBeenCalled()
  })
})
