<script setup lang="ts">
import { onMounted } from 'vue'

import { useContestAnnouncementRealtime } from '@/composables/useContestAnnouncementRealtime'

const props = defineProps<{
  contestId: string
}>()

const emit = defineEmits<{
  updated: []
}>()

const { start } = useContestAnnouncementRealtime(props.contestId, () => {
  emit('updated')
})

onMounted(() => {
  void start()
})
</script>

<template></template>
