import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import AppToast from '../AppToast.vue'
import { useToast, useToastState } from '@/composables/useToast'
import { useTheme } from '@/composables/useTheme'

function resetToastState() {
  const toast = useToast()
  const { toasts } = useToastState()

  for (const item of [...toasts.value]) {
    toast.dismiss(item.id)
  }
}

async function mountToastWithTheme(themeName: 'light' | 'dark') {
  const { theme } = useTheme()
  theme.value = themeName
  document.documentElement.setAttribute('data-theme', themeName)

  const wrapper = mount(AppToast)
  useToast().success('操作成功')
  await nextTick()

  return wrapper
}

describe('AppToast', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
    resetToastState()
  })

  afterEach(() => {
    resetToastState()
    vi.useRealTimers()
  })

  it('uses a white solid background in light theme', async () => {
    const wrapper = await mountToastWithTheme('light')
    const toastItem = wrapper.get('.app-toast-item')

    expect(toastItem.classes()).toContain('bg-white')
    expect(toastItem.classes()).not.toContain('bg-surface')
  })

  it('keeps the dark surface background in dark theme', async () => {
    const wrapper = await mountToastWithTheme('dark')
    const toastItem = wrapper.get('.app-toast-item')

    expect(toastItem.classes()).toContain('bg-surface')
    expect(toastItem.classes()).not.toContain('bg-white')
  })
})
