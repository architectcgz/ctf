<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch, type ComponentPublicInstance } from 'vue'
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
  Zap,
} from 'lucide-vue-next'

import type { AWDTeamServiceData, ScoreboardRow } from '@/api/contracts'
import {
  formatProjectorScore,
  formatProjectorTime,
  getServiceStatusLabel,
} from '@/components/platform/contest/projector/contestProjectorFormatters'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorServiceMatrixRow,
} from '@/components/platform/contest/projector/contestProjectorTypes'

const props = defineProps<{
  rows: ContestProjectorServiceMatrixRow[]
  edges: ContestProjectorAttackEdge[]
  scoreboardRows: ScoreboardRow[]
}>()

interface AttackBeam {
  id: string
  edge: ContestProjectorAttackEdge
  path: string
  markerX: number
  markerY: number
}

interface TeamPanel {
  row: ContestProjectorServiceMatrixRow
  rank?: number
  score: number
  compromisedCount: number
  receivedSuccess: number
}

const boardRef = ref<HTMLElement | null>(null)
const teamRefs = new Map<string, HTMLElement>()
const serviceRefs = new Map<string, HTMLElement>()
const beams = ref<AttackBeam[]>([])
let resizeObserver: ResizeObserver | null = null

const scoreMap = computed(() => new Map(props.scoreboardRows.map((row) => [row.team_id, row])))
const displayedRows = computed(() => props.rows.slice(0, 8))
const visibleEdges = computed(() => props.edges.slice(0, 12))
const successfulEdges = computed(() => visibleEdges.value.filter((edge) => edge.success > 0))
const failedEdgeCount = computed(() => visibleEdges.value.reduce((sum, edge) => sum + edge.failed, 0))
const compromisedServiceCount = computed(() =>
  displayedRows.value.reduce(
    (sum, row) => sum + row.services.filter((service) => service.service_status === 'compromised').length,
    0
  )
)
const aliveServiceCount = computed(() =>
  displayedRows.value.reduce(
    (sum, row) => sum + row.services.filter((service) => service.service_status === 'up').length,
    0
  )
)
const totalServiceCount = computed(() => displayedRows.value.reduce((sum, row) => sum + row.services.length, 0))

const teamPanels = computed<TeamPanel[]>(() =>
  displayedRows.value.map((row) => {
    const score = scoreMap.value.get(row.team_id)
    return {
      row,
      rank: score?.rank,
      score: score?.score ?? 0,
      compromisedCount: row.services.filter((service) => service.service_status === 'compromised').length,
      receivedSuccess: visibleEdges.value
        .filter((edge) => edge.victim_team_id === row.team_id)
        .reduce((sum, edge) => sum + edge.success, 0),
    }
  })
)

const rankingRows = computed(() =>
  props.scoreboardRows.slice(0, 8).map((row) => ({
    ...row,
    compromisedCount:
      teamPanels.value.find((panel) => panel.row.team_id === row.team_id)?.compromisedCount ?? 0,
  }))
)

