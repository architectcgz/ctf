<template>
  <div class="app-root">
    <RouterView v-slot="{ Component, route: resolvedRoute }">
      <Transition
        name="app-route"
        mode="out-in"
        appear
      >
        <component
          :is="Component"
          :key="resolvedRoute.matched[0]?.path || resolvedRoute.path"
        />
      </Transition>
    </RouterView>
    <AppToast />
    <AppDestructiveConfirm />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import AppDestructiveConfirm from '@/components/common/AppDestructiveConfirm.vue'
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
