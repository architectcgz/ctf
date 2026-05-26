<script setup lang="ts">
import { Plus, RefreshCw } from 'lucide-vue-next'

import PageHeader from '@/components/common/PageHeader.vue'

defineProps<{
  eyebrow: string
  title: string
  description: string
}>()

const emit = defineEmits<{
  reset: []
  refresh: []
}>()
</script>

<template>
  <div class="topology-page-header">
    <PageHeader
      :eyebrow="eyebrow"
      :title="title"
      :description="description"
    >
      <div class="template-library-toolbar">
        <button
          type="button"
          class="topology-toolbar-btn topology-toolbar-btn--primary"
          @click="emit('reset')"
        >
          <Plus class="h-4 w-4" />
          新建空白模板
        </button>
        <button
          type="button"
          class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
      </div>
    </PageHeader>
  </div>
</template>

<style scoped>
.template-library-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.topology-toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.45rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease;
}

.topology-toolbar-btn--ghost {
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-toolbar-btn--ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  color: var(--journal-accent);
}

.topology-toolbar-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
}

.topology-toolbar-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
}
</style>
