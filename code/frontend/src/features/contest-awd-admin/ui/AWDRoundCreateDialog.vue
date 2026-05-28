<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type { AWDRoundData } from '@/api/contracts'

const props = defineProps<{
  open: boolean
  nextRoundNumber: number
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      round_number: number
      status: AWDRoundData['status']
      attack_score: number
      defense_score: number
    },
  ]
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
      <div class="ui-field awd-round-field">
        <label
          class="ui-field__label"
          for="awd-round-number"
        >轮次编号</label>
        <span
          class="ui-control-wrap"
          :class="{ 'is-error': !!fieldErrors.round_number }"
        >
          <input
            id="awd-round-number"
            v-model.number="form.round_number"
            type="number"
            min="1"
            step="1"
            class="ui-control"
          >
        </span>
        <p
          v-if="fieldErrors.round_number"
          class="ui-field__error"
        >
          {{ fieldErrors.round_number }}
        </p>
      </div>

      <div class="ui-field awd-round-field">
        <label
          class="ui-field__label"
          for="awd-round-status"
        >初始状态</label>
        <span class="ui-control-wrap">
          <select
            id="awd-round-status"
            v-model="form.status"
            class="ui-control"
          >
            <option value="pending">待开始</option>
            <option value="running">进行中</option>
            <option value="finished">已结束</option>
          </select>
        </span>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="ui-field awd-round-field">
          <label
            class="ui-field__label"
            for="awd-attack-score"
          >攻击分</label>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.attack_score }"
          >
            <input
              id="awd-attack-score"
              v-model.number="form.attack_score"
              type="number"
              min="0"
              step="1"
              class="ui-control"
            >
          </span>
          <p
            v-if="fieldErrors.attack_score"
            class="ui-field__error"
          >
            {{ fieldErrors.attack_score }}
          </p>
        </div>

        <div class="ui-field awd-round-field">
          <label
            class="ui-field__label"
            for="awd-defense-score"
          >防守分</label>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.defense_score }"
          >
            <input
              id="awd-defense-score"
              v-model.number="form.defense_score"
              type="number"
              min="0"
              step="1"
              class="ui-control"
            >
          </span>
          <p
            v-if="fieldErrors.defense_score"
            class="ui-field__error"
          >
            {{ fieldErrors.defense_score }}
          </p>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="awd-round-dialog__footer">
        <button
          id="awd-round-create-cancel"
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-round-create-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving"
          @click="handleSubmit"
        >
          {{ saving ? '创建中...' : '创建轮次' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.awd-round-field {
  --ui-field-gap: var(--space-2);
}

.awd-round-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
  width: 100%;
}

@media (max-width: 767px) {
  .awd-round-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
