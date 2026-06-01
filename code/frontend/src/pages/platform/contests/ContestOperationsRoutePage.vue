<template>
  <PlatformContestOperationsPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestOperationsPage } from '@/features/platform/contests'

defineProps<{
  contestId: string
}>()
</script>
