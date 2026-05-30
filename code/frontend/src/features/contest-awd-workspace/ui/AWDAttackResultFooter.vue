<template>
  <footer class="ops-panel__footer">
    <div
      class="result-alert"
      :class="success ? 'success' : 'danger'"
    >
      <Terminal class="h-3.5 w-3.5" />
      <span>{{ message }}</span>
    </div>
  </footer>
</template>

<style scoped>
.ops-panel__footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
}

.result-alert {
  padding: 0.65rem 1rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: var(--font-size-12);
  font-weight: 900;
}

.result-alert.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border: 1px solid color-mix(in srgb, var(--color-success) 20%, transparent);
}

.result-alert.danger {
  background: var(--color-danger-soft);
  color: var(--color-danger);
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
}
</style>

<script setup lang="ts">
import { Terminal } from 'lucide-vue-next'

defineProps<{
  success: boolean
  message: string
}>()
</script>
