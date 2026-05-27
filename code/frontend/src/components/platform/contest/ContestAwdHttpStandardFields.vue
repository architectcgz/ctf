<script setup lang="ts">
import type {
  AwdConfigFieldErrors,
  AwdHttpActionDraft,
  AwdHttpActionSection,
  AwdHttpStandardPreset,
} from './contestAwdConfigTypes'

defineProps<{
  awdHttpMethodOptions: readonly string[]
  awdHttpStandardPresets: readonly AwdHttpStandardPreset[]
  applyHttpPreset: (presetId: string) => void
  fieldErrors: AwdConfigFieldErrors
  httpActionSections: readonly AwdHttpActionSection[]
  httpStandardDraft: Record<string, AwdHttpActionDraft>
}>()
</script>

<template>
  <div class="checker-preset-strip checker-preset-strip--compact">
    <button
      v-for="preset in awdHttpStandardPresets"
      :key="preset.id"
      type="button"
      class="ui-btn ui-btn--secondary checker-preset-button"
      @click="applyHttpPreset(preset.id)"
    >
      {{ preset.label }}
    </button>
  </div>

  <section
    v-for="action in httpActionSections"
    :key="action.key"
    class="checker-action-section checker-action-section--panel"
  >
    <header class="list-heading checker-action-section__head">
      <div class="checker-action-section__heading">
        <h4 class="list-heading__title checker-action-section__title">
          {{ action.title }}
        </h4>
        <span class="checker-action-section__hint">动作配置</span>
      </div>
    </header>
    <div class="checker-action-grid checker-action-grid--http">
      <label class="ui-field checker-field checker-field--method">
        <span class="ui-field__label">Method</span>
        <span class="ui-control-wrap">
          <select
            v-model="httpStandardDraft[action.key].method"
            class="ui-control"
          >
            <option
              v-for="method in awdHttpMethodOptions"
              :key="method"
              :value="method"
            >
              {{ method }}
            </option>
          </select>
        </span>
      </label>
      <label class="ui-field checker-field checker-field--path">
        <span class="ui-field__label">Path</span>
        <span
          class="ui-control-wrap"
          :class="{
            'is-error': action.pathErrorKey ? !!fieldErrors[action.pathErrorKey] : false,
          }"
        >
          <input
            v-model="httpStandardDraft[action.key].path"
            type="text"
            class="ui-control"
          >
        </span>
        <span
          v-if="action.pathErrorKey && fieldErrors[action.pathErrorKey]"
          class="ui-field__error"
        >
          {{ fieldErrors[action.pathErrorKey] }}
        </span>
      </label>
      <label class="ui-field checker-field checker-field--status">
        <span class="ui-field__label">状态码</span>
        <span
          class="ui-control-wrap"
          :class="{ 'is-error': !!fieldErrors[action.statusErrorKey] }"
        >
          <input
            v-model.number="httpStandardDraft[action.key].expected_status"
            type="number"
            min="1"
            step="1"
            class="ui-control"
          >
        </span>
        <span
          v-if="fieldErrors[action.statusErrorKey]"
          class="ui-field__error"
        >
          {{ fieldErrors[action.statusErrorKey] }}
        </span>
      </label>
    </div>
    <div class="checker-action-extra-grid checker-action-extra-grid--http">
      <label class="ui-field checker-field checker-field--wide">
        <span class="ui-field__label">Body Template</span>
        <span class="ui-control-wrap">
          <textarea
            v-model="httpStandardDraft[action.key].body_template"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
      </label>
      <label class="ui-field">
        <span class="ui-field__label">Expected Substring</span>
        <span class="ui-control-wrap">
          <input
            v-model="httpStandardDraft[action.key].expected_substring"
            type="text"
            class="ui-control awd-config-control--mono"
          >
        </span>
      </label>
      <label class="ui-field checker-action-extra-grid__wide">
        <span class="ui-field__label">Headers JSON</span>
        <span
          class="ui-control-wrap"
          :class="{ 'is-error': !!fieldErrors[action.headersErrorKey] }"
        >
          <textarea
            v-model="httpStandardDraft[action.key].headers_text"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
        <span
          v-if="fieldErrors[action.headersErrorKey]"
          class="ui-field__error"
        >
          {{ fieldErrors[action.headersErrorKey] }}
        </span>
      </label>
    </div>
  </section>
</template>
