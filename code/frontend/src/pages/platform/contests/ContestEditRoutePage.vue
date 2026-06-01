<template>
  <PlatformContestEditPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestEditPage } from '@/features/platform/contest-manage'

defineProps<{
  contestId: string
}>()
</script>
