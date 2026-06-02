<template>
  <footer class="panel-footer">
    <button type="button" class="view-all-btn" @click="$emit('viewAll')">
      <span class="footer-icon">
        <AlignLeft class="footer-icon__glyph" />
      </span>
      <span class="view-all-btn__label"> 查看全部通知 </span>
      <ChevronRight class="arrow-icon" />
    </button>
  </footer>
</template>

<script setup lang="ts">
import { AlignLeft, ChevronRight } from 'lucide-vue-next'

defineEmits<{
  viewAll: []
}>()
</script>

<style scoped>
.panel-footer {
  position: relative;
  flex: 0 0 auto;
  margin-top: auto;
  width: 100%;
  padding: var(--space-3) 0 calc(var(--space-3) + env(safe-area-inset-bottom, 0px));
  z-index: 1;
  border-top: 1px solid var(--notification-panel-edge-soft);
  background-color: var(--color-bg-surface);
  background-image: var(--notification-footer-bg);
  box-shadow: none;
}

:global([data-theme='dark']) .panel-footer {
  background-color: var(--color-bg-surface);
}

.view-all-btn {
  width: 100%;
  min-height: var(--ui-control-height-lg);
  display: grid;
  grid-template-columns: 2.25rem minmax(0, 1fr) 1.75rem;
  align-items: center;
  column-gap: var(--space-3);
  padding: 0 1.75rem 0 1.5rem;
  border: 0;
  background-color: var(--color-bg-surface);
  background-image: var(--notification-footer-bg);
  color: var(--notification-footer-text);
  cursor: pointer;
  text-align: left;
}

:global([data-theme='dark']) .view-all-btn {
  background-color: var(--color-bg-surface);
}

.view-all-btn:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 46%, var(--color-border-default));
  outline-offset: var(--space-1);
}

.footer-icon {
  width: 2.125rem;
  height: 2.125rem;
  border-radius: 50%;
  color: var(--notification-footer-text);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--notification-footer-icon-bg);
  border: 1px solid var(--notification-footer-icon-border);
}

.footer-icon__glyph {
  width: var(--space-4);
  height: var(--space-4);
}

.view-all-btn__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  line-height: 1;
  font-weight: 500;
  letter-spacing: 0.00625rem;
}

.arrow-icon {
  justify-self: end;
  width: var(--space-5);
  height: var(--space-5);
  color: var(--notification-footer-text);
}
</style>
