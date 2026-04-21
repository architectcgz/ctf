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
const emit = defineEmits<AWDRoundHeaderPanelEmits>()

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
          Operations Control / AWD Platform
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
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="ops-btn ops-btn--neutral"
            :disabled="loadingRounds || loadingRoundDetail"
            @click="emit('refresh')"
          >
            <RefreshCw
              class="h-3.5 w-3.5"
              :class="{ 'animate-spin': loadingRounds || loadingRoundDetail }"
            />
            <span>同步态势</span>
          </button>
          
          <button
            type="button"
            class="ops-btn ops-btn--neutral"
            @click="emit('openCreateRoundDialog')"
          >
            <PlusCircle class="h-3.5 w-3.5" />
            <span>创建轮次</span>
          </button>

          <div class="ops-divider" />

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
          <span class="pulse-dot" /> 实时追踪模式 (15s)
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
  padding: 0 0 2rem; /* 移除外框，调整边距对齐主内容 */
  background: transparent;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.awd-ops-header__top {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.awd-ops-header__overline {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.awd-ops-header__title {
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  color: #0f172a;
  margin: 0;
}

/* Round Switcher - 更加轻量 */
.round-switcher {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: white;
  padding: 0.25rem;
  border-radius: 0.85rem;
  border: 1px solid #e2e8f0;
}

.round-nav-btn {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.65rem;
  color: #64748b;
  transition: all 0.2s ease;
  cursor: pointer;
}

.round-nav-btn:hover:not(:disabled) {
  background: #f8fafc;
  color: #3b82f6;
}

.round-nav-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
}

.round-select-wrapper {
  position: relative;
  height: 2.25rem;
  padding: 0 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-left: 1px solid #f1f5f9;
  border-right: 1px solid #f1f5f9;
  min-width: 10rem;
}

.select-icon { color: #cbd5e1; }

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
  gap: 0.5rem;
  font-size: 13px;
  color: #1e293b;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #cbd5e1;
}
.status-dot.running { background: #22c55e; box-shadow: 0 0 10px rgba(34, 197, 94, 0.4); }

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  height: 2.5rem;
  padding: 0 1.25rem;
  border-radius: 0.85rem;
  font-size: 13px;
  font-weight: 700;
  transition: all 0.2s ease;
  cursor: pointer;
}

.ops-btn--neutral {
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
}

.ops-btn--neutral:hover:not(:disabled) {
  border-color: #cbd5e1;
  color: #0f172a;
}

.ops-btn--primary {
  background: #2563eb;
  color: white;
  border: none;
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.15);
}

.ops-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ops-divider {
  width: 1px;
  height: 1.5rem;
  background: #e2e8f0;
  margin: 0 0.5rem;
}

.awd-ops-header__bottom {
  padding-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--workspace-line-soft) 40%, transparent);
}

.hint-item { font-size: 11px; font-weight: 600; color: #94a3b8; }
.hint-item--live { color: #2563eb; display: flex; align-items: center; gap: 0.5rem; }
.pulse-dot { width: 6px; height: 6px; background: #2563eb; border-radius: 50%; animation: pulse 2s infinite; }

@keyframes pulse {
  0% { transform: scale(0.95); }
  70% { transform: scale(1.1); opacity: 0.5; }
  100% { transform: scale(0.95); }
}

@media (max-width: 1280px) {
  .awd-ops-header__top { flex-direction: column; align-items: flex-start; gap: 1.5rem; }
  .awd-ops-header__actions { width: 100%; overflow-x: auto; padding-bottom: 0.5rem; }
}
</style>

