<script setup lang="ts">
import { ArrowRight, XCircle, Zap } from 'lucide-vue-next'

import type { ScoreboardRow } from '@/api/contracts'

import { formatProjectorScore, formatProjectorTime } from '../model/projectorFormatters'
import type { ContestProjectorAttackEdge } from '../model/projectorTypes'
import type { AttackMapDetailPanel } from '../model'

defineProps<{
  rankingRows: Array<ScoreboardRow & { compromisedCount: number }>
  visibleEdges: ContestProjectorAttackEdge[]
  successfulEdgeCount: number
  failedEdgeCount: number
}>()

const emit = defineEmits<{
  openDetail: [panel: AttackMapDetailPanel]
}>()
</script>

<template>
  <aside class="attack-side attack-side--stats">
    <section
      class="rank-block panel-drilldown"
      role="button"
      tabindex="0"
      aria-label="展开完整团队排名"
      @click.stop="emit('openDetail', 'ranking')"
      @keydown.enter.stop.prevent="emit('openDetail', 'ranking')"
      @keydown.space.stop.prevent="emit('openDetail', 'ranking')"
    >
      <h4>
        团队排名
        <span>展开</span>
      </h4>
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

    <section
      class="attack-stat-block panel-drilldown"
      role="button"
      tabindex="0"
      aria-label="展开攻击统计"
      @click.stop="emit('openDetail', 'attacks')"
      @keydown.enter.stop.prevent="emit('openDetail', 'attacks')"
      @keydown.space.stop.prevent="emit('openDetail', 'attacks')"
    >
      <h4>
        攻击统计
        <span>展开</span>
      </h4>
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
        <span>成功 {{ successfulEdgeCount }}</span>
        <span>失败 {{ failedEdgeCount }}</span>
        <span v-if="visibleEdges.some((edge) => edge.reciprocalSuccess > 0)">
          <Zap /> 存在互攻
        </span>
      </div>
    </section>
  </aside>
</template>
