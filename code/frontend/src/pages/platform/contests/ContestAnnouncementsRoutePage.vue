<template>
  <PlatformContestAnnouncementsPage :contest-id="contestId" />
</template>

<script setup lang="ts">
import { PlatformContestAnnouncementsPage } from '@/features/platform/contest-announcements'

defineProps<{
  contestId: string
}>()
</script>
