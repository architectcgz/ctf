<template>
  <PlatformContestOperationsPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestOperationsPage } from '@/features/platform/contest-operations'

defineProps<{
  contestId: string
}>()
</script>
