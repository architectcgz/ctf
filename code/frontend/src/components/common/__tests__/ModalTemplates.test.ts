import { existsSync, readFileSync } from 'node:fs'

import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

const shellPath = `${process.cwd()}/src/components/common/modal-templates/ModalTemplateShell.vue`
const classicPath = `${process.cwd()}/src/components/common/modal-templates/ClassicCenteredModal.vue`
const drawerPath = `${process.cwd()}/src/components/common/modal-templates/SlideOverDrawer.vue`
const minimalPath = `${process.cwd()}/src/components/common/modal-templates/MinimalFloatingModal.vue`
const cTooltipPath = `${process.cwd()}/src/components/common/modal-templates/CContextTooltip.vue`
const cPopoverPath = `${process.cwd()}/src/components/common/modal-templates/CLightActionPopover.vue`
const cConfirmPath = `${process.cwd()}/src/components/common/modal-templates/CImmersiveConfirmDialog.vue`
const cInputPath = `${process.cwd()}/src/components/common/modal-templates/CFocusedInputDialog.vue`
const adminModalPath = `${process.cwd()}/src/components/common/modal-templates/AdminSurfaceModal.vue`
const adminDrawerPath = `${process.cwd()}/src/components/common/modal-templates/AdminSurfaceDrawer.vue`

function readSource(path: string): string {
  return existsSync(path) ? readFileSync(path, 'utf-8') : ''
}

async function loadClassicCenteredModal() {
  if (!existsSync(classicPath)) return null
  const modules = import.meta.glob('../modal-templates/*.vue')
  const loader = modules['../modal-templates/ClassicCenteredModal.vue']
  if (!loader) return null
  return ((await loader()) as { default: unknown }).default
}

async function loadComponent(path: string) {
  const modules = import.meta.glob('../modal-templates/*.vue')
  const loader = modules[path]
  if (!loader) return null
  return ((await loader()) as { default: unknown }).default
}

afterEach(() => {
  document.body.innerHTML = ''
})

