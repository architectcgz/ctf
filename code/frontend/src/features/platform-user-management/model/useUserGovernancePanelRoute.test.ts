import { beforeEach, describe, expect, it, vi } from 'vitest'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string | string[] | undefined>,
}))

const replaceMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    replace: replaceMock,
  }),
}))

import { useUserGovernancePanelRoute } from './useUserGovernancePanelRoute'

describe('useUserGovernancePanelRoute', () => {
  beforeEach(() => {
    routeState.query = {}
    replaceMock.mockReset()
  })

  it('应将缺省与旧的 directory query 归一到 overview', () => {
    expect(useUserGovernancePanelRoute().activePanel.value).toBe('overview')

    routeState.query = { panel: 'directory' }

    expect(useUserGovernancePanelRoute().activePanel.value).toBe('overview')
  })

  it('应保留 import panel', () => {
    routeState.query = { panel: 'import' }

    expect(useUserGovernancePanelRoute().activePanel.value).toBe('import')
  })

  it('切到 overview 时应移除 panel query', async () => {
    routeState.query = { panel: 'import', keyword: 'alice' }

    await useUserGovernancePanelRoute().switchPanel('overview')

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'UserManage',
      query: { keyword: 'alice' },
    })
  })

  it('切到 import 时应写入 panel query', async () => {
    routeState.query = { keyword: 'alice' }

    await useUserGovernancePanelRoute().switchPanel('import')

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'UserManage',
      query: { keyword: 'alice', panel: 'import' },
    })
  })

  it('切到当前 panel 时不应重复导航', async () => {
    routeState.query = { panel: 'import' }

    await useUserGovernancePanelRoute().switchPanel('import')

    expect(replaceMock).not.toHaveBeenCalled()
  })
})
