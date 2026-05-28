<script setup lang="ts">
import { Settings, Swords, Trophy } from 'lucide-vue-next'

import type { ContestFieldLocks, ContestFormDraft } from '../model'
import { getStatusLabel } from '@/utils/contest'
import PlatformContestFormSectionShell from './PlatformContestFormSectionShell.vue'

withDefaults(
  defineProps<{
    mode: 'create' | 'edit'
    contestMode: ContestFormDraft['mode']
    contestStatus: ContestFormDraft['status']
    statusOptions?: Array<{ label: string; value: ContestFormDraft['status'] }>
    fieldLocks: ContestFieldLocks
  }>(),
  {
    statusOptions: () => [],
  }
)

const emit = defineEmits<{
  'update:mode': [value: ContestFormDraft['mode']]
  'update:status': [value: ContestFormDraft['status']]
}>()
</script>

<template>
  <PlatformContestFormSectionShell
    title="赛制与状态"
    description="控制竞赛的底层逻辑模式与全平台生命周期。"
    :icon="Settings"
  >
    <div class="ui-field contest-form-field contest-form-row">
      <label class="contest-form-row__label">竞技模式</label>
      <div class="contest-form-row__control">
        <div class="contest-form-mode-options">
          <button
            type="button"
            class="contest-form-mode-card"
            :class="{ active: contestMode === 'jeopardy', disabled: fieldLocks.mode }"
            :disabled="fieldLocks.mode"
            @click="!fieldLocks.mode && emit('update:mode', 'jeopardy')"
          >
            <Trophy class="contest-form-mode-card__icon" />
            <span class="contest-form-mode-card__label">Jeopardy</span>
            <span class="contest-form-mode-card__description">经典夺旗解题赛</span>
          </button>
          <button
            type="button"
            class="contest-form-mode-card"
            :class="{ active: contestMode === 'awd', disabled: fieldLocks.mode }"
            :disabled="fieldLocks.mode"
            @click="!fieldLocks.mode && emit('update:mode', 'awd')"
          >
            <Swords class="contest-form-mode-card__icon" />
            <span class="contest-form-mode-card__label">AWD</span>
            <span class="contest-form-mode-card__description">实时攻防对抗赛</span>
          </button>
        </div>
        <p
          v-if="fieldLocks.mode"
          class="contest-form-field-hint contest-form-field-hint--warning contest-form-field-hint--strong"
        >
          竞赛已生效，模式锁定不可更改。
        </p>
      </div>
    </div>

    <div
      v-if="mode === 'edit'"
      class="ui-field contest-form-field contest-form-row"
    >
      <label class="contest-form-row__label">运行阶段</label>
      <div class="contest-form-row__control">
        <div class="ui-control-wrap contest-form-control-wrap">
          <select
            id="contest-status"
            :value="contestStatus"
            class="ui-control contest-form-select"
            @change="
              emit('update:status', ($event.target as HTMLSelectElement).value as ContestFormDraft['status'])
            "
          >
            <option
              v-for="option in statusOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ getStatusLabel(option.value) }}
            </option>
          </select>
        </div>
        <p class="contest-form-field-hint">
          手动控制竞赛在前端的可见性与交互状态。
        </p>
      </div>
    </div>
  </PlatformContestFormSectionShell>
</template>
