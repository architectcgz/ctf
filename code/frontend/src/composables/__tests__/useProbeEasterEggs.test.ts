import { createApp, type App } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'

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

const STORAGE_KEY = 'ctf.probe-easter-eggs'

describe('useProbeEasterEggs', () => {
  let app: App | null = null

  afterEach(() => {
    app?.unmount()
    app = null
    vi.restoreAllMocks()
  })

  beforeEach(() => {
    sessionStorage.clear()
  })

  it('达到阈值前不应产生彩蛋信号', () => {
    const [probes, testApp] = withSetup(() => useProbeEasterEggs())
    app = testApp

    expect(probes.track('login-brand', 3)).toMatchObject({
      count: 1,
      unlocked: false,
      activated: false,
    })
    expect(probes.track('login-brand', 3)).toMatchObject({
      count: 2,
      unlocked: false,
      activated: false,
    })
  })

  it('达到阈值后应返回一次性彩蛋信号', () => {
    const [probes, testApp] = withSetup(() => useProbeEasterEggs())
    app = testApp

    probes.track('notification-refresh', 2)

    expect(probes.track('notification-refresh', 2)).toMatchObject({
      count: 2,
      unlocked: true,
      activated: true,
    })
  })

  it('同一会话重复触发同一彩蛋时不应重复返回未消费信号', () => {
    const [probes, testApp] = withSetup(() => useProbeEasterEggs())
    app = testApp

    probes.track('notification-id', 2)
    probes.track('notification-id', 2)

    expect(probes.track('notification-id', 2)).toMatchObject({
      count: 3,
      unlocked: false,
      activated: true,
    })
  })

  it('应从 sessionStorage 恢复已记录计数', () => {
    sessionStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        version: 1,
        counts: {
          'challenge-side-rail': 2,
        },
        activated: {},
      })
    )

    const [probes, testApp] = withSetup(() => useProbeEasterEggs())
    app = testApp

    expect(probes.track('challenge-side-rail', 3)).toMatchObject({
      count: 3,
      unlocked: true,
      activated: true,
    })
  })

  it('sessionStorage 不可用时应静默降级到内存态', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked')
    })

    const [probes, testApp] = withSetup(() => useProbeEasterEggs())
    app = testApp

    expect(probes.track('error-status', 2)).toMatchObject({
      count: 1,
      unlocked: false,
      activated: false,
    })
    expect(probes.track('error-status', 2)).toMatchObject({
      count: 2,
      unlocked: true,
      activated: true,
    })
  })
})