function getServiceKey(teamId: string, service: AWDTeamServiceData): string {
  return service.service_id ? `${teamId}:service:${service.service_id}` : `${teamId}:challenge:${service.challenge_id}`
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

function getServiceIconName(service: AWDTeamServiceData): 'database' | 'globe' | 'server' | 'shield' {
  const label = getServiceDisplayName(service).toLowerCase()
  if (label.includes('drive') || label.includes('盘') || label.includes('data') || label.includes('db')) return 'database'
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
      const sourceX = sourceRect.left + sourceRect.width / 2 - boardRect.left
      const sourceY = sourceRect.top + sourceRect.height / 2 - boardRect.top
      const targetX = targetRect.left + targetRect.width / 2 - boardRect.left
      const targetY = targetRect.top + targetRect.height / 2 - boardRect.top
      const distanceX = Math.abs(targetX - sourceX)
      const curve = Math.max(96, distanceX * 0.46)
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
  () => [props.rows, props.edges],
  () => {
    void scheduleBeamUpdate()
  },
  { deep: true }
)

onMounted(() => {
  resizeObserver = new ResizeObserver(() => updateBeams())
  if (boardRef.value) {
    resizeObserver.observe(boardRef.value)
  }
  void scheduleBeamUpdate()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  teamRefs.clear()
  serviceRefs.clear()
})
</script>

<template>
  <section class="attack-map-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          Attack Map
        </div>
        <h3>实时攻击地图</h3>
      </div>
      <Crosshair class="panel-icon panel-icon--attack" />
    </header>

    <div class="attack-map-layout">
      <aside class="attack-side attack-side--teams">
        <section class="legend-block">
          <h4>图例说明</h4>
          <div class="legend-grid">
            <span><Server /> 服务主机</span>
            <span><Shield /> 所属团队</span>
            <span class="legend-success"><ArrowRight /> 成功攻击</span>
            <span class="legend-failed"><ArrowRight /> 攻击失败</span>
          </div>
        </section>

        <section class="team-list-block">
          <h4>团队及其服务列表</h4>
          <article
            v-for="panel in teamPanels"
            :key="panel.row.team_id"
            class="team-list-card"
            :class="{ 'team-list-card--hot': panel.receivedSuccess > 0 }"
          >
            <header>
              <strong>{{ panel.row.team_name }}</strong>
              <span>存活 {{ panel.score }} / 已攻破 {{ panel.compromisedCount }}</span>
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
          <strong>实时攻击地图</strong>
          <span>箭头指向：攻击方 → 目标服务</span>
        </div>

        <div
          ref="boardRef"
          class="attack-board"
        >
          <svg
            class="attack-beam-layer"
            aria-hidden="true"
          >
            <defs>
              <marker
                id="attack-arrow-success"
                markerHeight="8"
                markerWidth="8"
                orient="auto"
                refX="7"
                refY="4"
                viewBox="0 0 8 8"
              >
                <path
                  d="M0,0 L8,4 L0,8 Z"
                  class="attack-marker attack-marker--success"
                />
              </marker>
              <marker
                id="attack-arrow-failed"
                markerHeight="8"
                markerWidth="8"
                orient="auto"
                refX="7"
                refY="4"
                viewBox="0 0 8 8"
              >
                <path
                  d="M0,0 L8,4 L0,8 Z"
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
                :marker-end="beam.edge.success > 0 ? 'url(#attack-arrow-success)' : 'url(#attack-arrow-failed)'"
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
            v-for="panel in teamPanels"
            :key="panel.row.team_id"
            :ref="(element) => setTeamRef(panel.row.team_id, element)"
            class="map-team-node"
            :class="{
              'map-team-node--hot': panel.receivedSuccess > 0,
              'map-team-node--rank-one': panel.rank === 1,
            }"
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
                :ref="(element) => setServiceRef(getServiceKey(panel.row.team_id, service), element)"
                class="map-service"
                :class="[
                  `map-service--${service.service_status}`,
                  { 'map-service--hit': getServiceAttackCount(panel.row.team_id, service) > 0 },
                ]"
                :title="`${getServiceDisplayName(service)} · ${getServiceStatusLabel(service.service_status)}`"
              >
                <Database v-if="getServiceIconName(service) === 'database'" />
                <Globe2 v-else-if="getServiceIconName(service) === 'globe'" />
                <ShieldAlert v-else-if="getServiceIconName(service) === 'shield'" />
                <Server v-else />
                <span>{{ getServiceDisplayName(service) }}</span>
                <i v-if="getServiceAttackCount(panel.row.team_id, service) > 0">
                  {{ getServiceAttackCount(panel.row.team_id, service) }}
                </i>
              </span>
            </div>
          </article>
        </div>

        <footer class="selected-service-strip">
          <span>提示：服务图标显示最近攻击落点和受损状态</span>
          <strong>{{ visibleEdges[0]?.latest_service_label ?? '暂无目标服务' }}</strong>
        </footer>
      </main>

      <aside class="attack-side attack-side--stats">
        <section class="status-block">
          <h4>比赛状态</h4>
          <div class="status-grid">
            <span><strong>{{ displayedRows.length }}</strong> 存活团队</span>
            <span><strong>{{ aliveServiceCount }} / {{ totalServiceCount }}</strong> 存活服务</span>
            <span><strong>{{ compromisedServiceCount }}</strong> 已攻破服务</span>
          </div>
        </section>

        <section class="rank-block">
          <h4>团队排名</h4>
          <div class="rank-list">
            <div
              v-for="row in rankingRows"
              :key="row.team_id"
              class="rank-row"
            >
              <span>{{ row.rank }}</span>
              <strong>{{ row.team_name }}</strong>
              <em>{{ formatProjectorScore(row.score) }}</em>
              <small>{{ row.compromisedCount }}</small>
            </div>
          </div>
        </section>

        <section class="attack-stat-block">
          <h4>攻击统计</h4>
          <article
            v-for="edge in visibleEdges.slice(0, 5)"
            :key="edge.id"
            class="attack-stat-row"
          >
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
            <span v-if="visibleEdges.some((edge) => edge.reciprocalSuccess > 0)"><Zap /> 存在互攻</span>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.attack-map-panel,
.attack-side,
.legend-block,
.team-list-block,
.attack-map-main,
.status-block,
.rank-block,
.attack-stat-block {
  display: flex;
  flex-direction: column;
}

.attack-map-panel {
  gap: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 56%, transparent);
  padding: var(--space-4);
}

