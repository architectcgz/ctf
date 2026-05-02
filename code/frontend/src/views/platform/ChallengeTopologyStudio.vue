<script setup lang="ts">
import ChallengeTopologyStudioPage from '@/components/platform/topology/ChallengeTopologyStudioPage.vue'
import { useChallengeTopologyStudioRoutePage } from '@/features/platform-challenges'

const { challengeId, backToChallengeDetail } = useChallengeTopologyStudioRoutePage()
</script>

<template>
  <ChallengeTopologyStudioPage :challenge-id="challengeId" @back="backToChallengeDetail" />
</template>
