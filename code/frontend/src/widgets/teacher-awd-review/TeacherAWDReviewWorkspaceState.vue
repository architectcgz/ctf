<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'

defineProps<{
  loading: boolean
  error: string | null
  hasReview: boolean
}>()

const emit = defineEmits<{
  loadReview: []
}>()
</script>

<template>
  <div
    v-if="loading"
    class="teacher-empty-state workspace-directory-empty awd-review-loading"
  >
    <div class="academy-spinner" />
    <p>正在载入复盘分析数据...</p>
  </div>

  <AppEmpty
    v-else-if="error"
    title="复盘详情加载失败"
    :description="error"
    icon="AlertCircle"
    class="teacher-empty-state workspace-directory-empty"
  >
    <template #action>
      <button
        type="button"
        class="header-btn header-btn--primary"
        @click="emit('loadReview')"
      >
        重新加载
      </button>
    </template>
  </AppEmpty>

  <slot v-else-if="hasReview" />
</template>

<style scoped>
.awd-review-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  min-height: 14rem;
  color: var(--awd-review-muted);
}

.academy-spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid color-mix(in srgb, var(--awd-review-line) 88%, transparent);
  border-top-color: var(--awd-review-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
