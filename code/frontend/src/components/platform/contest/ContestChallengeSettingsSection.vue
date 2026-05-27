<script setup lang="ts">
import type {
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'

type DialogMode = 'create' | 'edit'

defineProps<{
  mode: DialogMode
  contestMode: ContestDetailData['mode']
  showContestSelector: boolean
  showContestSettings: boolean
  challengeOptions: AdminChallengeListItem[]
  draft?: AdminContestChallengeViewData | null
  loadingChallengeCatalog: boolean
  challengeId: string
  points: string
  order: string
  isVisible: string
  challengeError: string
  pointsError: string
  orderError: string
}>()

const emit = defineEmits<{
  'update:challenge-id': [value: string]
  'update:points': [value: string]
  'update:order': [value: string]
  'update:is-visible': [value: string]
}>()
</script>

<template>
  <label
    v-if="showContestSelector"
    class="ui-field contest-challenge-dialog__field"
    for="contest-challenge-select"
  >
    <span class="ui-field__label contest-challenge-dialog__label">{{
      contestMode === 'awd' ? 'AWD 服务' : '题目'
    }}</span>
    <template v-if="mode === 'create'">
      <span
        class="ui-control-wrap"
        :class="{
          'is-disabled': loadingChallengeCatalog || challengeOptions.length === 0,
          'is-error': !!challengeError,
        }"
      >
        <select
          id="contest-challenge-select"
          :value="challengeId"
          class="ui-control contest-challenge-dialog__control"
          :disabled="loadingChallengeCatalog || challengeOptions.length === 0"
          @change="emit('update:challenge-id', ($event.target as HTMLSelectElement).value)"
        >
          <option
            value=""
            disabled
          >
            {{ loadingChallengeCatalog ? '正在加载题目目录...' : '请选择题目' }}
          </option>
          <option
            v-for="challenge in challengeOptions"
            :key="challenge.id"
            :value="challenge.id"
          >
            {{ challenge.title }}
          </option>
        </select>
      </span>
    </template>
    <template v-else>
      <span class="ui-control-wrap contest-challenge-dialog__readonly">
        <span class="ui-control contest-challenge-dialog__control">
          {{ draft?.title || `Challenge #${draft?.challenge_id || ''}` }}
        </span>
      </span>
    </template>
    <span
      v-if="challengeError"
      class="ui-field__error contest-challenge-dialog__error"
    >
      {{ challengeError }}
    </span>
  </label>

  <div
    v-if="showContestSettings"
    class="contest-challenge-dialog__grid"
  >
    <label
      class="ui-field contest-challenge-dialog__field"
      for="contest-challenge-points"
    >
      <span class="ui-field__label contest-challenge-dialog__label">分值</span>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!pointsError }"
      >
        <input
          id="contest-challenge-points"
          :value="points"
          type="number"
          min="1"
          step="1"
          class="ui-control contest-challenge-dialog__control"
          @input="emit('update:points', ($event.target as HTMLInputElement).value)"
        >
      </span>
      <span
        v-if="pointsError"
        class="ui-field__error contest-challenge-dialog__error"
      >
        {{ pointsError }}
      </span>
    </label>

    <label
      class="ui-field contest-challenge-dialog__field"
      for="contest-challenge-order"
    >
      <span class="ui-field__label contest-challenge-dialog__label">顺序</span>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!orderError }"
      >
        <input
          id="contest-challenge-order"
          :value="order"
          type="number"
          min="0"
          step="1"
          class="ui-control contest-challenge-dialog__control"
          @input="emit('update:order', ($event.target as HTMLInputElement).value)"
        >
      </span>
      <span
        v-if="orderError"
        class="ui-field__error contest-challenge-dialog__error"
      >
        {{ orderError }}
      </span>
    </label>
  </div>

  <label
    v-if="showContestSettings"
    class="ui-field contest-challenge-dialog__field"
    for="contest-challenge-visibility"
  >
    <span class="ui-field__label contest-challenge-dialog__label">可见性</span>
    <span class="ui-control-wrap">
      <select
        id="contest-challenge-visibility"
        :value="isVisible"
        class="ui-control contest-challenge-dialog__control"
        @change="emit('update:is-visible', ($event.target as HTMLSelectElement).value)"
      >
        <option value="true">可见</option>
        <option value="false">隐藏</option>
      </select>
    </span>
  </label>
</template>

<style scoped>
.contest-challenge-dialog__field {
  --ui-field-gap: var(--space-2);
}

.contest-challenge-dialog__label {
  font-size: var(--font-size-0-875);
}

.contest-challenge-dialog__grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-challenge-dialog__control,
.contest-challenge-dialog__readonly {
  min-height: 2.75rem;
}

.contest-challenge-dialog__readonly {
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
}

.contest-challenge-dialog__error {
  font-size: var(--font-size-0-75);
}

@media (max-width: 767px) {
  .contest-challenge-dialog__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
