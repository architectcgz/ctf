<script setup lang="ts">
import {
  computed,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
  type ComponentPublicInstance,
} from 'vue'
import {
  ArrowRight,
  Box,
  Crosshair,
  Database,
  Globe2,
  Server,
  Shield,
  ShieldAlert,
  Skull,
  Sparkles,
  Trophy,
  XCircle,
  Zap,
} from 'lucide-vue-next'

import type { AWDAttackLogData, AWDTeamServiceData, ScoreboardRow } from '@/api/contracts'
import {
  formatProjectorScore,
  formatProjectorTime,
  getServiceStatusLabel,
} from '@/components/platform/contest/projector/contestProjectorFormatters'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackTeamPanel,
  ContestProjectorServiceMatrixRow,
} from '@/components/platform/contest/projector/contestProjectorTypes'
import ContestProjectorAttackDetailOverlay from '@/components/platform/contest/projector/ContestProjectorAttackDetailOverlay.vue'
import '@/components/platform/contest/projector/ContestProjectorAttackMap.css'
import '@/components/platform/contest/projector/ContestProjectorAttackMapResponsive.css'

type AttackMapDetailPanel = 'teams' | 'ranking' | 'attacks'

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

interface AttackBeam {
  id: string
  edge: ContestProjectorAttackEdge
  path: string
  markerX: number
  markerY: number
}

interface TeamDragOffset {
  x: number
  y: number
}

interface TeamDragState {
  teamId: string
  pointerId: number
  startX: number
  startY: number
  originX: number
  originY: number
  minX: number
  maxX: number
  minY: number
  maxY: number
}

const boardRingXRadius = 42
const boardRingYRadius = 36
const boardRef = ref<HTMLElement | null>(null)
const activeDetailPanel = ref<AttackMapDetailPanel | null>(null)
const teamDragOffsets = ref<Record<string, TeamDragOffset>>({})
const teamDragState = ref<TeamDragState | null>(null)
const teamRefs = new Map<string, HTMLElement>()
const serviceRefs = new Map<string, HTMLElement>()
const beams = ref<AttackBeam[]>([])
let resizeObserver: ResizeObserver | null = null

const scoreMap = computed(() => new Map(props.scoreboardRows.map((row) => [row.team_id, row])))
const displayedRows = computed(() => (props.expanded ? props.rows : props.rows.slice(0, 8)))
const visibleEdges = computed(() =>
  props.expanded ? props.edges.slice(0, 24) : props.edges.slice(0, 12)
)
const attackDetailRows = computed(() => props.edges)
const recentAttackEvents = computed(() => props.latestAttackEvents.slice(0, 6))
const successfulEdges = computed(() => visibleEdges.value.filter((edge) => edge.success > 0))
const failedEdgeCount = computed(() =>
  visibleEdges.value.reduce((sum, edge) => sum + edge.failed, 0)
)
const dragStorageKey = computed(() => {
  const teamScope = props.rows
    .map((row) => row.team_id)
    .sort()
    .join('|')
  return `contest-projector:attack-board:drag-offsets:${teamScope}`
})
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
  if (!event) return null
  return event.service_id
    ? `${event.victim_team_id}:service:${event.service_id}`
    : `${event.victim_team_id}:challenge:${event.challenge_id}`
})

function getServiceKey(teamId: string, service: AWDTeamServiceData): string {
  return service.service_id
    ? `${teamId}:service:${service.service_id}`
    : `${teamId}:challenge:${service.challenge_id}`
}

function getServiceDisplayName(service: AWDTeamServiceData): string {
  return (
    service.service_name?.trim() ||
    service.challenge_title?.trim() ||
    (service.service_id ? `服务 ${service.service_id}` : `题目 ${service.challenge_id}`)
  )
}

function setTeamRef(teamId: string, element: Element | ComponentPublicInstance | null): void {
  if (element instanceof HTMLElement) {
    teamRefs.set(teamId, element)
    resizeObserver?.observe(element)
  } else {
    teamRefs.delete(teamId)
  }
}

function setServiceRef(key: string, element: Element | ComponentPublicInstance | null): void {
  if (element instanceof HTMLElement) {
    serviceRefs.set(key, element)
    resizeObserver?.observe(element)
  } else {
    serviceRefs.delete(key)
  }
}

function getServiceAttackCount(teamId: string, service: AWDTeamServiceData): number {
  const serviceKey = getServiceKey(teamId, service)
  return visibleEdges.value
    .filter((edge) => edge.latest_target_key === serviceKey && edge.success > 0)
    .reduce((total, edge) => total + edge.success, 0)
}

function isFirstBloodTarget(teamId: string, service: AWDTeamServiceData): boolean {
  return firstBloodTargetKey.value === getServiceKey(teamId, service)
}

