<script setup lang="ts">
import type { AwdConfigFieldErrors, AwdScriptCheckerDraft } from './contestAwdConfigTypes'

defineProps<{
  fieldErrors: AwdConfigFieldErrors
  scriptCheckerDraft: AwdScriptCheckerDraft
}>()
</script>

<template>
  <div class="checker-action-grid checker-action-grid--script-meta">
    <label class="ui-field">
      <span class="ui-field__label">Runtime</span>
      <span class="ui-control-wrap">
        <select
          v-model="scriptCheckerDraft.runtime"
          class="ui-control"
        >
          <option value="python3">python3</option>
        </select>
      </span>
    </label>
    <label class="ui-field">
      <span class="ui-field__label">输出格式</span>
      <span class="ui-control-wrap">
        <select
          v-model="scriptCheckerDraft.output"
          class="ui-control"
        >
          <option value="exit_code">Exit Code</option>
          <option value="json">JSON</option>
        </select>
      </span>
    </label>
    <label class="ui-field">
      <span class="ui-field__label">超时时间</span>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.script_timeout }"
      >
        <input
          v-model.number="scriptCheckerDraft.timeout_sec"
          type="number"
          min="1"
          max="60"
          step="1"
          class="ui-control"
        >
      </span>
      <span
        v-if="fieldErrors.script_timeout"
        class="ui-field__error"
      >
        {{ fieldErrors.script_timeout }}
      </span>
    </label>
  </div>
  <label class="ui-field">
    <span class="ui-field__label">入口文件</span>
    <span
      class="ui-control-wrap"
      :class="{ 'is-error': !!fieldErrors.script_entry }"
    >
      <input
        v-model="scriptCheckerDraft.entry"
        type="text"
        class="ui-control"
      >
    </span>
    <span
      v-if="fieldErrors.script_entry"
      class="ui-field__error"
    >
      {{ fieldErrors.script_entry }}
    </span>
  </label>
  <div class="checker-action-extra-grid checker-action-extra-grid--script">
    <label class="ui-field checker-field checker-field--wide">
      <span class="ui-field__label">Args</span>
      <span class="ui-control-wrap">
        <textarea
          v-model="scriptCheckerDraft.args_text"
          rows="3"
          class="ui-control awd-config-control--mono"
        />
      </span>
    </label>
    <label class="ui-field checker-field checker-field--wide">
      <span class="ui-field__label">Env JSON</span>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.script_env_text }"
      >
        <textarea
          v-model="scriptCheckerDraft.env_text"
          rows="3"
          class="ui-control awd-config-control--mono"
        />
      </span>
      <span
        v-if="fieldErrors.script_env_text"
        class="ui-field__error"
      >
        {{ fieldErrors.script_env_text }}
      </span>
    </label>
  </div>
</template>
