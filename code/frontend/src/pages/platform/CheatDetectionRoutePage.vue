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

<script setup lang="ts">
import {
  CheatDetectionWorkspacePanel,
  useCheatDetectionPage,
} from '@/features/platform/cheat-detection'

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
