<script setup lang="ts">
import AppCard from '@/components/common/AppCard.vue'

defineProps<{
  title?: string
  subtitle?: string
}>()
</script>

<template>
  <AppCard variant="panel" :title="title" :subtitle="subtitle">
    <template #header>
      <slot name="header" />
    </template>
    <slot />
  </AppCard>
</template>
