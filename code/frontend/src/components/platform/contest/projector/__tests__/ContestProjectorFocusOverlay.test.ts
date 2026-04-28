import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'

import ContestProjectorFocusOverlay from '@/components/platform/contest/projector/ContestProjectorFocusOverlay.vue'
import focusOverlaySource from '@/components/platform/contest/projector/ContestProjectorFocusOverlay.vue?raw'

afterEach(() => {
  document.body.innerHTML = ''
})

describe('ContestProjectorFocusOverlay', () => {
  it('打开时居中承载聚焦内容，并支持背景点击关闭', async () => {
    const wrapper = mount(ContestProjectorFocusOverlay, {
      props: {
        activePanel: 'traffic',
      },
      slots: {
        default: '<section class="focused-panel-body">代理流量</section>',
      },
    })

    expect(wrapper.get('[role="dialog"]').attributes('aria-modal')).toBe('true')
    expect(wrapper.text()).toContain('代理流量')

    await wrapper.get('.projector-focus-overlay').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('支持 Escape 关闭，并在未打开时不渲染遮罩', async () => {
    const wrapper = mount(ContestProjectorFocusOverlay, {
      props: {
        activePanel: 'services',
      },
    })

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ activePanel: null })
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('聚焦弹出框内嵌面板应使用实体背景，避免内容区域透明', () => {
    expect(focusOverlaySource).toContain('background: var(--projector-focus-panel-surface);')
    expect(focusOverlaySource).toContain('background: var(--projector-focus-content-surface);')
    expect(focusOverlaySource).not.toContain('var(--color-bg-elevated) 68%, transparent')
  })
})
