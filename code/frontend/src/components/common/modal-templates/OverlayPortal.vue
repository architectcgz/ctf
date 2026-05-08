<script setup lang="ts">
import { toRef } from 'vue'
import type { HTMLAttributes } from 'vue'

import { useOverlayBehavior } from './useOverlayBehavior'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(
  defineProps<{
    open: boolean
    shellClass?: HTMLAttributes['class']
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
    lockScroll?: boolean
    transitionName?: string
  }>(),
  {
    shellClass: '',
    closeOnBackdrop: true,
    closeOnEscape: true,
    lockScroll: true,
    transitionName: 'overlay-portal-fade',
  }
)

const emit = defineEmits<{
  close: []
}>()

function requestClose(): void {
  emit('close')
}

function handleBackdropClick(): void {
  if (!props.closeOnBackdrop) return
  requestClose()
}

useOverlayBehavior({
  open: toRef(props, 'open'),
  close: requestClose,
  closeOnEscape: toRef(props, 'closeOnEscape'),
  lockScroll: toRef(props, 'lockScroll'),
})
</script>

<template>
  <Teleport to="body">
    <Transition :name="transitionName">
      <div v-if="open" v-bind="$attrs" :class="shellClass" @click.self="handleBackdropClick">
        <slot :close="requestClose" />
      </div>
    </Transition>
  </Teleport>
</template>
