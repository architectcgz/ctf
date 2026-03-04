<template>
  <div class="app-root">
    <RouterView />
    <AppToast />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import AppToast from '@/components/common/AppToast.vue'
import { useTheme } from '@/composables/useTheme'

const { initTheme } = useTheme()

onMounted(() => {
  initTheme()
})
</script>

<style scoped>
.app-root {
  min-height: 100vh;
  background-color: var(--color-bg-base);
  color: var(--color-text-primary);
}
</style>
