<script setup lang="ts">
import { ArrowRight, Database, Globe2, Server, ShieldAlert, X } from 'lucide-vue-next'

import type { ScoreboardRow } from '@/api/contracts'
import {
  formatProjectorScore,
  formatProjectorTime,
  getServiceStatusLabel,
} from '../model/projectorFormatters'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackTeamPanel,
} from '../model/projectorTypes'
import {
  getProjectorServiceDisplayName,
  getProjectorServiceIconName,
  type AttackMapDetailPanel,
} from '../model'
import './ContestProjectorAttackDetailOverlay.css'

defineProps<{
  activePanel: AttackMapDetailPanel | null
  teamPanels: ContestProjectorAttackTeamPanel[]
  rankingRows: Array<ScoreboardRow & { compromisedCount: number }>
  attackRows: ContestProjectorAttackEdge[]
}>()

const emit = defineEmits<{
  close: []
}>()

function getDetailPanelTitle(panel: AttackMapDetailPanel | null): string {
  if (panel === 'teams') return '队伍与服务列表'
  if (panel === 'ranking') return '完整团队排名'
  if (panel === 'attacks') return '攻击统计'
  return ''
}
</script>

<template>
  <Teleport to="body">
    <div v-if="activePanel" class="attack-detail-overlay" @click.self="emit('close')">
      <section
        class="attack-detail-panel"
        role="dialog"
        aria-modal="true"
        :aria-label="getDetailPanelTitle(activePanel)"
      >
        <header class="attack-detail-head">
          <div>
            <div class="projector-overline">Drilldown</div>
            <h3>{{ getDetailPanelTitle(activePanel) }}</h3>
          </div>
          <button
            type="button"
            class="attack-detail-close"
            aria-label="关闭详情"
            title="关闭"
            @click="emit('close')"
          >
            <X />
          </button>
        </header>

        <div v-if="activePanel === 'teams'" class="team-detail-grid">
          <article v-for="panel in teamPanels" :key="panel.row.team_id" class="team-detail-card">
            <header>
              <strong>{{ panel.row.team_name }}</strong>
              <span
                >{{ formatProjectorScore(panel.score) }} / 受损 {{ panel.compromisedCount }}</span
              >
            </header>
            <div class="team-detail-services">
              <span
                v-for="service in panel.row.services"
                :key="service.id"
                :class="`team-detail-service--${service.service_status}`"
              >
                <Database v-if="getProjectorServiceIconName(service) === 'database'" />
                <Globe2 v-else-if="getProjectorServiceIconName(service) === 'globe'" />
                <ShieldAlert v-else-if="getProjectorServiceIconName(service) === 'shield'" />
                <Server v-else />
                <strong>{{ getProjectorServiceDisplayName(service) }}</strong>
                <small>{{ getServiceStatusLabel(service.service_status) }}</small>
              </span>
            </div>
          </article>
        </div>

        <div v-else-if="activePanel === 'ranking'" class="ranking-detail-list">
          <div v-for="row in rankingRows" :key="row.team_id" class="ranking-detail-row">
            <span>{{ row.rank }}</span>
            <strong>{{ row.team_name }}</strong>
            <em>{{ formatProjectorScore(row.score) }}</em>
            <small>解题 {{ row.solved_count }}</small>
            <small>受损 {{ row.compromisedCount }}</small>
          </div>
        </div>

        <div v-else-if="activePanel === 'attacks'" class="attack-detail-list">
          <article
            v-for="edge in attackRows"
            :key="edge.id"
            class="attack-detail-row"
            :class="edge.success > 0 ? 'attack-detail-row--success' : 'attack-detail-row--failed'"
          >
            <div>
              <strong>{{ edge.attacker_team }}</strong>
              <ArrowRight />
              <span>{{ edge.victim_team }}</span>
            </div>
            <p>
              <span>{{ edge.latest_service_label }}</span>
              <em>成功 {{ edge.success }} / 失败 {{ edge.failed }}</em>
              <small>{{ formatProjectorTime(edge.latest_at) }}</small>
            </p>
          </article>
        </div>
      </section>
    </div>
  </Teleport>
</template>
