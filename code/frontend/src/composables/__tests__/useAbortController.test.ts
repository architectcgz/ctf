import { createApp, type App } from 'vue'
import { describe, it, expect, vi, afterEach } from 'vitest'

import { useAbortController } from '../useAbortController'

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

let app: App | null = null

afterEach(() => {
  app?.unmount()
  app = null
})

describe('useAbortController', () => {
  it('应该创建新的 AbortController', () => {
    const [composable, testApp] = withSetup(() => useAbortController())
    app = testApp
    const { createController, signal } = composable
    const controller = createController()

    expect(controller).toBeInstanceOf(AbortController)
    expect(signal()).toBeDefined()
  })

  it('应该在创建新 controller 时取消旧的', () => {
    const [composable, testApp] = withSetup(() => useAbortController())
    app = testApp
    const { createController } = composable
    const controller1 = createController()
    const abortSpy = vi.spyOn(controller1, 'abort')

    createController()

    expect(abortSpy).toHaveBeenCalled()
  })

  it('应该手动取消 controller', () => {
    const [composable, testApp] = withSetup(() => useAbortController())
    app = testApp
    const { createController, abort } = composable
    const controller = createController()
    const abortSpy = vi.spyOn(controller, 'abort')

    abort()

    expect(abortSpy).toHaveBeenCalled()
  })
})
