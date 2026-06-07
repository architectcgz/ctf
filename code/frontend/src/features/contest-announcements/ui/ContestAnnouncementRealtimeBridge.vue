<script setup lang="ts">
import { onMounted } from 'vue'

import { useAnnouncementSubscription } from '../model'

const props = defineProps<{
  contestId: string
}>()

const emit = defineEmits<{
  updated: []
}>()

const { start } = useAnnouncementSubscription(props.contestId, () => {
  emit('updated')
})

onMounted(() => {
  void start()
})
</script>

<template></template>
