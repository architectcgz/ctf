<script setup lang="ts">
import {
  RefreshCw,
  Sword,
  TimerReset,
  ShieldCheck,
  ChevronLeft,
  ChevronRight,
  History,
} from 'lucide-vue-next'
import { computed } from 'vue'

import type {
  AWDRoundHeaderPanelEmits,
  AWDRoundHeaderPanelProps,
} from '@/components/platform/contest/awdInspector.types'

const props = defineProps<AWDRoundHeaderPanelProps & { hideStudioLink?: boolean }>()
const emit = defineEmits<AWDRoundHeaderPanelEmits & { 'open:contest-edit': [] }>()

const roundOptions = computed(() =>
  [...props.rounds].sort((a, b) => b.round_number - a.round_number)
)

const hasPrev = computed(() => {
  if (!props.selectedRound) return false
  return props.rounds.some((r) => r.round_number === props.selectedRound!.round_number - 1)
})

const hasNext = computed(() => {
  if (!props.selectedRound) return false
  return props.rounds.some((r) => r.round_number === props.selectedRound!.round_number + 1)
})

function navigateRound(delta: number) {
  if (!props.selectedRound) return
  const target = props.rounds.find(
    (r) => r.round_number === props.selectedRound!.round_number + delta
  )
  if (target) {
    emit('update:selectedRoundId', target.id)
  }
}
</script>

<template>
  <div class="awd-ops-header">
    <div class="awd-ops-header__top">
      <div class="awd-ops-header__identity">
        <div
          v-if="rounds.length > 0"
          class="round-switcher"
        >
          <button
            class="round-nav-btn"
            :disabled="!hasPrev"
            title="上一轮"
            @click="navigateRound(-1)"
          >
            <ChevronLeft class="h-4 w-4" />
          </button>

          <div class="round-select-wrapper">
            <History class="h-3.5 w-3.5 select-icon" />
            <select
              :value="selectedRoundId"
              class="round-select-native"
              @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)"
            >
              <option
                v-for="round in roundOptions"
                :key="round.id"
                :value="round.id"
              >
                ROUND {{ String(round.round_number).padStart(2, '0') }} ·
                {{ getRoundStatusLabel(round.status).toUpperCase() }}
              </option>
            </select>
            <div class="round-select-display">
              <span class="font-black">ROUND {{ selectedRound?.round_number || '--' }}</span>
              <span
                class="status-dot"
                :class="selectedRound?.status"
              />
            </div>
          </div>

          <button
            class="round-nav-btn"
            :disabled="!hasNext"
            title="下一轮"
            @click="navigateRound(1)"
          >
            <ChevronRight class="h-4 w-4" />
          </button>
        </div>
      </div>

      <div class="awd-ops-header__actions">
        <button
          type="button"
          class="ops-btn ops-btn--neutral"
          title="刷新数据"
          :disabled="loadingRounds || loadingRoundDetail"
          @click="emit('refresh')"
        >
          <RefreshCw
            class="h-4 w-4"
            :class="{ 'animate-spin': loadingRounds || loadingRoundDetail }"
          />
        </button>

        <button
          v-if="!hideStudioLink"
          type="button"
          class="ops-btn ops-btn--neutral"
          @click="emit('open:contest-edit')"
        >
          竞赛工作室
        </button>

        <div
          v-if="!hideStudioLink"
          class="ops-divider"
        />

        <button
          type="button"
          class="ops-btn ops-btn--neutral"
          :disabled="!selectedRoundId || !canRecordServiceChecks"
          @click="emit('openServiceCheckDialog')"
        >
          <ShieldCheck class="h-3.5 w-3.5" />
          <span>补录检查</span>
        </button>

        <button
          type="button"
          class="ops-btn ops-btn--neutral"
          :disabled="!selectedRoundId || !canRecordAttackLogs"
          @click="emit('openAttackLogDialog')"
        >
          <Sword class="h-3.5 w-3.5" />
          <span>补录攻击</span>
        </button>

        <button
          type="button"
          class="ops-btn ops-btn--primary"
          :disabled="checking || !selectedRoundId"
          @click="emit('runSelectedRoundCheck')"
        >
          <TimerReset class="h-3.5 w-3.5" />
          <span>{{ checkButtonLabel }}</span>
        </button>
      </div>
    </div>

    <div
      v-if="shouldAutoRefresh || serviceCheckHint || attackLogHint"
      class="awd-ops-header__bottom"
    >
      <div class="flex items-center gap-4">
        <span
          v-if="shouldAutoRefresh"
          class="hint-item hint-item--live"
        >
          <span class="pulse-dot" /> 当前正在跟随 live 轮次，每 15 秒自动刷新一次
        </span>
        <span
          v-if="selectedRoundId && !canRecordServiceChecks && serviceCheckHint"
          class="hint-item"
        >
          {{ serviceCheckHint }}
        </span>
        <span
          v-if="selectedRoundId && !canRecordAttackLogs && attackLogHint"
          class="hint-item"
        >
          {{ attackLogHint }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.awd-ops-header {
  --round-switcher-control-size: calc(var(--ui-control-height-sm) + var(--space-1));
  --round-selector-min-width: calc(var(--ui-service-cell-min-width) + var(--space-8));
  --round-status-dot-size: var(--space-1-5);
  --round-status-glow-size: var(--space-2-5);
  --ops-divider-height: var(--space-6);
  --ops-button-shadow-y: var(--space-2);
  --ops-button-shadow-blur: var(--space-5);

  padding: 0 0 var(--space-4);
  background: transparent;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.awd-ops-header__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-4);
}

