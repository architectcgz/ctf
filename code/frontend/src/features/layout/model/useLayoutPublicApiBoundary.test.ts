import { describe, expect, it } from 'vitest'

import featureIndexSource from '../index.ts?raw'
import modelIndexSource from './index.ts?raw'
import appShellPageSource from '../../../pages/AppShellRoutePage.vue?raw'

describe('layout feature public API boundaries', () => {
  it('feature 顶层 index.ts 应只从 model/index.ts 重新导出，不深链 model 内部文件', () => {
    expect(featureIndexSource).toMatch(/export \* from '\.\/model'/)
    expect(featureIndexSource).not.toMatch(/from\s+['"]\.\/model\/useLayout/)
  })

  it('model/index.ts 应暴露三个 bridge composable 的公共入口', () => {
    expect(modelIndexSource).toContain("export { useLayoutSessionActionsBridge }")
    expect(modelIndexSource).toContain("export { useLayoutNotificationDrawerBridge }")
    expect(modelIndexSource).toContain("export { useLayoutNotificationRealtimeBridge }")
  })

  it('外部 consumer 应通过 public API import，不应存在深链到 model 内部文件的路径', () => {
    // AppShellRoutePage 应从 @/features/layout 顶层引入，不应深链
    expect(appShellPageSource).toContain("from '@/features/layout'")
    expect(appShellPageSource).not.toMatch(
      /from\s+['"]@\/features\/layout\/model\/useLayout/
    )
  })
})
