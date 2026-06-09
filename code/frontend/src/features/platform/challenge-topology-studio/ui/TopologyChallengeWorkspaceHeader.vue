<script setup lang="ts">
import { GitBranch, RefreshCw, Save } from 'lucide-vue-next'

import TopologySummaryGrid from './TopologySummaryGrid.vue'

type TopologySummary = {
  networks: number
  nodes: number
  links: number
  policies: number
}

defineProps<{
  eyebrow: string
  title: string
  description: string
  summary: TopologySummary
  exporting: boolean
  canExport: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  back: []
  refresh: []
  exportPackage: []
  save: []
}>()
</script>

<template>
  <div>
    <header class="workspace-topbar topology-workspace-topbar">
      <div class="topology-topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="topology-topbar-chip">{{ eyebrow }}</span>
      </div>
      <div class="topology-topbar-actions">
        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          @click="emit('back')"
        >
          返回题目详情
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          :disabled="exporting || !canExport"
          @click="emit('exportPackage')"
        >
          <GitBranch class="h-4 w-4" />
          {{ exporting ? '导出中...' : '导出题目包' }}
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary topology-action-btn"
          :disabled="saving"
          @click="emit('save')"
        >
          <Save class="h-4 w-4" />
          {{ saving ? '保存中...' : '保存拓扑' }}
        </button>
      </div>
    </header>

    <section class="workspace-tab-heading topology-page-heading">
      <div class="workspace-tab-heading__main">
        <div class="topology-page-kicker">
          {{ eyebrow }}
        </div>
        <h1 class="hero-title">
          {{ title }}
        </h1>
      </div>
      <p class="workspace-page-copy topology-page-copy">
        {{ description }}
      </p>

      <TopologySummaryGrid :summary="summary" mode="challenge" />
    </section>
  </div>
</template>

<style scoped>
.workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2-5);
}

.workspace-overline,
.topology-page-kicker {
  display: inline-flex;
  align-items: center;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.topology-topbar-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0 var(--space-3);
  font-size: var(--font-size-0-76);
  font-weight: 600;
  color: var(--journal-accent);
}

.topology-topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.topology-action-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-0-84);
  --ui-btn-secondary-border: var(--journal-border);
  --ui-btn-secondary-background: color-mix(
    in srgb,
    var(--journal-surface) 94%,
    var(--color-bg-base)
  );
  --ui-btn-secondary-color: var(--journal-ink);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-color: var(--journal-accent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-primary-border: transparent;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-color: var(--color-bg-base);
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--color-bg-base)
  );
  --ui-btn-primary-hover-shadow: 0 12px 28px
    color-mix(in srgb, var(--journal-accent) 16%, transparent);
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.topology-action-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.topology-page-heading {
  display: grid;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page-copy {
  max-width: 48rem;
}

@media (max-width: 767px) {
  .workspace-topbar {
    align-items: flex-start;
    padding-bottom: var(--space-5);
  }
}
</style>
