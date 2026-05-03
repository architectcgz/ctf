import { beforeEach, describe, expect, it, vi } from 'vitest'

const routeState = vi.hoisted(() => ({
  query: {
    redirect: undefined as unknown,
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
}))

import { useLoginViewPage } from './useLoginViewPage'

describe('useLoginViewPage', () => {
  beforeEach(() => {
    routeState.query.redirect = undefined
  })

  it('应保留站内跳转路径', () => {
    routeState.query.redirect = '/teacher/dashboard'

    const { redirectTo } = useLoginViewPage()

    expect(redirectTo.value).toBe('/teacher/dashboard')
  })

  it('应拦截外链跳转', () => {
    routeState.query.redirect = 'https://evil.example/phish'

    const { redirectTo } = useLoginViewPage()

    expect(redirectTo.value).toBe('/')
  })

  it('应回退空值与空白跳转', () => {
    routeState.query.redirect = '   '

    const blank = useLoginViewPage()
    expect(blank.redirectTo.value).toBe('/')

    routeState.query.redirect = undefined

    const missing = useLoginViewPage()
    expect(missing.redirectTo.value).toBe('/')
  })
})
