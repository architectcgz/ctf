<script setup lang="ts">
defineProps<{
  primary: string
  secondary: string
}>()
</script>

<template>
  <div class="teacher-directory-row-metrics">
    <span>{{ primary }}</span>
    <span>{{ secondary }}</span>
  </div>
</template>

<style scoped>
.teacher-directory-row-metrics {
  display: grid;
  gap: var(--space-1);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}
</style>
