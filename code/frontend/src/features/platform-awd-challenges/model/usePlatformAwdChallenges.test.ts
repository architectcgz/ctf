import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import { ApiError } from '@/api/request'
import { usePlatformAwdChallenges } from '@/features/platform-awd-challenges'

const adminApiMocks = vi.hoisted(() => ({
  createAdminAwdChallenge: vi.fn(),
  deleteAdminAwdChallenge: vi.fn(),
  listAdminAwdChallenges: vi.fn(),
  updateAdminAwdChallenge: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/awd-authoring', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('usePlatformAwdChallenges', () => {
  beforeEach(() => {
    adminApiMocks.createAdminAwdChallenge.mockReset()
    adminApiMocks.deleteAdminAwdChallenge.mockReset()
    adminApiMocks.listAdminAwdChallenges.mockReset()
    adminApiMocks.updateAdminAwdChallenge.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    confirmMock.mockReset()

    adminApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          name: 'Bank Portal AWD',
          slug: 'bank-portal-awd',
          category: 'web',
          difficulty: 'hard',
          description: 'desc',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'draft',
          readiness_status: 'pending',
          created_at: '2026-04-17T08:00:00.000Z',
          updated_at: '2026-04-17T09:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
  })

  it('loads list data and creates templates', async () => {
    let composable!: ReturnType<typeof usePlatformAwdChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformAwdChallenges()
        return () => null
      },
    })

    mount(Harness)
    await composable.refresh()
    await flushPromises()

    expect(adminApiMocks.listAdminAwdChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      keyword: undefined,
      service_type: undefined,
      status: undefined,
    })
    expect(composable.list.value[0]?.slug).toBe('bank-portal-awd')

    adminApiMocks.createAdminAwdChallenge.mockResolvedValue({
      id: '2',
      name: 'Binary Gate AWD',
      slug: 'binary-gate-awd',
      category: 'pwn',
      difficulty: 'medium',
      description: '',
      service_type: 'binary_tcp',
      deployment_mode: 'single_container',
      version: 'v1',
      status: 'draft',
      readiness_status: 'pending',
      created_at: '2026-04-17T10:00:00.000Z',
      updated_at: '2026-04-17T10:00:00.000Z',
    })

    await composable.saveChallenge({
      name: 'Binary Gate AWD',
      slug: 'binary-gate-awd',
      category: 'pwn',
      difficulty: 'medium',
      description: '',
      service_type: 'binary_tcp',
      deployment_mode: 'single_container',
      status: 'draft',
    })
    await flushPromises()

    expect(adminApiMocks.createAdminAwdChallenge).toHaveBeenCalledWith({
      name: 'Binary Gate AWD',
      slug: 'binary-gate-awd',
      category: 'pwn',
      difficulty: 'medium',
      description: undefined,
      service_type: 'binary_tcp',
      deployment_mode: 'single_container',
    })
    expect(toastMocks.success).toHaveBeenCalledWith('AWD 题目已创建')
  })

  it('preserves delete failure messages', async () => {
    let composable!: ReturnType<typeof usePlatformAwdChallenges>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformAwdChallenges()
        return () => null
      },
    })

    mount(Harness)
    await composable.refresh()
    await flushPromises()

    confirmMock.mockResolvedValue(true)
    adminApiMocks.deleteAdminAwdChallenge.mockRejectedValue(
      new ApiError('AWD 题目已被赛事配置引用，暂时不能删除', { status: 409 })
    )

    await composable.removeChallenge(composable.list.value[0]!)
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('AWD 题目已被赛事配置引用，暂时不能删除')
  })
})
