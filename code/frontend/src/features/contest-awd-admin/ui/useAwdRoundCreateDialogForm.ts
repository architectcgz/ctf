import { computed, reactive, watch, type Ref } from 'vue'

import type { AWDRoundData } from '@/api/contracts'

import type { AwdCreateRoundPayload } from './awdOperationsDialogContracts'

interface UseAwdRoundCreateDialogFormOptions {
  open: Readonly<Ref<boolean>>
  nextRoundNumber: Readonly<Ref<number>>
}

function createDefaultForm(roundNumber: number) {
  return {
    round_number: roundNumber,
    status: 'pending' as AWDRoundData['status'],
    attack_score: 50,
    defense_score: 50,
  }
}

export function useAwdRoundCreateDialogForm({
  open,
  nextRoundNumber,
}: UseAwdRoundCreateDialogFormOptions) {
  const form = reactive(createDefaultForm(nextRoundNumber.value))
  const fieldErrors = reactive({
    round_number: '',
    attack_score: '',
    defense_score: '',
  })

  const dialogTitle = computed(() => `创建第 ${nextRoundNumber.value} 轮`)

  function clearErrors() {
    fieldErrors.round_number = ''
    fieldErrors.attack_score = ''
    fieldErrors.defense_score = ''
  }

  function resetForm(roundNumber: number) {
    Object.assign(form, createDefaultForm(roundNumber))
    clearErrors()
  }

  function validate(): boolean {
    clearErrors()

    if (!Number.isInteger(form.round_number) || form.round_number <= 0) {
      fieldErrors.round_number = '轮次编号必须是大于 0 的整数'
    }
    if (!Number.isInteger(form.attack_score) || form.attack_score < 0) {
      fieldErrors.attack_score = '攻击分必须是大于等于 0 的整数'
    }
    if (!Number.isInteger(form.defense_score) || form.defense_score < 0) {
      fieldErrors.defense_score = '防守分必须是大于等于 0 的整数'
    }

    return !fieldErrors.round_number && !fieldErrors.attack_score && !fieldErrors.defense_score
  }

  function buildPayload(): AwdCreateRoundPayload | null {
    if (!validate()) {
      return null
    }

    return {
      round_number: form.round_number,
      status: form.status,
      attack_score: form.attack_score,
      defense_score: form.defense_score,
    }
  }

  watch(
    () => [open.value, nextRoundNumber.value] as const,
    ([isOpen, roundNumber]) => {
      if (!isOpen) {
        return
      }
      resetForm(roundNumber)
    },
    { immediate: true }
  )

  return {
    form,
    fieldErrors,
    dialogTitle,
    buildPayload,
  }
}
