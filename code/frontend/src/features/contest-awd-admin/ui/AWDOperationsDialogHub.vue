<script setup lang="ts">
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDReadinessData,
  AWDRoundData,
  AWDAttackLogData,
  AWDTeamServiceData,
} from '@/api/contracts'

import { AWDReadinessOverrideDialog } from '@/features/awd-readiness'

import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'

defineProps<{
  roundDialogOpen: boolean
  nextRoundNumber: number
  creatingRound: boolean
  serviceCheckDialogOpen: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeViewData[]
  savingServiceCheck: boolean
  attackLogDialogOpen: boolean
  savingAttackLog: boolean
  overrideDialogState: {
    open: boolean
    title: string
    readiness: AWDReadinessData | null
    confirmLoading: boolean
  }
}>()

const emit = defineEmits<{
  'update:roundDialogOpen': [value: boolean]
  saveRound: [
    payload: {
      round_number: number
      status: AWDRoundData['status']
      attack_score: number
      defense_score: number
    },
  ]
  'update:serviceCheckDialogOpen': [value: boolean]
  saveServiceCheck: [
    payload: {
      team_id: number
      service_id: number
      service_status: AWDTeamServiceData['service_status']
      check_result?: Record<string, unknown>
    },
  ]
  'update:attackLogDialogOpen': [value: boolean]
  saveAttackLog: [
    payload: {
      attacker_team_id: number
      victim_team_id: number
      service_id: number
      attack_type: AWDAttackLogData['attack_type']
      submitted_flag?: string
      is_success: boolean
    },
  ]
  'update:overrideDialogOpen': [value: boolean]
  confirmOverride: [reason: string]
}>()
</script>

<template>
  <div
    v-if="overrideDialogState.open"
    class="sr-only"
    aria-live="assertive"
  >
    {{ overrideDialogState.title }} 强制继续
  </div>

  <AWDRoundCreateDialog
    :open="roundDialogOpen"
    :next-round-number="nextRoundNumber"
    :saving="creatingRound"
    @update:open="emit('update:roundDialogOpen', $event)"
    @save="emit('saveRound', $event)"
  />

  <AWDServiceCheckDialog
    :open="serviceCheckDialogOpen"
    :teams="teams"
    :challenge-links="challengeLinks"
    :saving="savingServiceCheck"
    @update:open="emit('update:serviceCheckDialogOpen', $event)"
    @save="emit('saveServiceCheck', $event)"
  />

  <AWDAttackLogDialog
    :open="attackLogDialogOpen"
    :teams="teams"
    :challenge-links="challengeLinks"
    :saving="savingAttackLog"
    @update:open="emit('update:attackLogDialogOpen', $event)"
    @save="emit('saveAttackLog', $event)"
  />

  <AWDReadinessOverrideDialog
    :open="overrideDialogState.open"
    :title="overrideDialogState.title"
    :readiness="overrideDialogState.readiness"
    :confirm-loading="overrideDialogState.confirmLoading"
    @update:open="emit('update:overrideDialogOpen', $event)"
    @confirm="emit('confirmOverride', $event)"
  />
</template>
