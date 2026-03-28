<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

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
  <ElDialog
    :model-value="open"
    :title="dialogTitle"
    width="520px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-round-number">轮次编号</label>
        <input
          id="awd-round-number"
          v-model.number="form.round_number"
          type="number"
          min="1"
          step="1"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
        >
        <p v-if="fieldErrors.round_number" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.round_number }}</p>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-round-status">初始状态</label>
        <select
          id="awd-round-status"
          v-model="form.status"
          class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
        >
          <option value="pending">待开始</option>
          <option value="running">进行中</option>
          <option value="finished">已结束</option>
        </select>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-attack-score">攻击分</label>
          <input
            id="awd-attack-score"
            v-model.number="form.attack_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.attack_score" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.attack_score }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium text-[var(--color-text-primary)]" for="awd-defense-score">防守分</label>
          <input
            id="awd-defense-score"
            v-model.number="form.defense_score"
            type="number"
            min="0"
            step="1"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
          >
          <p v-if="fieldErrors.defense_score" class="text-xs text-[var(--color-danger)]">{{ fieldErrors.defense_score }}</p>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex items-center justify-end gap-2">
        <button
          id="awd-round-create-cancel"
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm text-[var(--color-text-primary)] transition hover:border-primary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-round-create-submit"
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="saving"
          @click="handleSubmit"
        >
          {{ saving ? '创建中...' : '创建轮次' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
