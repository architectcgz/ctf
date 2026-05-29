import { computed, ref, type Ref } from 'vue'

import type { AWDRoundData } from '@/api/contracts'

import type {
  AwdCreateAttackLogPayload,
  AwdCreateRoundPayload,
  AwdCreateServiceCheckPayload,
} from './awdOperationsDialogContracts'

interface UseAwdOperationsDialogStateOptions {
  runtimeStageReady: Readonly<Ref<boolean>>
  rounds: Readonly<Ref<Array<Pick<AWDRoundData, 'round_number'>>>>
  createRound: (payload: AwdCreateRoundPayload) => Promise<unknown>
  createServiceCheck: (payload: AwdCreateServiceCheckPayload) => Promise<unknown>
  createAttackLog: (payload: AwdCreateAttackLogPayload) => Promise<unknown>
  closeOverrideDialog: () => void
}

function openRuntimeDialog(
  runtimeStageReady: Readonly<Ref<boolean>>,
  dialogOpen: Ref<boolean>
) {
  if (!runtimeStageReady.value) {
    return
  }
  dialogOpen.value = true
}

async function runDialogMutationAndClose<T>(
  dialogOpen: Ref<boolean>,
  mutation: (payload: T) => Promise<unknown>,
  payload: T
) {
  await mutation(payload)
  dialogOpen.value = false
}

export function useAwdOperationsDialogState({
  runtimeStageReady,
  rounds,
  createRound,
  createServiceCheck,
  createAttackLog,
  closeOverrideDialog,
}: UseAwdOperationsDialogStateOptions) {
  const roundDialogOpen = ref(false)
  const serviceCheckDialogOpen = ref(false)
  const attackLogDialogOpen = ref(false)
  const nextRoundNumber = computed(() =>
    rounds.value.length === 0 ? 1 : Math.max(...rounds.value.map((item) => item.round_number)) + 1
  )

  function openRoundDialog() {
    openRuntimeDialog(runtimeStageReady, roundDialogOpen)
  }

  function updateRoundDialogOpen(value: boolean) {
    roundDialogOpen.value = value
  }

  function openServiceCheckDialog() {
    openRuntimeDialog(runtimeStageReady, serviceCheckDialogOpen)
  }

  function updateServiceCheckDialogOpen(value: boolean) {
    serviceCheckDialogOpen.value = value
  }

  function openAttackLogDialog() {
    openRuntimeDialog(runtimeStageReady, attackLogDialogOpen)
  }

  function updateAttackLogDialogOpen(value: boolean) {
    attackLogDialogOpen.value = value
  }

  async function handleCreateRound(payload: AwdCreateRoundPayload) {
    await runDialogMutationAndClose(roundDialogOpen, createRound, payload)
  }

  async function handleCreateServiceCheck(payload: AwdCreateServiceCheckPayload) {
    await runDialogMutationAndClose(serviceCheckDialogOpen, createServiceCheck, payload)
  }

  async function handleCreateAttackLog(payload: AwdCreateAttackLogPayload) {
    await runDialogMutationAndClose(attackLogDialogOpen, createAttackLog, payload)
  }

  function handleOverrideDialogOpenChange(value: boolean) {
    if (!value) {
      closeOverrideDialog()
    }
  }

  return {
    roundDialogOpen,
    serviceCheckDialogOpen,
    attackLogDialogOpen,
    nextRoundNumber,
    openRoundDialog,
    updateRoundDialogOpen,
    openServiceCheckDialog,
    updateServiceCheckDialogOpen,
    openAttackLogDialog,
    updateAttackLogDialogOpen,
    handleCreateRound,
    handleCreateServiceCheck,
    handleCreateAttackLog,
    handleOverrideDialogOpenChange,
  }
}
