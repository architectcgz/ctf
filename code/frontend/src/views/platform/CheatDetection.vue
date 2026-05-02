<script setup lang="ts">
import CheatDetectionWorkspacePanel from '@/components/platform/cheat/CheatDetectionWorkspacePanel.vue'
import { useCheatDetectionPage } from '@/features/platform-overview'

const { riskData, loading, error, quickActions, loadRiskData, openAudit, formatDateTime } =
  useCheatDetectionPage()
</script>

<template>
  <CheatDetectionWorkspacePanel
    :risk-data="riskData"
    :loading="loading"
    :error="error"
    :quick-actions="quickActions"
    :format-date-time="formatDateTime"
    @refresh="void loadRiskData()"
    @open-audit="void openAudit($event)"
  />
</template>
