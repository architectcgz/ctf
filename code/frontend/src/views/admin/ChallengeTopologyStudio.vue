<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

import ChallengeTopologyStudioPage from '@/components/admin/topology/ChallengeTopologyStudioPage.vue'

const route = useRoute()
const router = useRouter()
</script>

<template>
  <ChallengeTopologyStudioPage
    :challenge-id="String(route.params.id || '')"
    @back="router.push({ name: 'AdminChallengeDetail', params: { id: String(route.params.id || '') } })"
  />
</template>