.panel-head,
.map-title,
.team-list-card header,
.map-team-node header,
.attack-stat-row div,
.attack-stat-row p,
.attack-stat-summary {
  display: flex;
  align-items: center;
}

.panel-head {
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.panel-head h3 {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
  font-weight: 900;
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.panel-icon {
  width: var(--space-5);
  height: var(--space-5);
}

.panel-icon--attack {
  color: var(--color-danger);
}

.attack-map-layout {
  display: grid;
  grid-template-columns: minmax(18rem, 0.92fr) minmax(32rem, 1.42fr) minmax(18rem, 0.86fr);
  gap: var(--space-4);
}

.attack-side,
.attack-map-main {
  gap: var(--space-3);
  min-width: 0;
}

.legend-block,
.team-list-block,
.status-block,
.rank-block,
.attack-stat-block {
  gap: var(--space-3);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-surface) 48%, transparent);
  padding: var(--space-3);
}

.legend-block h4,
.team-list-block h4,
.status-block h4,
.rank-block h4,
.attack-stat-block h4 {
  margin: 0;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
}

.legend-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
}

.legend-grid span,
.team-list-service--up,
.team-list-service--down,
.team-list-service--compromised {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-1-5);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.legend-grid svg,
.team-list-services svg {
  width: var(--space-4);
  height: var(--space-4);
  flex: 0 0 auto;
}

.legend-success {
  color: var(--color-success);
}

.legend-failed {
  color: var(--color-danger);
}

.team-list-block {
  max-height: 41rem;
  overflow: auto;
}

.team-list-card {
  display: grid;
  gap: var(--space-3);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  padding: var(--space-3);
}

.team-list-card--hot {
  border-color: color-mix(in srgb, var(--color-danger) 42%, transparent);
}

.team-list-card header {
  justify-content: space-between;
  gap: var(--space-3);
}

.team-list-card strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.team-list-card header span,
.selected-service-strip,
.attack-stat-row p,
.attack-stat-summary {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.team-list-services {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
}

.team-list-services span {
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.375rem;
  padding: var(--space-1-5) var(--space-2);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.team-list-service--up {
  color: var(--color-success);
}

.team-list-service--down,
.team-list-service--compromised {
  color: var(--color-danger);
}

.map-title {
  justify-content: space-between;
  gap: var(--space-3);
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.map-title strong {
  color: var(--journal-ink);
  font-size: var(--font-size-14);
  font-weight: 900;
}

.attack-board {
  position: relative;
  min-height: 42rem;
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background:
    radial-gradient(circle at center, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 54%),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--color-border-subtle) 22%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      0deg,
      color-mix(in srgb, var(--color-border-subtle) 18%, transparent) 1px,
      transparent 1px
    ),
    color-mix(in srgb, var(--color-bg-surface) 52%, transparent);
  background-size: auto, var(--space-8) var(--space-8), var(--space-8) var(--space-8), auto;
}

.attack-beam-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.attack-beam__halo,
.attack-beam__line {
  fill: none;
}

.attack-beam__halo {
  stroke: color-mix(in srgb, var(--color-danger) 18%, transparent);
  stroke-width: var(--space-2);
}

.attack-beam__line {
  stroke: color-mix(in srgb, var(--color-warning) 72%, transparent);
  stroke-dasharray: var(--space-3) var(--space-2);
  stroke-linecap: round;
  stroke-width: var(--space-0-5);
  animation: attack-dash 1.8s linear infinite;
}

.attack-beam--success .attack-beam__line {
  stroke: color-mix(in srgb, var(--color-success) 78%, transparent);
  stroke-dasharray: var(--space-5) var(--space-2);
  stroke-width: var(--space-1);
}

.attack-beam--mutual .attack-beam__halo {
  stroke: color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.attack-beam__impact {
  fill: none;
  stroke: color-mix(in srgb, var(--color-success) 78%, transparent);
  stroke-width: var(--space-0-5);
  animation: impact-pulse 1.6s ease-out infinite;
}

.attack-marker--success {
  fill: var(--color-success);
}

.attack-marker--failed {
  fill: var(--color-danger);
}

.map-team-node {
  position: absolute;
  z-index: 2;
  width: min(32%, 16rem);
  min-width: 12rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 34%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 84%, transparent);
  padding: var(--space-3);
  box-shadow: 0 var(--space-2) var(--space-7) color-mix(in srgb, var(--color-shadow-strong) 14%, transparent);
}

.map-team-node:nth-of-type(1) {
  top: var(--space-5);
  left: 50%;
  transform: translateX(-50%);
}

.map-team-node:nth-of-type(2) {
  top: 37%;
  left: var(--space-5);
}

.map-team-node:nth-of-type(3) {
  top: 37%;
  right: var(--space-5);
}

.map-team-node:nth-of-type(4) {
  bottom: var(--space-5);
  left: var(--space-5);
}

.map-team-node:nth-of-type(5) {
  right: var(--space-5);
  bottom: var(--space-5);
}

.map-team-node:nth-of-type(n + 6) {
  display: none;
}

.map-team-node--hot {
  border-color: color-mix(in srgb, var(--color-danger) 46%, transparent);
}

.map-team-node--rank-one {
  box-shadow:
    0 var(--space-2) var(--space-7) color-mix(in srgb, var(--color-warning) 14%, transparent),
    inset 0 0 0 1px color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.map-team-node header {
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}

.team-emblem {
  display: inline-flex;
  width: var(--space-8);
  height: var(--space-8);
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 999rem;
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
  color: var(--journal-accent);
}

.team-emblem svg {
  width: var(--space-5);
  height: var(--space-5);
}

.map-team-node strong,
.map-team-node small {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.map-team-node strong {
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
}

.map-team-node small {
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
  font-weight: 900;
}

.map-service-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
}

.map-service {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 4.25rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 68%, transparent);
  padding: var(--space-2);
  color: var(--color-text-secondary);
  text-align: center;
}

.map-service svg {
  width: var(--space-5);
  height: var(--space-5);
}

.map-service span {
  width: 100%;
  overflow: hidden;
  font-size: var(--font-size-10);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.map-service i {
  position: absolute;
  top: calc(var(--space-1) * -1);
  right: calc(var(--space-1) * -1);
  display: inline-flex;
  min-width: var(--space-5);
  height: var(--space-5);
  align-items: center;
  justify-content: center;
  border-radius: 999rem;
  background: var(--color-danger);
  color: var(--color-bg-elevated);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
  font-style: normal;
  font-weight: 900;
}

.map-service--up {
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  color: var(--color-success);
}

.map-service--down,
.map-service--compromised {
  border-color: color-mix(in srgb, var(--color-danger) 38%, transparent);
  background: color-mix(in srgb, var(--color-danger) 12%, var(--color-bg-surface));
  color: var(--color-danger);
}

.map-service--hit {
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--color-danger) 44%, transparent),
    0 0 var(--space-5) color-mix(in srgb, var(--color-danger) 18%, transparent);
  animation: service-hit 1.8s ease-in-out infinite;
}

