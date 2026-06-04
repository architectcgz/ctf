<template>
  <section class="student-insight-timeline-section">
    <TrainingTimelineContent
      :timeline="timeline"
      :loading="loading"
      class="student-insight-timeline-content"
    />
  </section>
</template>

<style scoped>
.student-insight-timeline-content {
  min-width: 0;
}
</style>

<script setup lang="ts">
import type { TimelineEvent } from '@/api/contracts'
import { TrainingTimelineContent } from '@/entities/training-timeline'

defineProps<{
  timeline: TimelineEvent[]
  loading?: boolean
}>()
</script>
