<script setup lang="ts">
import { Code2, ShieldCheck } from 'lucide-vue-next'

defineProps<{
  displayName: string
  title: string
  protocolLabel: string
  checkerTypeLabel: string
}>()
</script>

<template>
  <header class="awd-config-editor-head">
    <div class="awd-config-editor-head__main">
      <h2>{{ displayName }}</h2>
      <p>{{ title }}</p>
    </div>
    <div class="awd-config-editor-head__meta">
      <span class="engine-lock">
        <ShieldCheck class="h-4 w-4" />
        {{ protocolLabel }}
      </span>
      <span class="engine-lock">
        <Code2 class="h-4 w-4" />
        {{ checkerTypeLabel }}
      </span>
    </div>
  </header>
</template>

<style scoped>
.awd-config-editor-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--awd-card-border);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: var(--awd-card-shadow);
}

.awd-config-editor-head__main {
  min-width: 0;
}

.awd-config-editor-head h2 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-20);
  font-weight: 800;
  line-height: 1.25;
  word-break: break-word;
}

.awd-config-editor-head p {
  margin: var(--space-1) 0 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  line-height: 1.4;
}

.awd-config-editor-head__meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.engine-lock {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: var(--ui-control-radius);
  background: var(--color-bg-surface);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
}

@media (max-width: 767px) {
  .awd-config-editor-head {
    grid-template-columns: 1fr;
  }
}
</style>
