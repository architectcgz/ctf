<script setup lang="ts">
import { AlertCircle, ShieldQuestion } from 'lucide-vue-next'

import type { AdminCheatDetectionData } from '@/api/contracts'
import AppLoading from '@/components/common/AppLoading.vue'
import CheatDetectionHeroPanel from '@/components/platform/cheat/CheatDetectionHeroPanel.vue'
import CheatDetectionReviewPanels from '@/components/platform/cheat/CheatDetectionReviewPanels.vue'

type CheatQuickAction = {
  title: string
  description: string
  actionLabel: string
  query: Record<string, string>
}

defineProps<{
  riskData: AdminCheatDetectionData | null
  loading: boolean
  error: string
  quickActions: ReadonlyArray<CheatQuickAction>
  formatDateTime: (value: string) => string
}>()

const emit = defineEmits<{
  refresh: []
  openAudit: [query: Record<string, string>]
}>()

function handleRefresh(): void {
  emit('refresh')
}

function handleOpenAudit(query: Record<string, string>): void {
  emit('openAudit', query)
}
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero cheat-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <CheatDetectionHeroPanel
          :generated-at-label="riskData ? formatDateTime(riskData.generated_at) : null"
          :loading="loading"
          :summary="riskData?.summary ?? null"
          @open-audit="handleOpenAudit({})"
          @refresh="handleRefresh"
        />

        <div class="journal-divider" />

        <AppLoading
          v-if="loading && !riskData"
          class="cheat-loading"
        >
          正在扫描合规风险...
        </AppLoading>

        <div
          v-else-if="riskData"
          class="cheat-workbench"
        >
          <CheatDetectionReviewPanels
            :risk-data="riskData"
            :quick-actions="quickActions"
            :format-date-time="formatDateTime"
            @open-audit="handleOpenAudit"
          />
        </div>

        <div
          v-else-if="error"
          class="cheat-error-box"
          role="alert"
        >
          <AlertCircle class="h-4 w-4" />
          <span>{{ error }}</span>
          <button
            type="button"
            class="ui-btn ui-btn--ghost ui-btn--sm"
            @click="handleRefresh"
          >
            重试
          </button>
        </div>

        <div
          v-else
          class="cheat-empty-shell"
        >
          <ShieldQuestion class="h-12 w-12" />
          <p>当前没有任何风险检出</p>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.cheat-shell {
  --journal-shell-dark-accent: var(--color-primary-hover);
  --cheat-card-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --cheat-divider: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider-border: 1px dashed var(--cheat-divider);
  --page-top-tabs-gap: var(--space-7);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}

.cheat-loading {
  padding-block: var(--space-7);
}

.cheat-workbench {
  display: grid;
  gap: var(--space-4);
}

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
}

.cheat-error-box {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-danger) 22%, var(--cheat-card-border));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: var(--space-4);
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.cheat-empty-shell {
  display: grid;
  justify-items: center;
  gap: var(--space-3);
  padding-block: var(--space-8);
  color: var(--journal-muted);
}
</style>
