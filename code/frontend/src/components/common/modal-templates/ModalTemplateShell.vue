<script setup lang="ts">
import OverlayPortal from './OverlayPortal.vue'

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

function closeDialog(): void {
  emit('update:open', false)
  emit('close')
}
</script>

<template>
  <OverlayPortal
    v-bind="$attrs"
    :open="open"
    :shell-class="[
      'modal-template-shell',
      overlayClass,
      { 'modal-template-shell--frosted': frosted },
    ]"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    transition-name="modal-template-fade"
    @close="closeDialog"
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
  </OverlayPortal>
</template>

<style>
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
