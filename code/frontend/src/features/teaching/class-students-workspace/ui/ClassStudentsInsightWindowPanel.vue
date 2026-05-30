<script setup lang="ts">
defineProps<{
  insightWindowFromDate: string
  insightWindowToDate: string
  insightWindowError: string | null
  insightWindowLabel: string
  canApplyInsightWindow: boolean
  canResetInsightWindow: boolean
}>()

const emit = defineEmits<{
  updateInsightWindowFromDate: [value: string]
  updateInsightWindowToDate: [value: string]
  applyInsightWindow: []
  resetInsightWindow: []
}>()
</script>

<template>
  <section class="teacher-window-shell" aria-label="班级训练时间段">
    <div class="teacher-window-head">
      <div class="teacher-window-copy">
        <div class="teacher-summary-title">
          <span>Training Window</span>
        </div>
        <h2 class="teacher-window-title">班级训练时间段</h2>
        <p class="teacher-window-description">
          当前统计窗口：{{ insightWindowLabel }}
        </p>
      </div>
      <span class="teacher-surface-chip">
        {{ insightWindowLabel }}
      </span>
    </div>

    <div class="teacher-window-grid">
      <label class="teacher-field">
        <span class="teacher-field-label">开始日期</span>
        <div class="teacher-field-control teacher-filter-control">
          <input
            :value="insightWindowFromDate"
            type="date"
            class="teacher-input"
            @input="emit('updateInsightWindowFromDate', ($event.target as HTMLInputElement).value)"
          >
        </div>
      </label>

      <label class="teacher-field">
        <span class="teacher-field-label">结束日期</span>
        <div class="teacher-field-control teacher-filter-control">
          <input
            :value="insightWindowToDate"
            type="date"
            class="teacher-input"
            @input="emit('updateInsightWindowToDate', ($event.target as HTMLInputElement).value)"
          >
        </div>
      </label>

      <div class="teacher-window-actions">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          :disabled="!canResetInsightWindow"
          @click="emit('resetInsightWindow')"
        >
          恢复默认
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="!canApplyInsightWindow"
          @click="emit('applyInsightWindow')"
        >
          应用时间段
        </button>
      </div>
    </div>

    <p v-if="insightWindowError" class="teacher-window-error" role="alert">
      {{ insightWindowError }}
    </p>
  </section>
</template>

<style scoped>
.teacher-window-shell {
  display: grid;
  gap: var(--space-4);
  margin-bottom: var(--space-5);
  padding: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: var(--radius-2xl);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 7%, transparent),
      transparent 36%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 78%, var(--color-bg-base))
    );
}

.teacher-window-head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: start;
}

.teacher-window-copy {
  display: grid;
  gap: var(--space-2);
}

.teacher-window-title {
  margin: 0;
  font-size: var(--font-size-1-20);
  line-height: 1.2;
}

.teacher-window-description {
  margin: 0;
  color: var(--color-text-secondary);
}

.teacher-window-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
  gap: var(--space-4);
  align-items: end;
}

.teacher-window-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  justify-content: flex-end;
}

.teacher-window-error {
  margin: 0;
  font-size: var(--font-size-0-82);
  color: var(--color-danger);
}

@media (max-width: 1080px) {
  .teacher-window-grid {
    grid-template-columns: 1fr;
  }

  .teacher-window-actions {
    justify-content: flex-start;
  }
}
</style>
