<script setup lang="ts">
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
} from '@/api/contracts'

import { AWDReadinessOverrideDialog } from '@/features/awd-readiness'

import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'
import type {
  AwdCreateAttackLogPayload,
  AwdCreateRoundPayload,
  AwdCreateServiceCheckPayload,
  AwdOperationsOverrideDialogState,
} from './awdOperationsDialogContracts'

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
  overrideDialogState: AwdOperationsOverrideDialogState
}>()

const emit = defineEmits<{
  'update:roundDialogOpen': [value: boolean]
  saveRound: [payload: AwdCreateRoundPayload]
  'update:serviceCheckDialogOpen': [value: boolean]
  saveServiceCheck: [payload: AwdCreateServiceCheckPayload]
  'update:attackLogDialogOpen': [value: boolean]
  saveAttackLog: [payload: AwdCreateAttackLogPayload]
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
