<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

import ChallengeWriteupViewPage from '@/components/platform/writeup/ChallengeWriteupViewPage.vue'

const route = useRoute()
const router = useRouter()
</script>

<template>
  <ChallengeWriteupViewPage
    :challenge-id="String(route.params.id || '')"
    @back="
      router.push({
        name: 'PlatformChallengeDetail',
        params: { id: String(route.params.id || '') },
        query: { panel: 'writeup' },
      })
    "
    @edit="
      router.push({ name: 'PlatformChallengeWriteup', params: { id: String(route.params.id || '') } })
    "
  />
</template>
