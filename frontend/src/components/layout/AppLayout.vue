<template>
  <div class="min-h-screen bg-base text-text-primary">
    <div class="flex">
      <Sidebar class="hidden md:block" />
      <div class="min-w-0 flex-1">
        <TopNav />
        <main class="mx-auto max-w-7xl px-4 py-6 md:px-6">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'

import Sidebar from '@/components/layout/Sidebar.vue'
import TopNav from '@/components/layout/TopNav.vue'
</script>

