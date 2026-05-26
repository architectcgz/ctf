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
