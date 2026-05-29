<script setup lang="ts">
import type { AWDAttackLogData } from '@/api/contracts'

defineProps<{
  form: {
    attack_type: AWDAttackLogData['attack_type']
    submitted_flag: string
    is_success: boolean
  }
  hasTargets: boolean
}>()
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-attack-type"
      >攻击类型</label>
      <span class="ui-control-wrap">
        <select
          id="awd-attack-type"
          v-model="form.attack_type"
          class="ui-control"
        >
          <option value="flag_capture">Flag 获取</option>
          <option value="service_exploit">服务利用</option>
        </select>
      </span>
    </div>
  </div>

  <div class="ui-field awd-operations-field">
    <label
      class="ui-field__label"
      for="awd-attack-flag"
    >提交 Flag</label>
    <span class="ui-control-wrap">
      <input
        id="awd-attack-flag"
        v-model="form.submitted_flag"
        type="text"
        class="ui-control"
        placeholder="可选，补录 flag_capture 时填写"
      >
    </span>
  </div>

  <label class="ui-control-wrap awd-operations-checkbox">
    <input
      v-model="form.is_success"
      type="checkbox"
      class="awd-operations-checkbox__box"
    >
    <span class="awd-operations-checkbox__label">本次攻击判定成功</span>
  </label>
  <p class="ui-field__hint">
    人工补录仅进入当前轮复盘记录，不写入正式排行榜与实时竞赛得分。
  </p>
  <p
    v-if="!hasTargets"
    class="ui-field__hint awd-operations-field__warning"
  >
    至少需要 2 支队伍且已关联题目后，才能补录攻击日志。
  </p>
</template>
