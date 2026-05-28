<script setup lang="ts">
import {
  ArrowRight,
  Database,
  Globe2,
  Server,
  Shield,
  ShieldAlert,
  Trophy,
  XCircle,
} from 'lucide-vue-next'

import type { AWDAttackLogData } from '@/api/contracts'

import { formatProjectorScore, formatProjectorTime } from '../model/projectorFormatters'
import type { ContestProjectorAttackTeamPanel } from '../model/projectorTypes'
import {
  getProjectorServiceDisplayName,
  getProjectorServiceIconName,
  type AttackMapDetailPanel,
} from '../model'

defineProps<{
  teamPanels: ContestProjectorAttackTeamPanel[]
  firstBlood: AWDAttackLogData | null
}>()

const emit = defineEmits<{
  openDetail: [panel: AttackMapDetailPanel]
}>()
</script>

<template>
  <aside class="attack-side attack-side--teams">
    <section class="legend-block">
      <h4>图例说明</h4>
      <div class="legend-grid">
        <span><Server /> 服务主机</span>
        <span><Shield /> 所属团队</span>
        <span class="legend-success"><ArrowRight /> 成功攻击</span>
        <span class="legend-failed"><XCircle /> 攻击失败</span>
      </div>
    </section>

    <section
      v-if="firstBlood"
      class="first-blood-block"
    >
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
      @click.stop="emit('openDetail', 'teams')"
      @keydown.enter.stop.prevent="emit('openDetail', 'teams')"
      @keydown.space.stop.prevent="emit('openDetail', 'teams')"
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
          <span>{{ formatProjectorScore(panel.score) }} / 受损 {{ panel.compromisedCount }}</span>
        </header>
        <div class="team-list-services">
          <span
            v-for="service in panel.row.services.slice(0, 4)"
            :key="service.id"
            :class="`team-list-service--${service.service_status}`"
          >
            <Database v-if="getProjectorServiceIconName(service) === 'database'" />
            <Globe2 v-else-if="getProjectorServiceIconName(service) === 'globe'" />
            <ShieldAlert v-else-if="getProjectorServiceIconName(service) === 'shield'" />
            <Server v-else />
            {{ getProjectorServiceDisplayName(service) }}
          </span>
        </div>
      </article>
    </section>
  </aside>
</template>
