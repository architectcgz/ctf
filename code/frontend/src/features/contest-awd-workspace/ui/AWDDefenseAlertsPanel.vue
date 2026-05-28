<script setup lang="ts">
defineProps<{
  alerts: Array<{
    challengeId: string
    challengeTitle: string
    statusLabel: string
    tone: 'danger' | 'warning'
    issues: string[]
  }>
}>()
</script>

<template>
  <div v-if="alerts.length > 0" class="defense-alerts">
    <div
      v-for="alert in alerts"
      :key="alert.challengeId"
      class="defense-alert"
      :class="alert.tone"
    >
      <div class="defense-alert__header">
        <span class="alert-title">{{ alert.challengeTitle }}</span>
        <span class="alert-badge">{{ alert.statusLabel }}</span>
      </div>
      <div class="alert-issues">
        <span v-for="issue in alert.issues" :key="issue">{{ issue }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.defense-alert {
  background: var(--color-warning-soft);
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent);
  border-left: 3px solid var(--color-warning);
  border-radius: 0.5rem;
  margin-bottom: 0.75rem;
  padding: 0.75rem 1rem;
}

.defense-alert.danger {
  background: var(--color-danger-soft);
  border-color: color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-left-color: var(--color-danger);
}

.defense-alert__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.alert-title {
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.alert-badge {
  font-size: var(--font-size-11);
  font-weight: 900;
}

.alert-issues {
  color: var(--color-text-secondary);
  display: flex;
  gap: 0.5rem;
  font-size: var(--font-size-11);
  font-weight: 700;
  margin-top: 0.35rem;
}
</style>
