<script setup lang="ts">
import { watch } from 'vue'
import { useRouter } from 'vue-router'

import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'

const props = defineProps<{
  to: AppRouteTarget | null
  replace?: boolean
}>()

const router = useRouter()

watch(
  () => props.to,
  async (to) => {
    if (!to) {
      return
    }
    if (props.replace) {
      await router.replace(to)
      return
    }
    await router.push(to)
  }
)
</script>

<template>
  <span v-if="false" />
</template>