function openDetailPanel(panel: AttackMapDetailPanel): void {
  activeDetailPanel.value = panel
}

function closeDetailPanel(): void {
  activeDetailPanel.value = null
}

function clampValue(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

function loadTeamDragOffsets(): void {
  if (typeof window === 'undefined') return
  try {
    const rawValue = window.localStorage.getItem(dragStorageKey.value)
    if (!rawValue) {
      teamDragOffsets.value = {}
      return
    }
    const parsed = JSON.parse(rawValue) as Record<string, TeamDragOffset>
    teamDragOffsets.value = Object.fromEntries(
      Object.entries(parsed).filter(([teamId]) => props.rows.some((row) => row.team_id === teamId))
    )
  } catch {
    teamDragOffsets.value = {}
  }
}

function saveTeamDragOffsets(): void {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(dragStorageKey.value, JSON.stringify(teamDragOffsets.value))
  } catch {
    // Ignore storage failures; dragging should still work for the current session.
  }
}

function getBoardNodePosition(index: number, total: number): { x: number; y: number } {
  if (total <= 1) return { x: 50, y: 50 }

  const angle = -Math.PI / 2 + (Math.PI * 2 * index) / total
  return {
    x: 50 + Math.cos(angle) * boardRingXRadius,
    y: 50 + Math.sin(angle) * boardRingYRadius,
  }
}

function getTeamNodeStyle(teamId: string, index: number): Record<string, string> | undefined {
  if (!props.expanded) return undefined
  const offset = teamDragOffsets.value[teamId]

  if (props.boardOnly) {
    const position = getBoardNodePosition(index, teamPanels.value.length)
    const dragTransform = offset ? ` translate3d(${offset.x}px, ${offset.y}px, 0)` : ''
    return {
      left: `${position.x}%`,
      top: `${position.y}%`,
      transform: `translate(-50%, -50%)${dragTransform}`,
    }
  }

  if (!offset) return undefined
  return {
    transform: `translate3d(${offset.x}px, ${offset.y}px, 0)`,
  }
}

function isDraggingTeam(teamId: string): boolean {
  return teamDragState.value?.teamId === teamId
}

function startTeamDrag(event: PointerEvent, teamId: string): void {
  if (!props.expanded) return

  const board = boardRef.value
  const target = event.currentTarget
  if (!board || !(target instanceof HTMLElement)) return

  event.preventDefault()
  event.stopPropagation()

  const boardRect = board.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  const currentOffset = teamDragOffsets.value[teamId] ?? { x: 0, y: 0 }
  const safeInset = 8

  teamDragState.value = {
    teamId,
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    originX: currentOffset.x,
    originY: currentOffset.y,
    minX: safeInset - (targetRect.left - currentOffset.x - boardRect.left),
    maxX:
      boardRect.width -
      safeInset -
      targetRect.width -
      (targetRect.left - currentOffset.x - boardRect.left),
    minY: safeInset - (targetRect.top - currentOffset.y - boardRect.top),
    maxY:
      boardRect.height -
      safeInset -
      targetRect.height -
      (targetRect.top - currentOffset.y - boardRect.top),
  }

  target.setPointerCapture(event.pointerId)
}

function moveTeamDrag(event: PointerEvent): void {
  const state = teamDragState.value
  if (!state || state.pointerId !== event.pointerId) return

  event.preventDefault()
  const nextOffset = {
    x: clampValue(state.originX + event.clientX - state.startX, state.minX, state.maxX),
    y: clampValue(state.originY + event.clientY - state.startY, state.minY, state.maxY),
  }
  teamDragOffsets.value = {
    ...teamDragOffsets.value,
    [state.teamId]: nextOffset,
  }
  requestAnimationFrame(updateBeams)
}

function endTeamDrag(event: PointerEvent): void {
  const state = teamDragState.value
  if (!state || state.pointerId !== event.pointerId) return

  const target = event.currentTarget
  if (target instanceof HTMLElement && target.hasPointerCapture(event.pointerId)) {
    target.releasePointerCapture(event.pointerId)
  }
  teamDragState.value = null
  saveTeamDragOffsets()
  requestAnimationFrame(updateBeams)
}

function resetTeamDrag(teamId: string): void {
  if (!props.expanded) return
  const { [teamId]: _removed, ...restOffsets } = teamDragOffsets.value
  teamDragOffsets.value = restOffsets
  saveTeamDragOffsets()
  void scheduleBeamUpdate()
}

