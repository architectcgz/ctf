import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, ref } from 'vue'

import { useChallengeInstance } from '@/composables/useChallengeInstance'
import { useInstanceListPage } from '@/composables/useInstanceListPage'
import { useTeacherInstances } from '@/composables/useTeacherInstances'
import { ApiError } from '@/api/request'

const instanceApiMocks = vi.hoisted(() => ({
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  getMyInstances: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  destroyTeacherInstance: vi.fn(),
}))

const challengeApiMocks = vi.hoisted(() => ({
  createInstance: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  info: vi.fn(),
  success: vi.fn(),
}))

vi.mock('@/api/instance', () => ({
  destroyInstance: instanceApiMocks.destroyInstance,
  extendInstance: instanceApiMocks.extendInstance,
  getMyInstances: instanceApiMocks.getMyInstances,
  requestInstanceAccess: instanceApiMocks.requestInstanceAccess,
}))

vi.mock('@/api/teacher', () => ({
  destroyTeacherInstance: teacherApiMocks.destroyTeacherInstance,
}))

vi.mock('@/api/challenge', () => ({
  createInstance: challengeApiMocks.createInstance,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copy: vi.fn(),
  }),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('instance action errors', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    instanceApiMocks.destroyInstance.mockReset()
    instanceApiMocks.extendInstance.mockReset()
    instanceApiMocks.getMyInstances.mockReset()
    instanceApiMocks.requestInstanceAccess.mockReset()
    teacherApiMocks.destroyTeacherInstance.mockReset()
    challengeApiMocks.createInstance.mockReset()
    toastMocks.error.mockReset()
    toastMocks.info.mockReset()
    toastMocks.success.mockReset()
    instanceApiMocks.getMyInstances.mockResolvedValue([])
    vi.stubGlobal('confirm', vi.fn(() => true))
  })

  it('实例列表销毁失败时应优先展示接口返回消息', async () => {
    instanceApiMocks.destroyInstance.mockRejectedValue(
      new ApiError('实例正在回收中，请稍后再试', { status: 409 })
    )

    let composable!: ReturnType<typeof useInstanceListPage>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceListPage()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.destroyInstance('inst-1')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('实例正在回收中，请稍后再试')
    expect(toastMocks.error).not.toHaveBeenCalledWith('销毁失败，请稍后重试')
    wrapper.unmount()
  })

  it('题目详情销毁实例失败时应优先展示接口返回消息', async () => {
    instanceApiMocks.destroyInstance.mockRejectedValue(
      new ApiError('实例仍在创建中，暂时不能销毁', { status: 409 })
    )

    let composable!: ReturnType<typeof useChallengeInstance>
    const challengeId = ref('challenge-1')
    const Harness = defineComponent({
      setup() {
        composable = useChallengeInstance(challengeId)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.instance.value = {
      id: 'inst-1',
      challenge_id: 'challenge-1',
      status: 'running',
      share_scope: 'per_user',
      access_url: 'http://target.test',
      flag_type: 'dynamic',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 1,
      created_at: '2026-04-09T00:00:00.000Z',
    }

    await composable.destroy()
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('实例仍在创建中，暂时不能销毁')
    expect(toastMocks.error).not.toHaveBeenCalledWith('销毁实例失败')
    wrapper.unmount()
  })

  it('教师实例管理销毁失败时应优先展示接口返回消息', async () => {
    teacherApiMocks.destroyTeacherInstance.mockRejectedValue(
      new ApiError('实例所属练习仍在结算中，暂时不能销毁', { status: 409 })
    )

    let composable!: ReturnType<typeof useTeacherInstances>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherInstances()
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    await composable.removeInstance('inst-2')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('实例所属练习仍在结算中，暂时不能销毁')
    expect(toastMocks.error).not.toHaveBeenCalledWith('销毁实例失败，请稍后重试')
  })
})
