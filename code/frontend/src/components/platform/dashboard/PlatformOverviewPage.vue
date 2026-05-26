<script setup lang="ts">
import { toRef } from 'vue'

import type { AdminDashboardData } from '@/api/contracts'
import { usePlatformOverviewWorkspace } from '@/features/platform-overview'

import PlatformOverviewAlertsSection from './PlatformOverviewAlertsSection.vue'
import PlatformOverviewHeroPanel from './PlatformOverviewHeroPanel.vue'
import PlatformOverviewHotspotsSection from './PlatformOverviewHotspotsSection.vue'

const props = defineProps<{
  dashboard: AdminDashboardData | null
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openAuditLog: []
  openCheatDetection: []
}>()

const {
  alertCount,
  sortedContainers,
  metaPills,
  overviewMetrics,
  railScore,
  railCopy,
  formatPercent,
  formatUsageBarWidth,
  formatBytes,
  usageTone,
} = usePlatformOverviewWorkspace(toRef(props, 'dashboard'))
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero overview-shell">
    <main class="content-pane overview-content">
      <PlatformOverviewHeroPanel
        :error="error"
        :show-skeleton="loading && !dashboard"
        :meta-pills="metaPills"
        :rail-score="railScore"
        :rail-copy="railCopy"
        :overview-metrics="overviewMetrics"
        @retry="emit('retry')"
        @open-audit-log="emit('openAuditLog')"
        @open-cheat-detection="emit('openCheatDetection')"
      />

      <PlatformOverviewAlertsSection
        :loading="loading"
        :alert-count="alertCount"
        :alerts="dashboard?.alerts ?? []"
      />

      <PlatformOverviewHotspotsSection
        :loading="loading"
        :sorted-containers="sortedContainers"
        :format-percent="formatPercent"
        :format-usage-bar-width="formatUsageBarWidth"
        :format-bytes="formatBytes"
        :usage-tone="usageTone"
      />
    </main>
  </div>
</template>

<style scoped>
.overview-shell {
  --journal-shell-dark-accent: var(--color-primary-hover);
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-success: var(--color-success);
  --workspace-warning: var(--color-warning);
  --workspace-danger: var(--color-danger);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-lg: 18px;
  --workspace-font-mono: var(--font-family-mono);
}

.overview-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

@media (max-width: 640px) {
  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }
}
</style>
