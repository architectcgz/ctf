<template>
  <div class="animate-pulse space-y-3">
    <div v-for="i in rows" :key="i" class="h-4 rounded bg-elevated" :style="{ width: width(i) }" />
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    rows?: number
  }>(),
  {
    rows: 6,
  }
)

function width(i: number): string {
  if (i % 3 === 0) return '68%'
  if (i % 3 === 1) return '92%'
  return '80%'
}
</script>
