<script setup lang="ts">
import {
  BarChart3,
  Download,
  LayoutGrid,
  Sword,
  Zap,
} from 'lucide-vue-next'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'

import AWDAttackLogPanel from './AWDAttackLogPanel.vue'
import AWDScoreboardSummaryPanel from './AWDScoreboardSummaryPanel.vue'
import AWDServiceStatusPanel from './AWDServiceStatusPanel.vue'
import AWDTrafficPanel from './AWDTrafficPanel.vue'
import type {
  AWDInspectorCanvasWorkspaceEmits,
  AWDInspectorCanvasWorkspaceProps,
} from './awdInspector.types'

const props = defineProps<AWDInspectorCanvasWorkspaceProps>()
const emit = defineEmits<AWDInspectorCanvasWorkspaceEmits>()
defineSlots<{
  'service-alerts'?: (props: {
    serviceAlerts: AWDInspectorCanvasWorkspaceProps['serviceAlerts']
    selectedAlertKey: string
    getServiceAlertClass: (alertKey: string) => string
    applyServiceAlertFilter: (alertKey: string) => void
  }) => unknown
}>()
</script>

<template>
  <div class="awd-detail-canvas">
    <header class="canvas-tabs-header">
      <nav class="sub-tabs">
        <button
          class="sub-tab"
          :class="{ active: activeSubTab === 'matrix' }"
          type="button"
          @click="emit('update:activeSubTab', 'matrix')"
        >
          <LayoutGrid class="h-3.5 w-3.5" />
          运行矩阵
        </button>
        <button
          class="sub-tab"
          :class="{ active: activeSubTab === 'scoreboard' }"
          type="button"
          @click="emit('update:activeSubTab', 'scoreboard')"
        >
          <BarChart3 class="h-3.5 w-3.5" />
          排行榜单
        </button>
        <button
          class="sub-tab"
          :class="{ active: activeSubTab === 'attacks' }"
          type="button"
          @click="emit('update:activeSubTab', 'attacks')"
        >
          <Sword class="h-3.5 w-3.5" />
          攻击流水
        </button>
        <button
          class="sub-tab"
          :class="{ active: activeSubTab === 'traffic' }"
          type="button"
          @click="emit('update:activeSubTab', 'traffic')"
        >
          <Zap class="h-3.5 w-3.5" />
          流量分析
        </button>
      </nav>

      <div class="canvas-actions">
        <button
          type="button"
          class="ui-btn ui-btn--secondary awd-inspector-export-button"
          @click="emit('exportReviewPackage')"
        >
          <Download class="h-3.5 w-3.5" />
          导出复盘包
        </button>
      </div>
    </header>

    <div class="canvas-content custom-scrollbar">
      <div
        v-if="loadingRoundDetail"
        class="canvas-loading-overlay"
      >
        <AppLoading>同步态势中...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="!selectedRound"
        title="等待激活"
        description="在上方选择轮次以展开战场监控。"
        icon="History"
        class="py-24"
      />

      <div
        v-else
        class="pane-wrap"
      >
        <div
          v-show="activeSubTab === 'matrix'"
          class="pane-matrix"
        >
          <slot
            name="service-alerts"
            :service-alerts="serviceAlerts"
            :selected-alert-key="selectedAlertKey"
            :get-service-alert-class="getServiceAlertClass"
            :apply-service-alert-filter="applyServiceAlertFilter"
          />
          <AWDServiceStatusPanel
            v-bind="serviceStatusPanel"
            @update-service-team-filter="emit('updateServiceTeamFilter', $event)"
            @update-service-status-filter="emit('updateServiceStatusFilter', $event)"
            @update-service-check-source-filter="emit('updateServiceCheckSourceFilter', $event)"
            @update-service-alert-reason-filter="emit('updateServiceAlertReasonFilter', $event)"
            @export-services="emit('exportServices')"
          />
        </div>

        <div
          v-show="activeSubTab === 'scoreboard'"
          class="pane-scoreboard"
        >
          <AWDScoreboardSummaryPanel v-bind="scoreboardSummaryPanel" />
        </div>

        <div
          v-show="activeSubTab === 'attacks'"
          class="pane-attacks"
        >
          <AWDAttackLogPanel
            v-bind="attackLogPanel"
            @update-attack-team-filter="emit('updateAttackTeamFilter', $event)"
            @update-attack-result-filter="emit('updateAttackResultFilter', $event)"
            @update-attack-source-filter="emit('updateAttackSourceFilter', $event)"
            @export-attacks="emit('exportAttacks')"
          />
        </div>

        <div
          v-show="activeSubTab === 'traffic'"
          class="pane-traffic"
        >
          <AWDTrafficPanel
            v-bind="trafficPanel"
            @apply-traffic-filters="emit('applyTrafficFilters', $event)"
            @change-traffic-page="emit('changeTrafficPage', $event)"
            @reset-traffic-filters="emit('resetTrafficFilters')"
          />
        </div>
      </div>
    </div>
  </div>
</template>
