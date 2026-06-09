<script setup lang="ts">
defineProps<{
  slaScore: number
  defenseScore: number
  slaError: string
  defenseError: string
}>()

const emit = defineEmits<{
  'update:slaScore': [value: number]
  'update:defenseScore': [value: number]
}>()
</script>

<template>
  <section class="awd-config-form-section awd-config-card awd-config-card--compact">
    <header class="list-heading awd-config-section-head">
      <div>
        <div class="journal-note-label">Score Weight</div>
        <h3 class="list-heading__title">权重设置</h3>
      </div>
    </header>
    <div class="awd-config-score-grid">
      <label class="ui-field">
        <span class="ui-field__label">SLA 分</span>
        <span class="ui-control-wrap" :class="{ 'is-error': !!slaError }">
          <input
            :value="slaScore"
            type="number"
            min="0"
            max="5"
            step="1"
            class="ui-control"
            @input="emit('update:slaScore', Number(($event.target as HTMLInputElement).value))"
          />
        </span>
        <span v-if="slaError" class="ui-field__error">{{ slaError }}</span>
      </label>
      <label class="ui-field">
        <span class="ui-field__label">防守分</span>
        <span class="ui-control-wrap" :class="{ 'is-error': !!defenseError }">
          <input
            :value="defenseScore"
            type="number"
            min="0"
            max="5"
            step="1"
            class="ui-control"
            @input="emit('update:defenseScore', Number(($event.target as HTMLInputElement).value))"
          />
        </span>
        <span v-if="defenseError" class="ui-field__error">{{ defenseError }}</span>
      </label>
    </div>
  </section>
</template>

<style scoped>
.awd-config-form-section {
  display: grid;
  gap: var(--space-3);
}

.awd-config-card {
  padding: var(--space-4);
  border: 1px solid var(--awd-card-border);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: var(--awd-card-shadow);
}

.awd-config-card--compact {
  gap: var(--space-2);
}

.awd-config-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.awd-config-score-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(2, minmax(0, 12rem));
}

:deep(.ui-field) {
  gap: var(--space-1);
}

:deep(.ui-field__label) {
  font-size: var(--font-size-12);
}

:deep(.ui-control) {
  min-height: 2.5rem;
}

@media (max-width: 767px) {
  .awd-config-score-grid {
    grid-template-columns: 1fr;
  }
}
</style>
