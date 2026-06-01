<template>
  <PlatformContestEditPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestEditPage } from '@/features/platform/contests'

defineProps<{
  contestId: string
}>()
</script>
