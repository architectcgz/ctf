<template>
  <section class="summary-row">
    <div class="summary-main">
      <span v-if="hasUnread" class="summary-number">
        {{ unreadCount }}
      </span>
      <span class="summary-text">{{ drawerSummary }}</span>
    </div>

    <nav class="summary-actions" aria-label="通知操作">
      <button
        type="button"
        class="text-action"
        :disabled="!hasUnread || isMarkingAllRead"
        :aria-busy="isMarkingAllRead ? 'true' : 'false'"
        @click="$emit('markAllRead')"
      >
        全部设为已读
      </button>
    </nav>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  hasUnread: boolean
  unreadCount: number
  drawerSummary: string
  isMarkingAllRead: boolean
}>()

defineEmits<{
  markAllRead: []
}>()
</script>

<style scoped>
.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.375rem;
  margin-top: var(--space-8);
  min-height: 2rem;
}

.summary-main {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  min-width: 0;
  white-space: nowrap;
}

.summary-number {
  color: var(--notification-signal);
  font-size: var(--font-size-26);
  line-height: 1;
  font-weight: 300;
  letter-spacing: -0.0625rem;
  text-shadow: 0 0 1.125rem color-mix(in srgb, var(--notification-signal) 32%, transparent);
}

.summary-text {
  min-width: 0;
  color: var(--notification-summary);
  font-size: var(--font-size-13);
  line-height: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
}

.summary-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  white-space: nowrap;
}

.text-action {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--notification-action);
  font-size: var(--font-size-13);
  font-weight: 500;
  cursor: pointer;
}

.text-action:hover {
  color: color-mix(in srgb, var(--notification-action) 86%, var(--notification-panel-text));
}

.text-action:disabled {
  cursor: default;
  opacity: 0.54;
}

.text-action:focus-visible {
  outline: var(--ui-focus-ring-width) solid rgb(114 184 255 / 0.46);
  outline-offset: var(--space-1);
}

@media (max-width: 768px) {
  .summary-row {
    align-items: flex-start;
    flex-direction: column;
    margin-top: var(--space-7);
  }
}
</style>
