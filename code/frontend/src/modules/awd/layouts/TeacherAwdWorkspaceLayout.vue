<script setup lang="ts">
import AwdWorkspaceLayoutFrame from './AwdWorkspaceLayoutFrame.vue'

defineProps<{
  contestTitle: string
  pageTitle: string
  pageDescription?: string
  pages: Array<{ key: string; label: string; description: string }>
  currentPage: string
  heroMetrics?: Array<{ label: string; value: string; helper?: string }>
  resolvePath?: (pageKey: string) => string
  loading?: boolean
  error?: string
  empty?: boolean
  emptyTitle?: string
  emptyDescription?: string
}>()
</script>

<template>
  <AwdWorkspaceLayoutFrame
    eyebrow="Teacher Review"
    shell-class="teacher-management-shell teacher-surface awd-workspace-layout--teacher"
    :contest-title="contestTitle"
    :page-title="pageTitle"
    :page-description="pageDescription"
    :pages="pages"
    :current-page="currentPage"
    :hero-metrics="heroMetrics"
    :resolve-path="resolvePath"
    :loading="loading"
    :error="error"
    :empty="empty"
    :empty-title="emptyTitle"
    :empty-description="emptyDescription"
  >
    <slot />
  </AwdWorkspaceLayoutFrame>
</template>
