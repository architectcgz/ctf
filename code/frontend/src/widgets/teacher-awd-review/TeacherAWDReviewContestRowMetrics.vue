<script setup lang="ts">
defineProps<{
  primary: string
  secondary: string
}>()
</script>

<template>
  <div class="teacher-directory-row-metrics" :class="'workspace-directory-compact-text'">
    <span>{{ primary }}</span>
    <span>{{ secondary }}</span>
  </div>
</template>

<style scoped>
.teacher-directory-row-metrics {
  display: grid;
  gap: var(--space-1);
}
</style>
