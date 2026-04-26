<script setup lang="ts">
import { 
  RefreshCw, 
  Sword, 
  TimerReset, 
  PlusCircle, 
  ShieldCheck, 
  ChevronLeft, 
  ChevronRight,
  History
} from 'lucide-vue-next'
import { computed } from 'vue'

import type {
  AWDRoundHeaderPanelEmits,
  AWDRoundHeaderPanelProps,
} from '@/components/platform/contest/awdInspector.types'

const props = defineProps<AWDRoundHeaderPanelProps>()
const emit = defineEmits<AWDRoundHeaderPanelEmits & { 'open:contest-edit': [] }>()

const roundOptions = computed(() => 
  [...props.rounds].sort((a, b) => b.round_number - a.round_number)
)

const hasPrev = computed(() => {
  if (!props.selectedRound) return false
  return props.rounds.some(r => r.round_number === props.selectedRound!.round_number - 1)
})

const hasNext = computed(() => {
  if (!props.selectedRound) return false
  return props.rounds.some(r => r.round_number === props.selectedRound!.round_number + 1)
})

function navigateRound(delta: number) {
  if (!props.selectedRound) return
  const target = props.rounds.find(r => r.round_number === props.selectedRound!.round_number + delta)
  if (target) {
    emit('update:selectedRoundId', target.id)
  }
}
</script>

<template>
  <div class="awd-ops-header">
    <div class="awd-ops-header__top">
      <div class="awd-ops-header__identity">
        <div class="awd-ops-header__overline">
          Command Center / AWD Operations
        </div>
        <div class="flex items-center gap-6">
          <h2 class="awd-ops-header__title">
            {{ contest.title }}
          </h2>
          
          <!-- Integrated Round Switcher -->
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
                  ROUND {{ String(round.round_number).padStart(2, '0') }} · {{ getRoundStatusLabel(round.status).toUpperCase() }}
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
      </div>

      <div class="awd-ops-header__actions">
        <!-- Secondary Actions Group -->
        <div class="flex items-center gap-2 mr-4">
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
            type="button"
            class="ops-btn ops-btn--neutral"
            @click="emit('open:contest-edit')"
          >
            竞赛工作室
          </button>
        </div>

        <div class="ops-divider" />

        <!-- Primary Operations Group -->
        <div class="flex items-center gap-2 ml-4">
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
  padding: 0 0 1.5rem;
  background: transparent;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.awd-ops-header__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.awd-ops-header__overline {
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--color-text-muted);
  margin-bottom: 0.25rem;
}

.awd-ops-header__title {
  font-size: var(--font-size-1-25);
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
  margin: 0;
}

/* Round Switcher */
.round-switcher {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  background: var(--color-bg-surface);
  padding: 0.25rem;
  border-radius: 0.85rem;
  border: 1px solid var(--color-border-default);
}

.round-nav-btn {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.65rem;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
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
  height: 2.25rem;
  padding: 0 var(--space-4);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  border-left: 1px solid var(--color-border-subtle);
  border-right: 1px solid var(--color-border-subtle);
  min-width: 10rem;
}

.select-icon { color: var(--color-text-muted); }

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
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-border-default);
}
.status-dot.running { background: var(--color-success); box-shadow: 0 0 10px color-mix(in srgb, var(--color-success) 40%, transparent); }

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  height: var(--ui-control-height-md);
  padding: 0 var(--space-5);
  border-radius: 0.85rem;
  font-size: var(--font-size-13);
  font-weight: 700;
  transition: all 0.2s ease;
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
  box-shadow: 0 8px 20px color-mix(in srgb, var(--color-primary) 15%, transparent);
}

.ops-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ops-divider {
  width: 1px;
  height: 1.5rem;
  background: var(--color-border-default);
  margin: 0 var(--space-2);
}

.awd-ops-header__bottom {
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border-subtle);
}

.hint-item { font-size: var(--font-size-11); font-weight: 600; color: var(--color-text-muted); }
.hint-item--live { color: var(--color-primary); display: flex; align-items: center; gap: var(--space-2); }
.pulse-dot { width: 6px; height: 6px; background: var(--color-primary); border-radius: 50%; animation: pulse 2s infinite; }

@keyframes pulse {
  0% { transform: scale(0.95); }
  70% { transform: scale(1.1); opacity: 0.5; }
  100% { transform: scale(0.95); }
}

@media (max-width: 1280px) {
  .awd-ops-header__top { flex-direction: column; align-items: flex-start; gap: var(--space-6); }
  .awd-ops-header__actions { width: 100%; overflow-x: auto; padding-bottom: 0.5rem; }
}
</style>
