<script setup lang="ts">
import {
  Activity,
  LayoutGrid,
  Radar,
  ShieldAlert,
} from 'lucide-vue-next'

import type { AWDInspectorStatsHudProps } from './awdInspector.types'

defineProps<AWDInspectorStatsHudProps>()
</script>

<template>
  <div class="awd-stats-hud">
    <div class="stat-card">
      <div class="stat-card__icon stat-card__icon--blue">
        <Activity class="h-4 w-4" />
      </div>
      <div class="stat-card__content">
        <div class="unit-label">
          Infrastructure
        </div>
        <div class="unit-value font-mono">
          {{ totalServiceCount }} <small>SRV</small>
        </div>
        <div class="unit-helper">
          ONLINE: {{ upCount }} · OFF: {{ downCount }}
        </div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-card__icon stat-card__icon--green">
        <Radar class="h-4 w-4" />
      </div>
      <div class="stat-card__content">
        <div class="unit-label">
          Battle Traffic
        </div>
        <div class="unit-value font-mono">
          {{ totalAttackCount }} <small>HITS</small>
        </div>
        <div class="unit-helper unit-helper--success">
          SUCCESS: {{ successfulAttackCount }}
        </div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-card__icon stat-card__icon--orange">
        <ShieldAlert class="h-4 w-4" />
      </div>
      <div class="stat-card__content">
        <div class="unit-label">
          Compromised
        </div>
        <div class="unit-value unit-value--warning font-mono">
          {{ compromisedCount }} <small>EXP</small>
        </div>
        <div class="unit-helper">
          AFFECTED: {{ attackedServiceCount }} TEAMS
        </div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-card__icon stat-card__icon--purple">
        <LayoutGrid class="h-4 w-4" />
      </div>
      <div class="stat-card__content">
        <div class="unit-label">
          Composition
        </div>
        <div class="unit-value">
          {{ getSourceOverviewLabel() }}
        </div>
        <div class="unit-helper">
          {{ getSourceOverviewDescription() }}
        </div>
      </div>
    </div>
  </div>
</template>
