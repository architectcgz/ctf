<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

import type {
  AwdConfigFieldErrors,
  AwdTcpCheckerStepDraft,
  AwdTcpStandardDraft,
} from './contestAwdConfigTypes'

defineProps<{
  addTcpCheckerStep: () => void
  expandedTcpCheckerStepIndex: number | null
  fieldErrors: AwdConfigFieldErrors
  removeTcpCheckerStep: (index: number) => void
  summarizeTcpCheckerStep: (step: AwdTcpCheckerStepDraft) => string
  tcpStandardDraft: AwdTcpStandardDraft
  toggleTcpCheckerStep: (index: number) => void
}>()
</script>

<template>
  <div class="checker-toolbar">
    <label class="ui-field awd-config-small-field">
      <span class="ui-field__label">总超时</span>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.tcp_timeout }"
      >
        <input
          v-model.number="tcpStandardDraft.timeout_ms"
          type="number"
          min="1"
          max="60000"
          step="100"
          class="ui-control"
        >
      </span>
      <span
        v-if="fieldErrors.tcp_timeout"
        class="ui-field__error"
      >
        {{ fieldErrors.tcp_timeout }}
      </span>
    </label>
    <button
      type="button"
      class="ui-btn ui-btn--secondary"
      @click="addTcpCheckerStep"
    >
      添加步骤
    </button>
  </div>
  <span
    v-if="fieldErrors.tcp_steps"
    class="ui-field__error"
  >
    {{ fieldErrors.tcp_steps }}
  </span>
  <section
    v-for="(step, index) in tcpStandardDraft.steps"
    :key="index"
    class="checker-action-section checker-action-section--panel checker-action-section--tcp"
    :class="{ 'is-collapsed': expandedTcpCheckerStepIndex !== index }"
  >
    <header class="list-heading checker-action-section__head">
      <button
        type="button"
        class="checker-step-toggle"
        :aria-expanded="expandedTcpCheckerStepIndex === index"
        @click="toggleTcpCheckerStep(index)"
      >
        <span class="checker-action-section__heading">
          <span class="list-heading__title checker-action-section__title">
            Step {{ index + 1 }}
          </span>
          <span class="checker-action-section__hint">
            {{ summarizeTcpCheckerStep(step) }}
          </span>
        </span>
        <ChevronDown class="h-4 w-4 checker-step-toggle__icon" />
      </button>
      <button
        v-if="tcpStandardDraft.steps.length > 1"
        type="button"
        class="ui-btn ui-btn--secondary"
        @click="removeTcpCheckerStep(index)"
      >
        删除
      </button>
    </header>
    <div
      v-show="expandedTcpCheckerStepIndex === index"
      class="checker-action-extra-grid checker-action-extra-grid--tcp"
    >
      <label class="ui-field checker-field checker-field--wide">
        <span class="ui-field__label">Send</span>
        <span class="ui-control-wrap">
          <textarea
            v-model="step.send"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
      </label>
      <label class="ui-field checker-field checker-field--wide">
        <span class="ui-field__label">Send Template</span>
        <span class="ui-control-wrap">
          <textarea
            v-model="step.send_template"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
      </label>
      <label class="ui-field">
        <span class="ui-field__label">Send Hex</span>
        <span class="ui-control-wrap">
          <textarea
            v-model="step.send_hex"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
      </label>
      <label class="ui-field checker-field checker-field--wide">
        <span class="ui-field__label">Expect Contains</span>
        <span class="ui-control-wrap">
          <textarea
            v-model="step.expect_contains"
            rows="2"
            class="ui-control awd-config-control--mono"
          />
        </span>
      </label>
      <label class="ui-field">
        <span class="ui-field__label">Expect Regex</span>
        <span class="ui-control-wrap">
          <input
            v-model="step.expect_regex"
            type="text"
            class="ui-control awd-config-control--mono"
          >
        </span>
      </label>
      <label class="ui-field">
        <span class="ui-field__label">Step Timeout</span>
        <span class="ui-control-wrap">
          <input
            v-model.number="step.timeout_ms"
            type="number"
            min="0"
            max="60000"
            step="100"
            class="ui-control"
          >
        </span>
      </label>
    </div>
  </section>
</template>
