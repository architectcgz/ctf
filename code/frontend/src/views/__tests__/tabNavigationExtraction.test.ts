import { existsSync, readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const tabNavigationHelperPath = `${process.cwd()}/src/composables/useTabKeyboardNavigation.ts`
const routeQueryTabsPath = `${process.cwd()}/src/composables/useRouteQueryTabs.ts`
const urlSyncedTabsPath = `${process.cwd()}/src/composables/useUrlSyncedTabs.ts`

describe('tab navigation extraction', () => {
  it('route query 和 url synced tabs 应共用按钮 ref 与键盘导航 helper，而不是各自维护一套实现', () => {
    expect(existsSync(tabNavigationHelperPath)).toBe(true)

    const tabNavigationHelperSource = readFileSync(tabNavigationHelperPath, 'utf-8')
    const routeQueryTabsSource = readFileSync(routeQueryTabsPath, 'utf-8')
    const urlSyncedTabsSource = readFileSync(urlSyncedTabsPath, 'utf-8')

    expect(tabNavigationHelperSource).toContain('export function useTabKeyboardNavigation')
    expect(tabNavigationHelperSource).toContain('setTabButtonRef')
    expect(tabNavigationHelperSource).toContain('handleTabKeydown')

    for (const source of [routeQueryTabsSource, urlSyncedTabsSource]) {
      expect(source).toContain(
        "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
      )
      expect(source).toContain('useTabKeyboardNavigation<')
      expect(source).not.toContain('const tabButtonRefs:')
      expect(source).not.toContain('function setTabButtonRef(')
      expect(source).not.toContain('function focusTab(')
      expect(source).not.toContain('function handleTabKeydown(')
    }
  })
})
