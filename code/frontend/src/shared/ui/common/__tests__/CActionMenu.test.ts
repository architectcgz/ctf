import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'

import CActionMenu from '../menus/CActionMenu.vue'

const ActionMenuHost = defineComponent({
  components: {
    CActionMenu,
  },
  setup() {
    const open = ref(false)

    return {
      open,
    }
  },
  template: `
    <CActionMenu v-model:open="open" title="管理操作" menu-label="更多操作">
      <template #trigger="{ open: menuOpen, toggle, setTriggerRef }">
        <button
          :ref="setTriggerRef"
          type="button"
          class="c-action-menu__trigger c-action-menu__trigger--icon"
          aria-haspopup="menu"
          :aria-expanded="menuOpen ? 'true' : 'false'"
          @click="toggle"
        >
          更多
        </button>
      </template>

      <template #default="{ close }">
        <button type="button" class="c-action-menu__item" role="menuitem" @click="close">
          编辑
        </button>
      </template>
    </CActionMenu>
  `,
})

describe('CActionMenu', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('应通过 Teleport 渲染到 body，并用共享浮层结构承接菜单内容', async () => {
    const wrapper = mount(ActionMenuHost, {
      attachTo: document.body,
    })

    await wrapper.get('button[aria-haspopup="menu"]').trigger('click')
    await nextTick()

    expect(document.body.querySelector('[data-action-menu-layer]')).not.toBeNull()
    expect(document.body.querySelector('[data-action-menu-panel]')).not.toBeNull()
    expect(document.body.textContent).toContain('管理操作')
    expect(document.body.textContent).toContain('编辑')

    wrapper.unmount()
  })

  it('点击 trigger、backdrop 和 Escape 时应通过 v-model 正确开关菜单', async () => {
    const wrapper = mount(ActionMenuHost, {
      attachTo: document.body,
    })
    const menuWrapper = wrapper.findComponent(CActionMenu)
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('click')
    await nextTick()
    expect(menuWrapper.emitted('update:open')?.[0]).toEqual([true])
    expect(trigger.attributes('aria-expanded')).toBe('true')

    document.body.querySelector<HTMLElement>('[data-action-menu-layer]')?.click()
    await nextTick()
    expect(menuWrapper.emitted('update:open')?.[1]).toEqual([false])
    expect(menuWrapper.emitted('close')?.[0]).toEqual([])
    expect(trigger.attributes('aria-expanded')).toBe('false')

    await trigger.trigger('click')
    await nextTick()
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()

    expect(menuWrapper.emitted('update:open')?.[2]).toEqual([true])
    expect(menuWrapper.emitted('update:open')?.[3]).toEqual([false])
    expect(menuWrapper.emitted('close')?.[1]).toEqual([])

    wrapper.unmount()
  })
})
