<script setup lang="ts">
import type { FlagType } from '@/api/contracts'
import type { PlatformChallengeFlagDraft } from '../model'

interface Props {
  draft: PlatformChallengeFlagDraft
}

defineProps<Props>()

const emit = defineEmits<{
  save: []
  'update:draft': [value: Partial<Pick<PlatformChallengeFlagDraft, 'flagPrefix' | 'flagRegex' | 'flagType' | 'flagValue'>>]
}>()

function updateFlagType(event: Event): void {
  emit('update:draft', { flagType: (event.target as HTMLSelectElement).value as FlagType })
}

function updateFlagValue(event: Event): void {
  emit('update:draft', { flagValue: (event.target as HTMLInputElement).value })
}

function updateFlagRegex(event: Event): void {
  emit('update:draft', { flagRegex: (event.target as HTMLInputElement).value })
}

function updateFlagPrefix(event: Event): void {
  emit('update:draft', { flagPrefix: (event.target as HTMLInputElement).value })
}
</script>

<template>
  <section class="journal-panel challenge-flag-panel p-5 md:p-6">
    <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
      <p class="challenge-flag-panel__copy">
        支持静态 Flag、动态前缀、正则判题和人工审核四种模式。保存后即时刷新当前题目配置。
      </p>
      <div class="flag-summary-chip">
        {{ draft.flagDraftSummary }}
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <label class="flag-field">
        <span class="flag-field-label">判题模式</span>
        <select
          :value="draft.flagType"
          class="flag-field-input"
          @change="updateFlagType"
        >
          <option value="static">静态 Flag</option>
          <option value="dynamic">动态前缀</option>
          <option value="regex">正则匹配</option>
          <option value="manual_review">人工审核</option>
        </select>
      </label>

      <label
        v-if="draft.flagType === 'dynamic' || draft.flagType === 'regex'"
        class="flag-field"
      >
        <span class="flag-field-label">Flag 前缀</span>
        <input
          :value="draft.flagPrefix"
          type="text"
          placeholder="例如：flag"
          class="flag-field-input"
          @input="updateFlagPrefix"
        >
      </label>

      <label
        v-if="draft.flagType === 'static'"
        class="flag-field md:col-span-2"
      >
        <span class="flag-field-label">静态 Flag</span>
        <input
          :value="draft.flagValue"
          type="text"
          placeholder="例如：flag{demo}"
          class="flag-field-input font-mono"
          @input="updateFlagValue"
        >
      </label>

      <label
        v-if="draft.flagType === 'regex'"
        class="flag-field md:col-span-2"
      >
        <span class="flag-field-label">正则表达式</span>
        <input
          :value="draft.flagRegex"
          type="text"
          placeholder="例如：^flag\{demo-[0-9]+\}$"
          class="flag-field-input font-mono"
          @input="updateFlagRegex"
        >
      </label>
    </div>

    <div
      v-if="draft.isSharedInstanceChallenge"
      class="challenge-flag-panel__warning"
    >
      共享实例只适用于无状态题。该模式不提供用户级答案隔离，静态/正则答案可能被转发；若需隔离答案，请使用
      per_user 或 per_team。
    </div>

    <div
      v-if="draft.flagType === 'manual_review'"
      class="challenge-flag-panel__warning"
    >
      学生提交的答案将进入教师审核队列。审核通过后才会计分并更新通过状态。
    </div>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="text-sm text-[var(--journal-muted)]">
        当前配置：{{ draft.flagConfigSummary }}
      </div>
      <button
        :disabled="draft.saving"
        class="ui-btn ui-btn--primary"
        type="button"
        @click="emit('save')"
      >
        {{ draft.saving ? '保存中...' : '保存配置' }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.challenge-flag-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-flag-panel__copy {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

.challenge-flag-panel__warning {
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}

.challenge-flag-panel .ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.flag-summary-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 20%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-accent);
}

.flag-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.flag-field-label {
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-field-input {
  min-height: 2.9rem;
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-92);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.flag-field-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}
</style>
