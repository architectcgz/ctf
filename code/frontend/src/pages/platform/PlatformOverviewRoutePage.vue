<script setup lang="ts">
import { PlatformOverviewPage, usePlatformOverviewPage } from '@/features/platform/overview'

const { dashboard, loading, error, loadDashboard, auditLogRoute, cheatDetectionRoute } =
  usePlatformOverviewPage()
</script>

<template>
  <PlatformOverviewPage
    :dashboard="dashboard"
    :loading="loading"
    :error="error"
    :audit-log-route="auditLogRoute"
    :cheat-detection-route="cheatDetectionRoute"
    @retry="loadDashboard"
  />
</template>
