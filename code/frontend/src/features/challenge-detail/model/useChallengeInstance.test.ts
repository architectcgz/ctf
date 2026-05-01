import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, ref } from 'vue'

import { useChallengeInstance } from '@/features/challenge-detail'
import { useInstanceListPage } from '@/features/instance-list'
import { useTeacherInstances } from '@/features/teacher-instances'
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

const confirmMock = vi.hoisted(() => vi.fn())

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

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
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
    confirmMock.mockReset()
    confirmMock.mockResolvedValue(true)
    instanceApiMocks.getMyInstances.mockResolvedValue([])
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

  it('实例列表销毁前应先经过统一危险确认，取消后不应继续请求', async () => {
    confirmMock.mockResolvedValue(false)

    let composable!: ReturnType<typeof useInstanceListPage>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceListPage()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.instances.value = [
      {
        id: 'inst-1',
        challenge_id: 'chal-1',
        challenge_title: 'SQL 注入基础',
        category: 'web',
        difficulty: 'easy',
        status: 'running',
        access_url: 'http://example.test',
        flag_type: 'static',
        share_scope: 'per_user',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-05T00:00:00Z',
        remaining: 1200,
      },
    ]

    await composable.destroyInstance('inst-1')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceApiMocks.destroyInstance).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('实例列表中的 AWD 队伍实例不应触发延时或销毁请求', async () => {
    let composable!: ReturnType<typeof useInstanceListPage>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceListPage()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.instances.value = [
      {
        id: 'awd-inst-1',
        challenge_id: 'awd-service-1',
        challenge_title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        status: 'running',
        access_url: '',
        flag_type: 'dynamic',
        share_scope: 'per_team',
        contest_mode: 'awd',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-05T00:00:00Z',
        remaining: 1200,
      },
    ]

    await composable.extendTime('awd-inst-1')
    await composable.destroyInstance('awd-inst-1')
    await flushPromises()

    expect(instanceApiMocks.extendInstance).not.toHaveBeenCalled()
    expect(instanceApiMocks.destroyInstance).not.toHaveBeenCalled()
    expect(confirmMock).not.toHaveBeenCalled()
    expect(toastMocks.error).toHaveBeenCalledWith('AWD 队伍实例不支持在此处延时或销毁')
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

  it('题目详情销毁前应先经过统一危险确认，取消后不应继续请求', async () => {
    confirmMock.mockResolvedValue(false)

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

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceApiMocks.destroyInstance).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('题目详情打开 TCP 实例时不应尝试用浏览器打开 tcp URL', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null)
    instanceApiMocks.requestInstanceAccess.mockResolvedValue({
      access_url: 'tcp://127.0.0.1:30001',
      access: {
        protocol: 'tcp',
        host: '127.0.0.1',
        port: 30001,
        command: 'nc 127.0.0.1 30001',
      },
    })

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
      id: 'inst-tcp',
      challenge_id: 'challenge-1',
      status: 'running',
      share_scope: 'per_user',
      access_url: 'tcp://127.0.0.1:30001',
      access: {
        protocol: 'tcp',
        host: '127.0.0.1',
        port: 30001,
        command: 'nc 127.0.0.1 30001',
      },
      flag_type: 'dynamic',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 1,
      created_at: '2026-04-09T00:00:00.000Z',
    }

    await composable.open()
    await flushPromises()

    expect(openSpy).not.toHaveBeenCalled()
    expect(toastMocks.info).toHaveBeenCalledWith('TCP 连接命令已复制')
    wrapper.unmount()
    openSpy.mockRestore()
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
