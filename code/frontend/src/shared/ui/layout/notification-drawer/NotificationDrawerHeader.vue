<template>
  <header class="panel-header">
    <div class="title-wrap">
      <div class="bell-wrap">
        <BellRing class="bell-icon" />
        <span v-if="hasUnread" class="bell-dot" />
      </div>

      <div>
        <h1>NOTIFICATIONS</h1>
        <p>通知中心</p>
      </div>
    </div>

    <button type="button" class="close-btn" aria-label="关闭抽屉" @click="$emit('close')">
      <X class="close-btn__icon" />
    </button>
  </header>
</template>

<script setup lang="ts">
import { BellRing, X } from 'lucide-vue-next'

defineProps<{
  hasUnread: boolean
}>()

defineEmits<{
  close: []
}>()
</script>

<style scoped>
.panel-header {
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  min-height: 3rem;
}

.title-wrap {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.bell-wrap {
  position: relative;
  width: 2rem;
  height: 2rem;
  color: var(--notification-icon-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.bell-icon {
  width: var(--space-5);
  height: var(--space-5);
  filter: drop-shadow(
    0 0 0.4375rem color-mix(in srgb, var(--color-primary) 16%, var(--color-bg-surface))
  );
}

.bell-dot {
  position: absolute;
  top: var(--space-0-5);
  right: var(--space-1);
  width: var(--space-1-5);
  height: var(--space-1-5);
  border-radius: 50%;
  background: var(--notification-signal);
  box-shadow: 0 0 0.875rem color-mix(in srgb, var(--notification-signal) 70%, transparent);
}

.panel-header h1 {
  margin: 0;
  color: var(--notification-title);
  font-size: var(--font-size-16);
  line-height: 1.18;
  font-weight: 500;
  letter-spacing: 0.14rem;
}

.panel-header p {
  margin: var(--space-2) 0 0;
  color: var(--notification-panel-muted);
  font-size: var(--font-size-12);
  line-height: 1.2;
  letter-spacing: 0.0125rem;
}

.close-btn {
  position: relative;
  top: 0.1875rem;
  width: 1.875rem;
  height: 1.875rem;
  padding: 0;
  color: var(--notification-panel-muted);
  border: none;
  border-radius: 0.75rem;
  background: transparent;
  cursor: pointer;
}

.close-btn__icon {
  width: var(--space-4-5);
  height: var(--space-4-5);
}

.close-btn:hover,
.close-btn:focus-visible {
  background: color-mix(in srgb, var(--notification-panel-text) 7%, transparent);
}

.close-btn:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 46%, var(--color-border-default));
  outline-offset: var(--space-1);
}
</style>
