import { describe, it, expect, vi } from 'vitest'
import { useAbortController } from '../useAbortController'

describe('useAbortController', () => {
  it('应该创建新的 AbortController', () => {
    const { createController, signal } = useAbortController()
    const controller = createController()

    expect(controller).toBeInstanceOf(AbortController)
    expect(signal()).toBeDefined()
  })

  it('应该在创建新 controller 时取消旧的', () => {
    const { createController } = useAbortController()
    const controller1 = createController()
    const abortSpy = vi.spyOn(controller1, 'abort')

    createController()

    expect(abortSpy).toHaveBeenCalled()
  })

  it('应该手动取消 controller', () => {
    const { createController, abort } = useAbortController()
    const controller = createController()
    const abortSpy = vi.spyOn(controller, 'abort')

    abort()

    expect(abortSpy).toHaveBeenCalled()
  })
})
