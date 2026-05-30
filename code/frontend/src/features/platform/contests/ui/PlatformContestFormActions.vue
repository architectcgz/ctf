<script setup lang="ts">
defineProps<{
  mode: 'create' | 'edit'
  saving: boolean
  showCancel: boolean
}>()

const emit = defineEmits<{
  cancel: []
}>()
</script>

<template>
  <div class="contest-form-actions">
    <button
      v-if="showCancel"
      type="button"
      class="ui-btn ui-btn--secondary contest-form-button contest-form-button--secondary"
      @click="emit('cancel')"
    >
      取消
    </button>
    <button
      type="submit"
      class="ui-btn ui-btn--primary contest-form-button contest-form-button--primary"
      :disabled="saving"
    >
      {{ saving ? '保存中...' : mode === 'create' ? '创建竞赛' : '保存变更' }}
    </button>
  </div>
</template>
