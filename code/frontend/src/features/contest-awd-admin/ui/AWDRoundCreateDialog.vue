<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type { AWDRoundData } from '@/api/contracts'

import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import AWDRoundCreateScoreSection from './AWDRoundCreateScoreSection.vue'
import AWDRoundCreateSettingsSection from './AWDRoundCreateSettingsSection.vue'
import type { AwdCreateRoundPayload } from './awdOperationsDialogContracts'
import './awdOperationsDialogs.css'

const props = defineProps<{
  open: boolean
  nextRoundNumber: number
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AwdCreateRoundPayload]
}>()

const form = reactive({
  round_number: 1,
  status: 'pending' as AWDRoundData['status'],
  attack_score: 50,
  defense_score: 50,
})

const fieldErrors = reactive({
  round_number: '',
  attack_score: '',
  defense_score: '',
})

const dialogTitle = computed(() => `创建第 ${props.nextRoundNumber} 轮`)

watch(
  () => [props.open, props.nextRoundNumber] as const,
  ([open, nextRoundNumber]) => {
    if (!open) {
      return
    }
    form.round_number = nextRoundNumber
    form.status = 'pending'
    form.attack_score = 50
    form.defense_score = 50
    clearErrors()
  },
  { immediate: true }
)

function clearErrors() {
  fieldErrors.round_number = ''
  fieldErrors.attack_score = ''
  fieldErrors.defense_score = ''
}

function closeDialog() {
  emit('update:open', false)
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

function handleSubmit() {
  if (props.saving) {
    return
  }

  if (!validate()) {
    return
  }

  emit('save', {
    round_number: form.round_number,
    status: form.status,
    attack_score: form.attack_score,
    defense_score: form.defense_score,
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    :title="dialogTitle"
    subtitle="设置轮次编号、初始状态和攻防分，提交后会进入赛事运维节奏。"
    eyebrow="AWD Operations"
    width="32.5rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="space-y-5"
      @submit.prevent="handleSubmit"
    >
      <AWDRoundCreateSettingsSection
        :form="form"
        :field-errors="fieldErrors"
      />

      <AWDRoundCreateScoreSection
        :form="form"
        :field-errors="fieldErrors"
      />
    </form>

    <template #footer>
      <AWDOperationsDialogFooter
        cancel-id="awd-round-create-cancel"
        submit-id="awd-round-create-submit"
        submit-text="创建轮次"
        saving-text="创建中..."
        :saving="saving"
        :disabled="saving"
        @cancel="closeDialog"
        @submit="handleSubmit"
      />
    </template>
  </AdminSurfaceModal>
</template>