describe('modal templates', () => {
  it('应该提供共享弹窗壳、既有模板和 C 端模板组件', () => {
    expect(existsSync(shellPath)).toBe(true)
    expect(existsSync(classicPath)).toBe(true)
    expect(existsSync(drawerPath)).toBe(true)
    expect(existsSync(minimalPath)).toBe(true)
    expect(existsSync(cTooltipPath)).toBe(true)
    expect(existsSync(cPopoverPath)).toBe(true)
    expect(existsSync(cConfirmPath)).toBe(true)
    expect(existsSync(cInputPath)).toBe(true)
    expect(existsSync(adminModalPath)).toBe(true)
    expect(existsSync(adminDrawerPath)).toBe(true)
  })

  it('经典居中弹窗应支持标题、插槽内容以及 backdrop 和 Escape 关闭', async () => {
    const ClassicCenteredModal = await loadClassicCenteredModal()

    expect(ClassicCenteredModal).not.toBeNull()
    if (!ClassicCenteredModal) return

    const wrapper = mount(ClassicCenteredModal, {
      props: {
        open: true,
        title: '编辑资源信息',
        subtitle: 'Resource Editor',
      },
      slots: {
        icon: '<span class="modal-icon-slot">I</span>',
        default: '<div class="modal-body-slot">字段内容</div>',
        footer: '<button class="modal-footer-slot">保存修改</button>',
      },
      global: {
        stubs: {
          teleport: true,
        },
      },
    })

    await nextTick()

    wrapper.get('.modal-template-shell')
    wrapper.get('.modal-template-panel--classic')
    expect(wrapper.text()).toContain('编辑资源信息')
    expect(wrapper.text()).toContain('Resource Editor')
    expect(wrapper.html()).toContain('modal-icon-slot')
    expect(wrapper.html()).toContain('modal-body-slot')
    expect(wrapper.html()).toContain('modal-footer-slot')

    await wrapper.get('.modal-template-shell').trigger('click')
    expect(wrapper.emitted('update:open')?.[0]).toEqual([false])
    expect(wrapper.emitted('close')?.[0]).toEqual([])

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(wrapper.emitted('update:open')?.[1]).toEqual([false])
  })

  it('后台弹窗模板应把外部 class 和 data 属性绑定到外层壳而不是触发 fallthrough warning', async () => {
    const AdminSurfaceModal = await loadComponent('../modal-templates/AdminSurfaceModal.vue')

    expect(AdminSurfaceModal).not.toBeNull()
    if (!AdminSurfaceModal) return

    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    try {
      const wrapper = mount(AdminSurfaceModal, {
        attrs: {
          class: 'admin-surface-tone',
          'data-dialog-surface': 'contest',
        },
        props: {
          open: true,
          title: '创建竞赛',
          subtitle: '填写竞赛基础信息',
        },
        slots: {
          default: '<div class="admin-surface-body">内容</div>',
        },
        global: {
          stubs: {
            teleport: true,
          },
        },
      })

      await nextTick()

      const shell = wrapper.get('.modal-template-shell')
      expect(shell.classes()).toContain('admin-surface-tone')
      expect(shell.attributes('data-dialog-surface')).toBe('contest')
      expect(
        warnSpy.mock.calls.some(([message]) =>
          String(message).includes('Extraneous non-props attributes')
        )
      ).toBe(false)
    } finally {
      warnSpy.mockRestore()
    }
  })

  it('模板组件应保留文档里的关键视觉骨架', () => {
    const shellSource = readSource(shellPath)
    const classicSource = readSource(classicPath)
    const drawerSource = readSource(drawerPath)
    const minimalSource = readSource(minimalPath)
    const tooltipSource = readSource(cTooltipPath)
    const popoverSource = readSource(cPopoverPath)
    const confirmSource = readSource(cConfirmPath)
    const inputSource = readSource(cInputPath)
    const adminModalSource = readSource(adminModalPath)
    const adminDrawerSource = readSource(adminDrawerPath)

    expect(shellSource).toContain('Teleport to="body"')
    expect(shellSource).toContain("emit('update:open', false)")
    expect(shellSource).toContain('.modal-template-shell')
    expect(shellSource).not.toContain('background: rgba(15, 23, 42, 0.4);')

    expect(classicSource).toContain('.modal-template-panel--classic')
    expect(classicSource).toContain('Resource Editor')
    expect(classicSource).toContain('.modal-template-classic__icon')
    expect(classicSource).not.toContain('background: #ffffff;')
    expect(classicSource).not.toContain('color: #2563eb;')
    expect(classicSource).not.toContain('background: rgba(219, 234, 254, 1);')

    expect(drawerSource).toContain('.modal-template-panel--drawer')
    expect(drawerSource).toContain('.modal-template-drawer__head-main')
    expect(drawerSource).toContain('.modal-template-drawer__title-block')
    expect(drawerSource).toContain('.modal-template-drawer__icon')
    expect(drawerSource).not.toContain("eyebrow: 'Advanced Editor'")
    expect(drawerSource).not.toContain('高度承载')
    expect(drawerSource).not.toContain('background: #ffffff;')
    expect(drawerSource).not.toContain('color: #059669;')
    expect(drawerSource).not.toContain('background: rgba(236, 253, 245, 1);')

    expect(minimalSource).toContain('.modal-template-panel--minimal')
    expect(minimalSource).toContain('.modal-template-minimal__title')
    expect(minimalSource).toContain('快捷编辑')
    expect(minimalSource).not.toContain('background: #ffffff;')
    expect(minimalSource).not.toContain('border-bottom-color: #7c3aed;')

    expect(tooltipSource).toContain('TLS 握手')
    expect(tooltipSource).toContain('.c-context-tooltip__trigger')
    expect(tooltipSource).toContain('.c-context-tooltip__panel')
    expect(tooltipSource).not.toContain('bg-slate-900')
    expect(tooltipSource).not.toContain('text-[13px]')
    expect(tooltipSource).not.toContain('border-slate-400')

    expect(popoverSource).toContain('发现题目问题？')
    expect(popoverSource).toContain('发送反馈')
    expect(popoverSource).toContain('.c-light-action-popover__panel')
    expect(popoverSource).not.toContain('text-[13px]')
    expect(popoverSource).not.toContain('text-[14px]')
    expect(popoverSource).not.toContain('text-[12px]')
    expect(popoverSource).not.toContain('shadow-[0_12px_40px_rgba(0,0,0,0.12)]')
    expect(popoverSource).not.toContain('bg-[#2a7a58]')
    expect(popoverSource).not.toContain('focus:border-[#2a7a58]')

    expect(confirmSource).toContain('确认重建靶机环境？')
    expect(confirmSource).toContain('backdrop-blur-sm')
    expect(confirmSource).toContain('确认销毁重建')
    expect(confirmSource).not.toContain('background: #ffffff;')
    expect(confirmSource).not.toContain('color: #ef4444;')
    expect(confirmSource).not.toContain('background: #dc2626;')

    expect(inputSource).toContain('创建新队伍')
    expect(inputSource).toContain('确认创建')
    expect(inputSource).toContain('.c-focused-input-dialog__surface')
    expect(inputSource).toContain('.c-focused-input-dialog__header')
    expect(inputSource).toContain('.c-focused-input-dialog__form')
    expect(inputSource).not.toContain('background: #ffffff;')
    expect(inputSource).not.toContain('border-color: #2a7a58;')
    expect(inputSource).not.toContain('background: #2a7a58;')
    expect(inputSource).not.toContain('background: #206346;')

    expect(adminModalSource).toContain('ClassicCenteredModal')
    expect(adminModalSource).toContain('Admin Workspace')
    expect(adminDrawerSource).toContain('SlideOverDrawer')
    expect(adminDrawerSource).toContain('Admin Actions')
  })

  it('C 端专注型输入弹窗应支持标题、描述、表单与页脚动作插槽', async () => {
    const FocusedInputDialog = await loadComponent('../modal-templates/CFocusedInputDialog.vue')

    expect(FocusedInputDialog).not.toBeNull()
    if (!FocusedInputDialog) return

    const wrapper = mount(FocusedInputDialog, {
      props: {
        open: true,
        title: '创建新队伍',
        description: '为你的战队起一个响亮的代号。',
      },
      slots: {
        default: '<input class="focused-input-slot" />',
        footer: '<button class="focused-footer-slot">确认创建</button>',
      },
      global: {
        stubs: {
          teleport: true,
        },
      },
    })

    await nextTick()

    expect(wrapper.text()).toContain('创建新队伍')
    expect(wrapper.text()).toContain('为你的战队起一个响亮的代号。')
    expect(wrapper.find('.c-focused-input-dialog__surface').exists()).toBe(true)
    expect(wrapper.find('.c-focused-input-dialog__header').exists()).toBe(true)
    expect(wrapper.find('.c-focused-input-dialog__form').html()).toContain('focused-input-slot')
    expect(wrapper.html()).toContain('focused-footer-slot')
  })
})
