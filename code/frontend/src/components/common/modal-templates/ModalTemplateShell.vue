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

watch(
  () => [props.open, props.closeOnEscape] as const,
  ([open, closeOnEscape]) => {
    if (typeof window === 'undefined') return

    window.removeEventListener('keydown', handleWindowKeydown)
    if (open && closeOnEscape) {
      window.addEventListener('keydown', handleWindowKeydown)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('keydown', handleWindowKeydown)
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
  z-index: 90;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--modal-template-shell-overlay);
}

.modal-template-shell--frosted {
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 60%, transparent);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.modal-template-panel {
  position: relative;
}

.modal-template-fade-enter-active,
.modal-template-fade-leave-active {
  transition: opacity 0.22s ease;
}

.modal-template-fade-enter-from,
.modal-template-fade-leave-to {
  opacity: 0;
}
</style>
