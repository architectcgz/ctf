<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch, type ComponentPublicInstance } from 'vue'
import { ArrowRight, Crosshair, Database, Globe2, Server, ShieldAlert, Zap } from 'lucide-vue-next'

import type { AWDTeamServiceData } from '@/api/contracts'
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
}>()

interface AttackBeam {
  id: string
  edge: ContestProjectorAttackEdge
  path: string
  markerX: number
  markerY: number
}

const boardRef = ref<HTMLElement | null>(null)
const teamRefs = new Map<string, HTMLElement>()
const serviceRefs = new Map<string, HTMLElement>()
const beams = ref<AttackBeam[]>([])
let resizeObserver: ResizeObserver | null = null

const displayedRows = computed(() => props.rows.slice(0, 8))
const visibleEdges = computed(() => props.edges.slice(0, 10))

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
  } else {
    teamRefs.delete(teamId)
  }
}

function setServiceRef(key: string, element: Element | ComponentPublicInstance | null): void {
  if (element instanceof HTMLElement) {
    serviceRefs.set(key, element)
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
  if (label.includes('drive') || label.includes('盘') || label.includes('data')) return 'database'
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
      const curve = Math.max(80, distanceX * 0.42)
      const controlAX = sourceX + (targetX >= sourceX ? curve : -curve)
      const controlBX = targetX - (targetX >= sourceX ? curve : -curve)
      const path = `M ${sourceX} ${sourceY} C ${controlAX} ${sourceY}, ${controlBX} ${targetY}, ${targetX} ${targetY}`

      return {
        id: edge.id,
        edge,
        path,
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
  for (const element of teamRefs.values()) {
    resizeObserver.observe(element)
  }
  for (const element of serviceRefs.values()) {
    resizeObserver.observe(element)
  }
  void scheduleBeamUpdate()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})
</script>

<template>
  <section class="attack-map-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          攻击关系
        </div>
        <h3>队伍服务拓扑</h3>
      </div>
      <Crosshair class="panel-icon panel-icon--attack" />
    </header>

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

      <div class="team-node-grid">
        <article
          v-for="row in displayedRows"
          :key="row.team_id"
          :ref="(element) => setTeamRef(row.team_id, element)"
          class="team-node"
        >
          <header class="team-node__head">
            <strong>{{ row.team_name }}</strong>
            <span>{{ row.services.length }} SVC</span>
          </header>

          <div class="service-icon-grid">
            <span
              v-for="service in row.services.slice(0, 8)"
              :key="service.id"
              :ref="(element) => setServiceRef(getServiceKey(row.team_id, service), element)"
              class="service-icon"
              :class="[
                `service-icon--${service.service_status}`,
                { 'service-icon--hit': getServiceAttackCount(row.team_id, service) > 0 },
              ]"
              :title="`${getServiceDisplayName(service)} · ${getServiceStatusLabel(service.service_status)}`"
            >
              <Globe2
                v-if="getServiceIconName(service) === 'globe'"
                class="service-icon__glyph"
              />
              <Database
                v-else-if="getServiceIconName(service) === 'database'"
                class="service-icon__glyph"
              />
              <ShieldAlert
                v-else-if="getServiceIconName(service) === 'shield'"
                class="service-icon__glyph"
              />
              <Server
                v-else
                class="service-icon__glyph"
              />
              <span>{{ getServiceDisplayName(service) }}</span>
              <i v-if="getServiceAttackCount(row.team_id, service) > 0">
                {{ getServiceAttackCount(row.team_id, service) }}
              </i>
            </span>
          </div>
        </article>
      </div>
    </div>

    <div class="attack-feed-strip">
      <article
        v-for="edge in visibleEdges.slice(0, 4)"
        :key="edge.id"
        class="attack-feed-card"
        :class="{ 'attack-feed-card--success': edge.success > 0 }"
      >
        <div class="attack-feed-card__line">
          <strong>{{ edge.attacker_team }}</strong>
          <ArrowRight />
          <strong>{{ edge.victim_team }}</strong>
        </div>
        <div class="attack-feed-card__meta">
          <span>{{ edge.latest_service_label }}</span>
          <span>{{ edge.success }} HIT / {{ edge.failed }} MISS</span>
          <span>{{ formatProjectorScore(edge.score) }} pts</span>
          <span>{{ formatProjectorTime(edge.latest_at) }}</span>
          <span
            v-if="edge.reciprocalSuccess > 0"
            class="attack-feed-card__mutual"
          >
            <Zap /> 互攻
          </span>
        </div>
      </article>

      <div
        v-if="visibleEdges.length === 0"
        class="panel-empty"
      >
        暂无队伍攻击关系
      </div>
    </div>
  </section>
