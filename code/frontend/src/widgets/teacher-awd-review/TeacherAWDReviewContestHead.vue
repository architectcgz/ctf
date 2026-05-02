<script setup lang="ts">
import { AWD_REVIEW_DIRECTORY_HEADERS } from './model/directory'
</script>

<template>
  <div class="teacher-directory-head">
    <span
      v-for="(label, index) in AWD_REVIEW_DIRECTORY_HEADERS"
      :key="label"
      :class="index === 0 ? 'teacher-directory-head-cell teacher-directory-head-cell-code' : index === 1 ? 'teacher-directory-head-cell teacher-directory-head-cell-name' : ''"
    >
      {{ label }}
    </span>
  </div>
</template>

<style scoped>
.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--awd-review-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px dashed var(--teacher-divider);
  color: var(--journal-muted);
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

@media (max-width: 1080px) {
  .teacher-directory-head {
    display: none;
  }
}
</style>
