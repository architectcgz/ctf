import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import AppToast from '../AppToast.vue'
import { useToast, useToastState } from '@/composables/useToast'

function resetToastState() {
  const toast = useToast()
  const { toasts } = useToastState()

  for (const item of [...toasts.value]) {
    toast.dismiss(item.id)
  }
}

async function mountToast() {
  const wrapper = mount(AppToast)
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

  it('uses tokenized surface class instead of hardcoded light/dark background utilities', async () => {
    const wrapper = await mountToast()
    useToast().success('操作成功')
    await nextTick()

    const toastItem = wrapper.get('.app-toast-item')

    expect(toastItem.classes()).toContain('app-toast-item')
    expect(toastItem.classes()).not.toContain('bg-white')
    expect(toastItem.classes()).not.toContain('bg-surface')
  })

  it('uses primary token in info toast close button style', async () => {
    const wrapper = await mountToast()
    useToast().info('主题适配检查')
    await nextTick()

    const closeButton = wrapper.get('.app-toast-close')
    const style = closeButton.attributes('style') ?? ''

    expect(style).toContain('var(--color-primary)')
    expect(style).not.toContain('8, 145, 178')
  })
})
