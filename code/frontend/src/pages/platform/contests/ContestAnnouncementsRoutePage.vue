<template>
  <PlatformContestAnnouncementsPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestAnnouncementsPage } from '@/features/platform/contests'

defineProps<{
  contestId: string
}>()
</script>
