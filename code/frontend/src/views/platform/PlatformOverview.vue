<script setup lang="ts">
import PlatformOverviewPage from '@/components/platform/dashboard/PlatformOverviewPage.vue'
import { usePlatformOverviewPage } from '@/features/platform-overview'

const { dashboard, loading, error, loadDashboard, openAuditLog, openCheatDetection } =
  usePlatformOverviewPage()
</script>

<template>
  <PlatformOverviewPage
    :dashboard="dashboard"
    :loading="loading"
    :error="error"
    @retry="loadDashboard"
    @open-audit-log="openAuditLog"
    @open-cheat-detection="openCheatDetection"
  />
</template>
