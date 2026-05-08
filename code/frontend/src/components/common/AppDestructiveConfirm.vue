<script setup lang="ts">
import { computed } from 'vue'

import DeleteConfirmModal from '@/components/common/DeleteConfirmModal.vue'
import { useDestructiveConfirmState } from '@/composables/useDestructiveConfirm'

const { current, visible, confirm, cancel } = useDestructiveConfirmState()

const modalValue = computed({
  get: () => visible.value,
  set: (next: boolean) => {
    if (!next) {
      cancel()
    }
  },
})
</script>

<template>
  <DeleteConfirmModal
    v-model="modalValue"
    :title="current?.title || ''"
    :description="current?.message || ''"
    :warning="current?.warning || ''"
    :confirm-text="current?.confirmButtonText || '确认'"
    :cancel-text="current?.cancelButtonText || '取消'"
    :close-on-backdrop="current?.closeOnBackdrop ?? true"
    :close-on-escape="current?.closeOnEscape ?? true"
    @confirm="confirm"
  />
</template>
