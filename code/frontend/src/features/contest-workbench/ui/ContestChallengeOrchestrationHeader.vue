<script setup lang="ts">
import { Plus, RefreshCw } from 'lucide-vue-next'

defineProps<{
  isAwdContest: boolean
  panelCopy: string
  loading: boolean
}>()

const emit = defineEmits<{
  refresh: []
  create: []
}>()
</script>

<template>
  <header class="studio-pane-header">
    <div class="header-main">
      <h1 class="pane-title">
        题目编排
      </h1>
      <p class="pane-description">
        {{ panelCopy }}
      </p>
    </div>
    <div class="header-actions">
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('refresh')"
      >
        <RefreshCw
          class="h-3.5 w-3.5"
          :class="{ 'animate-spin': loading }"
        />
        <span>同步数据</span>
      </button>
      <button
        id="contest-challenge-add"
        type="button"
        class="ui-btn ui-btn--primary"
        @click="emit('create')"
      >
        <Plus class="h-3.5 w-3.5" />
        <span>{{ isAwdContest ? '新增服务' : '关联新题目' }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.studio-pane-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
}

.header-main {
  min-width: 0;
}

.pane-title {
  margin: 0;
  font-size: var(--font-size-20);
  font-weight: 900;
  color: var(--color-text-primary);
}

.pane-description {
  margin: var(--space-2) 0 0;
  max-width: var(--ui-selector-width-lg);
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

@media (max-width: 720px) {
  .studio-pane-header {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
