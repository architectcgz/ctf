<template>
  <div class="relative min-h-screen overflow-x-hidden bg-base text-text-primary">
    <div class="pointer-events-none absolute inset-0">
      <div
        class="absolute inset-x-0 top-0 h-44 bg-[linear-gradient(180deg,rgba(8,145,178,0.12),transparent)]"
      />
      <div
        class="absolute inset-y-0 left-0 w-px bg-[linear-gradient(180deg,transparent,rgba(255,255,255,0.08),transparent)]"
      />
      <div
        class="absolute inset-x-0 bottom-0 h-56 bg-[linear-gradient(0deg,rgba(0,0,0,0.16),transparent)]"
      />
    </div>
    <div class="relative flex min-h-screen">
      <Sidebar
        :collapsed="sidebarCollapsed"
        :mobile-open="sidebarOpen"
        @close-mobile="sidebarOpen = false"
        @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      />
      <div class="min-w-0 flex-1">
        <TopNav
          :sidebar-collapsed="sidebarCollapsed"
          :notification-status="notificationStatus"
          @toggle-sidebar="sidebarOpen = true"
          @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
        />
        <main
          class="workspace-main mx-auto w-full max-w-[1600px] px-4 py-6 md:px-6 xl:px-8"
        >
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterView } from 'vue-router'
import { useRoute } from 'vue-router'

import Sidebar from '@/components/layout/Sidebar.vue'
import TopNav from '@/components/layout/TopNav.vue'
import { useNotificationRealtime } from '@/composables/useNotificationRealtime'

const route = useRoute()
const { start, status: notificationStatus } = useNotificationRealtime()
const sidebarCollapsed = ref(false)
const sidebarOpen = ref(false)

onMounted(() => {
  void start()
})

watch(
  () => route.fullPath,
  () => {
    sidebarOpen.value = false
  }
)
</script>

<style scoped>
.workspace-main {
  position: relative;
  isolation: isolate;
  min-height: calc(100vh - 5rem);
}

@media (max-width: 767px) {
  .workspace-main {
    min-height: auto;
  }
}
</style>
