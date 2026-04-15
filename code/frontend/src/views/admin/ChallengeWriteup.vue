<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

import ChallengeWriteupEditorPage from '@/components/admin/writeup/ChallengeWriteupEditorPage.vue'

const route = useRoute()
const router = useRouter()
</script>

<template>
  <ChallengeWriteupEditorPage
    :challenge-id="String(route.params.id || '')"
    @back="
      router.push({
        name: 'AdminChallengeDetail',
        params: { id: String(route.params.id || '') },
        query: { panel: 'writeup' },
      })
    "
  />
</template>