.selected-service-strip {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-surface) 48%, transparent);
  padding: var(--space-3);
}

.selected-service-strip strong {
  color: var(--journal-ink);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-2);
}

.status-grid span {
  display: grid;
  gap: var(--space-1);
  border-right: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
  text-align: center;
}

.status-grid span:last-child {
  border-right: 0;
}

.status-grid strong {
  color: var(--journal-ink);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-25);
  font-weight: 900;
}

.rank-list,
.attack-stat-block {
  gap: var(--space-2);
}

.rank-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  gap: var(--space-3);
  align-items: center;
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.rank-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-row span,
.rank-row em,
.rank-row small {
  font-family: var(--font-family-mono);
  font-style: normal;
  font-weight: 900;
}

.rank-row em {
  color: var(--journal-ink);
}

.attack-stat-row {
  display: grid;
  gap: var(--space-1);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
}

.attack-stat-row div,
.attack-stat-row p {
  gap: var(--space-2);
}

.attack-stat-row strong,
.attack-stat-row span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-stat-row strong {
  color: var(--journal-ink);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.attack-stat-row svg {
  width: var(--space-3);
  height: var(--space-3);
  color: var(--color-danger);
}

.result-success {
  color: var(--color-success);
}

.result-failed {
  color: var(--color-danger);
}

.attack-stat-summary {
  flex-wrap: wrap;
  gap: var(--space-2);
  padding-top: var(--space-1);
}

.attack-stat-summary span {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
}

.attack-stat-summary svg {
  width: var(--space-3);
  height: var(--space-3);
  color: var(--color-warning);
}

@keyframes attack-dash {
  to {
    stroke-dashoffset: calc(var(--space-8) * -1);
  }
}

@keyframes impact-pulse {
  0% {
    opacity: 0.86;
    r: 5;
  }

  100% {
    opacity: 0;
    r: 18;
  }
}

@keyframes service-hit {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(calc(var(--space-0-5) * -1));
  }
}

@media (max-width: 1280px) {
  .attack-map-layout {
    grid-template-columns: 1fr;
  }

  .attack-board {
    min-height: 36rem;
  }
}

@media (max-width: 900px) {
  .map-team-node {
    position: relative;
    inset: auto;
    width: 100%;
    min-width: 0;
    transform: none;
  }

  .attack-board {
    display: grid;
    gap: var(--space-3);
    min-height: auto;
    padding: var(--space-3);
  }

  .attack-beam-layer {
    display: none;
  }

  .status-grid,
  .team-list-services {
    grid-template-columns: 1fr;
  }
}
</style>
