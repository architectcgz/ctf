<script setup lang="ts">
import { ChevronLeft, Settings } from 'lucide-vue-next'

defineProps<{
  contestTitle: string
}>()

const emit = defineEmits<{
  (event: 'back'): void
  (event: 'open-studio'): void
}>()

function handleBack(): void {
  emit('back')
}

function handleOpenStudio(): void {
  emit('open-studio')
}
</script>

<template>
  <header class="ops-topbar">
    <div class="topbar-left">
      <button
        type="button"
        class="icon-btn"
        title="退出"
        @click="handleBack"
      >
        <ChevronLeft class="h-5 w-5" />
      </button>
      <div class="divider" />
      <div class="brand">
        <div class="brand-overline">
          Command Center
        </div>
        <h1 class="brand-title">
          {{ contestTitle }}
        </h1>
      </div>
    </div>

    <div class="topbar-right">
      <button
        type="button"
        class="studio-link-btn"
        @click="handleOpenStudio"
      >
        <Settings class="h-4 w-4" />
        <span>进入竞赛工作室</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.ops-topbar {
  height: 4.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-8);
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--color-border-default);
  z-index: 50;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: var(--space-6);
}

.divider {
  width: 1px;
  height: 1.5rem;
  background: var(--color-border-default);
}

.brand-overline {
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-primary);
  letter-spacing: 0.15em;
}

.brand-title {
  font-size: var(--font-size-1-00);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: 0;
}

.studio-link-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2-5);
  height: 2.5rem;
  padding: 0 var(--space-5);
  border-radius: 0.85rem;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.studio-link-btn:hover {
  background: var(--color-bg-surface);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.icon-btn {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.6rem;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all 0.2s ease;
  background: transparent;
  border: none;
}

.icon-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}
</style>
