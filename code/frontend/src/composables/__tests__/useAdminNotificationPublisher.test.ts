import { createApp, type App } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

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

function withSetup<T>(composable: () => T): [T, App] {
  let result!: T

  const app = createApp({
    setup() {
      result = composable()
      return () => null
    },
  })

  app.mount(document.createElement('div'))
  return [result, app]
}

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('useAdminNotificationPublisher', () => {
  let app: App | null = null

  afterEach(() => {
    app?.unmount()
    app = null
  })

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

    expect(adminApiMocks.getUsers).toHaveBeenCalledWith(
      {
        page: 1,
        page_size: 20,
        keyword: 'alice',
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
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

  it('空关键词时不应请求用户搜索，并且过期结果不应覆盖最新结果', async () => {
    const slowRequest = deferred<{
      list: Array<{
        id: string
        username: string
        name?: string
        status: 'active'
        roles: ['student']
        created_at: string
      }>
      total: number
      page: number
      page_size: number
    }>()

    adminApiMocks.getUsers
      .mockReset()
      .mockImplementationOnce(() => slowRequest.promise)
      .mockResolvedValueOnce({
        list: [
          {
            id: 'u-2',
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

    const publisher = useAdminNotificationPublisher()
    const slowSearch = publisher.searchUsers('bob')
    await publisher.searchUsers('alice')

    expect(publisher.userOptions.value).toEqual([expect.objectContaining({ username: 'alice' })])

    slowRequest.resolve({
      list: [
        {
          id: 'u-1',
          username: 'bob',
          name: 'Bob',
          status: 'active',
          roles: ['student'],
          created_at: '2026-03-31T08:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    await slowSearch

    expect(publisher.userOptions.value).toEqual([expect.objectContaining({ username: 'alice' })])

    adminApiMocks.getUsers.mockClear()
    await publisher.searchUsers('   ')
    expect(adminApiMocks.getUsers).not.toHaveBeenCalled()
    expect(publisher.userOptions.value).toEqual([])
  })

  it('新的用户搜索应中止上一次进行中的请求', async () => {
    const firstRequestCanceled = deferred<void>()

    adminApiMocks.getUsers
      .mockReset()
      .mockImplementationOnce(
        (_params: unknown, options?: { signal?: AbortSignal }) =>
          new Promise((_resolve, reject) => {
            options?.signal?.addEventListener(
              'abort',
              () => {
                firstRequestCanceled.resolve()
                reject(Object.assign(new Error('canceled'), { code: 'ERR_CANCELED' }))
              },
              { once: true }
            )
          })
      )
      .mockResolvedValueOnce({
        list: [
          {
            id: 'u-2',
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

    const publisher = useAdminNotificationPublisher()
    const firstSearch = publisher.searchUsers('bob')

    await Promise.resolve()
    const firstSignal = adminApiMocks.getUsers.mock.calls[0]?.[1]?.signal as AbortSignal | undefined

    await publisher.searchUsers('alice')
    await firstRequestCanceled.promise

    expect(firstSignal?.aborted).toBe(true)
    await expect(firstSearch).resolves.toBeUndefined()
    expect(publisher.userOptions.value).toEqual([expect.objectContaining({ username: 'alice' })])
  })

  it('清空搜索词时应中止进行中的请求并清空结果', async () => {
    const requestCanceled = deferred<void>()

    adminApiMocks.getUsers.mockReset().mockImplementationOnce(
      (_params: unknown, options?: { signal?: AbortSignal }) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener(
            'abort',
            () => {
              requestCanceled.resolve()
              reject(Object.assign(new Error('canceled'), { code: 'ERR_CANCELED' }))
            },
            { once: true }
          )
        })
    )

    const publisher = useAdminNotificationPublisher()
    publisher.userOptions.value = [
      {
        id: 'u-stale',
        username: 'stale',
        name: 'Stale',
        status: 'active',
        roles: ['student'],
        created_at: '2026-03-31T08:00:00Z',
      },
    ]

    const searchTask = publisher.searchUsers('stale')

    await Promise.resolve()
    const firstSignal = adminApiMocks.getUsers.mock.calls[0]?.[1]?.signal as AbortSignal | undefined

    await publisher.searchUsers('   ')
    await requestCanceled.promise

    expect(firstSignal?.aborted).toBe(true)
    await expect(searchTask).resolves.toBeUndefined()
    expect(publisher.userOptions.value).toEqual([])
    expect(publisher.loadingUsers.value).toBe(false)
  })

  it('组件卸载时应中止进行中的用户搜索请求', async () => {
    const requestCanceled = deferred<void>()

    adminApiMocks.getUsers.mockReset().mockImplementationOnce(
      (_params: unknown, options?: { signal?: AbortSignal }) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener(
            'abort',
            () => {
              requestCanceled.resolve()
              reject(Object.assign(new Error('canceled'), { code: 'ERR_CANCELED' }))
            },
            { once: true }
          )
        })
    )

    const [publisher, testApp] = withSetup(() => useAdminNotificationPublisher())
    app = testApp

    const searchTask = publisher.searchUsers('alice')
    await Promise.resolve()

    const signal = adminApiMocks.getUsers.mock.calls[0]?.[1]?.signal as AbortSignal | undefined
    app.unmount()
    app = null

    await requestCanceled.promise

    expect(signal?.aborted).toBe(true)
    await expect(searchTask).resolves.toBeUndefined()
  })
})
