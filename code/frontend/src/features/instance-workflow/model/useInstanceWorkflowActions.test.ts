import { describe, expect, it, beforeEach, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { ApiError } from '@/api/request'

import { useInstanceWorkflowActions } from './useInstanceWorkflowActions'

const instanceApiMocks = vi.hoisted(() => ({
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  info: vi.fn(),
  success: vi.fn(),
}))

const clipboardMocks = vi.hoisted(() => ({
  copy: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/instance', () => ({
  destroyInstance: instanceApiMocks.destroyInstance,
  extendInstance: instanceApiMocks.extendInstance,
  requestInstanceAccess: instanceApiMocks.requestInstanceAccess,
}))

vi.mock('@/shared/model/common/useClipboard', () => ({
  useClipboard: () => clipboardMocks,
}))

vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('useInstanceWorkflowActions', () => {
  beforeEach(() => {
    instanceApiMocks.destroyInstance.mockReset()
    instanceApiMocks.extendInstance.mockReset()
    instanceApiMocks.requestInstanceAccess.mockReset()
    toastMocks.error.mockReset()
    toastMocks.info.mockReset()
    toastMocks.success.mockReset()
    clipboardMocks.copy.mockReset()
    confirmMock.mockReset()
    confirmMock.mockResolvedValue(true)
  })

  it('打开 TCP 实例时应复制连接命令而不是调用浏览器', async () => {
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

    let composable!: ReturnType<typeof useInstanceWorkflowActions>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceWorkflowActions({
          resolveTarget: (id) => (id ? { id } : null),
          onExtended: () => undefined,
          onDestroyed: () => undefined,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.openInstance('inst-tcp')
    await flushPromises()

    expect(instanceApiMocks.requestInstanceAccess).toHaveBeenCalledWith('inst-tcp')
    expect(clipboardMocks.copy).toHaveBeenCalledWith('nc 127.0.0.1 30001')
    expect(toastMocks.info).toHaveBeenCalledWith('TCP 连接命令已复制')
    expect(openSpy).not.toHaveBeenCalled()

    wrapper.unmount()
    openSpy.mockRestore()
  })

  it('延时时应先尊重页面 owner 提供的阻止策略', async () => {
    const target = { id: 'shared-inst', share_scope: 'shared' as const }

    let composable!: ReturnType<typeof useInstanceWorkflowActions>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceWorkflowActions({
          resolveTarget: () => target,
          getExtendBlockedMessage: (instance) =>
            instance.share_scope === 'shared' ? '共享实例不支持手动延时' : null,
          onExtended: () => undefined,
          onDestroyed: () => undefined,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.extendInstance()
    await flushPromises()

    expect(instanceApiMocks.extendInstance).not.toHaveBeenCalled()
    expect(toastMocks.error).toHaveBeenCalledWith('共享实例不支持手动延时')

    wrapper.unmount()
  })

  it('销毁前取消确认后不应继续请求', async () => {
    confirmMock.mockResolvedValue(false)

    let destroyed = false
    let composable!: ReturnType<typeof useInstanceWorkflowActions>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceWorkflowActions({
          resolveTarget: () => ({ id: 'inst-1' }),
          onExtended: () => undefined,
          onDestroyed: () => {
            destroyed = true
          },
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.destroyInstance()
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceApiMocks.destroyInstance).not.toHaveBeenCalled()
    expect(destroyed).toBe(false)

    wrapper.unmount()
  })

  it('销毁失败时应优先展示接口返回消息', async () => {
    instanceApiMocks.destroyInstance.mockRejectedValue(
      new ApiError('实例仍在创建中，暂时不能销毁', { status: 409 })
    )

    const target = ref({ id: 'inst-1' })
    let composable!: ReturnType<typeof useInstanceWorkflowActions>
    const Harness = defineComponent({
      setup() {
        composable = useInstanceWorkflowActions({
          resolveTarget: () => target.value,
          onExtended: () => undefined,
          onDestroyed: () => undefined,
          destroyErrorMessage: '销毁实例失败',
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.destroyInstance()
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('实例仍在创建中，暂时不能销毁')
    expect(toastMocks.error).not.toHaveBeenCalledWith('销毁实例失败')

    wrapper.unmount()
  })
})
