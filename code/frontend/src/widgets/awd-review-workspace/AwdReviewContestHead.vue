<script setup lang="ts">
import { AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA } from './model/directory'
</script>

<template>
  <div class="teacher-directory-head" :class="'workspace-directory-grid-head'">
    <span
      v-for="column in AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA"
      :key="column.key"
      :class="['teacher-directory-head-cell', column.headClass]"
    >
      {{ column.label }}
    </span>
  </div>
</template>

<style scoped>
.teacher-directory-head {
  grid-template-columns: var(--awd-review-directory-columns);
}

@media (max-width: 1080px) {
  .teacher-directory-head {
    display: none;
  }
}
</style>
