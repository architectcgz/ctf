<script setup lang="ts">
import type { AwdConfigFieldErrors, AwdLegacyProbeDraft } from './contestAwdConfigTypes'

defineProps<{
  fieldErrors: AwdConfigFieldErrors
  legacyProbeDraft: AwdLegacyProbeDraft
}>()
</script>

<template>
  <label class="ui-field">
    <span class="ui-field__label">健康检查路径</span>
    <span
      class="ui-control-wrap"
      :class="{ 'is-error': !!fieldErrors.legacy_health_path }"
    >
      <input
        v-model="legacyProbeDraft.health_path"
        type="text"
        class="ui-control"
        placeholder="/healthz"
      >
    </span>
    <span
      v-if="fieldErrors.legacy_health_path"
      class="ui-field__error"
    >
      {{ fieldErrors.legacy_health_path }}
    </span>
  </label>
</template>
