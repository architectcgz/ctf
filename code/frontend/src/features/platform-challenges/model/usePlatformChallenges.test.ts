import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import { usePlatformChallenges } from '@/features/platform-challenges'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  createChallengePublishRequest: vi.fn(),
  deleteChallenge: vi.fn(),
  getChallenges: vi.fn(),
  getLatestChallengePublishRequest: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/authoring', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('element-plus', () => ({
  ElMessageBox: {
    confirm: confirmMock,
  },
}))

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('usePlatformChallenges', () => {
  beforeEach(() => {
    vi.useRealTimers()
    adminApiMocks.createChallengePublishRequest.mockReset()
    adminApiMocks.deleteChallenge.mockReset()
    adminApiMocks.getChallenges.mockReset()
    adminApiMocks.getLatestChallengePublishRequest.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    confirmMock.mockReset()

    adminApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Test Challenge',
          category: 'web',
          difficulty: 'easy',
          status: 'draft',
          points: 100,
          created_at: '2026-04-01T08:00:00.000Z',
          updated_at: '2026-04-01T08:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getLatestChallengePublishRequest.mockResolvedValue(null)
    adminApiMocks.createChallengePublishRequest.mockResolvedValue({
      id: 'req-1',
      challenge_id: '1',
      status: 'running',
      active: true,
      failure_summary: '',
      created_at: '2026-04-01T08:05:00.000Z',
      updated_at: '2026-04-01T08:05:00.000Z',
    })
  })

  it('submits publish checks and polls latest requests only while an active job exists', async () => {
    vi.useFakeTimers()

    adminApiMocks.getLatestChallengePublishRequest
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce({
        id: 'req-1',
        challenge_id: '1',
        status: 'running',
        active: true,
        failure_summary: '',
        created_at: '2026-04-01T08:05:00.000Z',
        updated_at: '2026-04-01T08:05:05.000Z',
      })
      .mockResolvedValueOnce({
        id: 'req-1',
        challenge_id: '1',
        status: 'failed',
        active: false,
        failure_summary: 'Flag 未配置',
        created_at: '2026-04-01T08:05:00.000Z',
        updated_at: '2026-04-01T08:05:30.000Z',
      })

    let composable!: ReturnType<typeof usePlatformChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformChallenges()
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getChallenges).toHaveBeenCalledTimes(1)
    expect(adminApiMocks.getLatestChallengePublishRequest).toHaveBeenCalledTimes(1)
    expect(composable.list.value[0]?.latestPublishRequest).toBeNull()

    await composable.publish(composable.list.value[0]!)
    await flushPromises()

    expect(adminApiMocks.createChallengePublishRequest).toHaveBeenCalledWith('1')
    expect(adminApiMocks.getLatestChallengePublishRequest).toHaveBeenCalledTimes(2)
    expect(composable.list.value[0]?.latestPublishRequest?.status).toBe('running')

    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(adminApiMocks.getLatestChallengePublishRequest).toHaveBeenCalledTimes(3)
    expect(composable.list.value[0]?.latestPublishRequest).toEqual({
      id: 'req-1',
      challenge_id: '1',
      status: 'failed',
      active: false,
      failure_summary: 'Flag 未配置',
      created_at: '2026-04-01T08:05:00.000Z',
      updated_at: '2026-04-01T08:05:30.000Z',
    })

    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(adminApiMocks.getLatestChallengePublishRequest).toHaveBeenCalledTimes(3)
  })

  it('refreshes challenge list after a publish check succeeds', async () => {
    vi.useFakeTimers()

    adminApiMocks.getChallenges
      .mockResolvedValueOnce({
        list: [
          {
            id: '1',
            title: 'Test Challenge',
            category: 'web',
            difficulty: 'easy',
            status: 'draft',
            points: 100,
            created_at: '2026-04-01T08:00:00.000Z',
            updated_at: '2026-04-01T08:00:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      })
      .mockResolvedValueOnce({
        list: [
          {
            id: '1',
            title: 'Test Challenge',
            category: 'web',
            difficulty: 'easy',
            status: 'published',
            points: 100,
            created_at: '2026-04-01T08:00:00.000Z',
            updated_at: '2026-04-01T08:05:30.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      })

    adminApiMocks.getLatestChallengePublishRequest
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce({
        id: 'req-1',
        challenge_id: '1',
        status: 'running',
        active: true,
        failure_summary: '',
        created_at: '2026-04-01T08:05:00.000Z',
        updated_at: '2026-04-01T08:05:05.000Z',
      })
      .mockResolvedValueOnce({
        id: 'req-1',
        challenge_id: '1',
        status: 'succeeded',
        active: false,
        failure_summary: '',
        created_at: '2026-04-01T08:05:00.000Z',
        updated_at: '2026-04-01T08:05:30.000Z',
      })

    let composable!: ReturnType<typeof usePlatformChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformChallenges()
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    expect(composable.list.value[0]?.status).toBe('draft')

    await composable.publish(composable.list.value[0]!)
    await flushPromises()

    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(adminApiMocks.getChallenges).toHaveBeenCalledTimes(2)
    expect(composable.list.value[0]?.latestPublishRequest?.status).toBe('succeeded')
    expect(composable.list.value[0]?.status).toBe('published')
  })

  it('preserves specific delete failure messages', async () => {
    confirmMock.mockResolvedValue(true)
    adminApiMocks.deleteChallenge.mockRejectedValue(
      new ApiError('还有学生正在解题，暂时不能删除', { code: 10007, status: 409 })
    )

    let composable!: ReturnType<typeof usePlatformChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformChallenges()
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    await composable.remove('1')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('还有学生正在解题，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除失败')
  })

  it('忽略翻页后才返回的旧发布状态结果', async () => {
    const stalePageOneRequest = deferred<null>()

    adminApiMocks.getChallenges
      .mockResolvedValueOnce({
        list: [
          {
            id: '1',
            title: 'Page One Challenge',
            category: 'web',
            difficulty: 'easy',
            status: 'draft',
            points: 100,
            created_at: '2026-04-01T08:00:00.000Z',
            updated_at: '2026-04-01T08:00:00.000Z',
          },
        ],
        total: 2,
        page: 1,
        page_size: 20,
      })
      .mockResolvedValueOnce({
        list: [
          {
            id: '2',
            title: 'Page Two Challenge',
            category: 'pwn',
            difficulty: 'medium',
            status: 'draft',
            points: 200,
            created_at: '2026-04-02T08:00:00.000Z',
            updated_at: '2026-04-02T08:00:00.000Z',
          },
        ],
        total: 2,
        page: 2,
        page_size: 20,
      })

    adminApiMocks.getLatestChallengePublishRequest
      .mockImplementationOnce(() => stalePageOneRequest.promise)
      .mockResolvedValueOnce({
        id: 'req-2',
        challenge_id: '2',
        status: 'succeeded',
        active: false,
        failure_summary: '',
        created_at: '2026-04-02T08:05:00.000Z',
        updated_at: '2026-04-02T08:05:10.000Z',
      })

    let composable!: ReturnType<typeof usePlatformChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformChallenges()
        return () => null
      },
    })

    mount(Harness)

    await composable.changePage(2)
    await flushPromises()

    expect(composable.list.value[0]?.id).toBe('2')
    expect(composable.list.value[0]?.latestPublishRequest?.id).toBe('req-2')

    stalePageOneRequest.resolve(null)
    await flushPromises()

    expect(composable.list.value[0]?.id).toBe('2')
    expect(composable.list.value[0]?.latestPublishRequest?.id).toBe('req-2')
  })
})
