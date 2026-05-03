<script setup lang="ts">
import ChallengeWriteupEditorPage from '@/components/platform/writeup/ChallengeWriteupEditorPage.vue'
import { useChallengeWriteupPage } from '@/features/platform-challenges'

const { challengeId, backToChallengeDetail } = useChallengeWriteupPage()
</script>

<template>
  <ChallengeWriteupEditorPage
    :challenge-id="challengeId"
    @back="backToChallengeDetail"
  />
</template>
