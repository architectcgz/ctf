import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { ApiError } from '@/api/request'

import { useManagedInstanceDestroyAction } from './useManagedInstanceDestroyAction'

const instanceAccessApiMocks = vi.hoisted(() => ({
  destroyManagedInstanceByRole: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/instances', () => ({
  destroyManagedInstanceByRole: instanceAccessApiMocks.destroyManagedInstanceByRole,
}))

vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('useManagedInstanceDestroyAction', () => {
  beforeEach(() => {
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockReset()
    confirmMock.mockReset()
    confirmMock.mockResolvedValue(true)
  })

  it('销毁前应先经过共享危险确认', async () => {
    let destroyed = false
    let composable!: ReturnType<typeof useManagedInstanceDestroyAction>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDestroyAction({
          role: 'teacher',
          resolveTarget: (id) => ({ id }),
          buildConfirmOptions: () => ({
            title: '确认销毁实例',
            message: '确定要销毁该实例吗？此操作不可恢复。',
          }),
          onDestroyed: () => {
            destroyed = true
          },
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.destroyManagedInstance('inst-1')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).toHaveBeenCalledWith(
      'teacher',
      'inst-1'
    )
    expect(destroyed).toBe(true)

    wrapper.unmount()
  })

  it('取消确认后不应继续请求', async () => {
    confirmMock.mockResolvedValue(false)

    let composable!: ReturnType<typeof useManagedInstanceDestroyAction>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDestroyAction({
          role: 'admin',
          resolveTarget: (id) => ({ id }),
          buildConfirmOptions: () => ({
            title: '强制销毁实例',
            message: '确认后立即销毁。',
          }),
          onDestroyed: () => undefined,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.destroyManagedInstance('inst-2')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).not.toHaveBeenCalled()

    wrapper.unmount()
  })

  it('销毁失败时应把接口返回消息交还页面 owner', async () => {
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockRejectedValue(
      new ApiError('实例所属练习仍在结算中，暂时不能销毁', { status: 409 })
    )

    const errorSpy = vi.fn()
    let composable!: ReturnType<typeof useManagedInstanceDestroyAction>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDestroyAction({
          role: 'teacher',
          resolveTarget: (id) => ({ id }),
          buildConfirmOptions: () => ({
            title: '确认销毁实例',
            message: '确定要销毁该实例吗？此操作不可恢复。',
          }),
          onDestroyed: () => undefined,
          onDestroyError: errorSpy,
          fallbackErrorMessage: '销毁实例失败，请稍后重试',
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.destroyManagedInstance('inst-3')
    await flushPromises()

    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        message: '实例所属练习仍在结算中，暂时不能销毁',
      })
    )

    wrapper.unmount()
  })

  it('重复点击时应在 handler 层阻止第二次销毁请求', async () => {
    let release!: () => void
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve
        })
    )

    let composable!: ReturnType<typeof useManagedInstanceDestroyAction>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDestroyAction({
          role: 'admin',
          resolveTarget: (id) => ({ id }),
          buildConfirmOptions: () => ({
            title: '强制销毁实例',
            message: '确认后立即销毁。',
          }),
          onDestroyed: () => undefined,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    const first = composable.destroyManagedInstance('inst-4')
    const second = composable.destroyManagedInstance('inst-4')
    await flushPromises()

    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).toHaveBeenCalledTimes(1)

    release()
    await first
    await second

    wrapper.unmount()
  })
})
