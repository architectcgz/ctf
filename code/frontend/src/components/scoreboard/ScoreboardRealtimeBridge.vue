<script setup lang="ts">
import { onMounted } from 'vue'

import { useContestScoreboardRealtime } from '@/composables/useContestScoreboardRealtime'

const props = defineProps<{
  contestId: string
}>()

const emit = defineEmits<{
  updated: []
}>()

const { start } = useContestScoreboardRealtime(props.contestId, () => {
  emit('updated')
})

onMounted(() => {
  void start()
})
</script>

<template></template>
