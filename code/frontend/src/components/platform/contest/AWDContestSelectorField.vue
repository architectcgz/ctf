<script setup lang="ts">
import type { ContestDetailData } from '@/api/contracts'

defineProps<{
  contests: ContestDetailData[]
  selectedContestId: string | null
}>()

const emit = defineEmits<{
  'update:selectedContestId': [contestId: string]
}>()
</script>

<template>
  <label class="ui-field awd-ops-selector-field">
    <span class="ui-field__label">选择 AWD 赛事</span>
    <span class="ui-control-wrap">
      <select
        id="awd-contest-selector"
        :value="selectedContestId || ''"
        class="ui-control"
        :disabled="contests.length === 0"
        @change="emit('update:selectedContestId', ($event.target as HTMLSelectElement).value)"
      >
        <option v-if="contests.length === 0" value="" disabled>暂无 AWD 赛事</option>
        <option v-for="contest in contests" :key="contest.id" :value="contest.id">
          {{ contest.title }}
        </option>
      </select>
    </span>
  </label>
</template>

<style scoped>
.awd-ops-selector-field {
  --ui-field-gap: var(--space-2);
}
</style>
