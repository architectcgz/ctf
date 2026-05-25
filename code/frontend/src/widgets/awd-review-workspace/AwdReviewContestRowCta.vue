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
  <div class="teacher-directory-row-cta" :class="'workspace-directory-row-btn'">
    <span>{{ label }}</span>
    <ArrowRight class="h-4 w-4" />
  </div>
</template>

<style scoped>
.teacher-directory-row-cta {
  gap: var(--space-1-5);
  color: var(--journal-accent-strong);
}
</style>
