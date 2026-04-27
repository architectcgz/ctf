<script setup lang="ts">
import { computed } from 'vue'
import { Maximize2, Minimize2, RefreshCw } from 'lucide-vue-next'

import type { AWDRoundData, ContestDetailData } from '@/api/contracts'
import {
  formatProjectorTime,
  getContestStatusLabel,
  getRoundStatusLabel,
} from '@/components/platform/contest/projector/contestProjectorFormatters'

const AUTO_FOLLOW_ROUND_VALUE = '__auto_follow_round__'

const props = defineProps<{
  contests: ContestDetailData[]
  rounds: AWDRoundData[]
  selectedContestId: string
  selectedRoundId: string
  roundAutoFollow: boolean
  lastUpdatedLabel: string
  fullscreenActive: boolean
  loadingContests: boolean
  loadingScoreboard: boolean
}>()

const emit = defineEmits<{
  refresh: []
  toggleFullscreen: []
  selectContest: [contestId: string]
  selectRound: [roundId: string]
  followCurrentRound: []
}>()

const roundSelectorValue = computed(() =>
  props.roundAutoFollow ? AUTO_FOLLOW_ROUND_VALUE : props.selectedRoundId
)
const orderedRounds = computed(() =>
  props.rounds.slice().sort((left, right) => right.round_number - left.round_number)
)

function handleContestSelect(event: Event): void {
  const target = event.target as HTMLSelectElement
  emit('selectContest', target.value)
}

function handleRoundSelect(event: Event): void {
  const target = event.target as HTMLSelectElement
  if (target.value === AUTO_FOLLOW_ROUND_VALUE) {
    emit('followCurrentRound')
    return
  }
  emit('selectRound', target.value)
}
</script>

<template>
  <header class="projector-header">
    <div>
      <div class="projector-overline">
        Contest Projector
      </div>
      <h1 class="projector-title">
        大屏展示
      </h1>
    </div>
    <div class="projector-actions">
      <span class="projector-sync">同步于 {{ lastUpdatedLabel }}</span>
      <button
        type="button"
        class="ops-btn ops-btn--neutral"
        @click="emit('toggleFullscreen')"
      >
        <Minimize2
          v-if="fullscreenActive"
          class="btn-icon"
        />
        <Maximize2
          v-else
          class="btn-icon"
        />
        <span>{{ fullscreenActive ? '退出全屏' : '全屏' }}</span>
      </button>
      <button
        type="button"
        class="ops-btn ops-btn--neutral"
        :disabled="loadingContests || loadingScoreboard"
        @click="emit('refresh')"
      >
        <RefreshCw
          class="btn-icon"
          :class="{ 'animate-spin': loadingContests || loadingScoreboard }"
        />
        <span>刷新</span>
      </button>
    </div>
  </header>

  <div
    v-if="contests.length > 0"
    class="projector-selectors"
  >
    <div class="projector-selector">
      <label
        class="projector-selector__label"
        for="projector-contest-select"
      >
        竞赛
      </label>
      <select
        id="projector-contest-select"
        class="projector-selector__control"
        :value="selectedContestId"
        :disabled="loadingScoreboard"
        @change="handleContestSelect"
      >
        <option
          v-for="contest in contests"
          :key="contest.id"
          :value="contest.id"
        >
          {{ contest.title }} · {{ getContestStatusLabel(contest.status) }} · {{ formatProjectorTime(contest.starts_at) }}
        </option>
      </select>
    </div>

    <div
      v-if="rounds.length > 0"
      class="projector-selector projector-selector--round"
    >
      <label
        class="projector-selector__label"
        for="projector-round-select"
      >
        轮次
      </label>
      <select
        id="projector-round-select"
        class="projector-selector__control"
        :value="roundSelectorValue"
        :disabled="loadingScoreboard"
        @change="handleRoundSelect"
      >
        <option :value="AUTO_FOLLOW_ROUND_VALUE">
          实时跟随当前轮次
        </option>
        <option
          v-for="round in orderedRounds"
          :key="round.id"
          :value="round.id"
        >
          R{{ round.round_number }} · {{ getRoundStatusLabel(round.status) }} · {{ formatProjectorTime(round.started_at ?? round.created_at) }}
        </option>
      </select>
    </div>
  </div>
</template>

<style scoped>
.projector-header,
.projector-actions {
  display: flex;
  align-items: center;
}

.projector-header {
  justify-content: space-between;
  gap: var(--space-4);
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.projector-title {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-45);
  font-weight: 900;
}

.projector-actions {
  gap: var(--space-3);
}

.projector-sync {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border-radius: var(--ui-control-radius-sm);
  font-size: var(--font-size-13);
  font-weight: 800;
  transition: all var(--ui-motion-fast);
}

.ops-btn--neutral {
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
}

.ops-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.btn-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.projector-selectors {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.projector-selector {
  display: flex;
  width: 100%;
  max-width: var(--ui-selector-width-lg);
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.projector-selector--round {
  max-width: var(--ui-selector-width-md);
}

.projector-selector__label {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.projector-selector__control {
  width: min(100%, var(--ui-selector-control-width));
  flex: 1 1 var(--ui-selector-control-min-width);
  min-height: var(--ui-control-height-md);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-bg-surface) 76%, transparent);
  padding: 0 var(--space-3);
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 800;
}

.projector-selector__control:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.projector-selector__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 56%, var(--color-border-default));
  outline: none;
  box-shadow: 0 0 0 var(--space-1) color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.projector-selector__control option {
  background: var(--color-bg-surface);
  color: var(--journal-ink);
}

.projector-selector__control,
.projector-selector__control option {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .projector-header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
