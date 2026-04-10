<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

import ChallengeWriteupViewPage from '@/components/admin/writeup/ChallengeWriteupViewPage.vue'

const route = useRoute()
const router = useRouter()
</script>

<template>
  <ChallengeWriteupViewPage
    :challenge-id="String(route.params.id || '')"
    @back="router.push({ name: 'AdminChallengeDetail', params: { id: String(route.params.id || '') }, query: { panel: 'writeup' } })"
    @edit="router.push({ name: 'AdminChallengeWriteup', params: { id: String(route.params.id || '') } })"
  />
</template>
