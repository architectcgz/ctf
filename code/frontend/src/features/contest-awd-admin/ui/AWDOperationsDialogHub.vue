<script setup lang="ts">
import { AWDReadinessOverrideDialog } from '@/features/awd-readiness'

import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'
import type {
  AwdAttackLogDialogBinding,
  AwdCreateAttackLogPayload,
  AwdCreateRoundPayload,
  AwdCreateServiceCheckPayload,
  AwdOperationsOverrideDialogState,
  AwdRoundCreateDialogBinding,
  AwdServiceCheckDialogBinding,
} from './awdOperationsDialogContracts'

defineProps<{
  roundDialog: AwdRoundCreateDialogBinding
  serviceCheckDialog: AwdServiceCheckDialogBinding
  attackLogDialog: AwdAttackLogDialogBinding
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
    v-bind="roundDialog"
    @update:open="emit('update:roundDialogOpen', $event)"
    @save="emit('saveRound', $event)"
  />

  <AWDServiceCheckDialog
    v-bind="serviceCheckDialog"
    @update:open="emit('update:serviceCheckDialogOpen', $event)"
    @save="emit('saveServiceCheck', $event)"
  />

  <AWDAttackLogDialog
    v-bind="attackLogDialog"
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
