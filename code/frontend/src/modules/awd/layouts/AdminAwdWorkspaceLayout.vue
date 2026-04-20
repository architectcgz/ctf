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
    eyebrow="Admin Operations"
    shell-class="journal-shell journal-shell-admin journal-hero awd-workspace-layout--admin"
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
