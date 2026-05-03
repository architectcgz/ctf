<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    label?: string
  }>(),
  {
    label: '进入复盘',
  }
)
</script>

<template>
  <div class="teacher-directory-row-cta">
    <span>{{ label }}</span>
    <ArrowRight class="h-4 w-4" />
  </div>
</template>

<style scoped>
.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  color: var(--journal-accent-strong);
  font-size: var(--font-size-0-82);
  font-weight: 700;
}
</style>
