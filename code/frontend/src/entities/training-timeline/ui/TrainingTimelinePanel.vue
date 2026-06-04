<template>
  <section
    class="journal-soft-surface journal-shell journal-hero timeline-shell-flat flex min-h-full flex-1 flex-col"
  >
    <TrainingTimelineContent
      :timeline="timeline"
      :page-size="pageSize"
      class="px-6 pt-0 pb-6 md:px-8"
    />
  </section>
</template>

<style scoped>
.timeline-shell-flat.journal-shell.journal-hero {
  border: 0;
  border-radius: 0 !important;
  box-shadow: none;
  overflow: visible;
}
</style>

<script setup lang="ts">
import type { TimelineEvent } from '../model'
import TrainingTimelineContent from './TrainingTimelineContent.vue'

const props = withDefaults(
  defineProps<{
    timeline: TimelineEvent[]
    pageSize?: number
  }>(),
  {
    pageSize: 10,
  }
)
</script>
