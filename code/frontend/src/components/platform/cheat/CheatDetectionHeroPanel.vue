<script setup lang="ts">
import { RefreshCw, SearchCheck } from 'lucide-vue-next'

import CheatDetectionSummaryPanel from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue'

defineProps<{
  generatedAtLabel: string | null
  loading: boolean
  summary: {
    submit_burst_users: number
    shared_ip_groups: number
    affected_users: number
  } | null
}>()

const emit = defineEmits<{
  openAudit: []
  refresh: []
}>()

function handleOpenAudit(): void {
  emit('openAudit')
}

function handleRefresh(): void {
  emit('refresh')
}
</script>

<template>
  <section class="workspace-hero">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">Integrity Workspace</div>
      <h1 class="hero-title">作弊检测</h1>
      <p class="hero-summary">
        基于提交爆发、IP 共享及行为指纹的多维度线索，快速定位需要继续审计复核的账号与行为。
      </p>
    </div>

    <div class="cheat-hero-actions">
      <div
        v-if="generatedAtLabel"
        class="hero-meta-badge"
      >
        <span class="hero-meta-badge__label">最近生成</span>
        <span class="hero-meta-badge__value">{{ generatedAtLabel }}</span>
      </div>
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="handleOpenAudit"
      >
        <SearchCheck class="h-4 w-4" />
        打开审计日志
      </button>
      <button
        type="button"
        class="ui-btn ui-btn--primary"
        @click="handleRefresh"
      >
        <RefreshCw
          class="h-4 w-4"
          :class="{ 'animate-spin': loading }"
        />
        刷新线索
      </button>
    </div>
  </section>

  <CheatDetectionSummaryPanel
    v-if="summary"
    :summary="summary"
  />
</template>

<style scoped>
.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 56rem;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.cheat-hero-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: flex-end;
  gap: var(--space-3);
  padding-bottom: 0.5rem;
}

.cheat-hero-actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-radius: 1rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.cheat-hero-actions > .ui-btn.ui-btn--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 46%, var(--journal-border));
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: var(--color-primary-hover);
}

.cheat-hero-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--journal-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.hero-meta-badge {
  display: grid;
  gap: var(--space-1);
  justify-items: end;
}

.hero-meta-badge__label {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.hero-meta-badge__value {
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--journal-ink);
}

@media (max-width: 720px) {
  .cheat-hero-actions {
    justify-content: flex-start;
  }
}
</style>
