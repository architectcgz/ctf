<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import { RouterLink } from 'vue-router'

defineOptions({
  inheritAttrs: false,
})

defineProps<{
  to: RouteLocationRaw
}>()
</script>

<template>
  <RouterLink :to="to" v-bind="$attrs">
    <slot />
  </RouterLink>
</template>
