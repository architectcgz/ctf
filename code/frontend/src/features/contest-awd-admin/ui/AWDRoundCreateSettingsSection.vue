<script setup lang="ts">
import type { AWDRoundData } from '@/api/contracts'

defineProps<{
  form: {
    round_number: number
    status: AWDRoundData['status']
  }
  fieldErrors: {
    round_number: string
  }
}>()
</script>

<template>
  <div class="ui-field awd-operations-field">
    <label
      class="ui-field__label"
      for="awd-round-number"
    >轮次编号</label>
    <span
      class="ui-control-wrap"
      :class="{ 'is-error': !!fieldErrors.round_number }"
    >
      <input
        id="awd-round-number"
        v-model.number="form.round_number"
        type="number"
        min="1"
        step="1"
        class="ui-control"
      >
    </span>
    <p
      v-if="fieldErrors.round_number"
      class="ui-field__error"
    >
      {{ fieldErrors.round_number }}
    </p>
  </div>

  <div class="ui-field awd-operations-field">
    <label
      class="ui-field__label"
      for="awd-round-status"
    >初始状态</label>
    <span class="ui-control-wrap">
      <select
        id="awd-round-status"
        v-model="form.status"
        class="ui-control"
      >
        <option value="pending">待开始</option>
        <option value="running">进行中</option>
        <option value="finished">已结束</option>
      </select>
    </span>
  </div>
</template>
