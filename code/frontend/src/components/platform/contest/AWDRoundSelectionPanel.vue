<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight, History } from 'lucide-vue-next'

import type { AWDRoundData } from '@/api/contracts'

const props = defineProps<{
  rounds: AWDRoundData[]
  selectedRound: AWDRoundData | null
  selectedRoundId: string | null
  getRoundStatusLabel: (status: AWDRoundData['status']) => string
}>()

const emit = defineEmits<{
  'update:selectedRoundId': [roundId: string]
}>()

const roundOptions = computed(() => [...props.rounds].sort((a, b) => b.round_number - a.round_number))
const hasPrev = computed(() => Boolean(props.selectedRound && props.rounds.some((round) => round.round_number === props.selectedRound!.round_number - 1)))
const hasNext = computed(() => Boolean(props.selectedRound && props.rounds.some((round) => round.round_number === props.selectedRound!.round_number + 1)))

function navigateRound(delta: number): void {
  if (!props.selectedRound) return
  const target = props.rounds.find((round) => round.round_number === props.selectedRound!.round_number + delta)
  if (target) {
    emit('update:selectedRoundId', target.id)
  }
}
</script>

<template>
  <section v-if="rounds.length > 0" class="awd-round-selection-panel">
    <label class="ui-field awd-round-filter-field">
      <span class="ui-field__label">Round Filter</span>
      <span class="ui-control-wrap awd-round-filter-control">
        <button type="button" class="ui-btn ui-btn--secondary awd-round-nav-button" :disabled="!hasPrev" @click="navigateRound(-1)">
          <ChevronLeft class="h-4 w-4" />
        </button>
        <History class="h-3.5 w-3.5 awd-round-filter-icon" />
        <select class="ui-control" :value="selectedRoundId ?? ''" @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)">
          <option v-for="round in roundOptions" :key="round.id" :value="round.id">
            ROUND {{ String(round.round_number).padStart(2, '0') }} · {{ getRoundStatusLabel(round.status).toUpperCase() }}
          </option>
        </select>
        <button type="button" class="ui-btn ui-btn--secondary awd-round-nav-button" :disabled="!hasNext" @click="navigateRound(1)">
          <ChevronRight class="h-4 w-4" />
        </button>
      </span>
    </label>
  </section>
</template>

<style scoped>
.awd-round-selection-panel {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-4);
}

.awd-round-filter-field {
  min-width: min(100%, 28rem);
}

.awd-round-filter-control {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-2);
}

.awd-round-filter-icon {
  color: var(--color-text-muted);
}

.awd-round-nav-button {
  --ui-btn-height: 2.5rem;
  --ui-btn-width: 2.5rem;
  --ui-btn-padding: 0;
}
</style>
