<script setup lang="ts">
import { computed, toRef } from 'vue'
import {
  ArrowRight,
  Box,
  Database,
  Globe2,
  Server,
  ShieldAlert,
  Skull,
  Sparkles,
  XCircle,
} from 'lucide-vue-next'

import type { AWDAttackLogData, AWDTeamServiceData } from '@/api/contracts'

import { formatProjectorTime, getServiceStatusLabel } from '../model/projectorFormatters'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackTeamPanel,
} from '../model/projectorTypes'
import {
  getProjectorAttackServiceKey,
  getProjectorServiceDisplayName,
  getProjectorServiceIconName,
  useProjectorAttackBoard,
} from '../model'

const props = defineProps<{
  teamPanels: ContestProjectorAttackTeamPanel[]
  visibleEdges: ContestProjectorAttackEdge[]
  recentAttackEvents: AWDAttackLogData[]
  firstBloodTargetKey: string | null
  expanded: boolean
  boardOnly: boolean
}>()

const visibleTeamPanels = computed(() => props.teamPanels)
const visibleEdgeRows = computed(() => props.visibleEdges)
const expandedRef = toRef(props, 'expanded')
const boardOnlyRef = toRef(props, 'boardOnly')
const firstBloodTargetKeyRef = toRef(props, 'firstBloodTargetKey')

const {
  boardRef,
  beams,
  setTeamRef,
  setServiceRef,
  getTeamNodeStyle,
  isDraggingTeam,
  startTeamDrag,
  moveTeamDrag,
  endTeamDrag,
  resetTeamDrag,
  getServiceAttackCount,
  isFirstBloodTarget,
} = useProjectorAttackBoard({
  teamPanels: visibleTeamPanels,
  visibleEdges: visibleEdgeRows,
  expanded: expandedRef,
  boardOnly: boardOnlyRef,
  firstBloodTargetKey: firstBloodTargetKeyRef,
})

const panelTitle = computed(() => (props.expanded ? '实时攻击面板' : '实时攻击地图'))
const panelSubtitle = computed(() => (props.expanded ? '全量队伍攻防矩阵' : '攻击方 → 目标服务'))

function getServiceRefKey(teamId: string, service: AWDTeamServiceData): string {
  return getProjectorAttackServiceKey(teamId, service)
}
</script>

<template>
  <main class="attack-map-main">
    <div class="map-title">
      <strong>{{ panelTitle }}</strong>
      <span>{{ panelSubtitle }}</span>
    </div>

    <div
      ref="boardRef"
      class="attack-board"
      :class="{ 'attack-board--drilldown': !expanded }"
      title="点击铺开攻击面板"
    >
      <svg
        class="attack-beam-layer"
        aria-hidden="true"
      >
        <defs>
          <marker
            id="attack-arrow-success"
            markerHeight="6"
            markerWidth="6"
            orient="auto"
            refX="5.25"
            refY="3"
            viewBox="0 0 6 6"
          >
            <path
              d="M0,0 L6,3 L0,6 Z"
              class="attack-marker attack-marker--success"
            />
          </marker>
          <marker
            id="attack-arrow-failed"
            markerHeight="6"
            markerWidth="6"
            orient="auto"
            refX="5.25"
            refY="3"
            viewBox="0 0 6 6"
          >
            <path
              d="M0,0 L6,3 L0,6 Z"
              class="attack-marker attack-marker--failed"
            />
          </marker>
        </defs>

        <g
          v-for="beam in beams"
          :key="beam.id"
          class="attack-beam"
          :class="{
            'attack-beam--success': beam.edge.success > 0,
            'attack-beam--failed': beam.edge.success === 0,
            'attack-beam--mutual': beam.edge.reciprocalSuccess > 0,
          }"
        >
          <path
            class="attack-beam__halo"
            :d="beam.path"
          />
          <path
            class="attack-beam__line"
            :d="beam.path"
            :marker-end="
              beam.edge.success > 0 ? 'url(#attack-arrow-success)' : 'url(#attack-arrow-failed)'
            "
          />
          <circle
            v-if="beam.edge.success > 0"
            class="attack-beam__impact"
            :cx="beam.markerX"
            :cy="beam.markerY"
            r="8"
          />
        </g>
      </svg>

      <article
        v-for="(panel, panelIndex) in teamPanels"
        :key="panel.row.team_id"
        :ref="(element) => setTeamRef(panel.row.team_id, element)"
        class="map-team-node"
        :class="{
          'map-team-node--hot': panel.receivedSuccess > 0,
          'map-team-node--rank-one': panel.rank === 1,
          'map-team-node--draggable': expanded,
          'map-team-node--dragging': isDraggingTeam(panel.row.team_id),
        }"
        :style="getTeamNodeStyle(panel.row.team_id, panelIndex)"
        :title="expanded ? '拖动调整队伍位置，双击恢复' : undefined"
        @pointerdown="startTeamDrag($event, panel.row.team_id)"
        @pointermove="moveTeamDrag"
        @pointerup="endTeamDrag"
        @pointercancel="endTeamDrag"
        @dblclick.stop="resetTeamDrag(panel.row.team_id)"
      >
        <header>
          <span class="team-emblem">
            <Skull v-if="panel.receivedSuccess > 0" />
            <Box v-else />
          </span>
          <div>
            <strong>{{ panel.row.team_name }}</strong>
            <small>{{ panel.score }} pts</small>
          </div>
        </header>

        <div class="map-service-grid">
          <span
            v-for="service in panel.row.services.slice(0, 6)"
            :key="service.id"
            :ref="(element) => setServiceRef(getServiceRefKey(panel.row.team_id, service), element)"
            class="map-service"
            :class="[
              `map-service--${service.service_status}`,
              {
                'map-service--hit': getServiceAttackCount(panel.row.team_id, service) > 0,
                'map-service--first-blood': isFirstBloodTarget(panel.row.team_id, service),
              },
            ]"
            :title="`${getProjectorServiceDisplayName(service)} · ${getServiceStatusLabel(service.service_status)}`"
          >
            <Database v-if="getProjectorServiceIconName(service) === 'database'" />
            <Globe2 v-else-if="getProjectorServiceIconName(service) === 'globe'" />
            <ShieldAlert v-else-if="getProjectorServiceIconName(service) === 'shield'" />
            <Server v-else />
            <span class="map-service-name">{{ getProjectorServiceDisplayName(service) }}</span>
            <small class="map-service-status">
              {{ getServiceStatusLabel(service.service_status) }}
            </small>
            <i v-if="getServiceAttackCount(panel.row.team_id, service) > 0">
              {{ getServiceAttackCount(panel.row.team_id, service) }}
            </i>
            <b v-if="isFirstBloodTarget(panel.row.team_id, service)">
              <Sparkles />
            </b>
          </span>
        </div>
      </article>
    </div>

    <footer class="attack-event-strip">
      <article
        v-for="event in recentAttackEvents"
        :key="event.id"
        :class="
          event.is_success ? 'attack-event-strip__item--success' : 'attack-event-strip__item--failed'
        "
      >
        <span>{{ event.attacker_team }}</span>
        <ArrowRight v-if="event.is_success" />
        <XCircle v-else />
        <strong>{{ event.victim_team }}</strong>
        <small>{{ formatProjectorTime(event.created_at) }}</small>
      </article>
      <strong v-if="recentAttackEvents.length === 0">暂无攻击事件</strong>
    </footer>
  </main>
</template>
