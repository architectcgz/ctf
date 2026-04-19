<script setup lang="ts">
import ClassicCenteredModal from './ClassicCenteredModal.vue'

withDefaults(
  defineProps<{
    open: boolean
    title: string
    subtitle?: string
    eyebrow?: string
    width?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
    frosted?: boolean
  }>(),
  {
    subtitle: '',
    eyebrow: 'Admin Workspace',
    width: '40rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
    frosted: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()
</script>

<template>
  <ClassicCenteredModal
    :open="open"
    :title="title"
    :subtitle="subtitle"
    :eyebrow="eyebrow"
    :width="width"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    :frosted="frosted"
    @update:open="emit('update:open', $event)"
    @close="emit('close')"
  >
    <template v-if="$slots.icon" #icon>
      <slot name="icon" />
    </template>

    <slot />

    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </ClassicCenteredModal>
</template>
