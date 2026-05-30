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

interface RuntimeDialogController<TPayload> {
  open: Ref<boolean>
  requestOpen: () => void
  setOpen: (value: boolean) => void
  submitAndClose: (payload: TPayload) => Promise<void>
}

function createRuntimeDialogController<TPayload>(
  runtimeStageReady: Readonly<Ref<boolean>>,
  mutation: (payload: TPayload) => Promise<unknown>
): RuntimeDialogController<TPayload> {
  const open = ref(false)

  function requestOpen() {
    if (!runtimeStageReady.value) {
      return
    }
    open.value = true
  }

  function setOpen(value: boolean) {
    open.value = value
  }

  async function submitAndClose(payload: TPayload) {
    await mutation(payload)
    open.value = false
  }

  return {
    open,
    requestOpen,
    setOpen,
    submitAndClose,
  }
}

export function useAwdOperationsDialogState({
  runtimeStageReady,
  rounds,
  createRound,
  createServiceCheck,
  createAttackLog,
  closeOverrideDialog,
}: UseAwdOperationsDialogStateOptions) {
  const roundDialog = createRuntimeDialogController(runtimeStageReady, createRound)
  const serviceCheckDialog = createRuntimeDialogController(runtimeStageReady, createServiceCheck)
  const attackLogDialog = createRuntimeDialogController(runtimeStageReady, createAttackLog)
  const nextRoundNumber = computed(() =>
    rounds.value.length === 0 ? 1 : Math.max(...rounds.value.map((item) => item.round_number)) + 1
  )

  function handleOverrideDialogOpenChange(value: boolean) {
    if (!value) {
      closeOverrideDialog()
    }
  }

  return {
    roundDialog,
    serviceCheckDialog,
    attackLogDialog,
    nextRoundNumber,
    handleOverrideDialogOpenChange,
  }
}
