<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Crosshair } from 'lucide-vue-next'

import type { AWDAttackLogData, ScoreboardRow } from '@/api/contracts'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackTeamPanel,
  ContestProjectorServiceMatrixRow,
} from '../model/projectorTypes'
import type { AttackMapDetailPanel } from '../model'
import ContestProjectorAttackBoard from './ContestProjectorAttackBoard.vue'
import ContestProjectorAttackDetailOverlay from './ContestProjectorAttackDetailOverlay.vue'
import ContestProjectorAttackMapStatsSidebar from './ContestProjectorAttackMapStatsSidebar.vue'
import ContestProjectorAttackMapTeamSidebar from './ContestProjectorAttackMapTeamSidebar.vue'
import './ContestProjectorAttackMap.css'
import './ContestProjectorAttackMapResponsive.css'

const props = withDefaults(
  defineProps<{
    rows: ContestProjectorServiceMatrixRow[]
    edges: ContestProjectorAttackEdge[]
    scoreboardRows: ScoreboardRow[]
    firstBlood: AWDAttackLogData | null
    latestAttackEvents: AWDAttackLogData[]
    expanded?: boolean
    boardOnly?: boolean
  }>(),
  {
    expanded: false,
    boardOnly: false,
  }
)

const activeDetailPanel = ref<AttackMapDetailPanel | null>(null)

const scoreMap = computed(() => new Map(props.scoreboardRows.map((row) => [row.team_id, row])))
const displayedRows = computed(() => (props.expanded ? props.rows : props.rows.slice(0, 8)))
const visibleEdges = computed(() =>
  props.expanded ? props.edges.slice(0, 24) : props.edges.slice(0, 12)
)
const attackDetailRows = computed(() => props.edges)
const recentAttackEvents = computed(() => props.latestAttackEvents.slice(0, 6))
const successfulEdgeCount = computed(() => visibleEdges.value.filter((edge) => edge.success > 0).length)
const failedEdgeCount = computed(() =>
  visibleEdges.value.reduce((sum, edge) => sum + edge.failed, 0)
)
const teamPanels = computed<ContestProjectorAttackTeamPanel[]>(() =>
  displayedRows.value.map((row) => {
    const score = scoreMap.value.get(row.team_id)
    return {
      row,
      rank: score?.rank,
      score: score?.score ?? 0,
      compromisedCount: row.services.filter((service) => service.service_status === 'compromised')
        .length,
      receivedSuccess: visibleEdges.value
        .filter((edge) => edge.victim_team_id === row.team_id)
        .reduce((sum, edge) => sum + edge.success, 0),
    }
  })
)

const detailTeamPanels = computed<ContestProjectorAttackTeamPanel[]>(() =>
  props.rows.map((row) => {
    const score = scoreMap.value.get(row.team_id)
    return {
      row,
      rank: score?.rank,
      score: score?.score ?? 0,
      compromisedCount: row.services.filter((service) => service.service_status === 'compromised')
        .length,
      receivedSuccess: props.edges
        .filter((edge) => edge.victim_team_id === row.team_id)
        .reduce((sum, edge) => sum + edge.success, 0),
    }
  })
)

const rankingRows = computed(() =>
  props.scoreboardRows.slice(0, props.expanded ? 20 : 8).map((row) => ({
    ...row,
    compromisedCount:
      detailTeamPanels.value.find((panel) => panel.row.team_id === row.team_id)?.compromisedCount ??
      0,
  }))
)

const detailRankingRows = computed(() =>
  props.scoreboardRows.map((row) => ({
    ...row,
    compromisedCount:
      detailTeamPanels.value.find((panel) => panel.row.team_id === row.team_id)?.compromisedCount ??
      0,
  }))
)

const firstBloodTargetKey = computed(() => {
  const event = props.firstBlood
  if (!event) {
    return null
  }
  return event.service_id
    ? `${event.victim_team_id}:service:${event.service_id}`
    : `${event.victim_team_id}:challenge:${event.awd_challenge_id}`
})

function openDetailPanel(panel: AttackMapDetailPanel): void {
  activeDetailPanel.value = panel
}

function closeDetailPanel(): void {
  activeDetailPanel.value = null
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    closeDetailPanel()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <section
    class="attack-map-panel"
    :class="{
      'attack-map-panel--expanded': expanded,
      'attack-map-panel--board-only': boardOnly,
    }"
  >
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          {{ expanded ? 'Attack Board' : 'Attack Map' }}
        </div>
        <h3>{{ expanded ? '实时攻击面板' : '实时攻击地图' }}</h3>
      </div>
      <Crosshair class="panel-icon panel-icon--attack" />
    </header>

    <div class="attack-map-layout">
      <ContestProjectorAttackMapTeamSidebar
        v-if="!boardOnly"
        :team-panels="teamPanels"
        :first-blood="firstBlood"
        @open-detail="openDetailPanel"
      />

      <ContestProjectorAttackBoard
        :team-panels="teamPanels"
        :visible-edges="visibleEdges"
        :recent-attack-events="recentAttackEvents"
        :first-blood-target-key="firstBloodTargetKey"
        :expanded="expanded"
        :board-only="boardOnly"
      />

      <ContestProjectorAttackMapStatsSidebar
        v-if="!boardOnly"
        :ranking-rows="rankingRows"
        :visible-edges="visibleEdges"
        :successful-edge-count="successfulEdgeCount"
        :failed-edge-count="failedEdgeCount"
        @open-detail="openDetailPanel"
      />
    </div>

    <ContestProjectorAttackDetailOverlay
      :active-panel="activeDetailPanel"
      :team-panels="detailTeamPanels"
      :ranking-rows="detailRankingRows"
      :attack-rows="attackDetailRows"
      @close="closeDetailPanel"
    />
  </section>
</template>
