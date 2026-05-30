<script setup lang="ts">
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
  <section class="student-directory-filters notification-category-filter" aria-label="消息分类筛选">
    <div class="student-directory-filter-grid notification-category-filter__grid">
      <label class="student-directory-filter-field notification-filter-field">
        <span class="student-directory-filter-label notification-filter-label">分类</span>
        <span class="ui-control-wrap student-directory-filter-control">
          <select
            :value="selectedCategory"
            class="ui-control notification-filter-control"
            @change="handleCategoryChange"
          >
            <option v-for="option in categoryOptions" :key="option.key" :value="option.key">
              {{ option.label }}
            </option>
          </select>
        </span>
      </label>

      <div class="student-directory-filter-actions notification-filter-actions">
        <span
          class="student-directory-filter-label student-directory-filter-label--ghost"
          aria-hidden="true"
        >
          操作
        </span>
        <div class="student-directory-filter-action-row">
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            :disabled="selectedCategory === 'all'"
            @click="emit('selectCategory', 'all')"
          >
            查看全部
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.notification-category-filter__grid {
  grid-template-columns: minmax(12rem, 16rem) auto;
}

.notification-filter-actions {
  justify-items: start;
}
</style>
