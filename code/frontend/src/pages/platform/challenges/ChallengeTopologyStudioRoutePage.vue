<script setup lang="ts">
import { ChallengeTopologyStudioPage } from '@/features/challenge-topology-studio'
import { useChallengeTopologyStudioRoutePage } from '@/features/platform/challenges'

const { challengeId, backToChallengeDetail } = useChallengeTopologyStudioRoutePage()
</script>

<template>
  <ChallengeTopologyStudioPage :challenge-id="challengeId" @back="backToChallengeDetail" />
</template>
