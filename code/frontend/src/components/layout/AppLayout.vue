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
          class="workspace-main decard-content mx-auto w-full max-w-[1600px] px-4 py-6 md:px-6 xl:px-8"
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
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 68%, transparent);
  border-radius: 0;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base)),
    color-mix(in srgb, var(--color-bg-base) 100%, var(--color-bg-surface))
  );
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.04);
}

.workspace-main > :deep(*) {
  position: relative;
  min-height: calc(100vh - 7rem);
  padding: 1rem;
}

:global([data-theme='light']) .workspace-main {
  border-color: rgba(226, 232, 240, 0.78);
  background: linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 10px 24px rgba(148, 163, 184, 0.1);
}

@media (max-width: 767px) {
  .workspace-main {
    min-height: auto;
  }

  .workspace-main > :deep(*) {
    min-height: auto;
    padding: 0.75rem;
  }
}
</style>
