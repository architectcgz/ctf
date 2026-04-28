<script setup lang="ts">
import type { AWDServiceAlertView } from '@/components/platform/contest/awdInspector.types'

defineProps<{
  alerts: AWDServiceAlertView[]
  selectedAlertKey: string
  getAlertClass: (alertKey: string) => string
}>()

const emit = defineEmits<{
  'select-alert': [alertKey: string]
}>()
</script>

<template>
  <div
    v-if="alerts.length > 0"
    class="alert-banner"
  >
    <span class="banner-tag">重点异常告警</span>
    <div class="alert-pills">
      <button
        v-for="alert in alerts"
        :key="alert.key"
        class="alert-pill"
        :class="[getAlertClass(alert.key), { 'is-active': selectedAlertKey === alert.key }]"
        type="button"
        @click="emit('select-alert', alert.key)"
      >
        {{ alert.label }} ({{ alert.count }})
      </button>
    </div>
  </div>
</template>

<style scoped>
.alert-banner {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  padding: var(--space-3) var(--space-5);
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface));
}

.banner-tag {
  color: var(--color-warning);
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
}

.alert-pills {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.alert-pill {
  padding: var(--space-1) var(--space-3);
  border: 1px solid
    var(--awd-service-alert-border, color-mix(in srgb, var(--color-warning) 30%, transparent));
  border-radius: 0.375rem;
  background: var(--awd-service-alert-bg, transparent);
  color: var(--awd-service-alert-color, var(--color-warning));
  cursor: pointer;
  font-size: var(--font-size-11);
  font-weight: 700;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.alert-pill:hover {
  background: var(--color-bg-elevated);
}

.alert-pill.is-active {
  border-color: var(--awd-service-alert-active-bg, var(--color-warning));
  background: var(--awd-service-alert-active-bg, var(--color-warning));
  color: var(--color-text-inverse);
}

.awd-service-alert--danger {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-danger) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-danger) 20%, transparent);
  --awd-service-alert-color: var(--color-danger);
  --awd-service-alert-active-bg: var(--color-danger);
}

.awd-service-alert--warning {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-warning) 20%, transparent);
  --awd-service-alert-color: var(--color-warning);
  --awd-service-alert-active-bg: var(--color-warning);
}

.awd-service-alert--neutral {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-text-muted) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-text-muted) 20%, transparent);
  --awd-service-alert-color: var(--color-text-primary);
  --awd-service-alert-active-bg: var(--color-text-secondary);
}
</style>
