<script setup lang="ts">
import { RefreshCw, ShieldAlert, ShieldCheck, Sword, TimerReset } from 'lucide-vue-next'

import type {
  AWDRoundHeaderPanelEmits,
  AWDRoundHeaderPanelProps,
} from '@/components/admin/contest/awdInspector.types'
import AppCard from '@/components/common/AppCard.vue'

defineProps<AWDRoundHeaderPanelProps>()
const emit = defineEmits<AWDRoundHeaderPanelEmits>()
</script>

<template>
  <section class="awd-round-shell-grid grid gap-4">
    <div class="awd-round-hero border p-6">
      <div class="awd-round-hero-overline flex flex-wrap items-center gap-2">
        <span>AWD Operations</span>
        <span class="awd-round-hero-chip rounded-full px-2 py-1">真实接口</span>
      </div>
      <div class="mt-3 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-3xl font-semibold tracking-tight text-white">{{ contest.title }}</h2>
          <p class="awd-round-hero-description mt-3 text-sm leading-7">
            针对当前 AWD 赛事查看轮次态势、服务健康、攻击记录，并支持立即触发当前轮巡检。
          </p>
        </div>
        <span
          v-if="selectedRound"
          class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
          :class="getRoundStatusClass(selectedRound.status)"
        >
          第 {{ selectedRound.round_number }} 轮 · {{ getRoundStatusLabel(selectedRound.status) }}
        </span>
      </div>

      <div class="mt-6 flex flex-wrap items-center gap-3 awd-round-toolbar">
        <button
          type="button"
          class="ui-btn ui-btn--secondary awd-round-toolbar__button"
          :disabled="loadingRounds || loadingRoundDetail"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新 AWD 数据
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--secondary awd-round-toolbar__button"
          @click="emit('openCreateRoundDialog')"
        >
          <TimerReset class="h-4 w-4" />
          创建轮次
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--secondary awd-round-toolbar__button"
          :disabled="!selectedRoundId || !canRecordServiceChecks"
          @click="emit('openServiceCheckDialog')"
        >
          <ShieldCheck class="h-4 w-4" />
          录入服务检查
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--secondary awd-round-toolbar__button"
          :disabled="!selectedRoundId || !canRecordAttackLogs"
          @click="emit('openAttackLogDialog')"
        >
          <Sword class="h-4 w-4" />
          补录攻击日志
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary awd-round-toolbar__button"
          :disabled="checking || !selectedRoundId"
          @click="emit('runSelectedRoundCheck')"
        >
          <TimerReset class="h-4 w-4" />
          {{ checkButtonLabel }}
        </button>
      </div>
      <p v-if="shouldAutoRefresh" class="awd-round-hint mt-3 text-xs">
        当前正在跟随 live 轮次，面板会每 15 秒自动刷新一次。
      </p>
      <p
        v-if="selectedRoundId && !canRecordServiceChecks && serviceCheckHint"
        class="awd-round-hint mt-1 text-xs"
      >
        {{ serviceCheckHint }}
      </p>
      <p
        v-if="selectedRoundId && !canRecordAttackLogs && attackLogHint"
        class="awd-round-hint mt-1 text-xs"
      >
        {{ attackLogHint }}
      </p>
    </div>

    <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
      <AppCard
        variant="metric"
        accent="primary"
        eyebrow="轮次数量"
        :title="String(roundsCount)"
        subtitle="当前赛事已创建的 AWD 轮次。"
      >
        <template #header>
          <div class="awd-metric-icon awd-metric-icon--primary">
            <TimerReset class="h-5 w-5" />
          </div>
        </template>
      </AppCard>

      <AppCard
        variant="metric"
        accent="warning"
        eyebrow="失陷服务"
        :title="String(compromisedCount)"
        subtitle="当前所选轮次中已被攻破的服务数。"
      >
        <template #header>
          <div class="awd-metric-icon awd-metric-icon--danger">
            <ShieldAlert class="h-5 w-5" />
          </div>
        </template>
      </AppCard>

      <AppCard
        variant="metric"
        accent="success"
        eyebrow="攻击流量"
        :title="String(totalAttackCount)"
        :subtitle="`成功 ${successfulAttackCount} / 失败 ${failedAttackCount}`"
      >
        <template #header>
          <div class="awd-metric-icon awd-metric-icon--success">
            <Sword class="h-5 w-5" />
          </div>
        </template>
      </AppCard>
    </div>
  </section>
</template>

<style scoped>
.awd-round-hero {
  border-radius: 1.75rem;
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 15%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base))
  );
  box-shadow: 0 24px 70px var(--color-shadow-soft);
}

.awd-round-hero-overline {
  font-size: var(--font-size-11);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.22em;
  color: color-mix(in srgb, var(--color-primary-hover) 75%, transparent);
}

.awd-round-hero-description {
  color: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
}

.awd-round-hero-chip {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 24%, transparent);
}

.awd-round-hint {
  color: color-mix(in srgb, var(--color-primary-hover) 70%, transparent);
}

.awd-round-toolbar__button {
  white-space: nowrap;
}

.awd-metric-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid transparent;
}

.awd-metric-icon--primary {
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: var(--color-primary);
}

.awd-metric-icon--danger {
  border-color: color-mix(in srgb, var(--color-danger) 20%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: var(--color-danger);
}

.awd-metric-icon--success {
  border-color: color-mix(in srgb, var(--color-success) 20%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

@media (min-width: 1280px) {
  .awd-round-shell-grid {
    grid-template-columns: 1.05fr 0.95fr;
  }
}
</style>