.awd-ops-header__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: nowrap;
}

/* Round Switcher */
.round-switcher {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  background: var(--color-bg-surface);
  padding: var(--space-1);
  border-radius: var(--ui-control-radius-lg);
  border: 1px solid var(--color-border-default);
}

.round-nav-btn {
  width: var(--round-switcher-control-size);
  height: var(--round-switcher-control-size);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--ui-control-radius-md);
  color: var(--color-text-secondary);
  transition:
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast);
  cursor: pointer;
  background: transparent;
  border: none;
}

.round-nav-btn:hover:not(:disabled) {
  background: var(--color-bg-elevated);
  color: var(--color-primary);
}

.round-nav-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
}

.round-select-wrapper {
  position: relative;
  height: var(--round-switcher-control-size);
  padding: 0 var(--space-4);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  border-left: 1px solid var(--color-border-subtle);
  border-right: 1px solid var(--color-border-subtle);
  min-width: var(--round-selector-min-width);
}

.select-icon {
  color: var(--color-text-muted);
}

.round-select-native {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  width: 100%;
}

.round-select-display {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-13);
  color: var(--color-text-primary);
}

.status-dot {
  width: var(--round-status-dot-size);
  height: var(--round-status-dot-size);
  border-radius: 50%;
  background: var(--color-border-default);
}
.status-dot.running {
  background: var(--color-success);
  box-shadow: 0 0 var(--round-status-glow-size)
    color-mix(in srgb, var(--color-success) 40%, transparent);
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  height: var(--ui-control-height-md);
  padding: 0 var(--space-5);
  border-radius: var(--ui-control-radius-lg);
  font-size: var(--font-size-13);
  font-weight: 700;
  transition:
    background-color var(--ui-motion-fast),
    border-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
  cursor: pointer;
}

.ops-btn--neutral {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  color: var(--color-text-secondary);
}

.ops-btn--neutral:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-text-primary);
}

.ops-btn--primary {
  background: var(--color-primary);
  color: var(--color-bg-base);
  border: none;
  box-shadow: 0 var(--ops-button-shadow-y) var(--ops-button-shadow-blur)
    color-mix(in srgb, var(--color-primary) 15%, transparent);
}

.ops-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ops-divider {
  width: 1px;
  height: var(--ops-divider-height);
  background: var(--color-border-default);
  margin: 0 var(--space-2);
}

.awd-ops-header__bottom {
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border-subtle);
}

.hint-item {
  font-size: var(--font-size-11);
  font-weight: 600;
  color: var(--color-text-muted);
}
.hint-item--live {
  color: var(--color-primary);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.pulse-dot {
  width: var(--round-status-dot-size);
  height: var(--round-status-dot-size);
  background: var(--color-primary);
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(0.95);
  }
  70% {
    transform: scale(1.1);
    opacity: 0.5;
  }
  100% {
    transform: scale(0.95);
  }
}

@media (max-width: 1280px) {
  .awd-ops-header__top {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-6);
  }
  .awd-ops-header__actions {
    width: 100%;
    justify-content: flex-start;
    overflow-x: auto;
    padding-bottom: var(--space-2);
  }
}
</style>