</template>

<style scoped>
.attack-map-panel,
.team-node,
.service-icon,
.attack-feed-strip {
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

.panel-head {
  display: flex;
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

.attack-board {
  position: relative;
  min-height: 24rem;
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background:
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
    color-mix(in srgb, var(--color-bg-surface) 48%, transparent);
  background-size: var(--space-8) var(--space-8);
  padding: var(--space-4);
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
  stroke: color-mix(in srgb, var(--color-danger) 76%, transparent);
  stroke-dasharray: var(--space-5) var(--space-2);
  stroke-width: var(--space-1);
}

.attack-beam--mutual .attack-beam__halo {
  stroke: color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.attack-beam__impact {
  fill: none;
  stroke: color-mix(in srgb, var(--color-danger) 78%, transparent);
  stroke-width: var(--space-0-5);
  animation: impact-pulse 1.6s ease-out infinite;
}

.attack-marker--success {
  fill: var(--color-danger);
}

.attack-marker--failed {
  fill: var(--color-warning);
}

.team-node-grid {
  position: relative;
  z-index: 2;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.team-node {
  min-width: 0;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 78%, transparent);
  padding: var(--space-3);
  box-shadow: 0 var(--space-2) var(--space-6) color-mix(in srgb, var(--color-shadow-strong) 10%, transparent);
}

.team-node__head {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.team-node__head strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.team-node__head span {
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
  font-weight: 900;
}

.service-icon-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-2);
}

.service-icon {
  position: relative;
  min-width: 0;
  min-height: 4.75rem;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 60%, transparent);
  padding: var(--space-2);
  color: var(--color-text-secondary);
  text-align: center;
}

.service-icon__glyph {
  width: var(--space-5);
  height: var(--space-5);
}

.service-icon span {
  width: 100%;
  overflow: hidden;
  font-size: var(--font-size-10);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-icon i {
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

.service-icon--up {
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  color: var(--color-success);
}

.service-icon--down,
.service-icon--compromised {
  border-color: color-mix(in srgb, var(--color-danger) 36%, transparent);
  background: color-mix(in srgb, var(--color-danger) 12%, var(--color-bg-surface));
  color: var(--color-danger);
}

.service-icon--hit {
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--color-danger) 44%, transparent),
    0 0 var(--space-5) color-mix(in srgb, var(--color-danger) 18%, transparent);
  animation: service-hit 1.8s ease-in-out infinite;
}

.attack-feed-strip {
  gap: var(--space-2);
}

.attack-feed-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-2);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
}

.attack-feed-card--success {
  color: var(--color-danger);
}

.attack-feed-card__line,
.attack-feed-card__meta,
.attack-feed-card__mutual {
  display: flex;
  min-width: 0;
  align-items: center;
}

.attack-feed-card__line {
  gap: var(--space-2);
  color: var(--journal-ink);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.attack-feed-card__line strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-feed-card__line svg {
  width: var(--space-4);
  height: var(--space-4);
  flex: 0 0 auto;
  color: var(--color-danger);
}

.attack-feed-card__meta {
  flex-wrap: wrap;
  gap: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.attack-feed-card__mutual {
  gap: var(--space-1);
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-warning) 14%, transparent);
  padding: 0 var(--space-2);
  color: var(--color-warning);
}

.attack-feed-card__mutual svg {
  width: var(--space-3);
  height: var(--space-3);
}

.panel-empty {
  display: flex;
  min-height: 8rem;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
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

@media (max-width: 900px) {
  .team-node-grid {
    grid-template-columns: 1fr;
  }

  .service-icon-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
