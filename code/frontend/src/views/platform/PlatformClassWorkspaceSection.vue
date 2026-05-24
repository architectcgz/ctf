<script setup lang="ts">
import { useClassWorkspaceSection } from '@/features/class-workspace-redirect'
import PlatformClassStudents from '@/views/platform/PlatformClassStudents.vue'

useClassWorkspaceSection()
</script>

<template>
  <PlatformClassStudents />
</template>
