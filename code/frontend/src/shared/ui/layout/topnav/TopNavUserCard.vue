<template>
  <div
    class="topnav-user-card flex items-center gap-3 px-2.5 py-1.5 sm:px-3"
    :class="{ 'topnav-user-card--admin': true }"
  >
    <div class="topnav-user-mark">
      {{ userInitial }}
    </div>
    <div class="topnav-user-identity hidden min-w-0 sm:block">
      <div class="topnav-user-name truncate text-sm font-semibold text-text-primary">
        {{ userDisplayName }}
      </div>
      <div class="topnav-user-role truncate">
        {{ roleCaption }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  userInitial: string
  userDisplayName: string
  roleCaption: string
}>()
</script>
