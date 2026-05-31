<script setup lang="ts">
import { toRef } from 'vue'

import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
} from '@/api/contracts'

import AWDAttackLogDetailsSection from './AWDAttackLogDetailsSection.vue'
import AWDAttackLogTargetSection from './AWDAttackLogTargetSection.vue'
import AWDOperationsDialogFooter from './AWDOperationsDialogFooter.vue'
import type { AwdCreateAttackLogPayload } from './awdOperationsDialogContracts'
import './awdOperationsDialogs.css'
import { useAwdAttackLogDialogForm } from './useAwdAttackLogDialogForm'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeViewData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [value: AwdCreateAttackLogPayload]
}>()

const { form, fieldErrors, challengeOptions, hasTargets, buildPayload } =
  useAwdAttackLogDialogForm({
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
    title="补录攻击日志"
    subtitle="将线下核实过的攻击事件补录进复盘记录，不直接改写正式排行榜。"
    eyebrow="AWD Operations"
    width="35rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="space-y-5"
      @submit.prevent="handleSubmit"
    >
      <AWDAttackLogTargetSection
        :teams="teams"
        :challenge-options="challengeOptions"
        :form="form"
        :field-errors="fieldErrors"
      />

      <AWDAttackLogDetailsSection
        :form="form"
        :has-targets="hasTargets"
      />
    </form>

    <template #footer>
      <AWDOperationsDialogFooter
        submit-id="awd-attack-log-submit"
        submit-text="保存攻击日志"
        saving-text="保存中..."
        :saving="saving"
        :disabled="saving || !hasTargets"
        @cancel="closeDialog"
        @submit="handleSubmit"
      />
    </template>
  </AdminSurfaceModal>
</template>
