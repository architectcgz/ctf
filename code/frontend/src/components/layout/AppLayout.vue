<template>
  <div class="relative min-h-screen overflow-x-hidden bg-base text-text-primary">
    <div class="pointer-events-none absolute inset-0">
      <div class="app-layout-top-glow absolute inset-x-0 top-0 h-44" />
      <div class="app-layout-side-rail absolute inset-y-0 left-0 w-px" />
      <div class="app-layout-bottom-shadow absolute inset-x-0 bottom-0 h-56" />
    </div>
    <div class="relative flex min-h-screen">
      <Sidebar
        :collapsed="sidebarCollapsed"
        :mobile-open="sidebarOpen"
        @close-mobile="sidebarOpen = false"
        @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      />
      <div class="min-w-0 flex flex-1 flex-col">
        <TopNav
          :sidebar-collapsed="sidebarCollapsed"
          :notification-status="notificationStatus"
          @toggle-sidebar="sidebarOpen = true"
          @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
        />
        <main
          class="workspace-main mx-auto w-full"
          :class="[mainShellClass, backofficeMainClass]"
        >
          <div
            class="workspace-page"
            :class="pageShellClass"
          >
            <RouterView v-slot="{ Component, route: resolvedRoute }">
              <Transition
                name="workspace-route"
                mode="out-in"
                appear
              >
                <component
                  :is="Component"
                  :key="resolvedRoute.path"
                  class="workspace-route-root"
                  :class="routeRootClass"
                />
              </Transition>
            </RouterView>
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { RouterView } from 'vue-router'

import Sidebar from '@/components/layout/Sidebar.vue'
import TopNav from '@/components/layout/TopNav.vue'
import { useNotificationRealtime } from '@/composables/useNotificationRealtime'
import { isBackofficeRoute } from '@/utils/backofficeRouteMeta'

const route = useRoute()
const { start, status: notificationStatus } = useNotificationRealtime()
const sidebarCollapsed = ref(false)
const sidebarOpen = ref(false)
const isBackofficeLayout = computed(() => isBackofficeRoute(route.path))
const mainShellClass = computed(() =>
  route.meta.contentLayout === 'bleed' ? 'workspace-main--bleed' : 'workspace-main--default'
)
const backofficeMainClass = computed(() =>
  isBackofficeLayout.value ? 'workspace-main--backoffice' : ''
)
const pageShellClass = computed(() =>
  route.meta.contentLayout === 'bleed' ? 'workspace-page--bleed' : ''
)
const routeRootClass = computed(() =>
  route.meta.contentLayout === 'bleed' ? 'workspace-route-root--bleed' : ''
)

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
.app-layout-top-glow {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary-hover) 12%, transparent),
    transparent
  );
}

.app-layout-side-rail {
  background: linear-gradient(
    180deg,
    transparent,
    color-mix(in srgb, var(--color-text-primary) 8%, transparent),
    transparent
  );
}

.app-layout-bottom-shadow {
  background: linear-gradient(
    0deg,
    color-mix(in srgb, var(--color-shadow-strong) 16%, transparent),
    transparent
  );
}

.workspace-main {
  flex: 1 1 auto;
  position: relative;
  isolation: isolate;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.workspace-main--default {
  max-width: 1600px;
  padding-block: 1.5rem;
  padding-inline: 1rem;
}

.workspace-main--bleed {
  max-width: none;
  padding-block: 0;
  padding-inline: 0;
}

.workspace-main--backoffice {
  max-width: none;
}

.workspace-main--backoffice :deep(.workspace-shell > .workspace-topbar) {
  display: none;
}

.workspace-page {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.workspace-page--bleed {
  margin-inline: 0;
}

.workspace-page--bleed :deep(.workspace-route-root--bleed) {
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.workspace-page--bleed :deep(.dashboard-view.workspace-route-root--bleed > .journal-shell) {
  flex: 1 1 auto;
  min-height: 0;
}

@media (min-width: 768px) {
  .workspace-main--default {
    padding-inline: 1.5rem;
  }
}

@media (min-width: 1280px) {
  .workspace-main--default {
    padding-inline: 2rem;
  }
}

@media (max-width: 767px) {
  .workspace-main {
    min-height: 0;
  }
}
</style>
