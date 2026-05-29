<script setup lang="ts">
withDefaults(
  defineProps<{
    cancelId?: string
    submitId?: string
    cancelText?: string
    submitText: string
    savingText: string
    saving: boolean
    disabled?: boolean
  }>(),
  {
    cancelId: undefined,
    submitId: undefined,
    cancelText: '取消',
    disabled: false,
  }
)

const emit = defineEmits<{
  cancel: []
  submit: []
}>()
</script>

<template>
  <div class="awd-operations-dialog__footer">
    <button
      :id="cancelId"
      type="button"
      class="ui-btn ui-btn--secondary"
      @click="emit('cancel')"
    >
      {{ cancelText }}
    </button>
    <button
      :id="submitId"
      type="button"
      class="ui-btn ui-btn--primary"
      :disabled="disabled"
      @click="emit('submit')"
    >
      {{ saving ? savingText : submitText }}
    </button>
  </div>
</template>
