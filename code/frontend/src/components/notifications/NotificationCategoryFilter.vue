<script setup lang="ts">
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import type { NotificationType } from '@/api/contracts'

defineProps<{
  total: number
  selectedCategory: NotificationType | 'all'
  selectedCategoryLabel: string
  categoryOptions: Array<{ key: NotificationType | 'all'; label: string }>
}>()

const emit = defineEmits<{
  selectCategory: [value: NotificationType | 'all']
}>()

function handleCategoryChange(event: Event): void {
  const target = event.target
  const next = target instanceof HTMLSelectElement ? target.value : 'all'
  emit('selectCategory', next as NotificationType | 'all')
}
</script>

<template>
  <WorkspaceDirectoryToolbar
    model-value=""
    :total="total"
    selected-sort-label=""
    :sort-options="[]"
    :show-search="false"
    :show-total="false"
    :filter-button-label="`分类：${selectedCategoryLabel}`"
    filter-panel-title="消息分类"
    reset-label="查看全部"
    :reset-disabled="selectedCategory === 'all'"
    @reset-filters="emit('selectCategory', 'all')"
  >
    <template #filter-panel>
      <label class="notification-filter-field">
        <span class="notification-filter-label">分类</span>
        <select
          :value="selectedCategory"
          class="workspace-directory-filter-control notification-filter-control"
          @change="handleCategoryChange"
        >
          <option
            v-for="option in categoryOptions"
            :key="option.key"
            :value="option.key"
          >
            {{ option.label }}
          </option>
        </select>
      </label>
    </template>
  </WorkspaceDirectoryToolbar>
</template>

<style scoped>
.notification-filter-field {
  display: grid;
  gap: var(--space-2);
}

.notification-filter-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

</style>
