import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import { useAdminChallenges } from '@/composables/useAdminChallenges'

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

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('element-plus', () => ({
  ElMessageBox: {
    confirm: confirmMock,
  },
}))

describe('useAdminChallenges', () => {
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

    let composable!: ReturnType<typeof useAdminChallenges>
    const Harness = defineComponent({
      setup() {
        composable = useAdminChallenges()
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

    let composable!: ReturnType<typeof useAdminChallenges>
    const Harness = defineComponent({
      setup() {
        composable = useAdminChallenges()
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
})
