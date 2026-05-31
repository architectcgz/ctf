<script setup lang="ts">
import { toRef } from 'vue'

import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
} from '@/api/contracts'

import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import AWDServiceCheckResultSection from './AWDServiceCheckResultSection.vue'
import AWDServiceCheckTargetSection from './AWDServiceCheckTargetSection.vue'
import type { AwdCreateServiceCheckPayload } from './awdOperationsDialogContracts'
import './awdOperationsDialogs.css'
import { useAwdServiceCheckDialogForm } from './useAwdServiceCheckDialogForm'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeViewData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AwdCreateServiceCheckPayload]
}>()

const { form, fieldErrors, challengeOptions, hasTargets, buildPayload } =
  useAwdServiceCheckDialogForm({
    open: toRef(props, 'open'),
    teams: toRef(props, 'teams'),
    challengeLinks: toRef(props, 'challengeLinks'),
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
    title="录入服务检查"
    subtitle="针对当前轮的队伍服务状态补录检查结果，便于赛后复盘和运维对账。"
    eyebrow="AWD Operations"
    width="35rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="space-y-5"
      @submit.prevent="handleSubmit"
    >
      <AWDServiceCheckTargetSection
        :teams="teams"
        :challenge-options="challengeOptions"
        :form="form"
        :field-errors="fieldErrors"
      />

      <AWDServiceCheckResultSection
        :form="form"
        :field-errors="fieldErrors"
        :has-targets="hasTargets"
      />
    </form>

    <template #footer>
      <AWDOperationsDialogFooter
        submit-id="awd-service-check-submit"
        submit-text="保存检查结果"
        saving-text="保存中..."
        :saving="saving"
        :disabled="saving || !hasTargets"
        @cancel="closeDialog"
        @submit="handleSubmit"
      />
    </template>
  </AdminSurfaceModal>
</template>
