<script setup lang="ts">
import { Clock } from 'lucide-vue-next'

import type { ContestFieldLocks } from '../model'
import PlatformContestFormSectionShell from './PlatformContestFormSectionShell.vue'

defineProps<{
  startsAt: string
  endsAt: string
  fieldLocks: ContestFieldLocks
  startsAtError: string
  endsAtError: string
}>()

const emit = defineEmits<{
  'update:startsAt': [value: string]
  'update:endsAt': [value: string]
}>()
</script>

<template>
  <PlatformContestFormSectionShell
    title="赛制与时间"
    description="精确配置比赛的启停节点，系统将按此时钟自动调度。"
    :icon="Clock"
  >
    <div class="contest-form-row">
      <label class="contest-form-row__label">赛程时间轴</label>
      <div class="contest-form-row__control">
        <div class="contest-form-timeline-fields">
          <div class="contest-form-timeline-field">
            <div
              class="ui-control-wrap contest-form-control-wrap"
              :class="{ 'is-disabled': fieldLocks.starts_at }"
            >
              <input
                id="contest-starts-at"
                :value="startsAt"
                type="datetime-local"
                class="ui-control contest-form-input"
                :disabled="fieldLocks.starts_at"
                @input="emit('update:startsAt', ($event.target as HTMLInputElement).value)"
              >
            </div>
            <p class="contest-form-field-hint contest-form-field-hint--compact">
              开始时间
            </p>
          </div>
          <div class="contest-form-timeline-divider">
            ——
          </div>
          <div class="contest-form-timeline-field">
            <div
              class="ui-control-wrap contest-form-control-wrap"
              :class="{ 'is-disabled': fieldLocks.ends_at }"
            >
              <input
                id="contest-ends-at"
                :value="endsAt"
                type="datetime-local"
                class="ui-control contest-form-input"
                :disabled="fieldLocks.ends_at"
                @input="emit('update:endsAt', ($event.target as HTMLInputElement).value)"
              >
            </div>
            <p class="contest-form-field-hint contest-form-field-hint--compact">
              结束时间
            </p>
          </div>
        </div>
        <p
          v-if="startsAtError || endsAtError"
          class="contest-form-field-error contest-form-field-error--spaced"
        >
          {{ startsAtError || endsAtError }}
        </p>
      </div>
    </div>
  </PlatformContestFormSectionShell>
</template>
