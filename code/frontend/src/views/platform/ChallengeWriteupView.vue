<script setup lang="ts">
import ChallengeWriteupViewPage from '@/components/platform/writeup/ChallengeWriteupViewPage.vue'
import { useChallengeWriteupViewPage } from '@/features/platform-challenges'

const { challengeId, backToChallengeDetail, goToWriteupEditor } = useChallengeWriteupViewPage()
</script>

<template>
  <ChallengeWriteupViewPage
    :challenge-id="challengeId"
    @back="backToChallengeDetail"
    @edit="goToWriteupEditor"
  />
</template>
