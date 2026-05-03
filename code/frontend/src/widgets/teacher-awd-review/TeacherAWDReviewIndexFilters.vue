<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'

type ContestStatusOption = {
  value: '' | TeacherAWDReviewContestItemData['status']
  label: string
}

defineProps<{
  statusOptions: readonly ContestStatusOption[]
  statusFilter: '' | TeacherAWDReviewContestItemData['status']
  keywordFilter: string
}>()

const emit = defineEmits<{
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()
</script>

<template>
  <section
    class="teacher-directory-filters"
    aria-label="赛事过滤"
  >
    <div class="awd-review-filter-grid">
      <label class="awd-review-field">
        <span class="awd-review-field__label">赛事状态</span>
        <select
          :value="statusFilter"
          class="awd-review-field__control"
          @change="emit('updateStatusFilter', ($event.target as HTMLSelectElement).value as '' | TeacherAWDReviewContestItemData['status'])"
        >
          <option
            v-for="option in statusOptions"
            :key="option.value || 'all'"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
      </label>

      <label class="awd-review-field awd-review-field--wide">
        <span class="awd-review-field__label">关键词</span>
        <input
          :value="keywordFilter"
          type="text"
          class="awd-review-field__control"
          placeholder="搜索赛事标题"
          @input="emit('updateKeywordFilter', ($event.target as HTMLInputElement).value)"
        >
      </label>
    </div>
  </section>
</template>

<style scoped>
.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
  padding: var(--workspace-directory-gap-top) 0 var(--space-4);
}

.awd-review-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(14rem, 16rem) minmax(16rem, 1fr);
}

.awd-review-field {
  display: grid;
  gap: var(--space-2);
}

.awd-review-field__label {
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-muted);
}

.awd-review-field__control {
  min-height: 2.8rem;
  width: 100%;
  border: 1px solid var(--teacher-control-border);
  border-radius: 16px;
  background: var(--journal-surface);
  padding: 0 0.95rem;
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 160ms ease,
    box-shadow 160ms ease;
}

.awd-review-field__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 46%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.awd-review-field__control::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 80%, transparent);
}

@media (max-width: 1080px) {
  .awd-review-filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
