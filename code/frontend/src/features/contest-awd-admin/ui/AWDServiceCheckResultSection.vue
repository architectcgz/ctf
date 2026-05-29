<script setup lang="ts">
defineProps<{
  form: {
    check_result_text: string
  }
  fieldErrors: {
    check_result_text: string
  }
  hasTargets: boolean
}>()
</script>

<template>
  <div class="ui-field awd-operations-field">
    <label
      class="ui-field__label"
      for="awd-service-check-result"
    >检查结果 JSON</label>
    <span
      class="ui-control-wrap"
      :class="{ 'is-error': !!fieldErrors.check_result_text }"
    >
      <textarea
        id="awd-service-check-result"
        v-model="form.check_result_text"
        rows="6"
        class="ui-control awd-operations-field__textarea"
        placeholder="{&quot;http_status&quot;:200,&quot;latency_ms&quot;:38}"
      />
    </span>
    <p
      v-if="fieldErrors.check_result_text"
      class="ui-field__error"
    >
      {{ fieldErrors.check_result_text }}
    </p>
    <p
      v-else-if="!hasTargets"
      class="ui-field__hint awd-operations-field__warning"
    >
      当前赛事缺少队伍或题目，暂时无法录入服务检查。
    </p>
  </div>
</template>
