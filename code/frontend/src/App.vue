<template>
  <div class="min-h-screen bg-base text-text-primary">
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
