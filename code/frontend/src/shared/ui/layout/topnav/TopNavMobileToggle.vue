<template>
  <button
    type="button"
    class="topnav-icon-button"
    :aria-label="sidebarCollapsed ? '展开导航' : '折叠导航'"
    @click="$emit('toggle')"
  >
    <Menu class="h-4 w-4 md:hidden" />
    <PanelLeftClose
      v-if="!sidebarCollapsed"
      class="hidden h-4 w-4 md:block"
    />
    <PanelLeftOpen
      v-else
      class="hidden h-4 w-4 md:block"
    />
  </button>
</template>

<script setup lang="ts">
import { Menu, PanelLeftClose, PanelLeftOpen } from 'lucide-vue-next'

defineProps<{
  sidebarCollapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()
</script>
