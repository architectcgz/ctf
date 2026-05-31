<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'

defineOptions({
  inheritAttrs: false,
})

defineProps<{
  to: AppRouteTarget
}>()
</script>

<template>
  <RouterLink :to="to" v-bind="$attrs">
    <slot />
  </RouterLink>
</template>
