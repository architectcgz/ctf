<script setup lang="ts">
import { toRef } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'

import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import AWDRoundCreateScoreSection from './AWDRoundCreateScoreSection.vue'
import AWDRoundCreateSettingsSection from './AWDRoundCreateSettingsSection.vue'
import type { AwdCreateRoundPayload } from './awdOperationsDialogContracts'
import './awdOperationsDialogs.css'
import { useAwdRoundCreateDialogForm } from './useAwdRoundCreateDialogForm'

const props = defineProps<{
  open: boolean
  nextRoundNumber: number
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AwdCreateRoundPayload]
}>()

const { form, fieldErrors, dialogTitle, buildPayload } = useAwdRoundCreateDialogForm({
  open: toRef(props, 'open'),
  nextRoundNumber: toRef(props, 'nextRoundNumber'),
})

function closeDialog() {
  emit('update:open', false)
}

function handleSubmit() {
  if (props.saving) {
    return
  }

  const payload = buildPayload()
  if (!payload) {
    return
  }

  emit('save', payload)
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
