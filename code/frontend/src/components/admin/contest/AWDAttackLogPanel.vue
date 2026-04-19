<script setup lang="ts">
import { ShieldCheck } from 'lucide-vue-next'

import type {
  AWDAttackLogPanelEmits,
  AWDAttackLogPanelProps,
} from '@/components/admin/contest/awdInspector.types'

const props = defineProps<AWDAttackLogPanelProps>()
const emit = defineEmits<AWDAttackLogPanelEmits>()

function updateAttackResultFilter(value: string): void {
  if (value !== 'all' && value !== 'success' && value !== 'failed') {
    return
  }
  emit('updateAttackResultFilter', value)
}

function updateAttackSourceFilter(value: string): void {
  if (value !== 'all' && !props.attackSourceOptions.includes(value as typeof props.attackSourceOptions[number])) {
    return
  }
  emit('updateAttackSourceFilter', value as 'all' | (typeof props.attackSourceOptions)[number])
}
</script>

<template>
  <div class="overflow-hidden rounded-2xl border border-border">
    <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3">
      <div class="text-sm font-semibold text-[var(--color-text-primary)]">攻击日志</div>
      <button
        id="awd-export-attacks"
        type="button"
        class="ui-btn ui-btn--secondary awd-attack-export-button"
        :disabled="filteredAttacks.length === 0"
        @click="emit('exportAttacks')"
      >
        导出当前筛选
      </button>
    </div>
    <div class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-3">
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">队伍</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-attack-filter-team"
            :value="attackTeamFilter"
            class="ui-control"
            @change="emit('updateAttackTeamFilter', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">全部队伍</option>
            <option v-for="team in attackTeamOptions" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
        </span>
      </label>
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">结果</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-attack-filter-result"
            :value="attackResultFilter"
            class="ui-control"
            @change="updateAttackResultFilter(($event.target as HTMLSelectElement).value)"
          >
            <option value="all">全部结果</option>
            <option value="success">仅成功</option>
            <option value="failed">仅失败</option>
          </select>
        </span>
      </label>
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">记录来源</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-attack-filter-source"
            :value="attackSourceFilter"
            class="ui-control"
            @change="updateAttackSourceFilter(($event.target as HTMLSelectElement).value)"
          >
            <option value="all">全部来源</option>
            <option v-for="source in attackSourceOptions" :key="source" :value="source">
              {{ getAttackSourceLabel(source) }}
            </option>
          </select>
        </span>
      </label>
    </div>
    <table class="min-w-full divide-y divide-border">
      <thead
        class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
      >
        <tr>
          <th class="px-4 py-3">时间</th>
          <th class="px-4 py-3">攻击方</th>
          <th class="px-4 py-3">受害方</th>
          <th class="px-4 py-3">类型</th>
          <th class="px-4 py-3">结果</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border bg-surface/70">
        <tr v-for="attack in filteredAttacks" :key="attack.id">
          <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
            {{ formatDateTime(attack.created_at) }}
          </td>
          <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">
            {{ attack.attacker_team }}
          </td>
          <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
            {{ attack.victim_team }}
          </td>
          <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
            <div>{{ getAttackTypeLabel(attack.attack_type) }}</div>
            <div class="mt-1 text-xs text-[var(--color-text-muted)]">
              {{ getChallengeTitle(attack.challenge_id) }}
            </div>
            <div class="mt-1 text-xs text-[var(--color-text-muted)]">
              {{ getAttackSourceLabel(attack.source) }}
            </div>
          </td>
          <td class="px-4 py-4 text-sm">
            <span
              class="inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold"
              :class="
                attack.is_success
                  ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                  : 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)]'
              "
            >
              <ShieldCheck v-if="attack.is_success" class="h-3.5 w-3.5" />
              {{ attack.is_success ? `成功 +${attack.score_gained}` : '失败' }}
            </span>
          </td>
        </tr>
        <tr v-if="filteredAttacks.length === 0">
          <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
            {{
              attacks.length === 0 ? '当前轮次还没有攻击记录。' : '当前筛选条件下没有攻击记录。'
            }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
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
