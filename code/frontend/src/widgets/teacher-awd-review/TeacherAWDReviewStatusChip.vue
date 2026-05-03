<script setup lang="ts">
defineProps<{
  status: string
  label: string
}>()
</script>

<template>
  <span
    class="awd-review-status-chip"
    :class="`awd-review-status-chip--${status || 'idle'}`"
  >
    {{ label }}
  </span>
</template>

<style scoped>
.awd-review-status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  padding: 0 var(--space-3);
  border-radius: 999px;
  border: 1px solid var(--awd-review-line);
  background: color-mix(in srgb, var(--awd-review-line) 14%, var(--awd-review-surface));
  color: var(--awd-review-muted);
  font-size: var(--font-size-0-74);
  font-weight: 700;
}

.awd-review-status-chip--running {
  border-color: color-mix(in srgb, var(--awd-review-primary) 24%, transparent);
  background: color-mix(in srgb, var(--awd-review-primary) 10%, transparent);
  color: var(--awd-review-primary-strong);
}

.awd-review-status-chip--ended,
.awd-review-status-chip--frozen {
  border-color: color-mix(in srgb, var(--awd-review-line-strong) 86%, transparent);
  background: color-mix(in srgb, var(--awd-review-line) 18%, var(--awd-review-surface-subtle));
  color: var(--awd-review-muted);
}
</style>
