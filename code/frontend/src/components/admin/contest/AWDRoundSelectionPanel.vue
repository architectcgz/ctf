<script setup lang="ts">
import type {
  AWDRoundSelectionPanelEmits,
  AWDRoundSelectionPanelProps,
} from '@/components/admin/contest/awdInspector.types'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import SectionCard from '@/components/common/SectionCard.vue'

defineProps<AWDRoundSelectionPanelProps>()
const emit = defineEmits<AWDRoundSelectionPanelEmits>()
</script>

<template>
  <SectionCard title="轮次切换" subtitle="查看当前轮的基础参数与状态。">
    <div class="space-y-4">
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">选择轮次</span>
        <span
          class="ui-control-wrap awd-round-filter-control"
          :class="{ 'is-disabled': loadingRounds || rounds.length === 0 }"
        >
          <select
            id="awd-round-selector"
            :value="selectedRoundId || ''"
            class="ui-control"
            :disabled="loadingRounds || rounds.length === 0"
            @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="round in rounds" :key="round.id" :value="round.id">
              第 {{ round.round_number }} 轮 · {{ getRoundStatusLabel(round.status) }}
            </option>
          </select>
        </span>
      </label>

      <div v-if="loadingRounds" class="flex justify-center py-8">
        <AppLoading>正在同步轮次...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="rounds.length === 0"
        title="当前赛事还没有 AWD 轮次"
        description="先让后台调度创建轮次，随后这里会展示服务状态和攻击数据。"
        icon="Flag"
      />

      <div v-else-if="selectedRound" class="grid gap-3">
        <AppCard
          variant="action"
          accent="neutral"
          eyebrow="轮次状态"
          :subtitle="getRoundStatusLabel(selectedRound.status)"
        >
          <template #default>
            <div class="awd-round-selection-copy text-sm">
              <p>攻击分值：{{ selectedRound.attack_score }}</p>
              <p class="mt-1">防守分值：{{ selectedRound.defense_score }}</p>
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="action"
          accent="neutral"
          eyebrow="时间窗口"
          :subtitle="formatDateTime(selectedRound.started_at)"
        >
          <template #default>
            <div class="awd-round-selection-copy text-sm">
              <p>开始：{{ formatDateTime(selectedRound.started_at) }}</p>
              <p class="mt-1">结束：{{ formatDateTime(selectedRound.ended_at) }}</p>
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="action"
          accent="warning"
          eyebrow="异常速览"
          :subtitle="`下线 ${downCount} · 失陷 ${compromisedCount}`"
        >
          <template #default>
            <div class="awd-round-selection-copy text-sm">
              <p>服务总数：{{ totalServiceCount }}</p>
              <p class="mt-1">最后巡检：{{ formatDateTime(selectedRound.updated_at) }}</p>
            </div>
          </template>
        </AppCard>
      </div>
    </div>
  </SectionCard>
</template>

<style scoped>
.awd-round-selection-copy {
  color: var(--color-text-secondary);
}

.awd-round-filter-field {
  --ui-field-gap: var(--space-2);
  --ui-field-label-size: var(--font-size-11);
  --ui-field-label-weight: 700;
  --ui-field-label-color: var(--color-text-muted);
  min-width: 0;
}

.awd-round-filter-field .ui-field__label {
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.awd-round-filter-control {
  width: 100%;
}
</style>
