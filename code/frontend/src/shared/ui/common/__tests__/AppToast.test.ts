import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import AppToast from '../AppToast.vue'
import appToastSource from '../AppToast.vue?raw'
import { useToast, useToastState } from '@/shared/model/common/useToast'

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

  it('renders the close control as an accessible icon button and dismisses the toast', async () => {
    const wrapper = await mountToast()
    useToast().warning('环境即将到期')
    await nextTick()

    const closeButton = wrapper.get('button[aria-label="关闭提示"]')

    expect(closeButton.attributes('aria-label')).toBe('关闭提示')
    expect(closeButton.text()).toBe('')
    expect(closeButton.find('.app-toast-close-icon').exists()).toBe(true)

    await nextTick()
    await closeButton.trigger('click')
    await nextTick()

    expect(wrapper.text()).not.toContain('环境即将到期')
    expect(appToastSource).toContain('aria-label="关闭提示"')
  })
})
