<script setup lang="ts">
import { ArrowLeft, RefreshCw } from 'lucide-vue-next'

defineProps<{
  contestTitle: string
  serviceName: string
  refreshing: boolean
}>()

const emit = defineEmits<{
  back: []
  refresh: []
}>()
</script>

<template>
  <header class="awd-config-page__topbar">
    <button type="button" class="ui-btn ui-btn--ghost" @click="emit('back')">
      <ArrowLeft class="h-4 w-4" />
      返回工作台
    </button>
    <div class="awd-config-page__topbar-main">
      <h1 class="awd-config-page__title">AWD 服务配置</h1>
      <div class="awd-config-page__crumbs">
        <span class="awd-config-page__crumb awd-config-page__crumb--muted">{{ contestTitle }}</span>
        <span class="awd-config-page__crumb-separator">/</span>
        <span class="awd-config-page__crumb awd-config-page__crumb--active">
          {{ serviceName }}
        </span>
      </div>
    </div>
    <button type="button" class="ui-btn ui-btn--secondary" :disabled="refreshing" @click="emit('refresh')">
      <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': refreshing }" />
      刷新
    </button>
  </header>
</template>

<style scoped>
.awd-config-page__topbar {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-4);
  min-height: 3.5rem;
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: var(--color-bg-surface);
}

.awd-config-page__topbar-main {
  min-width: 0;
  display: grid;
  gap: var(--space-1);
}

.awd-config-page__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-16);
  font-weight: 800;
}

.awd-config-page__crumbs {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.awd-config-page__crumb,
.awd-config-page__crumb-separator {
  min-width: 0;
}

.awd-config-page__crumb--muted,
.awd-config-page__crumb--active {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.awd-config-page__crumb--active {
  color: var(--color-text-primary);
  font-weight: 700;
}

@media (max-width: 767px) {
  .awd-config-page__topbar {
    grid-template-columns: 1fr;
    align-items: stretch;
  }
}
</style>
