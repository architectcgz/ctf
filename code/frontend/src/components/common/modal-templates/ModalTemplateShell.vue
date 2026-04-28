<script setup lang="ts">
import { onBeforeUnmount, useAttrs, watch } from 'vue'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(
  defineProps<{
    open: boolean
    panelClass?: string | string[]
    overlayClass?: string | string[]
    panelStyle?: Record<string, string>
    panelTag?: 'div' | 'aside' | 'section'
    ariaLabel?: string
    role?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
    frosted?: boolean
  }>(),
  {
    panelClass: '',
    overlayClass: '',
    panelStyle: () => ({}),
    panelTag: 'div',
    ariaLabel: '对话框',
    role: 'dialog',
    closeOnBackdrop: true,
    closeOnEscape: true,
    frosted: false,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const attrs = useAttrs()

function closeDialog(): void {
  emit('update:open', false)
  emit('close')
}

function handleBackdropClick(): void {
  if (!props.closeOnBackdrop) return
  closeDialog()
}

function handleWindowKeydown(event: KeyboardEvent): void {
  if (!props.open || !props.closeOnEscape || event.key !== 'Escape') return
  closeDialog()
}

// 统一管理副作用：键盘监听与滚动锁定
watch(
  () => [props.open, props.closeOnEscape] as const,
  ([open, closeOnEscape]) => {
    if (typeof window === 'undefined') return

    // 处理键盘事件
    window.removeEventListener('keydown', handleWindowKeydown)
    if (open && closeOnEscape) {
      window.addEventListener('keydown', handleWindowKeydown)
    }

    // 处理背景滚动锁定
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('keydown', handleWindowKeydown)
  // 确保组件销毁时恢复滚动
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-template-fade">
      <div
        v-if="open"
        v-bind="attrs"
        :class="[
          'modal-template-shell',
          overlayClass,
          { 'modal-template-shell--frosted': frosted },
        ]"
        @click.self="handleBackdropClick"
      >
        <component
          :is="panelTag"
          :class="['modal-template-panel', panelClass]"
          :style="panelStyle"
          :role="role"
          :aria-label="ariaLabel"
          aria-modal="true"
        >
          <slot />
        </component>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-template-shell {
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 40%, transparent);
  position: fixed;
  inset: 0;
  z-index: var(--ui-dialog-z-index);
  display: flex;

  /* 优雅做法：使用变量控制布局，默认居中 */
  align-items: var(--modal-shell-align, center);
  justify-content: var(--modal-shell-justify, center);
  padding: var(--modal-shell-padding, var(--space-4));

  background: var(--modal-template-shell-overlay);
}

.modal-template-shell--frosted {
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 60%, transparent);
  backdrop-filter: blur(var(--modal-shell-blur, var(--space-3)));
  -webkit-backdrop-filter: blur(var(--modal-shell-blur, var(--space-3)));
}

.modal-template-panel {
  position: relative;
}

.modal-template-fade-enter-active,
.modal-template-fade-leave-active {
  transition: opacity var(--ui-motion-normal);
}

.modal-template-fade-enter-from,
.modal-template-fade-leave-to {
  opacity: 0;
}
</style>
