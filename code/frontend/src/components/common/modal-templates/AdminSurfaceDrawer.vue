<script setup lang="ts">
import SlideOverDrawer from './SlideOverDrawer.vue'

withDefaults(
  defineProps<{
    open: boolean
    title: string
    subtitle?: string
    eyebrow?: string
    width?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    subtitle: '',
    eyebrow: 'Admin Actions',
    width: '32rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()
</script>

<template>
  <SlideOverDrawer
    :open="open"
    :title="title"
    :subtitle="subtitle"
    :eyebrow="eyebrow"
    :width="width"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    @update:open="emit('update:open', $event)"
    @close="emit('close')"
  >
    <template
      v-if="$slots.icon"
      #icon
    >
      <slot name="icon" />
    </template>

    <slot />

    <template
      v-if="$slots.footer"
      #footer
    >
      <slot name="footer" />
    </template>
  </SlideOverDrawer>
</template>