function getServiceIconName(
  service: AWDTeamServiceData
): 'database' | 'globe' | 'server' | 'shield' {
  const label = getServiceDisplayName(service).toLowerCase()
  if (
    label.includes('drive') ||
    label.includes('盘') ||
    label.includes('data') ||
    label.includes('db')
  )
    return 'database'
  if (label.includes('web') || label.includes('ticket') || label.includes('工单')) return 'globe'
  if (service.service_status === 'compromised' || service.service_status === 'down') return 'shield'
  return 'server'
}

function updateBeams(): void {
  const board = boardRef.value
  if (!board) {
    beams.value = []
    return
  }

  const boardRect = board.getBoundingClientRect()
  beams.value = visibleEdges.value
    .map((edge) => {
      const source = teamRefs.get(edge.attacker_team_id)
      const target = serviceRefs.get(edge.latest_target_key)
      if (!source || !target) return null

      const sourceRect = source.getBoundingClientRect()
      const targetRect = target.getBoundingClientRect()
      const sourceCenterX = sourceRect.left + sourceRect.width / 2 - boardRect.left
      const targetCenterX = targetRect.left + targetRect.width / 2 - boardRect.left
      const sourceX =
        sourceCenterX <= targetCenterX
          ? sourceRect.right - boardRect.left
          : sourceRect.left - boardRect.left
      const sourceY = sourceRect.top + sourceRect.height / 2 - boardRect.top
      const targetX =
        sourceCenterX <= targetCenterX
          ? targetRect.left - boardRect.left
          : targetRect.right - boardRect.left
      const targetY = targetRect.top + targetRect.height / 2 - boardRect.top
      const distanceX = Math.abs(targetX - sourceX)
      const curve = Math.max(72, distanceX * 0.42)
      const controlAX = sourceX + (targetX >= sourceX ? curve : -curve)
      const controlBX = targetX - (targetX >= sourceX ? curve : -curve)

      return {
        id: edge.id,
        edge,
        path: `M ${sourceX} ${sourceY} C ${controlAX} ${sourceY}, ${controlBX} ${targetY}, ${targetX} ${targetY}`,
        markerX: targetX,
        markerY: targetY,
      }
    })
    .filter((item): item is AttackBeam => item !== null)
}

async function scheduleBeamUpdate(): Promise<void> {
  await nextTick()
  updateBeams()
}

watch(
  () => [props.rows, props.edges, props.expanded],
  () => {
    void scheduleBeamUpdate()
  },
  { deep: true }
)

watch(
  () => [dragStorageKey.value, props.expanded],
  () => {
    if (props.expanded) {
      loadTeamDragOffsets()
      void scheduleBeamUpdate()
    }
  }
)

function handleKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return
  closeDetailPanel()
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  resizeObserver = new ResizeObserver(() => updateBeams())
  if (boardRef.value) {
    resizeObserver.observe(boardRef.value)
  }
  if (props.expanded) {
    loadTeamDragOffsets()
  }
  void scheduleBeamUpdate()
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  resizeObserver?.disconnect()
  resizeObserver = null
  teamRefs.clear()
  serviceRefs.clear()
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
        <div class="projector-overline">{{ expanded ? 'Attack Board' : 'Attack Map' }}</div>
        <h3>{{ expanded ? '实时攻击面板' : '实时攻击地图' }}</h3>
      </div>
      <Crosshair class="panel-icon panel-icon--attack" />
    </header>

    <div class="attack-map-layout">
      <aside v-if="!boardOnly" class="attack-side attack-side--teams">
        <section class="legend-block">
          <h4>图例说明</h4>
          <div class="legend-grid">
            <span><Server /> 服务主机</span>
            <span><Shield /> 所属团队</span>
            <span class="legend-success"><ArrowRight /> 成功攻击</span>
            <span class="legend-failed"><XCircle /> 攻击失败</span>
          </div>
        </section>

        <section v-if="firstBlood" class="first-blood-block">
          <div class="first-blood-icon">
            <Trophy />
          </div>
          <div>
            <h4>首血</h4>
            <strong>{{ firstBlood.attacker_team }}</strong>
            <span>攻破 {{ firstBlood.victim_team }}</span>
            <small
              >{{ formatProjectorScore(firstBlood.score_gained) }} pts ·
              {{ formatProjectorTime(firstBlood.created_at) }}</small
            >
          </div>
        </section>

        <section
          class="team-list-block panel-drilldown"
          role="button"
          tabindex="0"
          aria-label="展开队伍与服务列表"
          @click.stop="openDetailPanel('teams')"
          @keydown.enter.stop.prevent="openDetailPanel('teams')"
          @keydown.space.stop.prevent="openDetailPanel('teams')"
        >
          <h4>
            团队及其服务列表
            <span>展开</span>
          </h4>
          <article
            v-for="panel in teamPanels"
            :key="panel.row.team_id"
            class="team-list-card"
            :class="{ 'team-list-card--hot': panel.receivedSuccess > 0 }"
          >
            <header>
              <strong>{{ panel.row.team_name }}</strong>
              <span
                >{{ formatProjectorScore(panel.score) }} / 受损 {{ panel.compromisedCount }}</span
              >
            </header>
            <div class="team-list-services">
              <span
                v-for="service in panel.row.services.slice(0, 4)"
                :key="service.id"
                :class="`team-list-service--${service.service_status}`"
              >
                <Database v-if="getServiceIconName(service) === 'database'" />
                <Globe2 v-else-if="getServiceIconName(service) === 'globe'" />
                <ShieldAlert v-else-if="getServiceIconName(service) === 'shield'" />
                <Server v-else />
                {{ getServiceDisplayName(service) }}
              </span>
            </div>
          </article>
        </section>
      </aside>

      <main class="attack-map-main">
        <div class="map-title">
          <strong>{{ expanded ? '实时攻击面板' : '实时攻击地图' }}</strong>
          <span>{{ expanded ? '全量队伍攻防矩阵' : '攻击方 → 目标服务' }}</span>
        </div>

        <div
          ref="boardRef"
          class="attack-board"
          :class="{ 'attack-board--drilldown': !expanded }"
          title="点击铺开攻击面板"
        >
          <svg class="attack-beam-layer" aria-hidden="true">
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
                <path d="M0,0 L6,3 L0,6 Z" class="attack-marker attack-marker--success" />
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
                <path d="M0,0 L6,3 L0,6 Z" class="attack-marker attack-marker--failed" />
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
              <path class="attack-beam__halo" :d="beam.path" />
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
                :ref="
                  (element) => setServiceRef(getServiceKey(panel.row.team_id, service), element)
                "
                class="map-service"
                :class="[
                  `map-service--${service.service_status}`,
                  {
                    'map-service--hit': getServiceAttackCount(panel.row.team_id, service) > 0,
                    'map-service--first-blood': isFirstBloodTarget(panel.row.team_id, service),
                  },
                ]"
                :title="`${getServiceDisplayName(service)} · ${getServiceStatusLabel(service.service_status)}`"
              >
                <Database v-if="getServiceIconName(service) === 'database'" />
                <Globe2 v-else-if="getServiceIconName(service) === 'globe'" />
                <ShieldAlert v-else-if="getServiceIconName(service) === 'shield'" />
                <Server v-else />
                <span class="map-service-name">{{ getServiceDisplayName(service) }}</span>
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
              event.is_success
                ? 'attack-event-strip__item--success'
                : 'attack-event-strip__item--failed'
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

      <aside v-if="!boardOnly" class="attack-side attack-side--stats">
        <section
          class="rank-block panel-drilldown"
          role="button"
          tabindex="0"
          aria-label="展开完整团队排名"
          @click.stop="openDetailPanel('ranking')"
          @keydown.enter.stop.prevent="openDetailPanel('ranking')"
          @keydown.space.stop.prevent="openDetailPanel('ranking')"
        >
          <h4>
            团队排名
            <span>展开</span>
          </h4>
          <div class="rank-list">
            <div v-for="row in rankingRows" :key="row.team_id" class="rank-row">
              <span>{{ row.rank }}</span>
              <strong>{{ row.team_name }}</strong>
              <em>{{ formatProjectorScore(row.score) }}</em>
              <small>{{ row.compromisedCount }}</small>
            </div>
          </div>
        </section>

        <section
          class="attack-stat-block panel-drilldown"
          role="button"
          tabindex="0"
          aria-label="展开攻击统计"
          @click.stop="openDetailPanel('attacks')"
          @keydown.enter.stop.prevent="openDetailPanel('attacks')"
          @keydown.space.stop.prevent="openDetailPanel('attacks')"
        >
          <h4>
            攻击统计
            <span>展开</span>
          </h4>
          <article v-for="edge in visibleEdges.slice(0, 5)" :key="edge.id" class="attack-stat-row">
            <div>
              <strong>{{ edge.attacker_team }}</strong>
              <ArrowRight />
              <span>{{ edge.latest_service_label }}</span>
            </div>
            <p>
              <span :class="edge.success > 0 ? 'result-success' : 'result-failed'">
                {{ edge.success > 0 ? '成功' : '失败' }}
              </span>
              <time>{{ formatProjectorTime(edge.latest_at) }}</time>
            </p>
          </article>
          <div class="attack-stat-summary">
            <span>成功 {{ successfulEdges.length }}</span>
            <span>失败 {{ failedEdgeCount }}</span>
            <span v-if="visibleEdges.some((edge) => edge.reciprocalSuccess > 0)"
              ><Zap /> 存在互攻</span
            >
          </div>
        </section>
      </aside>
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
