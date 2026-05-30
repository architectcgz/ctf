<script setup lang="ts">
import CheatDetectionWorkspacePanel from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue'
import { useCheatDetectionPage } from '@/features/platform/overview'

const {
  riskData,
  loading,
  error,
  auditLogRoute,
  buildAuditRoute,
  quickActions,
  loadRiskData,
  formatDateTime,
} = useCheatDetectionPage()
</script>

<template>
  <CheatDetectionWorkspacePanel
    :risk-data="riskData"
    :loading="loading"
    :error="error"
    :audit-log-route="auditLogRoute"
    :build-audit-route="buildAuditRoute"
    :quick-actions="quickActions"
    :format-date-time="formatDateTime"
    @refresh="void loadRiskData()"
  />
</template>
