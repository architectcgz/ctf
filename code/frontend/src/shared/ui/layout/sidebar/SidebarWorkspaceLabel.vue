<template>
  <div
    class="backoffice-sidebar__workspace px-6 py-5 overflow-hidden whitespace-nowrap transition-all duration-200"
    :class="workspaceClass"
  >
    <span class="backoffice-sidebar__workspace-label font-black uppercase tracking-widest">
      Workspace
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  collapsed: boolean
  mobile: boolean
}>()

const workspaceClass = computed(() => {
  if (props.mobile) {
    return ''
  }
  return props.collapsed ? 'opacity-0 h-0 p-0' : 'opacity-100 h-14'
})
</script>
