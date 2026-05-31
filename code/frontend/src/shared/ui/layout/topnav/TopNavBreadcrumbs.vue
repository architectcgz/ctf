<template>
  <div class="topnav-breadcrumb flex min-w-0 items-center text-sm font-bold">
    <button
      type="button"
      class="topnav-breadcrumb__link topnav-breadcrumb__root whitespace-nowrap"
      @click="$emit('navigate', breadcrumb.workspacePath)"
    >
      Workspace
    </button>
    <span class="topnav-breadcrumb__divider mx-2">/</span>
    <button
      type="button"
      class="topnav-breadcrumb__link whitespace-nowrap"
      @click="$emit('navigate', breadcrumb.modulePath)"
    >
      {{ breadcrumb.moduleLabel }}
    </button>
    <span class="topnav-breadcrumb__divider mx-2">/</span>
    <button
      type="button"
      class="topnav-breadcrumb__link truncate"
      :class="{ 'topnav-breadcrumb__current font-black': !breadcrumb.detailLabel }"
      :aria-current="breadcrumb.detailLabel ? undefined : 'page'"
      @click="$emit('navigate', breadcrumb.secondaryPath)"
    >
      {{ breadcrumb.secondaryLabel }}
    </button>
    <template v-if="breadcrumb.detailLabel">
      <span class="topnav-breadcrumb__divider mx-2">/</span>
      <button
        type="button"
        class="topnav-breadcrumb__link topnav-breadcrumb__current truncate font-black"
        aria-current="page"
        @click="$emit('navigate', breadcrumb.detailPath)"
      >
        {{ breadcrumb.detailLabel }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { WorkspaceShellBreadcrumb } from '@/shared/model/layout/useWorkspaceShellNavigation'

defineProps<{
  breadcrumb: WorkspaceShellBreadcrumb
}>()

defineEmits<{
  navigate: [path: string]
}>()
</script>
